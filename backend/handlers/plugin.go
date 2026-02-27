package handlers

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/game"
	"chemistryuno/backend/plugins"
	"chemistryuno/backend/repository"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// --- 公开接口 ---

// GetPluginCards 获取所有激活插件的卡牌列表（供 deck builder 使用）
func GetPluginCards(c *gin.Context) {
	cards, err := repository.PluginRepo.GetAllActiveCards()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取插件卡牌失败"})
		return
	}
	c.JSON(http.StatusOK, cards)
}

// --- 管理员接口：Plugin CRUD ---

// ListPlugins 列出所有插件
func ListPlugins(c *gin.Context) {
	plugins, err := repository.PluginRepo.GetAllPlugins()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取插件列表失败"})
		return
	}
	c.JSON(http.StatusOK, plugins)
}

// CreatePlugin 创建插件
func CreatePlugin(c *gin.Context) {
	uid, _ := c.Get("uid")
	authorUID := 0
	switch v := uid.(type) {
	case int:
		authorUID = v
	case uint:
		authorUID = int(v)
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	plugin := &database.Plugin{
		Name:        req.Name,
		Description: req.Description,
		AuthorUID:   authorUID,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}

	if err := repository.PluginRepo.CreatePlugin(plugin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建插件失败: " + err.Error()})
		return
	}

	game.LoadPluginCards()
	plugins.LoadServerScripts()
	c.JSON(http.StatusCreated, plugin)
}

// UpdatePlugin 更新插件（名称、描述、激活状态）
func UpdatePlugin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}

	plugin, err := repository.PluginRepo.GetPlugin(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "插件不存在"})
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		IsActive    *bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	if req.Name != nil {
		plugin.Name = *req.Name
	}
	if req.Description != nil {
		plugin.Description = *req.Description
	}
	if req.IsActive != nil {
		plugin.IsActive = *req.IsActive
	}

	if err := repository.PluginRepo.UpdatePlugin(plugin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新插件失败: " + err.Error()})
		return
	}

	game.LoadPluginCards()
	plugins.LoadServerScripts()
	c.JSON(http.StatusOK, plugin)
}

// DeletePlugin 删除插件及其所有卡牌
func DeletePlugin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}

	if err := repository.PluginRepo.DeletePlugin(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除插件失败: " + err.Error()})
		return
	}

	game.LoadPluginCards()
	plugins.LoadServerScripts()
	c.JSON(http.StatusOK, gin.H{"message": "插件已删除"})
}

// --- 管理员接口：PluginCard CRUD ---

// ListPluginCards 获取指定插件的卡牌列表
func ListPluginCards(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的插件 ID"})
		return
	}

	cards, err := repository.PluginRepo.GetCardsByPlugin(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取卡牌失败"})
		return
	}
	c.JSON(http.StatusOK, cards)
}

// CreatePluginCard 向插件添加卡牌
func CreatePluginCard(c *gin.Context) {
	idStr := c.Param("id")
	pluginID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的插件 ID"})
		return
	}

	// 确认插件存在
	if _, err := repository.PluginRepo.GetPlugin(uint(pluginID)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "插件不存在"})
		return
	}

	var req struct {
		Symbol       string          `json:"symbol" binding:"required"`
		DisplayName  string          `json:"display_name"`
		EffectType   string          `json:"effect_type" binding:"required"`
		EffectConfig json.RawMessage `json:"effect_config" binding:"required"`
		DefaultCount int             `json:"default_count"`
		Color        string          `json:"color"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	// 校验 effect_type
	validTypes := map[string]bool{"swap": true, "force_play": true, "convert": true}
	if !validTypes[req.EffectType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "effect_type 必须为 swap / force_play / convert"})
		return
	}

	// 校验 effect_config 结构
	if err := validateEffectConfig(req.EffectType, req.EffectConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "effect_config 格式错误: " + err.Error()})
		return
	}

	defaultCount := req.DefaultCount
	if defaultCount <= 0 {
		defaultCount = 2
	}
	if defaultCount > 20 {
		defaultCount = 20
	}

	card := &database.PluginCard{
		PluginID:     uint(pluginID),
		Symbol:       req.Symbol,
		DisplayName:  req.DisplayName,
		EffectType:   req.EffectType,
		EffectConfig: string(req.EffectConfig),
		DefaultCount: defaultCount,
		Color:        req.Color,
		CreatedAt:    time.Now(),
	}

	if err := repository.PluginRepo.CreateCard(card); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建卡牌失败: " + err.Error()})
		return
	}

	game.LoadPluginCards()
	plugins.LoadServerScripts()
	c.JSON(http.StatusCreated, card)
}

// UpdatePluginCard 更新卡牌定义
func UpdatePluginCard(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的卡牌 ID"})
		return
	}

	card, err := repository.PluginRepo.GetCard(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "卡牌不存在"})
		return
	}

	var req struct {
		Symbol       *string          `json:"symbol"`
		DisplayName  *string          `json:"display_name"`
		EffectType   *string          `json:"effect_type"`
		EffectConfig *json.RawMessage `json:"effect_config"`
		DefaultCount *int             `json:"default_count"`
		Color        *string          `json:"color"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	if req.Symbol != nil {
		newSymbol := strings.ToUpper(strings.TrimSpace(*req.Symbol))
		if newSymbol == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "symbol 不能为空"})
			return
		}
		// 若 symbol 发生变更，检查新符号是否已被其他卡牌占用
		if newSymbol != card.Symbol {
			if existing, err := repository.PluginRepo.GetCardBySymbol(newSymbol); err == nil && existing != nil {
				c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("symbol「%s」已被卡牌 ID %d 使用", newSymbol, existing.ID)})
				return
			}
		}
		card.Symbol = newSymbol
	}
	if req.DisplayName != nil {
		card.DisplayName = *req.DisplayName
	}
	if req.EffectType != nil {
		validTypes := map[string]bool{"swap": true, "force_play": true, "convert": true}
		if !validTypes[*req.EffectType] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "effect_type 必须为 swap / force_play / convert"})
			return
		}
		card.EffectType = *req.EffectType
	}
	if req.EffectConfig != nil {
		if err := validateEffectConfig(card.EffectType, *req.EffectConfig); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "effect_config 格式错误: " + err.Error()})
			return
		}
		card.EffectConfig = string(*req.EffectConfig)
	}
	if req.DefaultCount != nil {
		v := *req.DefaultCount
		if v <= 0 {
			v = 2
		}
		if v > 20 {
			v = 20
		}
		card.DefaultCount = v
	}
	if req.Color != nil {
		card.Color = *req.Color
	}

	if err := repository.PluginRepo.UpdateCard(card); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新卡牌失败: " + err.Error()})
		return
	}

	game.LoadPluginCards()
	plugins.LoadServerScripts()
	c.JSON(http.StatusOK, card)
}

// DeletePluginCard 删除卡牌
func DeletePluginCard(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的卡牌 ID"})
		return
	}

	if err := repository.PluginRepo.DeleteCard(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除卡牌失败: " + err.Error()})
		return
	}

	game.LoadPluginCards()
	plugins.LoadServerScripts()
	c.JSON(http.StatusOK, gin.H{"message": "卡牌已删除"})
}

// ReloadPlugins 热重载插件 registry（无需重启服务）
func ReloadPlugins(c *gin.Context) {
	game.LoadPluginCards()
	plugins.LoadServerScripts()
	c.JSON(http.StatusOK, gin.H{"message": "插件 registry 已重载", "count": len(game.GetAllPluginCards())})
}

// validateEffectConfig 校验 effect_config 的 JSON 结构是否符合对应 effectType 的要求
func validateEffectConfig(effectType string, raw json.RawMessage) error {
	switch effectType {
	case "swap":
		var cfg struct {
			Count int `json:"count"`
		}
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return err
		}
		if cfg.Count <= 0 {
			return errors.New("swap.count 必须大于 0")
		}
		if cfg.Count > 20 {
			return errors.New("swap.count 不能超过 20")
		}
	case "force_play":
		var cfg struct {
			Count int `json:"count"`
		}
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return err
		}
		if cfg.Count <= 0 {
			return errors.New("force_play.count 必须大于 0")
		}
		if cfg.Count > 10 {
			return errors.New("force_play.count 不能超过 10")
		}
	case "convert":
		var cfg struct {
			SourceCount int `json:"source_count"`
			TargetCount int `json:"target_count"`
		}
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return err
		}
		if cfg.SourceCount <= 0 {
			return errors.New("convert.source_count 必须大于 0")
		}
		if cfg.SourceCount > 10 {
			return errors.New("convert.source_count 不能超过 10")
		}
		if cfg.TargetCount <= 0 {
			return errors.New("convert.target_count 必须大于 0")
		}
		if cfg.TargetCount > 20 {
			return errors.New("convert.target_count 不能超过 20")
		}
	}
	return nil
}
