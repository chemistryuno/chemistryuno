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

// GameHistoryWithWinner 带胜者信息的游戏历史
type GameHistoryWithWinner struct {
	ID         uint   `json:"id"`
	RoomID     string `json:"room_id"`
	WinnerUID  *uint  `json:"winner_uid"`
	WinnerName string `json:"winner_name"`
	Players    string `json:"players"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at"`
}

// FindAllWithWinner 获取游戏历史（带胜者名称）
func (r *GameRepository) FindAllWithWinner(limit int) ([]GameHistoryWithWinner, error) {
	var results []GameHistoryWithWinner

	query := r.db.Table("game_history").
		Select("game_history.id, game_history.room_id, game_history.winner_uid, users.username as winner_name, game_history.players, game_history.started_at, game_history.finished_at").
		Joins("LEFT JOIN users ON game_history.winner_uid = users.uid").
		Order("game_history.created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Scan(&results).Error
	return results, err
}
