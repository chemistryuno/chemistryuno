package repository

import (
	"chemistryuno/backend/database"
	"fmt"
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

// FindByUserUID 查找用户的游戏历史
func (r *GameRepository) FindByUserUID(uid uint) ([]database.GameHistory, error) {
	var histories []database.GameHistory
	query := r.db.Order("created_at DESC").Limit(50)

	// 跨数据库兼容的 JSON 查询
	if r.db.Dialector.Name() == "mysql" {
		query = query.Where("JSON_CONTAINS(players, ?)", uid)
	} else {
		// SQLite：精确边界匹配，避免 UID=12 误命中 UID=123
		// players 存储格式为 JSON 数组，如 [1,2,12,123]
		id := fmt.Sprintf("%d", uid)
		query = query.Where(
			"players = ? OR players LIKE ? OR players LIKE ? OR players LIKE ?",
			"["+id+"]",          // 唯一元素 [12]
			"["+id+",%",         // 首元素   [12,...
			"%,"+id+"]",         // 末元素   ...,12]
			"%,"+id+",%",        // 中间元素  ...,12,...
		)
	}

	err := query.Find(&histories).Error
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

// FindRecentByUserUID 查找用户最近的游戏历史
func (r *GameRepository) FindRecentByUserUID(uid uint, limit int) ([]database.GameHistory, error) {
	var histories []database.GameHistory
	query := r.db.Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	// 跨数据库兼容的 JSON 查询
	if r.db.Dialector.Name() == "mysql" {
		query = query.Where("JSON_CONTAINS(players, ?)", uid)
	} else {
		// SQLite：精确边界匹配，避免 UID=12 误命中 UID=123
		id := fmt.Sprintf("%d", uid)
		query = query.Where(
			"players = ? OR players LIKE ? OR players LIKE ? OR players LIKE ?",
			"["+id+"]",
			"["+id+",%",
			"%,"+id+"]",
			"%,"+id+",%",
		)
	}

	err := query.Find(&histories).Error
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
	ID                  uint          `json:"id"`
	RoomID              string        `json:"room_id"`
	WinnerUID           *uint         `json:"winner_uid"`
	WinnerName          string        `json:"winner_name"`
	Players             database.JSON `json:"players"`
	OriginalPlayerCount int           `json:"original_player_count"`
	QuittedCount        int           `json:"quitted_count"`
	StartedAt           string        `json:"started_at"`
	FinishedAt          string        `json:"finished_at"`
}

// FindAllWithWinner 获取游戏历史（带胜者名称）
func (r *GameRepository) FindAllWithWinner(limit int) ([]GameHistoryWithWinner, error) {
	var results []GameHistoryWithWinner

	query := r.db.Table("game_history").
		Select("game_history.id, game_history.room_id, game_history.winner_uid, users.username as winner_name, game_history.players, game_history.original_player_count, game_history.quitted_count, game_history.started_at, game_history.finished_at").
		Joins("LEFT JOIN users ON game_history.winner_uid = users.uid").
		Order("game_history.created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Scan(&results).Error
	return results, err
}
