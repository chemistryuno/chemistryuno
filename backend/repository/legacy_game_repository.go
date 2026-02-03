package repository

import (
	"database/sql"
)

// LegacyGameRepository 处理旧表结构的游戏历史Repository
type LegacyGameRepository struct {
	db *sql.DB
}

func NewLegacyGameRepository(db *sql.DB) *LegacyGameRepository {
	return &LegacyGameRepository{db: db}
}

// GameHistoryInfo 游戏历史信息
type GameHistoryInfo struct {
	ID          int
	RoomID      string
	WinnerUID   int
	WinnerName  string
	PlayersJSON string
	Players     []int
	StartedAt   string
	FinishedAt  string
}

// GetGameHistory 获取游戏历史（最近100条）
func (r *LegacyGameRepository) GetGameHistory(limit int) ([]GameHistoryInfo, error) {
	rows, err := r.db.Query(`
		SELECT gh.id, gh.room_id, COALESCE(gh.winner_uid, 0), COALESCE(u.username, '未结算'), gh.players, COALESCE(gh.started_at, ''), COALESCE(gh.finished_at, '')
		FROM game_history gh
		LEFT JOIN users u ON gh.winner_uid = u.UID
		ORDER BY gh.finished_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []GameHistoryInfo
	for rows.Next() {
		var info GameHistoryInfo
		err := rows.Scan(&info.ID, &info.RoomID, &info.WinnerUID, &info.WinnerName,
			&info.PlayersJSON, &info.StartedAt, &info.FinishedAt)
		if err != nil {
			continue
		}
		results = append(results, info)
	}
	return results, nil
}
