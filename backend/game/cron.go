package game

import (
	"chemistryuno/backend/repository"
	"chemistryuno/backend/websocket"
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

	// 4. 每小时清除24小时前的大厅聊天消息
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			cleanupOldChatMessages()
		}
	}()

	// 5. 每小时清理已过期的游戏回放（永久保留除外）
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			cleanupExpiredGameReplays()
		}
	}()
}

func countExpired() {
	// 查找并禁用过期的公告
	count, err := repository.AnnouncementRepo.DeactivateExpired()
	if err == nil && count > 0 {
		log.Printf("⚖️ Cron: 已自动清理 %d 条过期公告", count)
	}
}

func cleanupSessions() {
	// 清理超过 24 小时未活动的会话
	count, err := repository.SessionRepo.CleanupInactive()
	if err == nil && count > 0 {
		log.Printf("⚖️ Cron: 已清理 %d 个过期会话", count)
	}
}

func cleanupOldChatMessages() {
	// 清理24小时前的大厅聊天消息
	count, err := repository.ChatRepo.DeleteOldMessages()
	if err != nil {
		log.Printf("⚠️ Cron: 清理过期大厅聊天失败: %v", err)
	} else if count > 0 {
		log.Printf("✅ Cron: 已清理 %d 条超过24小时的大厅聊天消息", count)
	}
}

func cleanupExpiredGameReplays() {
	count, err := repository.GameRepo.CleanupExpiredReplays(time.Now())
	if err != nil {
		log.Printf("⚠️ Cron: 清理过期回放失败: %v", err)
	} else if count > 0 {
		log.Printf("✅ Cron: 已清理 %d 条超过保留期的游戏回放", count)
	}
}

// ProcessScheduledAnnouncements 检查所有活跃公告的定时触发情况
func ProcessScheduledAnnouncements() {
	if websocket.GlobalHub == nil {
		return
	}

	// 查找所有设置了定时任务且当前处于活跃状态的公告
	announcements, err := repository.AnnouncementRepo.FindCronAnnouncements()
	if err != nil {
		return
	}

	now := time.Now()
	for _, announcement := range announcements {
		// 检查是否过期
		if announcement.ExpiresAt != nil && announcement.ExpiresAt.Before(now) {
			continue
		}

		shouldBroadcast := false
		if announcement.LastBroadcastAt == nil {
			shouldBroadcast = true
		} else if now.Sub(*announcement.LastBroadcastAt).Minutes() >= float64(announcement.CronInterval) {
			shouldBroadcast = true
		}

		if shouldBroadcast {
			// 更新最后广播时间
			_ = repository.AnnouncementRepo.UpdateLastBroadcast(announcement.ID, now)

			msg := websocket.Message{
				Type: "system_announcement",
				Data: map[string]interface{}{
					"id":            announcement.ID,
					"content":       announcement.Content,
					"type":          announcement.Type,
					"is_ticker":     announcement.IsTicker,
					"is_persistent": announcement.IsPersistent,
					"close_delay":   announcement.CloseDelay,
				},
			}
			if announcement.Title != "" {
				msg.Data.(map[string]interface{})["title"] = announcement.Title
			}
			websocket.GlobalHub.BroadcastToAll(msg)
		}
	}
}

// PushOnJoinAnnouncements 向新连接的玩家推送 "加入时触发" 的公告
func PushOnJoinAnnouncements(client *websocket.Client) {
	announcements, err := repository.AnnouncementRepo.FindOnJoinAnnouncements()
	if err != nil {
		return
	}

	for _, announcement := range announcements {
		msg := websocket.Message{
			Type: "system_announcement",
			Data: map[string]interface{}{
				"id":            announcement.ID,
				"content":       announcement.Content,
				"type":          announcement.Type,
				"is_ticker":     announcement.IsTicker,
				"is_persistent": announcement.IsPersistent,
				"close_delay":   announcement.CloseDelay,
				"on_join":       true,
			},
		}
		if announcement.Title != "" {
			msg.Data.(map[string]interface{})["title"] = announcement.Title
		}

		client.Send(msg)
	}
}
