package repository

import (
	"chemistryuno/database"

	"gorm.io/gorm"
)

type BountyRepository struct {
	db *gorm.DB
}

type Bounty struct {
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	TargetUID uint   `gorm:"not null;index" json:"target_uid"`
	Amount    int    `gorm:"not null" json:"amount"`
	CreatedBy uint   `gorm:"not null" json:"created_by"`
	Status    string `gorm:"size:20;default:active" json:"status"`
}

func (Bounty) TableName() string {
	return "bounties"
}

func NewBountyRepository() *BountyRepository {
	return &BountyRepository{db: database.DB}
}

// GetTotalBounty 获取用户的总悬赏金额
func (r *BountyRepository) GetTotalBounty(targetUID uint) (int, error) {
	var total int
	err := r.db.Model(&Bounty{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("target_uid = ? AND status = ?", targetUID, "active").
		Scan(&total).Error
	return total, err
}

// FindActiveByTarget 查找针对目标用户的活跃悬赏
func (r *BountyRepository) FindActiveByTarget(targetUID uint) ([]Bounty, error) {
	var bounties []Bounty
	err := r.db.Where("target_uid = ? AND status = ?", targetUID, "active").
		Find(&bounties).Error
	return bounties, err
}

// UpdateStatus 更新悬赏状态
func (r *BountyRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&Bounty{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// Create 创建悬赏
func (r *BountyRepository) Create(bounty *Bounty) error {
	return r.db.Create(bounty).Error
}
