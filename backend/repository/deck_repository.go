package repository

import (
	"chemistryuno/database"

	"gorm.io/gorm"
)

type DeckRepository struct {
	db *gorm.DB
}

func NewDeckRepository() *DeckRepository {
	return &DeckRepository{db: database.DB}
}

// FindGlobalDecks 查找所有全局牌组
func (r *DeckRepository) FindGlobalDecks() ([]database.DeckConfig, error) {
	var decks []database.DeckConfig
	err := r.db.Where("is_global = ?", true).Find(&decks).Error
	return decks, err
}

// FindGlobalDeck 查找第一个全局牌组（用于默认选择）
func (r *DeckRepository) FindGlobalDeck() (*database.DeckConfig, error) {
	var deck database.DeckConfig
	err := r.db.Where("is_global = ?", true).First(&deck).Error
	if err != nil {
		return nil, err
	}
	return &deck, nil
}

// FindByUserUID 查找用户的所有牌组
func (r *DeckRepository) FindByUserUID(uid uint) ([]database.DeckConfig, error) {
	var decks []database.DeckConfig
	err := r.db.Where("created_by_uid = ? AND is_global = ?", uid, false).
		Order("created_at DESC").
		Find(&decks).Error
	return decks, err
}

// FindByID 根据ID查找牌组
func (r *DeckRepository) FindByID(id uint) (*database.DeckConfig, error) {
	var deck database.DeckConfig
	err := r.db.First(&deck, id).Error
	if err != nil {
		return nil, err
	}
	return &deck, nil
}

// Create 创建牌组
func (r *DeckRepository) Create(deck *database.DeckConfig) error {
	return r.db.Create(deck).Error
}

// Update 更新牌组
func (r *DeckRepository) Update(deck *database.DeckConfig) error {
	return r.db.Save(deck).Error
}

// Delete 删除牌组
func (r *DeckRepository) Delete(id uint) error {
	return r.db.Delete(&database.DeckConfig{}, id).Error
}

// UpdateGlobalDeck 更新全局牌组
func (r *DeckRepository) UpdateGlobalDeck(name string, cards []byte) error {
	return r.db.Model(&database.DeckConfig{}).
		Where("is_global = ?", true).
		Updates(map[string]interface{}{
			"name":  name,
			"cards": cards,
		}).Error
}
