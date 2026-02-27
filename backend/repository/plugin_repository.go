package repository

import (
	"chemistryuno/backend/database"

	"gorm.io/gorm"
)

// PluginRepository 插件数据访问层
type PluginRepository struct {
	db *gorm.DB
}

// NewPluginRepository 创建 PluginRepository 实例
func NewPluginRepository() *PluginRepository {
	return &PluginRepository{db: database.DB}
}

// --- Plugin 操作 ---

// GetAllPlugins 获取所有插件
func (r *PluginRepository) GetAllPlugins() ([]database.Plugin, error) {
	var plugins []database.Plugin
	err := r.db.Order("created_at DESC").Find(&plugins).Error
	return plugins, err
}

// GetPlugin 根据 ID 获取插件
func (r *PluginRepository) GetPlugin(id uint) (*database.Plugin, error) {
	var plugin database.Plugin
	err := r.db.First(&plugin, id).Error
	if err != nil {
		return nil, err
	}
	return &plugin, nil
}

// CreatePlugin 创建插件
func (r *PluginRepository) CreatePlugin(p *database.Plugin) error {
	return r.db.Create(p).Error
}

// UpdatePlugin 更新插件
func (r *PluginRepository) UpdatePlugin(p *database.Plugin) error {
	return r.db.Save(p).Error
}

// DeletePlugin 删除插件及其卡牌
func (r *PluginRepository) DeletePlugin(id uint) error {
	if err := r.db.Where("plugin_id = ?", id).Delete(&database.PluginCard{}).Error; err != nil {
		return err
	}
	return r.db.Delete(&database.Plugin{}, id).Error
}

// --- PluginCard 操作 ---

// GetCardsByPlugin 获取插件的所有卡牌
func (r *PluginRepository) GetCardsByPlugin(pluginID uint) ([]database.PluginCard, error) {
	var cards []database.PluginCard
	err := r.db.Where("plugin_id = ?", pluginID).Order("created_at ASC").Find(&cards).Error
	return cards, err
}

// GetAllActiveCards 获取所有激活插件的卡牌
func (r *PluginRepository) GetAllActiveCards() ([]database.PluginCard, error) {
	var cards []database.PluginCard
	err := r.db.
		Joins("JOIN plugins ON plugins.id = plugin_cards.plugin_id").
		Where("plugins.is_active = ?", true).
		Find(&cards).Error
	return cards, err
}

// GetCard 根据 ID 获取卡牌
func (r *PluginRepository) GetCard(id uint) (*database.PluginCard, error) {
	var card database.PluginCard
	err := r.db.First(&card, id).Error
	if err != nil {
		return nil, err
	}
	return &card, nil
}

// GetPluginByHash 通过 .cumod 文件哈希查找插件（防重复安装）
func (r *PluginRepository) GetPluginByHash(hash string) (*database.Plugin, error) {
	if hash == "" {
		return nil, nil
	}
	var plugin database.Plugin
	err := r.db.Where("cumod_hash = ?", hash).First(&plugin).Error
	if err != nil {
		return nil, err
	}
	return &plugin, nil
}

// GetCardBySymbol 通过 symbol 查找卡牌（用于唯一性前置检查）
func (r *PluginRepository) GetCardBySymbol(symbol string) (*database.PluginCard, error) {
	var card database.PluginCard
	err := r.db.Where("symbol = ?", symbol).First(&card).Error
	if err != nil {
		return nil, err
	}
	return &card, nil
}

// CreateCard 创建插件卡牌
func (r *PluginRepository) CreateCard(c *database.PluginCard) error {
	return r.db.Create(c).Error
}

// UpdateCard 更新插件卡牌
func (r *PluginRepository) UpdateCard(c *database.PluginCard) error {
	return r.db.Save(c).Error
}

// DeleteCard 删除插件卡牌
func (r *PluginRepository) DeleteCard(id uint) error {
	return r.db.Delete(&database.PluginCard{}, id).Error
}
