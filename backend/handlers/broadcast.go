package handlers

import (
	"chemistryuno/backend/game"
	"chemistryuno/backend/websocket"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AdminBroadcast 向不同范围推送管理广播
//
//	scope   : "global" | "room" | "user"
//	target  : roomID (room) | UID 字符串 (user) | 空 (global)
//	msg_type: "info" | "warning" | "success" | "error"
//	title   : 可选，标题
//	content : 消息正文（必填）
func AdminBroadcast(c *gin.Context) {
	var req struct {
		Scope   string `json:"scope"   binding:"required"`
		Target  string `json:"target"`
		MsgType string `json:"msg_type" binding:"required"`
		Title   string `json:"title"`
		Content string `json:"content"  binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	validScopes := map[string]bool{"global": true, "room": true, "user": true}
	if !validScopes[req.Scope] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scope 必须为 global / room / user"})
		return
	}

	validTypes := map[string]bool{"info": true, "warning": true, "success": true, "error": true}
	if !validTypes[req.MsgType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "msg_type 必须为 info / warning / success / error"})
		return
	}

	hub := websocket.GlobalHub
	if hub == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket hub 未就绪"})
		return
	}

	msg := websocket.Message{
		Type: "admin_broadcast",
		Data: gin.H{
			"scope":    req.Scope,
			"msg_type": req.MsgType,
			"title":    req.Title,
			"content":  req.Content,
		},
	}

	switch req.Scope {
	case "global":
		hub.BroadcastToAll(msg)
	case "room":
		if req.Target == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "room scope 需要提供 target (roomID)"})
			return
		}
		hub.BroadcastToRoom(req.Target, msg)
	case "user":
		if req.Target == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user scope 需要提供 target (UID)"})
			return
		}
		uid, err := strconv.Atoi(req.Target)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user scope 的 target 必须为有效 UID"})
			return
		}
		hub.SendToUID(uid, msg)
	}

	c.JSON(http.StatusOK, gin.H{"message": "广播已发送", "scope": req.Scope})
}

// GetActiveRooms 获取当前所有活跃房间（含私密房间，管理员专用）
func GetActiveRooms(c *gin.Context) {
	rooms := game.GetAllRoomsAdmin()
	c.JSON(http.StatusOK, rooms)
}
