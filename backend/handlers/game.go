package handlers

import (
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
		Name       string `json:"name" binding:"required"`
		MaxPlayers int    `json:"max_players" binding:"required,min=2,max=8"`
		DeckID     int    `json:"deck_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid := c.GetInt("uid")
	username := c.GetString("username")

	room, err := game.CreateRoom(req.Name, uid, username, req.MaxPlayers, req.DeckID)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
