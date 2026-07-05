package anticheat

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	config     *RiskScoringConfig
	configPath string
	lock       sync.RWMutex
	watchers   []func(*RiskScoringConfig)
}

// NewConfigManager 创建配置管理器
func NewConfigManager(configPath string) (*ConfigManager, error) {
	cm := &ConfigManager{
		configPath: configPath,
		watchers:   make([]func(*RiskScoringConfig), 0),
	}

	// 尝试从文件加载配置
	if configPath != "" {
		if err := cm.LoadFromFile(configPath); err != nil {
			log.Printf("[配置] 无法从文件加载: %v，使用默认配置", err)
			cm.config = NewDefaultConfig()
		}
	} else {
		cm.config = NewDefaultConfig()
	}

	return cm, nil
}

// LoadFromFile 从YAML文件加载配置
func (cm *ConfigManager) LoadFromFile(filePath string) error {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	config := NewDefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return err
	}

	cm.config = config
	cm.notifyWatchers()

	log.Printf("[配置] 已从文件加载配置: %s", filePath)
	return nil
}

// LoadFromJSON 从JSON字符串加载配置
func (cm *ConfigManager) LoadFromJSON(jsonStr string) error {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	config := NewDefaultConfig()
	if err := json.Unmarshal([]byte(jsonStr), config); err != nil {
		return err
	}

	cm.config = config
	cm.notifyWatchers()

	log.Printf("[配置] 已从JSON加载配置")
	return nil
}

// SaveToFile 保存配置到YAML文件
func (cm *ConfigManager) SaveToFile(filePath string) error {
	cm.lock.RLock()
	data, err := yaml.Marshal(cm.config)
	cm.lock.RUnlock()

	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	log.Printf("[配置] 已保存配置到: %s", filePath)
	return nil
}

// GetConfig 获取当前配置（线程安全）
func (cm *ConfigManager) GetConfig() *RiskScoringConfig {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	// 返回深拷贝以防止外部修改
	configCopy := &RiskScoringConfig{
		Dimensions:         make(map[string]DimensionConfig),
		SanctionThresholds: cm.config.SanctionThresholds,
		EnabledStrategies:  make([]string, len(cm.config.EnabledStrategies)),
		UnbanConfig:        cm.config.UnbanConfig,
		Optimization:       cm.config.Optimization,
	}

	for k, v := range cm.config.Dimensions {
		configCopy.Dimensions[k] = v
	}

	copy(configCopy.EnabledStrategies, cm.config.EnabledStrategies)

	return configCopy
}

// UpdateDimensionWeight 更新维度的权重
func (cm *ConfigManager) UpdateDimensionWeight(dimension string, weight float64) error {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	dimConfig, exists := cm.config.Dimensions[dimension]
	if !exists {
		return ErrDimensionNotFound
	}

	dimConfig.Weight = weight
	cm.config.Dimensions[dimension] = dimConfig

	cm.notifyWatchers()
	log.Printf("[配置] 已更新维度 %s 的权重: %.2f", dimension, weight)

	return nil
}

// UpdateDimensionThreshold 更新维度的阈值
func (cm *ConfigManager) UpdateDimensionThreshold(dimension string, threshold int64) error {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	dimConfig, exists := cm.config.Dimensions[dimension]
	if !exists {
		return ErrDimensionNotFound
	}

	dimConfig.Threshold = threshold
	cm.config.Dimensions[dimension] = dimConfig

	cm.notifyWatchers()
	log.Printf("[配置] 已更新维度 %s 的阈值: %d", dimension, threshold)

	return nil
}

// ReplaceConfig atomically replaces the active risk scoring configuration.
func (cm *ConfigManager) ReplaceConfig(config *RiskScoringConfig) error {
	if config == nil {
		return ErrInvalidConfig
	}
	if err := ValidateConfig(config); err != nil {
		return err
	}

	cm.lock.Lock()
	defer cm.lock.Unlock()

	configCopy := &RiskScoringConfig{
		Dimensions:         make(map[string]DimensionConfig),
		SanctionThresholds: config.SanctionThresholds,
		EnabledStrategies:  make([]string, len(config.EnabledStrategies)),
		UnbanConfig:        config.UnbanConfig,
		Optimization:       config.Optimization,
	}
	for k, v := range config.Dimensions {
		configCopy.Dimensions[k] = v
	}
	copy(configCopy.EnabledStrategies, config.EnabledStrategies)

	cm.config = configCopy
	cm.notifyWatchers()
	log.Printf("[配置] 已替换反作弊配置")
	return nil
}

// ValidateConfig checks risk scoring and compensation configuration ranges.
func ValidateConfig(config *RiskScoringConfig) error {
	if config == nil {
		return ErrInvalidConfig
	}
	if len(config.Dimensions) == 0 {
		return ErrInvalidConfig
	}
	for _, dim := range config.Dimensions {
		if dim.Weight < 0 || dim.Threshold < 0 || dim.Percentile < 0 || dim.Percentile > 1 {
			return ErrInvalidConfig
		}
	}

	thresholds := config.SanctionThresholds
	if thresholds.ObserveMin < 0 || thresholds.WarningMin < 0 || thresholds.MuteMin < 0 || thresholds.BanMin < 0 ||
		thresholds.ObserveMax > 100 || thresholds.WarningMax > 100 || thresholds.MuteMax > 100 || thresholds.BanMax > 100 ||
		thresholds.ObserveMin > thresholds.ObserveMax || thresholds.WarningMin > thresholds.WarningMax ||
		thresholds.MuteMin > thresholds.MuteMax || thresholds.BanMin > thresholds.BanMax {
		return ErrInvalidConfig
	}

	unban := config.UnbanConfig
	if unban.MinAmount < 0 || unban.MaxAmount < unban.MinAmount ||
		unban.CompensationAmount < unban.MinAmount || unban.CompensationAmount > unban.MaxAmount ||
		unban.MessageMaxLength <= 0 || len(unban.DefaultMessage) > unban.MessageMaxLength ||
		unban.IdempotencyTTL <= 0 {
		return ErrInvalidConfig
	}

	if err := validateOptimizationConfig(config.Optimization); err != nil {
		return err
	}

	return nil
}

// validateOptimizationConfig checks ranges for the optimization feature toggles.
// Validation runs regardless of enable flags so that disabled-but-misconfigured
// values are caught before a later enable.
func validateOptimizationConfig(opt OptimizationConfig) error {
	at := opt.AdaptiveThreshold
	if at.BaselineWindow < 0 || at.MinSamples < 0 ||
		at.PersonalWeight < 0 || at.PersonalWeight > 1 ||
		at.GlobalSuperhumanZ < 0 || at.ContributionWeight < 0 {
		return ErrInvalidConfig
	}
	if at.BaselineWindowKind != "" && at.BaselineWindowKind != "count" && at.BaselineWindowKind != "time" {
		return ErrInvalidConfig
	}

	zs := opt.ZScore
	if zs.Threshold < 0 || zs.Weight < 0 {
		return ErrInvalidConfig
	}

	np := opt.NewPlayer
	if np.MinGames < 0 || np.MinAgeDays < 0 ||
		np.RelaxationFactor < 0 || np.RelaxationFactor > 1 {
		return ErrInvalidConfig
	}

	rd := opt.RiskDecay
	if rd.DecayFactor < 0 || rd.DecayFactor > 1 || rd.MinFloorHours < 0 {
		return ErrInvalidConfig
	}

	return nil
}

// UpdateSanctionThreshold 更新处罚阈值
func (cm *ConfigManager) UpdateSanctionThreshold(sanctionType string, minScore, maxScore float64) error {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	switch sanctionType {
	case "observe":
		cm.config.SanctionThresholds.ObserveMin = minScore
		cm.config.SanctionThresholds.ObserveMax = maxScore
	case "warning":
		cm.config.SanctionThresholds.WarningMin = minScore
		cm.config.SanctionThresholds.WarningMax = maxScore
	case "mute":
		cm.config.SanctionThresholds.MuteMin = minScore
		cm.config.SanctionThresholds.MuteMax = maxScore
	case "ban":
		cm.config.SanctionThresholds.BanMin = minScore
		cm.config.SanctionThresholds.BanMax = maxScore
	default:
		return ErrInvalidSanctionType
	}

	cm.notifyWatchers()
	log.Printf("[配置] 已更新处罚 %s 的阈值: [%.1f, %.1f]", sanctionType, minScore, maxScore)

	return nil
}

// EnableStrategy 启用策略
func (cm *ConfigManager) EnableStrategy(strategyName string) error {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	for _, s := range cm.config.EnabledStrategies {
		if s == strategyName {
			return nil // 已启用
		}
	}

	cm.config.EnabledStrategies = append(cm.config.EnabledStrategies, strategyName)

	cm.notifyWatchers()
	log.Printf("[配置] 已启用策略: %s", strategyName)

	return nil
}

// DisableStrategy 禁用策略
func (cm *ConfigManager) DisableStrategy(strategyName string) error {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	newStrategies := make([]string, 0)
	found := false
	for _, s := range cm.config.EnabledStrategies {
		if s != strategyName {
			newStrategies = append(newStrategies, s)
		} else {
			found = true
		}
	}

	if !found {
		return ErrStrategyNotFound
	}

	cm.config.EnabledStrategies = newStrategies

	cm.notifyWatchers()
	log.Printf("[配置] 已禁用策略: %s", strategyName)

	return nil
}

// Watch 注册配置变化监听器
func (cm *ConfigManager) Watch(callback func(*RiskScoringConfig)) {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	cm.watchers = append(cm.watchers, callback)
}

// notifyWatchers 通知所有监听器
func (cm *ConfigManager) notifyWatchers() {
	// 在持有锁的情况下调用回调可能导致死锁，所以先拷贝配置
	config := &RiskScoringConfig{
		Dimensions:         make(map[string]DimensionConfig),
		SanctionThresholds: cm.config.SanctionThresholds,
		EnabledStrategies:  make([]string, len(cm.config.EnabledStrategies)),
		UnbanConfig:        cm.config.UnbanConfig,
		Optimization:       cm.config.Optimization,
	}

	for k, v := range cm.config.Dimensions {
		config.Dimensions[k] = v
	}

	copy(config.EnabledStrategies, cm.config.EnabledStrategies)

	// 释放锁后通知
	for _, watcher := range cm.watchers {
		go func(w func(*RiskScoringConfig)) {
			w(config)
		}(watcher)
	}
}

// GetDimensions 获取所有维度配置
func (cm *ConfigManager) GetDimensions() map[string]DimensionConfig {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	result := make(map[string]DimensionConfig)
	for k, v := range cm.config.Dimensions {
		result[k] = v
	}
	return result
}

// GetSanctionThresholds 获取处罚阈值
func (cm *ConfigManager) GetSanctionThresholds() SanctionThresholds {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	return cm.config.SanctionThresholds
}

// GetEnabledStrategies 获取启用的策略列表
func (cm *ConfigManager) GetEnabledStrategies() []string {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	result := make([]string, len(cm.config.EnabledStrategies))
	copy(result, cm.config.EnabledStrategies)
	return result
}

// ExportJSON 导出配置为JSON字符串
func (cm *ConfigManager) ExportJSON() (string, error) {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Reset 重置为默认配置
func (cm *ConfigManager) Reset() {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	cm.config = NewDefaultConfig()
	cm.notifyWatchers()
	log.Printf("[配置] 已重置为默认配置")
}
