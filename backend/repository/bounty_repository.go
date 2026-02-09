package repository

import (
	"chemistryuno/backend/database"

	"gorm.io/gorm"
)

type BountyRepository struct {
	db *gorm.DB
}

func NewBountyRepository() *BountyRepository {
	return &BountyRepository{db: database.DB}
}

// GetTotalBounty 获取用户的总悬赏金额
func (r *BountyRepository) GetTotalBounty(targetUID uint) (int, error) {
	var total int
	err := r.db.Model(&database.Bounty{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("target_uid = ? AND status = ?", targetUID, "active").
		Scan(&total).Error
	return total, err
}

// FindActiveByTarget 查找针对目标用户的活跃悬赏
func (r *BountyRepository) FindActiveByTarget(targetUID uint) ([]database.Bounty, error) {
	var bounties []database.Bounty
	err := r.db.Where("target_uid = ? AND status = ?", targetUID, "active").
		Find(&bounties).Error
	return bounties, err
}

// UpdateStatus 更新悬赏状态
func (r *BountyRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&database.Bounty{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// Create 创建悬赏
func (r *BountyRepository) Create(bounty *database.Bounty) error {
	return r.db.Create(bounty).Error
}
