package game

import (
	"chemistryuno/database"
	"chemistryuno/websocket"
	"database/sql"
	"log"
	"time"
)

// StartCron 启动定时任务触发器
func StartCron() {
	// 1. 每分钟清理过期公告的任务
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			countExpired()
		}
	}()

	// 2. 每 10 分钟自动广播一次活跃的滚动公告
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			BroadcastActiveTicker()
		}
	}()

	// 3. 每小时清理过期会话
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			cleanupSessions()
		}
	}()
}

func countExpired() {
	// 查找并禁用过期的公告
	res, err := database.DB.Exec("UPDATE announcements SET active = 0 WHERE active = 1 AND expires_at IS NOT NULL AND expires_at < ?", time.Now())
	if err == nil {
		if count, _ := res.RowsAffected(); count > 0 {
			log.Printf("⚖️ Cron: 已自动清理 %d 条过期公告", count)
		}
	}
}

func cleanupSessions() {
	// 清理超过 24 小时未活动的会话
	res, err := database.DB.Exec("DELETE FROM user_sessions WHERE last_active < DATETIME('now', '-24 hours')")
	if err == nil {
		if count, _ := res.RowsAffected(); count > 0 {
			log.Printf("⚖️ Cron: 已清理 %d 个过期会话", count)
		}
	}
}

// BroadcastActiveTicker 广播所有标记为 ticker 的活跃公告
func BroadcastActiveTicker() {
	if websocket.GlobalHub == nil {
		return
	}

	rows, err := database.DB.Query("SELECT id, title, content, type, is_ticker FROM announcements WHERE active = 1 AND is_ticker = 1")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var title sql.NullString
		var content, aType string
		var isTicker bool
		if err := rows.Scan(&id, &title, &content, &aType, &isTicker); err == nil {
			msg := websocket.Message{
				Type: "system_announcement",
				Data: map[string]interface{}{
					"id":        id,
					"content":   content,
					"type":      aType,
					"is_ticker": isTicker,
				},
			}
			if title.Valid {
				msg.Data.(map[string]interface{})["title"] = title.String
			}
			websocket.GlobalHub.BroadcastToAll(msg)
		}
	}
}
