package handlers

import (
	"chemistryuno/database"
	"chemistryuno/game"
	"chemistryuno/models"
	"chemistryuno/websocket"
	"net/http"

	"github.com/gin-gonic/gin"
)

func broadcastUpdate(roomID string) {
	if websocket.GlobalHub != nil {
		websocket.GlobalHub.BroadcastToRoom(roomID, websocket.Message{
			Type: "game_update",
		})
	}
}

func broadcastPlayerJoint(roomID string) {
	if websocket.GlobalHub != nil {
		websocket.GlobalHub.BroadcastToRoom(roomID, websocket.Message{
			Type: "player_joined",
		})
	}
}

// 获取房间列表
func GetRooms(c *gin.Context) {
	rooms := game.GetAllRooms()
	c.JSON(http.StatusOK, rooms)
}

// 创建房间
func CreateRoom(c *gin.Context) {
	var req struct {
		Name         string `json:"name"`
		MaxPlayers   int    `json:"max_players" binding:"required,min=2,max=8"`
		DeckID       int    `json:"deck_id"`
		IsPointsMode bool   `json:"is_points_mode"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid := c.GetInt("uid")
	username := c.GetString("username")

	room, err := game.CreateRoom(req.Name, uid, username, req.MaxPlayers, req.DeckID, req.IsPointsMode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, room)
}

// 加入房间
func JoinRoom(c *gin.Context) {
	roomID := c.Param("id")
	uid := c.GetInt("uid")
	username := c.GetString("username")

	err := game.JoinRoom(roomID, uid, username)
	if err != nil {
		if err.Error() == "房间不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	broadcastPlayerJoint(roomID)
	c.JSON(http.StatusOK, gin.H{"message": "加入房间成功"})
}

// 离开房间
func LeaveRoom(c *gin.Context) {
	roomID := c.Param("id")
	uid := c.GetInt("uid")

	err := game.LeaveRoom(roomID, uid)
	if err != nil {
		if err.Error() == "房间不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	broadcastUpdate(roomID)
	c.JSON(http.StatusOK, gin.H{"message": "离开房间成功"})
}

// 开始游戏
func StartGame(c *gin.Context) {
	roomID := c.Param("id")
	uid := c.GetInt("uid")

	err := game.StartGame(roomID, uid)
	if err != nil {
		if err.Error() == "房间不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	broadcastUpdate(roomID)
	c.JSON(http.StatusOK, gin.H{"message": "游戏开始"})
}

// 获取房间状态
func GetRoomState(c *gin.Context) {
	roomID := c.Param("id")
	uid := c.GetInt("uid")

	state, err := game.GetRoomState(roomID, uid)
	if err != nil {
		if err.Error() == "房间不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == "你不在该房间中" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, state)
}

// 出牌
func PlayCard(c *gin.Context) {
	roomID := c.Param("id")
	uid := c.GetInt("uid")

	var req struct {
		Card      models.Card `json:"card" binding:"required"`
		Substance string      `json:"substance" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := game.PlayCard(roomID, uid, req.Card, req.Substance)
	if err != nil {
		if err.Error() == "房间不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	broadcastUpdate(roomID)
	c.JSON(http.StatusOK, gin.H{"message": "出牌成功"})
}

// 摸牌
func DrawCard(c *gin.Context) {
	roomID := c.Param("id")
	uid := c.GetInt("uid")

	err := game.DrawCard(roomID, uid, 2)
	if err != nil {
		if err.Error() == "房间不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	broadcastUpdate(roomID)
	c.JSON(http.StatusOK, gin.H{"message": "摸牌成功"})
}

// 获取可用物质列表
func GetAvailableSubstances(c *gin.Context) {
	roomID := c.Param("id")
	uid := c.GetInt("uid")

	substances, err := game.GetAvailableSubstances(roomID, uid)
	if err != nil {
		if err.Error() == "房间不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, substances)
}

// 查验反应是否成立
func VerifyReaction(c *gin.Context) {
	var req struct {
		R1 string `json:"r1" binding:"required"`
		R2 string `json:"r2" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	canReact := game.CanReact(req.R1, req.R2)
	c.JSON(http.StatusOK, gin.H{
		"can_react": canReact,
	})
}

// 发动双联反应
func DoublePlay(c *gin.Context) {
	roomID := c.Param("id")
	uid := c.GetInt("uid")

	var req struct {
		Sub1 string `json:"sub1" binding:"required"`
		Sub2 string `json:"sub2" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := game.DoublePlay(roomID, uid, req.Sub1, req.Sub2)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	broadcastUpdate(roomID)
	c.JSON(http.StatusOK, gin.H{"message": "双联反应发动成功！"})
}

// InitiateDuel 发起单挑
func InitiateDuel(c *gin.Context) {
	var req struct {
		TargetUID int `json:"target_uid" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	challengerUID := c.GetInt("uid")
	challengerName := c.GetString("username")

	if challengerUID == req.TargetUID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能挑战自己"})
		return
	}

	// 检查目标是否存在且在线
	if websocket.GlobalHub == nil || !websocket.GlobalHub.IsUIDOnline(req.TargetUID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "目标用户不在线"})
		return
	}

	// 获取目标用户名
	var targetName string
	err := database.DB.QueryRow("SELECT username FROM users WHERE UID = ?", req.TargetUID).Scan(&targetName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "目标用户不存在"})
		return
	}

	// 创建单挑房间
	room, err := game.StartDuel(challengerUID, challengerName, req.TargetUID, targetName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 通过 WebSocket 通知双方玩家进入房间
	msg := websocket.Message{
		Type:   "duel_start",
		RoomID: room.ID,
	}

	// 发送给发起者
	websocket.GlobalHub.SendToUID(challengerUID, msg)
	// 发送给被挑战者
	websocket.GlobalHub.SendToUID(req.TargetUID, msg)

	c.JSON(http.StatusOK, room)
}
