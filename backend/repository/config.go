package repository

import (
	"chemistryuno/backend/database"
	"strconv"
	"time"
)

// ConfigRepository 配置仓库
type ConfigRepository struct{}

// NewConfigRepository 创建配置仓库实例
func NewConfigRepository() *ConfigRepository {
	return &ConfigRepository{}
}

// GetValue 获取配置值
func (r *ConfigRepository) GetValue(key string) (string, error) {
	var config database.SystemConfig
	err := database.DB.Where("`key` = ?", key).Take(&config).Error
	if err != nil {
		return "", err
	}
	return config.Value, nil
}

// GetIntValue 获取整数配置值（秒）
func (r *ConfigRepository) GetIntValue(key string, defaultValue int) int {
	value, err := r.GetValue(key)
	if err != nil {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// GetDurationValue 获取时间间隔配置值
func (r *ConfigRepository) GetDurationValue(key string, defaultValue time.Duration) time.Duration {
	value, err := r.GetValue(key)
	if err != nil {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return time.Duration(intValue) * time.Second
}

// GetBoolValue 获取布尔配置值
func (r *ConfigRepository) GetBoolValue(key string, defaultValue bool) bool {
	value, err := r.GetValue(key)
	if err != nil {
		return defaultValue
	}
	return value == "true"
}

// SetValue 设置配置值
func (r *ConfigRepository) SetValue(key, value string) error {
	config := database.SystemConfig{
		Key:   key,
		Value: value,
	}
	return database.DB.Save(&config).Error
}

// GetAll 获取所有配置
func (r *ConfigRepository) GetAll() (map[string]string, error) {
	var configs []database.SystemConfig
	err := database.DB.Find(&configs).Error
	if err != nil {
		return nil, err
	}

	configMap := make(map[string]string)
	for _, cfg := range configs {
		configMap[cfg.Key] = cfg.Value
	}
	return configMap, nil
}

// InitDefaultConfigs 初始化默认配置
func (r *ConfigRepository) InitDefaultConfigs() error {
	// 迁移逻辑：如果存在旧的键名，尝试将其迁移到新的键名
	migrationMap := map[string]string{
		"game_turn_timeout":      "player_action_timeout",
		"reconnect_grace_period": "player_kick_timeout",
	}

	for oldKey, newKey := range migrationMap {
		var oldConfig database.SystemConfig
		if err := database.DB.Where("`key` = ?", oldKey).Take(&oldConfig).Error; err == nil {
			// 如果旧键存在，检查新键是否存在
			var newConfig database.SystemConfig
			if err := database.DB.Where("`key` = ?", newKey).Take(&newConfig).Error; err != nil {
				// 新键不存在，则迁移
				database.DB.Create(&database.SystemConfig{
					Key:   newKey,
					Value: oldConfig.Value,
				})
			}
			// 删除旧键
			database.DB.Where("`key` = ?", oldKey).Delete(&database.SystemConfig{})
		}
	}

	defaults := map[string]string{
		"player_kick_timeout":    "30",   // 玩家离线踢出时间（秒）
		"player_action_timeout":  "30",   // 玩家操作时间（秒）
		"auto_start_timeout":     "10",   // 自动开始倒计时（秒）
		"half_ready_timeout":     "60",   // 半数准备倒计时（秒）
		"reconnect_grace_period": "30",   // 掉线重连宽限期（秒）- 预留
		"points_scaling_enabled": "true", // 积分动态缩放 - 预留
	}

	for key, value := range defaults {
		var existing database.SystemConfig
		// 使用原生 SQL 查询键，确保兼容性
		err := database.DB.Where("`key` = ?", key).Take(&existing).Error
		if err != nil {
			// 配置不存在，创建默认值
			config := database.SystemConfig{
				Key:   key,
				Value: value,
			}
			if err := database.DB.Create(&config).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
