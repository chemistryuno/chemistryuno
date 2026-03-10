package handlers

import (
	"chemistryuno/backend/database"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const (
	pluginFieldTypeText      = "text"
	pluginFieldTypeTextarea  = "textarea"
	pluginFieldTypeNumber    = "number"
	pluginFieldTypeSwitch    = "switch"
	pluginFieldTypeJSON      = "json"
	pluginFieldTypeImage     = "image"
	pluginFieldTypeFile      = "file"
	pluginFieldTypeRouteList = "route_list"
)

var pluginConfigTypes = map[string]bool{
	pluginFieldTypeText:      true,
	pluginFieldTypeTextarea:  true,
	pluginFieldTypeNumber:    true,
	pluginFieldTypeSwitch:    true,
	pluginFieldTypeJSON:      true,
	pluginFieldTypeImage:     true,
	pluginFieldTypeFile:      true,
	pluginFieldTypeRouteList: true,
}

var pluginConfigKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_.-]{1,64}$`)

type pluginConfigField struct {
	Key         string          `json:"key"`
	Label       string          `json:"label"`
	Type        string          `json:"type"`
	Description string          `json:"description"`
	Default     json.RawMessage `json:"default"`
	Required    bool            `json:"required"`
	ReadOnly    bool            `json:"read_only"`
	Accept      string          `json:"accept"`
	MaxSizeKB   int             `json:"max_size_kb"`
	MinLength   int             `json:"min_length"`
	MaxLength   int             `json:"max_length"`
	Min         *float64        `json:"min"`
	Max         *float64        `json:"max"`
	Pattern     string          `json:"pattern"`
}

type pluginRoutePage struct {
	Path         string `json:"path"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	ContentHTML  string `json:"content_html"`
	RequiresAuth bool   `json:"requires_auth"`
	AdminOnly    bool   `json:"admin_only"`
	CoWorkerOnly bool   `json:"co_worker_only"`
}

func parsePluginConfigSchema(raw string) ([]pluginConfigField, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []pluginConfigField{}, nil
	}
	var fields []pluginConfigField
	if err := json.Unmarshal([]byte(trimmed), &fields); err != nil {
		return nil, err
	}
	if err := validatePluginConfigSchema(fields); err != nil {
		return nil, err
	}
	return fields, nil
}

func normalizePluginConfigSchema(raw json.RawMessage) (string, []pluginConfigField, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return "[]", []pluginConfigField{}, nil
	}
	fields, err := parsePluginConfigSchema(trimmed)
	if err != nil {
		return "", nil, err
	}
	normalized, err := json.Marshal(fields)
	if err != nil {
		return "", nil, err
	}
	return string(normalized), fields, nil
}

func validatePluginConfigSchema(fields []pluginConfigField) error {
	if len(fields) > 64 {
		return fmt.Errorf("config_schema 字段过多，最多 64 项")
	}

	seen := make(map[string]bool, len(fields))
	for i, field := range fields {
		key := strings.TrimSpace(field.Key)
		fieldType := strings.ToLower(strings.TrimSpace(field.Type))

		if key == "" {
			return fmt.Errorf("config_schema[%d].key 不能为空", i)
		}
		if !pluginConfigKeyPattern.MatchString(key) {
			return fmt.Errorf("config_schema[%d].key 非法: %s", i, key)
		}
		if seen[key] {
			return fmt.Errorf("config_schema 存在重复 key: %s", key)
		}
		seen[key] = true

		if fieldType == "" {
			fieldType = pluginFieldTypeText
		}
		if !pluginConfigTypes[fieldType] {
			return fmt.Errorf("config_schema[%d].type 不支持: %s", i, field.Type)
		}
		if field.MinLength < 0 || field.MaxLength < 0 {
			return fmt.Errorf("config_schema[%d] 长度限制非法", i)
		}
		if field.MaxLength > 0 && field.MinLength > field.MaxLength {
			return fmt.Errorf("config_schema[%d] min_length 不能大于 max_length", i)
		}
		if field.Min != nil && field.Max != nil && *field.Min > *field.Max {
			return fmt.Errorf("config_schema[%d] min 不能大于 max", i)
		}
		if strings.TrimSpace(field.Pattern) != "" {
			if _, err := regexp.Compile(strings.TrimSpace(field.Pattern)); err != nil {
				return fmt.Errorf("config_schema[%d] pattern 无效: %v", i, err)
			}
		}

		if len(field.Default) > (1 << 20) {
			return fmt.Errorf("config_schema[%d].default 超过大小限制", i)
		}
	}
	return nil
}

func convertDefaultValue(field pluginConfigField) string {
	if len(field.Default) == 0 {
		return ""
	}
	switch strings.ToLower(strings.TrimSpace(field.Type)) {
	case pluginFieldTypeNumber:
		var f float64
		if err := json.Unmarshal(field.Default, &f); err == nil {
			return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%f", f), "0"), ".")
		}
	case pluginFieldTypeSwitch:
		var b bool
		if err := json.Unmarshal(field.Default, &b); err == nil {
			if b {
				return "true"
			}
			return "false"
		}
	case pluginFieldTypeJSON, pluginFieldTypeRouteList:
		var generic interface{}
		if err := json.Unmarshal(field.Default, &generic); err == nil {
			normalized, marshalErr := json.Marshal(generic)
			if marshalErr == nil {
				return string(normalized)
			}
		}
	default:
		var s string
		if err := json.Unmarshal(field.Default, &s); err == nil {
			return s
		}
	}
	return string(field.Default)
}

func upsertPluginSetting(pluginID uint, key string, value string) error {
	storageKey := pluginSettingsStorageKey(pluginID, key)
	config := database.SystemConfig{
		Key:   storageKey,
		Value: value,
	}
	return database.DB.Save(&config).Error
}

func validatePluginSettingValue(field pluginConfigField, value string) error {
	fieldType := strings.ToLower(strings.TrimSpace(field.Type))
	if fieldType == "" {
		fieldType = pluginFieldTypeText
	}

	if field.Required && strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s 为必填项", field.Key)
	}
	if field.MaxLength > 0 && len(value) > field.MaxLength {
		return fmt.Errorf("%s 长度不能超过 %d", field.Key, field.MaxLength)
	}
	if field.MinLength > 0 && len(value) < field.MinLength && strings.TrimSpace(value) != "" {
		return fmt.Errorf("%s 长度不能小于 %d", field.Key, field.MinLength)
	}
	if strings.TrimSpace(field.Pattern) != "" && strings.TrimSpace(value) != "" {
		re, err := regexp.Compile(strings.TrimSpace(field.Pattern))
		if err != nil {
			return fmt.Errorf("%s pattern 非法: %v", field.Key, err)
		}
		if !re.MatchString(value) {
			return fmt.Errorf("%s 格式不符合要求", field.Key)
		}
	}
	if field.MaxSizeKB > 0 && len(value) > field.MaxSizeKB*1024*2 {
		return fmt.Errorf("%s 超过大小限制（%dKB）", field.Key, field.MaxSizeKB)
	}

	switch fieldType {
	case pluginFieldTypeNumber:
		if strings.TrimSpace(value) == "" {
			return nil
		}
		var num float64
		if _, err := fmt.Sscanf(strings.TrimSpace(value), "%f", &num); err != nil {
			return fmt.Errorf("%s 不是有效数字", field.Key)
		}
		if field.Min != nil && num < *field.Min {
			return fmt.Errorf("%s 不能小于 %v", field.Key, *field.Min)
		}
		if field.Max != nil && num > *field.Max {
			return fmt.Errorf("%s 不能大于 %v", field.Key, *field.Max)
		}
	case pluginFieldTypeSwitch:
		v := strings.TrimSpace(strings.ToLower(value))
		if v != "" && v != "true" && v != "false" {
			return fmt.Errorf("%s 必须为 true/false", field.Key)
		}
	case pluginFieldTypeJSON:
		if strings.TrimSpace(value) == "" {
			return nil
		}
		var generic interface{}
		if err := json.Unmarshal([]byte(value), &generic); err != nil {
			return fmt.Errorf("%s JSON 格式错误", field.Key)
		}
	case pluginFieldTypeRouteList:
		if strings.TrimSpace(value) == "" {
			return nil
		}
		var pages []pluginRoutePage
		if err := json.Unmarshal([]byte(value), &pages); err != nil {
			return fmt.Errorf("%s route_list JSON 格式错误", field.Key)
		}
		seen := map[string]bool{}
		for i, page := range pages {
			path := strings.TrimSpace(page.Path)
			if path == "" {
				return fmt.Errorf("%s route_list[%d].path 不能为空", field.Key, i)
			}
			if !strings.HasPrefix(path, "/") {
				return fmt.Errorf("%s route_list[%d].path 必须以 / 开头", field.Key, i)
			}
			if seen[path] {
				return fmt.Errorf("%s route_list 存在重复 path: %s", field.Key, path)
			}
			seen[path] = true
		}
	case pluginFieldTypeImage:
		v := strings.TrimSpace(value)
		if v == "" {
			return nil
		}
		if !strings.HasPrefix(v, "data:image/") && !strings.HasPrefix(v, "http://") && !strings.HasPrefix(v, "https://") {
			return fmt.Errorf("%s 必须为图片 data URL 或 http(s) URL", field.Key)
		}
	}

	return nil
}
