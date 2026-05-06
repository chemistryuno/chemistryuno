package handlers

import (
	"time"

	"chemistryuno/backend/websocket"

	"github.com/gin-gonic/gin"
)

func banNotificationMessage(bannedUntil *time.Time, reason string) websocket.Message {
	message := websocket.Message{
		Type:    "ban_notification",
		Message: "您的账号已被封禁，当前登录状态将保留。系统将返回大厅，封禁期间部分功能会受到限制。",
		Data: gin.H{
			"ban_reason":  reason,
			"redirect_to": "/",
		},
	}
	if bannedUntil != nil {
		message.Data.(gin.H)["banned_until"] = bannedUntil.Format(time.RFC3339)
	}
	return message
}

func sendBanNotification(playerUID uint, bannedUntil *time.Time, reason string) {
	if websocket.GlobalHub == nil || bannedUntil == nil {
		return
	}

	websocket.GlobalHub.SendToUID(int(playerUID), banNotificationMessage(bannedUntil, reason))
}
