package repository

import (
	"chemistryuno/database"
	"time"

	"gorm.io/gorm"
)

type GameRepository struct {
	db *gorm.DB
}

func NewGameRepository() *GameRepository {
	return &GameRepository{db: database.DB}
}

// Create 创建游戏历史记录
func (r *GameRepository) Create(history *database.GameHistory) error {
	return r.db.Create(history).Error
}

// FindByRoomID 根据房间ID查找游戏历史
func (r *GameRepository) FindByRoomID(roomID string) ([]database.GameHistory, error) {
	var histories []database.GameHistory
	err := r.db.Where("room_id = ?", roomID).
		Order("created_at DESC").
		Find(&histories).Error
	return histories, err
}

// FindByUserID 查找用户的游戏历史
func (r *GameRepository) FindByUserID(uid uint) ([]database.GameHistory, error) {
	var histories []database.GameHistory
	// 注意：players字段是JSON，需要使用JSON查询
	err := r.db.Where("JSON_CONTAINS(players, ?)", uid).
		Order("created_at DESC").
		Limit(50).
		Find(&histories).Error
	return histories, err
}

// FindAll 查找所有游戏历史（管理员）
func (r *GameRepository) FindAll(limit int) ([]database.GameHistory, error) {
	var histories []database.GameHistory
	query := r.db.Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&histories).Error
	return histories, err
}

// FindRecentByUserID 查找用户最近的游戏历史
func (r *GameRepository) FindRecentByUserID(uid uint, limit int) ([]database.GameHistory, error) {
	var histories []database.GameHistory
	err := r.db.Where("JSON_CONTAINS(players, ?)", uid).
		Order("created_at DESC").
		Limit(limit).
		Find(&histories).Error
	return histories, err
}

// GetGameStatsByDateRange 获取指定日期范围内的游戏统计
func (r *GameRepository) GetGameStatsByDateRange(startDate, endDate time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&database.GameHistory{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&count).Error
	return count, err
}
