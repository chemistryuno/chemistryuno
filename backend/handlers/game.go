package handlers

import (
	"chemistryuno/game"
	"chemistryuno/models"
	"chemistryuno/repository"
	"chemistryuno/websocket"
	"fmt"
	"math/rand"
	"net/http"
	"time"

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
		IsPrivate    bool   `json:"is_private"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid := c.GetInt("uid")
	username := c.GetString("username")

	room, err := game.CreateRoom(req.Name, uid, username, req.MaxPlayers, req.DeckID, req.IsPointsMode, req.IsPrivate)
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

// ToggleReady 准备或取消准备
func ToggleReady(c *gin.Context) {
	roomID := c.Param("id")
	uid := c.GetInt("uid")

	err := game.ToggleReady(roomID, uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "开始游戏"})
}

// 获取当前玩家的游戏历史
func GetMyGameHistory(c *gin.Context) {
	uid := c.GetInt("uid")
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}

	history, err := repository.GameRepo.FindByUserUID(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
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

	// 1. 检查目标是否空闲
	if !game.IsPlayerIdle(req.TargetUID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "目标用户正在游戏中，请稍后再试"})
		return
	}

	// 2. 检查目标是否存在且在线
	if websocket.GlobalHub == nil || !websocket.GlobalHub.IsUIDOnline(req.TargetUID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "目标用户不在线"})
		return
	}

	// 获取目标用户名
	user, err := repository.UserRepo.FindByUID(uint(req.TargetUID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "目标用户不存在"})
		return
	}
	targetName := user.Username

	// 生成挑战 ID
	challengeID := fmt.Sprintf("challenge_%d_%d", time.Now().Unix(), rand.Intn(1000))

	// 存储挑战信息 (临时使用同步锁或管理类)
	// 这里我们直接在 game 包里处理逻辑，或者在这里手动维护
	// 为了简单，我们先在 InitiateDuel 发送一个 invite 消息

	msg := websocket.Message{
		Type: "duel_invite",
		Data: gin.H{
			"challenge_id":    challengeID,
			"challenger_uid":  challengerUID,
			"challenger_name": challengerName,
			"target_name":     targetName,
			"timeout":         20,
		},
	}

	// 发送给被挑战者
	websocket.GlobalHub.SendToUID(req.TargetUID, msg)

	c.JSON(http.StatusOK, gin.H{
		"message":      "挑战邀请已发送",
		"challenge_id": challengeID,
	})
}

// RespondToDuel 响应单挑邀请
func RespondToDuel(c *gin.Context) {
	var req struct {
		TargetUID int  `json:"target_uid" binding:"required"` // 发起挑战的人 (这里的 Target 是指请求的目标，即 Challenger)
		Accept    bool `json:"accept"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	responderUID := c.GetInt("uid")
	responderName := c.GetString("username")
	challengerUID := req.TargetUID

	if !req.Accept {
		// 通知挑战者被拒绝
		websocket.GlobalHub.SendToUID(challengerUID, websocket.Message{
			Type: "duel_declined",
			Data: gin.H{"username": responderName},
		})
		c.JSON(http.StatusOK, gin.H{"message": "已拒绝挑战"})
		return
	}

	// 检查双方是否仍然空闲且在线
	if !game.IsPlayerIdle(responderUID) || !game.IsPlayerIdle(challengerUID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "某方已进入游戏"})
		return
	}

	if !websocket.GlobalHub.IsUIDOnline(challengerUID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "挑战者已离线"})
		return
	}

	// 获取挑战者名称
	var challengerName string
	if user, err := repository.UserRepo.FindByUID(uint(challengerUID)); err == nil {
		challengerName = user.Username
	}

	// 创建单挑房间
	room, err := game.StartDuel(challengerUID, challengerName, responderUID, responderName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 通知双方进入
	msg := websocket.Message{
		Type:   "duel_start",
		RoomID: room.ID,
	}
	websocket.GlobalHub.SendToUID(challengerUID, msg)
	websocket.GlobalHub.SendToUID(responderUID, msg)

	c.JSON(http.StatusOK, gin.H{"message": "挑战已接受", "room_id": room.ID})
}
