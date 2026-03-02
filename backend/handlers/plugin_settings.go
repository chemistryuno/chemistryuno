package handlers

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	pluginSettingsPrefix        = "plugin_settings"
	pluginSettingsHistoryPrefix = "plugin_settings_history"
	maxPluginSettingsPerUpdate  = 128
	maxPluginSettingValueBytes  = 1 << 20 // 1MB per value
	maxPluginSettingsHistory    = 20
)

var pluginSettingKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_.-]{1,64}$`)

func pluginSettingsStorageKey(pluginID uint, key string) string {
	return fmt.Sprintf("%s.%d.%s", pluginSettingsPrefix, pluginID, key)
}

func pluginSettingsHistoryKey(pluginID uint) string {
	return fmt.Sprintf("%s.%d", pluginSettingsHistoryPrefix, pluginID)
}

type pluginSettingsSnapshot struct {
	ID        string            `json:"id"`
	CreatedAt string            `json:"created_at"`
	CreatedBy int               `json:"created_by"`
	Settings  map[string]string `json:"settings"`
}

func parsePluginID(c *gin.Context) (uint, bool) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的插件 ID"})
		return 0, false
	}
	return uint(id), true
}

func getPluginForSettings(pluginID uint) (*database.Plugin, error) {
	plugin, err := repository.PluginRepo.GetPlugin(pluginID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return plugin, nil
}

func getPluginSettingsMapWithDB(db *gorm.DB, pluginID uint) (map[string]string, error) {
	likePattern := fmt.Sprintf("%s.%d.%%", pluginSettingsPrefix, pluginID)
	var configs []database.SystemConfig
	if err := db.Where("`key` LIKE ?", likePattern).Find(&configs).Error; err != nil {
		return nil, err
	}

	prefix := fmt.Sprintf("%s.%d.", pluginSettingsPrefix, pluginID)
	settings := make(map[string]string, len(configs))
	for _, cfg := range configs {
		if !strings.HasPrefix(cfg.Key, prefix) {
			continue
		}
		key := strings.TrimPrefix(cfg.Key, prefix)
		if key == "" {
			continue
		}
		settings[key] = cfg.Value
	}
	return settings, nil
}

func getPluginSettingsMap(pluginID uint) (map[string]string, error) {
	return getPluginSettingsMapWithDB(database.DB, pluginID)
}

func getPluginSettingsHistoryWithDB(db *gorm.DB, pluginID uint) ([]pluginSettingsSnapshot, error) {
	var config database.SystemConfig
	if err := db.Where("`key` = ?", pluginSettingsHistoryKey(pluginID)).Take(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []pluginSettingsSnapshot{}, nil
		}
		return nil, err
	}
	raw := strings.TrimSpace(config.Value)
	if raw == "" {
		return []pluginSettingsSnapshot{}, nil
	}
	var snapshots []pluginSettingsSnapshot
	if err := json.Unmarshal([]byte(raw), &snapshots); err != nil {
		return nil, err
	}
	return snapshots, nil
}

func getPluginSettingsHistory(pluginID uint) ([]pluginSettingsSnapshot, error) {
	return getPluginSettingsHistoryWithDB(database.DB, pluginID)
}

func savePluginSettingsHistoryWithDB(db *gorm.DB, pluginID uint, snapshots []pluginSettingsSnapshot) error {
	if len(snapshots) > maxPluginSettingsHistory {
		snapshots = snapshots[len(snapshots)-maxPluginSettingsHistory:]
	}
	data, err := json.Marshal(snapshots)
	if err != nil {
		return err
	}
	return db.Save(&database.SystemConfig{
		Key:   pluginSettingsHistoryKey(pluginID),
		Value: string(data),
	}).Error
}

func savePluginSettingsHistory(pluginID uint, snapshots []pluginSettingsSnapshot) error {
	return savePluginSettingsHistoryWithDB(database.DB, pluginID, snapshots)
}

func appendPluginSettingsSnapshotWithDB(db *gorm.DB, pluginID uint, uid int, settings map[string]string) error {
	history, err := getPluginSettingsHistoryWithDB(db, pluginID)
	if err != nil {
		return err
	}
	snapshotCopy := make(map[string]string, len(settings))
	for k, v := range settings {
		snapshotCopy[k] = v
	}
	history = append(history, pluginSettingsSnapshot{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		CreatedAt: time.Now().Format(time.RFC3339),
		CreatedBy: uid,
		Settings:  snapshotCopy,
	})
	return savePluginSettingsHistoryWithDB(db, pluginID, history)
}

func appendPluginSettingsSnapshot(pluginID uint, uid int, settings map[string]string) error {
	return appendPluginSettingsSnapshotWithDB(database.DB, pluginID, uid, settings)
}

// GetPluginSettings 获取插件配置（用户可读）
// GET /api/plugins/:id/settings
func GetPluginSettings(c *gin.Context) {
	pluginID, ok := parsePluginID(c)
	if !ok {
		return
	}

	plugin, err := getPluginForSettings(pluginID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if plugin == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "插件不存在"})
		return
	}

	settings, err := getPluginSettingsMap(pluginID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取插件配置失败"})
		return
	}

	schemaFields, schemaErr := parsePluginConfigSchema(plugin.ConfigSchema)
	if schemaErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "插件配置 schema 解析失败"})
		return
	}

	for _, field := range schemaFields {
		if _, exists := settings[field.Key]; exists {
			continue
		}
		defaultValue := convertDefaultValue(field)
		settings[field.Key] = defaultValue
	}

	c.JSON(http.StatusOK, gin.H{
		"plugin_id": pluginID,
		"settings":  settings,
		"schema":    schemaFields,
	})
}

// UpdatePluginSettings 更新插件配置（管理员）
// PUT /api/admin/plugins/:id/settings
func UpdatePluginSettings(c *gin.Context) {
	pluginID, ok := parsePluginID(c)
	if !ok {
		return
	}

	plugin, err := getPluginForSettings(pluginID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if plugin == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "插件不存在"})
		return
	}

	schemaFields, schemaErr := parsePluginConfigSchema(plugin.ConfigSchema)
	if schemaErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "插件配置 schema 解析失败"})
		return
	}
	schemaFieldMap := make(map[string]pluginConfigField, len(schemaFields))
	for _, field := range schemaFields {
		schemaFieldMap[field.Key] = field
	}
	uid, _ := c.Get("uid")
	operatorUID := 0
	switch v := uid.(type) {
	case int:
		operatorUID = v
	case uint:
		operatorUID = int(v)
	}

	var req struct {
		Settings map[string]string `json:"settings"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	if len(req.Settings) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "settings 不能为空"})
		return
	}
	if len(req.Settings) > maxPluginSettingsPerUpdate {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("单次最多更新 %d 个配置项", maxPluginSettingsPerUpdate)})
		return
	}

	normalizedSettings := make(map[string]string, len(req.Settings))
	for key, value := range req.Settings {
		normalizedKey := strings.TrimSpace(key)
		if !pluginSettingKeyPattern.MatchString(normalizedKey) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("配置项 key 非法: %s", key)})
			return
		}
		if len(value) > maxPluginSettingValueBytes {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("配置项 %s 超过大小限制", normalizedKey)})
			return
		}
		if field, ok := schemaFieldMap[normalizedKey]; ok {
			if err := validatePluginSettingValue(field, value); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}
		normalizedSettings[normalizedKey] = value
	}

	var settings map[string]string
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		for key, value := range normalizedSettings {
			storageKey := pluginSettingsStorageKey(pluginID, key)
			_, isSchemaField := schemaFieldMap[key]
			if value == "" && !isSchemaField {
				if err := tx.Where("`key` = ?", storageKey).Delete(&database.SystemConfig{}).Error; err != nil {
					return err
				}
				continue
			}
			config := database.SystemConfig{
				Key:   storageKey,
				Value: value,
			}
			if err := tx.Save(&config).Error; err != nil {
				return err
			}
		}

		// schema 自带配置项只允许读改，不允许删除：未提交的 schema key 保持原值或补默认
		currentSettings, err := getPluginSettingsMapWithDB(tx, pluginID)
		if err != nil {
			return err
		}
		for _, field := range schemaFields {
			if _, exists := currentSettings[field.Key]; exists {
				continue
			}
			defaultValue := convertDefaultValue(field)
			if err := tx.Save(&database.SystemConfig{
				Key:   pluginSettingsStorageKey(pluginID, field.Key),
				Value: defaultValue,
			}).Error; err != nil {
				return err
			}
		}

		settings, err = getPluginSettingsMapWithDB(tx, pluginID)
		if err != nil {
			return err
		}
		for _, field := range schemaFields {
			if _, exists := settings[field.Key]; exists {
				continue
			}
			settings[field.Key] = convertDefaultValue(field)
		}

		if err := appendPluginSettingsSnapshotWithDB(tx, pluginID, operatorUID, settings); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新插件配置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "插件配置更新成功",
		"plugin_id": pluginID,
		"settings":  settings,
		"schema":    schemaFields,
	})
}

// GetPluginSettingsHistory 获取插件配置历史（管理员）
// GET /api/admin/plugins/:id/settings/history
func GetPluginSettingsHistory(c *gin.Context) {
	pluginID, ok := parsePluginID(c)
	if !ok {
		return
	}
	plugin, err := getPluginForSettings(pluginID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if plugin == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "插件不存在"})
		return
	}
	history, err := getPluginSettingsHistory(pluginID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取插件配置历史失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"plugin_id": pluginID,
		"history":   history,
	})
}

// RollbackPluginSettings 回滚插件配置（管理员）
// POST /api/admin/plugins/:id/settings/rollback
func RollbackPluginSettings(c *gin.Context) {
	pluginID, ok := parsePluginID(c)
	if !ok {
		return
	}
	plugin, err := getPluginForSettings(pluginID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if plugin == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "插件不存在"})
		return
	}

	var req struct {
		SnapshotID string `json:"snapshot_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	history, err := getPluginSettingsHistory(pluginID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取配置历史失败"})
		return
	}
	var snapshot *pluginSettingsSnapshot
	for i := range history {
		if history[i].ID == req.SnapshotID {
			snapshot = &history[i]
			break
		}
	}
	if snapshot == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到对应快照"})
		return
	}

	schemaFields, schemaErr := parsePluginConfigSchema(plugin.ConfigSchema)
	if schemaErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "插件配置 schema 解析失败"})
		return
	}
	schemaFieldMap := make(map[string]pluginConfigField, len(schemaFields))
	for _, field := range schemaFields {
		schemaFieldMap[field.Key] = field
	}

	for key, value := range snapshot.Settings {
		if field, ok := schemaFieldMap[key]; ok {
			if err := validatePluginSettingValue(field, value); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("快照配置校验失败: %v", err)})
				return
			}
		}
	}

	var settings map[string]string
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		currentSettings, err := getPluginSettingsMapWithDB(tx, pluginID)
		if err != nil {
			return err
		}
		for key := range currentSettings {
			if _, exists := snapshot.Settings[key]; exists {
				continue
			}
			if _, isSchema := schemaFieldMap[key]; isSchema {
				continue
			}
			if err := tx.Where("`key` = ?", pluginSettingsStorageKey(pluginID, key)).Delete(&database.SystemConfig{}).Error; err != nil {
				return err
			}
		}

		for key, value := range snapshot.Settings {
			if err := tx.Save(&database.SystemConfig{
				Key:   pluginSettingsStorageKey(pluginID, key),
				Value: value,
			}).Error; err != nil {
				return err
			}
		}
		for _, field := range schemaFields {
			if _, exists := snapshot.Settings[field.Key]; exists {
				continue
			}
			defaultValue := convertDefaultValue(field)
			if err := tx.Save(&database.SystemConfig{
				Key:   pluginSettingsStorageKey(pluginID, field.Key),
				Value: defaultValue,
			}).Error; err != nil {
				return err
			}
		}

		settings, err = getPluginSettingsMapWithDB(tx, pluginID)
		if err != nil {
			return err
		}
		for _, field := range schemaFields {
			if _, exists := settings[field.Key]; exists {
				continue
			}
			settings[field.Key] = convertDefaultValue(field)
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "回滚配置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "插件配置已回滚",
		"plugin_id":   pluginID,
		"snapshot_id": req.SnapshotID,
		"settings":    settings,
		"schema":      schemaFields,
	})
}
