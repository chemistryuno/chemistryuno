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

	// 2. 检查并处理定时公告触发器
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			ProcessScheduledAnnouncements()
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
	res, err := database.LegacyDB.Exec("UPDATE announcements SET active = 0 WHERE active = 1 AND expires_at IS NOT NULL AND expires_at < ?", time.Now())
	if err == nil {
		if count, _ := res.RowsAffected(); count > 0 {
			log.Printf("⚖️ Cron: 已自动清理 %d 条过期公告", count)
		}
	}
}

func cleanupSessions() {
	// 清理超过 24 小时未活动的会话
	res, err := database.LegacyDB.Exec("DELETE FROM user_sessions WHERE last_active < DATE_SUB(NOW(), INTERVAL 24 HOUR)")
	if err == nil {
		if count, _ := res.RowsAffected(); count > 0 {
			log.Printf("⚖️ Cron: 已清理 %d 个过期会话", count)
		}
	}
}

// ProcessScheduledAnnouncements 检查所有活跃公告的定时触发情况
func ProcessScheduledAnnouncements() {
	if websocket.GlobalHub == nil {
		return
	}

	// 查找所有设置了定时任务且当前处于活跃状态的公告
	rows, err := database.LegacyDB.Query(`
		SELECT id, title, content, type, cron_interval, last_broadcast_at, is_ticker, is_persistent, close_delay 
		FROM announcements 
		WHERE active = 1 AND cron_interval > 0 AND (expires_at IS NULL OR expires_at > ?)`, time.Now())
	if err != nil {
		return
	}
	defer rows.Close()

	now := time.Now()
	for rows.Next() {
		var id, interval, closeDelay int
		var title sql.NullString
		var content, aType string
		var lastBroadcast sql.NullTime
		var isTicker, isPersistent bool
		if err := rows.Scan(&id, &title, &content, &aType, &interval, &lastBroadcast, &isTicker, &isPersistent, &closeDelay); err == nil {
			shouldBroadcast := false
			if !lastBroadcast.Valid {
				shouldBroadcast = true
			} else if now.Sub(lastBroadcast.Time).Minutes() >= float64(interval) {
				shouldBroadcast = true
			}

			if shouldBroadcast {
				msg := websocket.Message{
					Type: "system_announcement",
					Data: map[string]interface{}{
						"id":            id,
						"content":       content,
						"type":          aType,
						"is_ticker":     isTicker,
						"is_persistent": isPersistent,
						"close_delay":   closeDelay,
					},
				}
				if title.Valid {
					msg.Data.(map[string]interface{})["title"] = title.String
				}
				websocket.GlobalHub.BroadcastToAll(msg)

				// 更新最后广播时间
				database.LegacyDB.Exec("UPDATE announcements SET last_broadcast_at = ? WHERE id = ?", now, id)
			}
		}
	}
}

// PushOnJoinAnnouncements 向新连接的玩家推送 "加入时触发" 的公告
func PushOnJoinAnnouncements(client *websocket.Client) {
	rows, err := database.LegacyDB.Query(`
		SELECT id, title, content, type, is_ticker, is_persistent, close_delay 
		FROM announcements 
		WHERE active = 1 AND on_join = 1 AND (expires_at IS NULL OR expires_at > ?)`, time.Now())
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var title sql.NullString
		var content, aType string
		var isTicker, isPersistent bool
		var closeDelay int
		if err := rows.Scan(&id, &title, &content, &aType, &isTicker, &isPersistent, &closeDelay); err == nil {
			msg := websocket.Message{
				Type: "system_announcement",
				Data: map[string]interface{}{
					"id":            id,
					"content":       content,
					"type":          aType,
					"is_ticker":     isTicker,
					"is_persistent": isPersistent,
					"close_delay":   closeDelay,
					"on_join":       true,
				},
			}
			if title.Valid {
				msg.Data.(map[string]interface{})["title"] = title.String
			}

			client.Send(msg)
		}
	}
}
