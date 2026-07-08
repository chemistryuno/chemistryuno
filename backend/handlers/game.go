package handlers

import (
	"chemistryuno/backend/cache"
	"chemistryuno/backend/game"
	"chemistryuno/backend/models"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/websocket"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 检查用户是否被封禁（适用于拦截创建房间、加入房间、观战、决斗等）
func checkBanForGame(c *gin.Context, uid int) bool {
	cachedUser, err := repository.GetUserWithCache(c.Request.Context(), uint(uid))
	if err == nil && cachedUser != nil && cachedUser.BannedUntil != nil && cachedUser.BannedUntil.After(time.Now()) {
		reason := cachedUser.BanReason
		if reason == "" {
			reason = "账号已被封禁"
		}
		c.JSON(http.StatusForbidden, gin.H{"error": reason + "，无法进行游戏相关操作"})
		return true
	}
	return false
}

func broadcastUpdate(roomID string) {
	if websocket.GlobalHub != nil {
		log.Printf("📡 广播游戏更新到房间 %s", roomID)
		websocket.GlobalHub.BroadcastToRoom(roomID, websocket.Message{
			Type: "game_update",
		})
	} else {
		log.Printf("⚠️  WebSocket Hub 不可用，无法广播到房间 %s", roomID)
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
	uid := c.GetInt("uid")
	rooms := game.GetAllRooms(uid)
	c.JSON(http.StatusOK, rooms)
}

// 创建房间
func CreateRoom(c *gin.Context) {
	var req struct {
		Name          string `json:"name"`
		MaxPlayers    int    `json:"max_players" binding:"required,min=2,max=8"`
		DeckID        int    `json:"deck_id"`
		IsPointsMode  bool   `json:"is_points_mode"`
		IsPrivate     bool   `json:"is_private"`
		AccessKey     string `json:"access_key"`
		IsPvE         bool   `json:"is_pve"`
		PvEDifficulty int    `json:"pve_difficulty"`
		AICount       int    `json:"ai_count"`

		// AI补位功能配置
		EnableAIBackfill     bool `json:"enable_ai_backfill"`     // 是否启用AI补位
		AIBackfillDifficulty int  `json:"ai_backfill_difficulty"` // 补位AI难度

		// 等级匹配系统
		IsRanked   bool `json:"is_ranked"`   // 是否为排位模式
		LevelRange int  `json:"level_range"` // 允许的等级范围（默认5）

		// 教学脚本系统
		TutorialScript bool `json:"tutorial_script"` // 是否启用脚本化教学
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid := c.GetInt("uid")

	if checkBanForGame(c, uid) {
		return
	}

	// 分布式锁：防止同一用户并发创建多个房间
	lockResource := fmt.Sprintf("room:create:%d", uid)
	lockToken, lockErr := cache.AcquireLock(c.Request.Context(), lockResource, 10*time.Second)
	if lockErr == cache.ErrLockNotAcquired {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "请勿重复提交，上一个请求正在处理中"})
		return
	}
	if lockErr == nil {
		defer func() { _ = cache.ReleaseLock(c.Request.Context(), lockResource, lockToken) }()
	}
	// If ErrRedisUnavailable, proceed without lock (graceful degradation)

	// 玩家限购检查：每人同时只能在一个房间
	if !game.IsPlayerIdle(uid) {
		currentRoomID := game.GetUserRoomID(uid)
		c.JSON(http.StatusConflict, gin.H{
			"error":   "您当前已在实验室中（只能同时进行一个实验）",
			"room_id": currentRoomID,
		})
		return
	}

	// PvE 模式下私密性校验特殊处理：PvE 房间通常默认为私密或不公开，但这里沿用前段传参
	// 如果是积分模式且是私密房间（非 PvE），则禁止
	if req.IsPointsMode && req.IsPrivate && !req.IsPvE {
		c.JSON(http.StatusBadRequest, gin.H{"error": "积分模式下不可创建私密房间"})
		return
	}

	// AI补位参数验证
	if req.EnableAIBackfill && !req.IsPvE {
		if req.AIBackfillDifficulty < 1 || req.AIBackfillDifficulty > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "AI难度必须在1-100之间"})
			return
		}
	}

	// PvE模式不允许同时启用补位功能（避免冲突）
	if req.IsPvE && req.EnableAIBackfill {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PvE模式不支持AI补位功能"})
		return
	}

	// 等级匹配验证
	if req.IsRanked {
		// 排位模式下，等级范围必须在3-10之间
		if req.LevelRange < 3 || req.LevelRange > 10 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "排位模式等级范围必须在3-10之间"})
			return
		}
	} else if req.LevelRange == 0 {
		// 如果未指定，默认等级范围为5
		req.LevelRange = 5
	}

	room, err := game.CreateRoomWithKey(req.Name, uid, req.MaxPlayers, req.DeckID, req.IsPointsMode, req.IsPrivate, req.AccessKey, req.IsPvE, req.PvEDifficulty, req.AICount, req.EnableAIBackfill, req.AIBackfillDifficulty, req.IsRanked, req.LevelRange, req.TutorialScript)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, room)
	// 广播房间列表更新
	go game.BroadcastRoomsUpdate()
}

// 加入房间
func JoinRoom(c *gin.Context) {
	roomID := c.Param("id")
	uid := c.GetInt("uid")

	if checkBanForGame(c, uid) {
		return
	}

	isAdmin := models.RoleHasAdminAccess(c.GetString("role"))

	// 从查询参数获取访问密钥和观战模式
	accessKey := c.Query("key")
	asSpectator := c.Query("spectator") == "true" || c.Query("spectator") == "1"

	// 如果是以玩家身份加入，检查是否已经在其他房间
	if !asSpectator {
		currentRoomID := game.GetUserRoomID(uid)
		if currentRoomID != "" && currentRoomID != roomID {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "您正在另一个实验室进行实验，请先完成或离开当前房间",
				"room_id": currentRoomID,
			})
			return
		}
	}

	err := game.JoinRoomWithKeyAsSpectator(roomID, uid, accessKey, asSpectator, isAdmin)
	if err != nil {
		if err.Error() == "房间不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	broadcastPlayerJoint(roomID)
	go game.BroadcastRoomsUpdate() // 广播房间列表更新
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
	go game.BroadcastRoomsUpdate() // 广播房间列表更新到主界面
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

	go game.BroadcastRoomsUpdate() // 广播房间列表更新，用于主界面显示倒计时和状态变化
	c.JSON(http.StatusOK, gin.H{"message": "准备状态已更新"})
}

// 获取当前玩家的游戏历史
func GetMyGameHistory(c *gin.Context) {
	uid := c.GetInt("uid")
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}

	historyList, err := repository.GameRepo.FindByUserUID(uint(uid))
	if err != nil {
		log.Printf("查询游戏历史失败 (uid=%d): %v", uid, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取游戏历史失败，请稍后重试"})
		return
	}

	// 收集所有涉及到的 WinnerUID
	winnerUIDs := make([]uint, 0)
	for _, h := range historyList {
		if h.WinnerUID != nil {
			winnerUIDs = append(winnerUIDs, *h.WinnerUID)
		}
	}

	// 获取胜者名称映射
	winnerNames := make(map[uint]string)
	if len(winnerUIDs) > 0 {
		users, _ := repository.UserRepo.FindByUIDs(winnerUIDs)
		for uid, user := range users {
			if user.Nickname != "" {
				winnerNames[uid] = user.Nickname
			} else {
				winnerNames[uid] = user.Username
			}
		}
	}

	var history []map[string]interface{}
	for _, h := range historyList {
		// 解析玩家列表 JSON
		var players []int
		if err := json.Unmarshal([]byte(h.Players), &players); err != nil {
			players = []int{}
		}

		winnerName := "AI"
		if h.IsInvalid {
			winnerName = "无效对局"
		} else if h.WinnerUID != nil && int(*h.WinnerUID) > 0 {
			if name, ok := winnerNames[*h.WinnerUID]; ok {
				winnerName = name
			} else {
				winnerName = "未知用户/AI"
			}
		}

		history = append(history, map[string]interface{}{
			"id":                    h.ID,
			"room_id":               h.RoomID,
			"winner_uid":            h.WinnerUID,
			"winner_name":           winnerName,
			"is_invalid":            h.IsInvalid,
			"invalid_reason":        h.InvalidReason,
			"has_replay":            h.ReplayLog != "",
			"replay_expires_at":     h.ReplayExpiresAt,
			"replay_cleared_at":     h.ReplayClearedAt,
			"players":               players,
			"original_player_count": h.OriginalPlayerCount,
			"quitted_count":         h.QuittedCount,
			"started_at":            h.StartedAt,
			"finished_at":           h.FinishedAt,
			"created_at":            h.CreatedAt,
		})
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
	log.Printf("✅ 游戏已启动: 房间 %s", roomID)
	go game.BroadcastRoomsUpdate() // 广播房间列表更新
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
		ThinkMs   int64       `json:"think_ms"` // 客户端上报的本回合真实思考耗时（毫秒），可选
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := game.PlayCard(roomID, uid, req.Card, req.Substance, req.ThinkMs)
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

// DiscardAndDraw 弃置 2 张手牌并从摸牌堆补 2 张
func DiscardAndDraw(c *gin.Context) {
	roomID := c.Param("id")
	uid := c.GetInt("uid")

	var req struct {
		Cards []string `json:"cards" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := game.DiscardAndDraw(roomID, uid, req.Cards); err != nil {
		if err.Error() == "房间不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	broadcastUpdate(roomID)
	c.JSON(http.StatusOK, gin.H{"message": "弃牌成功"})
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

	r1 := game.NormalizeSubscripts(req.R1)
	r2 := game.NormalizeSubscripts(req.R2)

	canReact := game.CanReact(r1, r2)
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

	if checkBanForGame(c, challengerUID) {
		return
	}

	challengerName := ""
	challengerNickname := "玩家"

	// 获取挑战者完整信息
	cUser, err := repository.UserRepo.FindByUID(uint(challengerUID))
	if err == nil && cUser != nil {
		challengerName = cUser.Username
		challengerNickname = cUser.Nickname
	}

	if challengerUID == req.TargetUID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能挑战自己"})
		return
	}

	// 分布式锁：防止重复发起单挑邀请（TTL 与邀请超时一致：20s）
	duelLockRes := fmt.Sprintf("duel:%d:%d", challengerUID, req.TargetUID)
	duelToken, duelLockErr := cache.AcquireLock(c.Request.Context(), duelLockRes, 20*time.Second)
	if duelLockErr == cache.ErrLockNotAcquired {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "邀请已发送，请等待对方响应"})
		return
	}
	if duelLockErr == nil {
		defer func() { _ = cache.ReleaseLock(c.Request.Context(), duelLockRes, duelToken) }()
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

	// 获取目标用户信息
	user, err := repository.UserRepo.FindByUID(uint(req.TargetUID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "目标用户不存在"})
		return
	}
	targetName := user.Username
	targetNickname := user.Nickname

	// 生成挑战 ID
	challengeID := fmt.Sprintf("challenge_%d_%d", time.Now().Unix(), rand.Intn(1000))

	// 存储挑战信息 (临时使用同步锁或管理类)
	// 这里我们直接在 game 包里处理逻辑，或者在这里手动维护
	// 为了简单，我们先在 InitiateDuel 发送一个 invite 消息

	msg := websocket.Message{
		Type: "duel_invite",
		Data: gin.H{
			"challenge_id":        challengeID,
			"challenger_uid":      challengerUID,
			"challenger_name":     challengerName,
			"challenger_nickname": challengerNickname,
			"target_name":         targetName,
			"target_nickname":     targetNickname,
			"timeout":             20,
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

	if checkBanForGame(c, responderUID) {
		return
	}

	responderName := ""
	responderNickname := "玩家" // 默认

	// 获取完整用户信息以获取昵称
	user, err := repository.UserRepo.FindByUID(uint(responderUID))
	if err == nil && user != nil {
		responderName = user.Username
		responderNickname = user.Nickname
	}

	challengerUID := req.TargetUID

	if !req.Accept {
		// 通知挑战者被拒绝
		websocket.GlobalHub.SendToUID(challengerUID, websocket.Message{
			Type: "duel_declined",
			Data: gin.H{
				"uid":      responderUID,
				"nickname": responderNickname,
			},
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

// GetReactionHints 获取基于场上物质的反应提示
func GetReactionHints(c *gin.Context) {
	roomID := c.Param("id")
	uid := c.GetInt("uid")

	hints, err := game.GetReactionHints(roomID, uid)
	if err != nil {
		if err.Error() == "房间不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	game.MarkHintUsed(roomID, uid)

	c.JSON(http.StatusOK, hints)
}

// CheckRoomStatus 检查房间状态（无需加入房间）
func CheckRoomStatus(c *gin.Context) {
	roomID := c.Param("id")

	exists, status := game.GetRoomStatus(roomID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"exists": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"exists": true,
		"status": status,
	})
}
