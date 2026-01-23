package handlers

import (
	"chemistryuno/game"
	"chemistryuno/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

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

	userID := c.GetInt("user_id")
	username := c.GetString("username")

	room, err := game.CreateRoom(req.Name, userID, username, req.MaxPlayers, req.DeckID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, room)
}

// 加入房间
func JoinRoom(c *gin.Context) {
	roomID := c.Param("id")
	userID := c.GetInt("user_id")
	username := c.GetString("username")

	err := game.JoinRoom(roomID, userID, username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "加入房间成功"})
}

// 离开房间
func LeaveRoom(c *gin.Context) {
	roomID := c.Param("id")
	userID := c.GetInt("user_id")

	err := game.LeaveRoom(roomID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "离开房间成功"})
}

// 开始游戏
func StartGame(c *gin.Context) {
	roomID := c.Param("id")
	userID := c.GetInt("user_id")

	err := game.StartGame(roomID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "游戏开始"})
}

// 出牌
func PlayCard(c *gin.Context) {
	roomID := c.Param("id")
	userID := c.GetInt("user_id")

	var req struct {
		Card      models.Card `json:"card" binding:"required"`
		Substance string      `json:"substance" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := game.PlayCard(roomID, userID, req.Card, req.Substance)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "出牌成功"})
}

// 摸牌
func DrawCard(c *gin.Context) {
	roomID := c.Param("id")
	userID := c.GetInt("user_id")

	err := game.DrawCard(roomID, userID, 2)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "摸牌成功"})
}

// 获取可用物质列表
func GetAvailableSubstances(c *gin.Context) {
	roomID := c.Param("id")
	userID := c.GetInt("user_id")

	substances, err := game.GetAvailableSubstances(roomID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, substances)
}
