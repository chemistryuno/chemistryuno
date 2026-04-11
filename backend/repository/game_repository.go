package repository

import (
	"chemistryuno/backend/database"
	"encoding/json"
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
	idStr := fmt.Sprintf("%d", uid)

	// 无论数据库环境，先通过 LIKE 进行初步筛选以利用索引（如果存在）
	// 同时使用四重边界锁定确保初步筛选的覆盖度
	query := r.db.Order("created_at DESC").Limit(50).Where(
		"players LIKE ? OR players LIKE ? OR players LIKE ? OR players LIKE ?",
		"["+idStr+"]",   // 唯一元素
		"["+idStr+",%",  // 数组开头
		"%,"+idStr+"]",  // 数组结尾
		"%,"+idStr+",%", // 数组中间
	)

	err := query.Find(&histories).Error
	if err != nil {
		return nil, err
	}

	// 在内存中进行二次严格逻辑校验，确保 100% 精确
	var result []database.GameHistory
	for _, h := range histories {
		var players []int
		if err := json.Unmarshal([]byte(h.Players), &players); err == nil {
			for _, pid := range players {
				if pid == int(uid) {
					result = append(result, h)
					break
				}
			}
		}
	}

	return result, nil
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
	IsInvalid           bool          `json:"is_invalid"`
	InvalidReason       string        `json:"invalid_reason"`
	ReplayPermanent     bool          `json:"replay_permanent"`
	ReplayExpiresAt     *time.Time    `json:"replay_expires_at"`
	ReplayClearedAt     *time.Time    `json:"replay_cleared_at"`
	CheatDetected       bool          `json:"cheat_detected"`
	CheatUIDs           database.JSON `json:"cheat_uids"`
	HasReplay           bool          `json:"has_replay"`
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
		Select("game_history.id, game_history.room_id, game_history.winner_uid, users.username as winner_name, game_history.is_invalid, game_history.invalid_reason, game_history.replay_permanent, game_history.replay_expires_at, game_history.replay_cleared_at, game_history.cheat_detected, game_history.cheat_uids, CASE WHEN game_history.replay_log IS NULL OR game_history.replay_log = '' THEN FALSE ELSE TRUE END as has_replay, game_history.players, game_history.original_player_count, game_history.quitted_count, game_history.started_at, game_history.finished_at").
		Joins("LEFT JOIN users ON game_history.winner_uid = users.uid").
		Order("game_history.created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Scan(&results).Error
	return results, err
}

// FindByID 按历史记录 ID 查询
func (r *GameRepository) FindByID(id uint) (*database.GameHistory, error) {
	var history database.GameHistory
	err := r.db.First(&history, id).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// ClearReplayByID 清空某条历史的回放内容（管理员手动消除）
func (r *GameRepository) ClearReplayByID(id uint) error {
	now := time.Now()
	result := r.db.Model(&database.GameHistory{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"replay_log":        "",
			"replay_permanent":  false,
			"replay_expires_at": nil,
			"replay_cleared_at": &now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CleanupExpiredReplays 清理超过保留期限的普通回放（永久回放不受影响）
func (r *GameRepository) CleanupExpiredReplays(now time.Time) (int64, error) {
	result := r.db.Model(&database.GameHistory{}).
		Where("replay_permanent = ?", false).
		Where("replay_log IS NOT NULL AND replay_log != ''").
		Where("replay_expires_at IS NOT NULL AND replay_expires_at < ?", now).
		Updates(map[string]interface{}{
			"replay_log":        "",
			"replay_expires_at": nil,
			"replay_cleared_at": &now,
		})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}
