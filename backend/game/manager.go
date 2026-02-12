package game

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/models"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/websocket"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"
)

var (
	rooms      = make(map[string]*GameRoom)
	roomMutex  sync.RWMutex
	configRepo *repository.ConfigRepository
)

type GameRoom struct {
	Room       *models.Room
	GameState  *models.GameState
	mutex      sync.RWMutex
	OfflineAt  map[int]time.Time // UID -> 离线起始时间
	StartTimer *time.Timer       // 游戏开始倒计时器
}

func (gr *GameRoom) cancelStartTimer() {
	if gr.StartTimer != nil {
		gr.StartTimer.Stop()
		gr.StartTimer = nil
	}
	gr.Room.Countdown = 0
}

func (gr *GameRoom) checkAutoStart() {
	roomID := gr.Room.ID
	numPlayers := len(gr.Room.Players)
	maxPlayers := gr.Room.MaxPlayers

	// 统计准备的玩家（只要还在房间内就计数，不检查实时在线状态以防刷新导致的频繁重置）
	numReady := len(gr.Room.ReadyUIDs)

	// 确定目标倒计时（必须至少有2名玩家才能开始倒计时）
	targetCountdown := 0
	if numPlayers >= 2 {
		if numPlayers == maxPlayers && numReady == maxPlayers {
			// 满员且全部准备 -> 快速开始（从配置读取）
			targetCountdown = getAutoStartTimeout()
		} else if numReady >= (maxPlayers+1)/2 {
			// 至少2人且准备人数过半 -> 倒计时（从配置读取）
			targetCountdown = getHalfReadyTimeout()
		}
	}

	log.Printf("[自动开始检查] 房间 %s: 玩家数=%d/%d, 准备数=%d, 倒计时=%d秒",
		roomID, numPlayers, maxPlayers, numReady, targetCountdown)

	// 如果不再满足任何倒计时条件
	if targetCountdown == 0 {
		if gr.StartTimer != nil {
			gr.cancelStartTimer()
			gr.broadcastRoomUpdate()
		}
		return
	}

	// 如果没有定时器，或者当前倒计时比目标倒计时长（例如从60s变10s），则更新
	if gr.StartTimer == nil || (targetCountdown < gr.Room.Countdown) {
		if gr.StartTimer != nil {
			gr.StartTimer.Stop()
		}

		gr.Room.Countdown = targetCountdown
		gr.broadcastRoomUpdate()

		gr.StartTimer = time.AfterFunc(time.Duration(targetCountdown)*time.Second, func() {
			gr.mutex.Lock()
			// 二次检查，防止竞争
			if gr.Room.Status != "waiting" || gr.StartTimer == nil {
				gr.mutex.Unlock()
				return
			}

			// 踢出未准备的玩家
			readyMap := make(map[int]bool)
			for _, uid := range gr.Room.ReadyUIDs {
				readyMap[uid] = true
			}

			var playersToKeep []int
			var playersToKick []int
			for _, uid := range gr.Room.Players {
				if readyMap[uid] {
					playersToKeep = append(playersToKeep, uid)
				} else {
					playersToKick = append(playersToKick, uid)
				}
			}

			gr.Room.Players = playersToKeep
			gr.Room.ReadyUIDs = []int{} // 清空准备状态
			gr.Room.Countdown = 0
			gr.StartTimer = nil

			// 如果剩下的人还够，就开始游戏
			if len(gr.Room.Players) >= 2 {
				log.Printf("[自动开始] 房间 %s 倒计时结束，准备开始游戏，玩家数：%d", roomID, len(gr.Room.Players))
				gr.mutex.Unlock()

				// 执行踢出
				for _, uid := range playersToKick {
					log.Printf("[自动开始] 踢出未准备玩家：%d", uid)
					gr.kickPlayer(uid, "由于未准备，您已被移出游戏")
				}

				log.Printf("[自动开始] 调用StartGame for room %s", roomID)
				err := StartGame(roomID, 0)
				if err != nil {
					log.Printf("[自动开始] StartGame失败：%v", err)
				} else {
					log.Printf("[自动开始] StartGame成功，广播更新")
					gr.broadcastRoomUpdate()
				}
			} else {
				gr.mutex.Unlock()
				for _, uid := range playersToKick {
					gr.kickPlayer(uid, "由于未准备，您已被移出游戏")
				}
				gr.broadcastRoomUpdate()
			}
		})

		// 倒计时显示逻辑（每秒减少）
		go func(roomID string, startVal int) {
			for i := startVal - 1; i > 0; i-- {
				time.Sleep(1 * time.Second)
				roomMutex.RLock()
				g, exists := rooms[roomID]
				roomMutex.RUnlock()
				if !exists {
					return
				}
				g.mutex.Lock()
				// 如果定时器被取消，或者被重置为不同的值（说明有新的倒计时开始了），则退出当前协程
				if g.StartTimer == nil || g.Room.Status != "waiting" || g.Room.Countdown != i+1 {
					g.mutex.Unlock()
					return
				}
				g.Room.Countdown = i
				g.broadcastRoomUpdate()
				g.mutex.Unlock()
			}
		}(roomID, targetCountdown)
	}
}

func (gr *GameRoom) broadcastRoomUpdate() {
	if websocket.GlobalHub != nil {
		websocket.GlobalHub.BroadcastToRoom(gr.Room.ID, websocket.Message{
			Type: "game_update",
			Data: gr.Room.ID, // 前端通常收到房间ID后会调用 GetRoomState
		})
	}
}

// 记录当前玩家回合开始时间到数据库
func (gr *GameRoom) recordTurnStart() {
	if gr.GameState != nil && len(gr.GameState.Players) > 0 {
		uid := gr.GameState.Players[gr.GameState.CurrentPlayer].UID
		repository.UserRepo.UpdateTurnStartedAt(uint(uid), time.Now())
	}
}

func isBanned(uid int) (bool, time.Time, string, error) {
	userRepo := repository.NewUserRepository()
	bannedUntil, _, reason, err := userRepo.CheckBanStatus(uint(uid))
	if err != nil {
		return false, time.Time{}, "", err
	}
	if bannedUntil != nil && time.Now().Before(*bannedUntil) {
		return true, *bannedUntil, reason, nil
	}
	return false, time.Time{}, "", nil
}

// IsPlayerIdle 检查玩家是否由于已在游戏中而忙碌
func IsPlayerIdle(uid int) bool {
	roomMutex.RLock()
	defer roomMutex.RUnlock()

	for _, gr := range rooms {
		if gr.Room.Status == "playing" {
			for _, puid := range gr.Room.Players {
				if puid == uid {
					return false
				}
			}
		}
	}
	return true
}

// 初始化默认牌组配置
func getDefaultDeckConfig() map[string]int {
	return map[string]int{
		"H": 12, "O": 12,
		"C": 4, "N": 4, "F": 4, "Na": 4, "Mg": 4, "Al": 4,
		"Si": 4, "P": 4, "S": 4, "Cl": 4, "K": 4, "Ca": 4,
		"Mn": 4, "Fe": 4, "Cu": 4, "Zn": 4, "Br": 4, "I": 4, "Ag": 4,
		"+2": 8, "+4": 4,
		"He": 1, "Ne": 1, "Ar": 1, "Kr": 1,
		"Au": 4,
		// "Choice": 4,  // 已移除
	}
}

// 获取当前全局牌组配置
func getGlobalDeckConfigFromDB() (map[string]int, string, int) {
	deckRepo := repository.NewDeckRepository()
	deck, err := deckRepo.FindGlobalDeck()
	if err != nil {
		return getDefaultDeckConfig(), "默认牌组", 10
	}

	var cards map[string]int
	if err := json.Unmarshal([]byte(deck.Cards), &cards); err != nil {
		return getDefaultDeckConfig(), "默认牌组", 10
	}
	return cards, deck.Name, deck.InitialCards
}

// 创建房间
func CreateRoom(name string, creatorUID int, creatorName string, maxPlayers int, deckID int, isPointsMode bool, isPrivate bool) (*models.Room, error) {
	return CreateRoomWithKey(name, creatorUID, creatorName, maxPlayers, deckID, isPointsMode, isPrivate, "")
}

// 创建房间（支持自定义访问密钥）
func CreateRoomWithKey(name string, creatorUID int, creatorName string, maxPlayers int, deckID int, isPointsMode bool, isPrivate bool, customKey string) (*models.Room, error) {
	banned, until, reason, _ := isBanned(creatorUID)
	if banned {
		if reason == "" {
			reason = "您的账号由于多次消极游戏已被封禁"
		}
		return nil, fmt.Errorf("%s，直到 %s", reason, until.Format("2006-01-02 15:04:05"))
	}

	if name == "" {
		const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		b := make([]byte, 6)
		for i := range b {
			b[i] = charset[rand.Intn(len(charset))]
		}
		name = "LAB-" + string(b)
	}

	// 生成简单的房间ID
	roomID := fmt.Sprintf("room_%d_%d", time.Now().Unix(), rand.Intn(1000))

	// 加载牌组配置
	var deckConfig models.DeckConfig
	// 积分模式强制使用默认牌组
	if isPointsMode || deckID <= 1 { // deckID <= 1 意味着使用全局默认牌组 (ID=1)
		cards, dname, initialCards := getGlobalDeckConfigFromDB()
		deckConfig.Cards = cards
		deckConfig.Name = dname
		deckConfig.InitialCards = initialCards
		deckConfig.IsGlobal = true
		deckConfig.ID = 1
	} else {
		deck, err := repository.DeckRepo.FindByID(uint(deckID))
		if err != nil {
			cards, dname, initialCards := getGlobalDeckConfigFromDB()
			deckConfig.Cards = cards
			deckConfig.Name = dname
			deckConfig.InitialCards = initialCards
			deckConfig.IsGlobal = true
			deckConfig.ID = 1
		} else {
			deckConfig.ID = int(deck.ID)
			deckConfig.Name = deck.Name
			deckConfig.InitialCards = deck.InitialCards
			// 解析JSON字符串到map
			var cards map[string]int
			if err := json.Unmarshal([]byte(deck.Cards), &cards); err != nil {
				// 解析失败，使用默认牌组
				cards, dname, initialCards := getGlobalDeckConfigFromDB()
				deckConfig.Cards = cards
				deckConfig.Name = dname
				deckConfig.InitialCards = initialCards
				deckConfig.IsGlobal = true
				deckConfig.ID = 1
			} else {
				deckConfig.Cards = cards
			}
		}
	}

	// 生成私密房间访问密钥
	var accessKey string
	if isPrivate {
		if customKey != "" {
			// 使用自定义密钥（需验证格式）
			if len(customKey) < 4 || len(customKey) > 20 {
				return nil, fmt.Errorf("访问密钥长度必须在4-20个字符之间")
			}
			accessKey = customKey
		} else {
			// 自动生成8位密钥
			const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
			b := make([]byte, 8)
			for i := range b {
				b[i] = charset[rand.Intn(len(charset))]
			}
			accessKey = string(b)
		}
	}

	room := &models.Room{
		ID:           roomID,
		Name:         name,
		Players:      []int{creatorUID},
		ReadyUIDs:    []int{},
		Countdown:    0,
		Spectators:   []int{},
		MaxPlayers:   maxPlayers,
		DeckConfig:   &deckConfig,
		Status:       "waiting",
		IsPointsMode: isPointsMode,
		IsPrivate:    isPrivate,
		AccessKey:    accessKey,
		CreatedAt:    time.Now(),
	}

	gameRoom := &GameRoom{
		Room:      room,
		OfflineAt: make(map[int]time.Time),
	}

	roomMutex.Lock()
	rooms[roomID] = gameRoom
	roomMutex.Unlock()

	return room, nil
}

// StartDuel 创建单挑房间
func StartDuel(challengerUID int, challengerName string, targetUID int, targetName string) (*models.Room, error) {
	// 默认配置
	cards, name, initialCards := getGlobalDeckConfigFromDB()
	deckConfig := models.DeckConfig{
		Cards:        cards,
		Name:         name,
		InitialCards: initialCards,
		IsGlobal:     true,
		ID:           1,
	}

	roomID := fmt.Sprintf("duel_%d_%d", time.Now().Unix(), rand.Intn(1000))
	room := &models.Room{
		ID:            roomID,
		Name:          fmt.Sprintf("Duel: %s VS %s", challengerName, targetName),
		Players:       []int{challengerUID, targetUID},
		ReadyUIDs:     []int{},
		Countdown:     0,
		Spectators:    []int{},
		MaxPlayers:    2,
		DeckConfig:    &deckConfig,
		Status:        "waiting",
		IsPointsMode:  true, // 单挑默认积分模式
		IsDuel:        true,
		ChallengerUID: challengerUID,
		TargetUID:     targetUID,
		CreatedAt:     time.Now(),
	}

	gameRoom := &GameRoom{
		Room:      room,
		OfflineAt: make(map[int]time.Time),
	}

	roomMutex.Lock()
	rooms[roomID] = gameRoom
	roomMutex.Unlock()

	return room, nil
}

// 获取所有房间（对于已加入的私密房间，也会返回）
func GetAllRooms(uid int) []*models.Room {
	roomMutex.RLock()
	defer roomMutex.RUnlock()

	result := []*models.Room{}
	for _, gr := range rooms {
		// 消除已结束的房间
		if gr.Room.Status == "finished" {
			continue
		}

		// 非私密房间：直接显示
		if !gr.Room.IsPrivate {
			result = append(result, gr.Room)
			continue
		}

		// 私密房间：仅当用户在房间中时才显示（支持快速重连）
		userInRoom := false
		for _, pid := range gr.Room.Players {
			if pid == uid {
				userInRoom = true
				break
			}
		}
		if userInRoom {
			result = append(result, gr.Room)
		}
	}

	// 排序逻辑：waiting 优先，然后按创建时间从新到旧
	sort.Slice(result, func(i, j int) bool {
		if result[i].Status != result[j].Status {
			if result[i].Status == "waiting" {
				return true
			}
			if result[j].Status == "waiting" {
				return false
			}
		}
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	return result
}

// GetRoomStatus 检查房间是否存在及其状态
func GetRoomStatus(roomID string) (exists bool, status string) {
	roomMutex.RLock()
	defer roomMutex.RUnlock()

	gr, exists := rooms[roomID]
	if !exists {
		return false, ""
	}

	return true, gr.Room.Status
}

// 积分结算逻辑
func handlePointsCalculation(gr *GameRoom) {
	finished := gr.GameState.FinishedPlayers
	count := len(finished)
	if count < 2 {
		return
	}

	// 计算积分倍率：每有一个未完成玩家离开，结算减少 1/总人数
	multiplier := 1.0
	if gr.GameState.OriginalPlayerCount > 0 {
		multiplier = 1.0 - (float64(gr.GameState.QuittedCount) / float64(gr.GameState.OriginalPlayerCount))
		if multiplier < 0 {
			multiplier = 0
		}
	}

	changes := make(map[int]int)

	for i, uid := range finished {
		points := 0
		if i == count-1 && count > 1 {
			// 最后一名得5分
			points = 5
		} else {
			// 100 / 名次
			rank := i + 1
			points = 100 / rank
		}

		// 应用倍率
		points = int(float64(points) * multiplier)

		changes[uid] = points
		repository.UserRepo.IncrementPoints(uint(uid), points)
		repository.UserRepo.IncrementMonthlyPoints(uint(uid), points)
	}
	gr.GameState.PointsChanges = changes

	// 悬赏逻辑处理
	winnerUID := finished[0]
	playerUIDs := []int{}
	for _, p := range gr.GameState.Players {
		playerUIDs = append(playerUIDs, p.UID)
	}

	// 查找针对这些玩家的悬赏
	totalBountyForWinner := 0
	for _, targetUID := range playerUIDs {
		bounties, err := repository.BountyRepo.FindActiveByTarget(uint(targetUID))
		if err != nil {
			continue
		}
		for _, bounty := range bounties {
			if gr.Room.IsDuel && targetUID == gr.Room.TargetUID {
				// 单挑模式特别处理
				if winnerUID == gr.Room.ChallengerUID {
					// 发起者赢：获得全部悬赏
					totalBountyForWinner += bounty.Amount
					repository.BountyRepo.UpdateStatus(bounty.ID, "claimed")
				} else if winnerUID == gr.Room.TargetUID {
					// 被挑战者赢：获得一半悬赏
					reward := bounty.Amount / 2
					totalBountyForWinner += reward
					repository.BountyRepo.UpdateStatus(bounty.ID, "claimed")
				}
			} else {
				// 普通模式：只要被悬赏者输了（不是第一名），胜者就能获得该悬赏
				if targetUID != winnerUID {
					totalBountyForWinner += bounty.Amount
					repository.BountyRepo.UpdateStatus(bounty.ID, "claimed")
				}
			}
		}
	}

	if totalBountyForWinner > 0 {
		repository.UserRepo.IncrementPoints(uint(winnerUID), totalBountyForWinner)
		repository.UserRepo.IncrementMonthlyPoints(uint(winnerUID), totalBountyForWinner)
		changes[winnerUID] += totalBountyForWinner
	}
}

// StartRoomMonitor 启动房间监控协程
func StartRoomMonitor() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			checkAllRooms()
		}
	}()

	// 启动定期维护协程（每周衰减、每月重置）
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			performPeriodicMaintenance()
		}
	}()
}

func performPeriodicMaintenance() {
	// 1. 每月重置月榜积分 (ResetMonthlyPointsIfNeeded 内部会判断是否是新月份)
	repository.UserRepo.ResetMonthlyPointsIfNeeded()

	// 2. 每周衰减前10%玩家积分2%
	repository.UserRepo.DecayTopPlayersPoints(10) // 10% 的玩家

	// 3. 清理过期的好友请求 (7天过期)
	repository.FriendshipRepo.CleanupExpiredRequests()
}

func checkAllRooms() {
	roomMutex.RLock()
	roomList := make([]*GameRoom, 0, len(rooms))
	for _, gr := range rooms {
		roomList = append(roomList, gr)
	}
	roomMutex.RUnlock()

	for _, gr := range roomList {
		gr.checkInactivity()
	}
}

func (gr *GameRoom) checkInactivity() {
	gr.mutex.Lock()
	roomID := gr.Room.ID
	now := time.Now()
	playersToKick := []int{}

	// 1. 匹配超时检测 (5分钟，针对空闲长久的房间)
	if gr.Room.Status == "waiting" {
		if time.Since(gr.Room.CreatedAt) > 5*time.Minute && len(gr.Room.Players) == 0 {
			gr.mutex.Unlock()
			gr.terminateRoom("匹配超时，房间已自动关闭")
			return
		}
	}

	// 2. 检测离线超过30秒的玩家
	for _, uid := range gr.Room.Players {
		isOnline := false
		if websocket.GlobalHub != nil {
			isOnline = websocket.GlobalHub.IsUIDInRoom(roomID, uid)
		}

		if !isOnline {
			// 如果玩家刚被检测到离线，先记录离线时间，暂不踢出
			offlineTime, exists := gr.OfflineAt[uid]
			if !exists {
				gr.OfflineAt[uid] = now
				repository.UserRepo.UpdateLastOfflineAt(uint(uid), now)
				continue // 下一次检查再来判断
			}

			// 计算 SQL 中的到期时间进行判断
			turnStart, lastOffline, err := repository.UserRepo.GetUserReconnectionData(uint(uid))
			if err == nil {
				// 判定离线超时：离线时间、回合开始时间、DB中的最后离线时间，三者中最新的那个作为起点
				expiryBase := offlineTime
				if turnStart != nil && turnStart.After(expiryBase) {
					expiryBase = *turnStart
				}
				if lastOffline != nil && lastOffline.After(expiryBase) {
					expiryBase = *lastOffline
				}

				kickTimeout := getPlayerKickTimeout()
				if now.Sub(expiryBase) > kickTimeout {
					playersToKick = append(playersToKick, uid)
				}
			} else {
				// 回退逻辑
				kickTimeout := getPlayerKickTimeout()
				if now.Sub(offlineTime) > kickTimeout {
					playersToKick = append(playersToKick, uid)
				}
			}
		} else {
			delete(gr.OfflineAt, uid)
		}
	}
	gr.mutex.Unlock()

	// 3. 执行踢出操作
	for _, uid := range playersToKick {
		reason := "由于断开连接超时，您已被移出房间"
		if gr.Room.Status == "playing" {
			reason = "由于消极游戏，您已被踢出"
		}
		gr.kickPlayer(uid, reason)
	}

	// 4. 检测后续状态
	gr.mutex.Lock()
	if gr.Room.Status == "waiting" {
		gr.checkAutoStart()
	} else if gr.Room.Status == "playing" && len(gr.Room.Players) < 2 {
		gr.mutex.Unlock()
		gr.terminateRoom("由于玩家人数不足，房间已被关闭")
		return
	}
	gr.mutex.Unlock()
}

func (gr *GameRoom) kickPlayer(uid int, reason string) {
	gr.mutex.Lock()
	roomID := gr.Room.ID

	// 通知被踢出的玩家
	if websocket.GlobalHub != nil {
		websocket.GlobalHub.SendToUID(uid, websocket.Message{
			Type:    "player_kicked",
			Message: reason,
		})
	}

	// 记录消极游戏行为并处理封禁（仅在游戏开始后计入）
	if reason == "由于消极游戏，您已被踢出" && gr.Room.Status == "playing" {
		count, _ := repository.UserRepo.GetNegativePlayCount(uint(uid))
		count++
		if count >= 3 {
			bannedUntil := time.Now().Add(30 * time.Minute)
			repository.UserRepo.UpdateBanStatus(uint(uid), &bannedUntil)
			repository.UserRepo.UpdateNegativePlayCount(uint(uid), 0)
			if websocket.GlobalHub != nil {
				websocket.GlobalHub.SendToUID(uid, websocket.Message{
					Type:    "player_banned",
					Message: "由于多次消极游戏，您的账号已被封禁 30 分钟。请健康游戏。",
				})
			}
		} else {
			repository.UserRepo.UpdateNegativePlayCount(uint(uid), count)
		}
	}

	// 如果是竞技模式，对被踢出的玩家进行积分惩罚
	if gr.Room.IsPointsMode && gr.Room.Status == "playing" {
		// 被踢出者扣除 30 积分作为惩罚
		repository.UserRepo.DeductPoints(uint(uid), 30)
	}

	// 移除玩家
	newPlayers := []int{}
	for _, pid := range gr.Room.Players {
		if pid != uid {
			newPlayers = append(newPlayers, pid)
		}
	}
	gr.Room.Players = newPlayers

	// 如果所有玩家都离开了，关闭房间
	if len(gr.Room.Players) == 0 {
		gr.mutex.Unlock()
		gr.terminateRoom("由于没有研究员留守，实验室已自动关闭")
		return
	}

	// 如果游戏正在进行，也从 GameState 中移除
	if gr.GameState != nil {
		newPS := []*models.PlayerState{}
		kickedIndex := -1
		isFinished := false
		for _, fuid := range gr.GameState.FinishedPlayers {
			if fuid == uid {
				isFinished = true
				break
			}
		}

		for i, ps := range gr.GameState.Players {
			if ps.UID != uid {
				newPS = append(newPS, ps)
			} else {
				kickedIndex = i
			}
		}

		if !isFinished && kickedIndex != -1 {
			gr.GameState.QuittedCount++
		}

		gr.GameState.Players = newPS

		// 调整当前玩家索引
		if kickedIndex != -1 {
			if gr.GameState.CurrentPlayer > kickedIndex {
				gr.GameState.CurrentPlayer--
			}
			if gr.GameState.CurrentPlayer >= len(gr.GameState.Players) {
				gr.GameState.CurrentPlayer = 0
			}
		}
	}

	// 如果是正在游戏中且玩家少于2个，则解散
	if gr.Room.Status == "playing" && len(gr.Room.Players) < 2 {
		gr.mutex.Unlock()
		gr.terminateRoom("由于实验样本不足（少于2人），本次反应宣告失败，实验室已关闭")
		return
	}

	gr.mutex.Unlock()

	// 广播玩家离开消息
	if websocket.GlobalHub != nil {
		websocket.GlobalHub.BroadcastToRoom(roomID, websocket.Message{
			Type: "player_left",
			UID:  uid,
			Data: fmt.Sprintf("研究员 %d 已离开实验室", uid),
		})
	}
}

func (gr *GameRoom) terminateRoom(reason string) {
	roomID := gr.Room.ID
	roomMutex.Lock()
	delete(rooms, roomID)
	roomMutex.Unlock()

	if websocket.GlobalHub != nil {
		websocket.GlobalHub.BroadcastToRoom(roomID, websocket.Message{
			Type:    "room_terminated",
			Message: reason,
		})
	}
}

// ToggleReady 切换玩家准备状态
func ToggleReady(roomID string, uid int) error {
	roomMutex.RLock()
	gr, exists := rooms[roomID]
	roomMutex.RUnlock()

	if !exists {
		return errors.New("房间不存在")
	}

	gr.mutex.Lock()
	defer gr.mutex.Unlock()

	if gr.Room.Status != "waiting" {
		return errors.New("游戏已开始，无法更改准备状态")
	}

	// 检查玩家是否在房间中
	isInRoom := false
	for _, pid := range gr.Room.Players {
		if pid == uid {
			isInRoom = true
			break
		}
	}
	if !isInRoom {
		return errors.New("您不在该房间中")
	}

	foundIdx := -1
	for i, ruid := range gr.Room.ReadyUIDs {
		if ruid == uid {
			foundIdx = i
			break
		}
	}

	if foundIdx >= 0 {
		// 取消准备
		gr.Room.ReadyUIDs = append(gr.Room.ReadyUIDs[:foundIdx], gr.Room.ReadyUIDs[foundIdx+1:]...)
		repository.UserRepo.UpdateRoomReadyStatus(uint(uid), false)
	} else {
		// 准备
		gr.Room.ReadyUIDs = append(gr.Room.ReadyUIDs, uid)
		repository.UserRepo.UpdateRoomReadyStatus(uint(uid), true)
	}

	gr.checkAutoStart()
	gr.broadcastRoomUpdate()
	return nil
}

// 加入房间
func JoinRoom(roomID string, uid int, username string) error {
	return JoinRoomWithKey(roomID, uid, username, "")
}

// JoinRoomWithKey 携带密钥加入房间
func JoinRoomWithKey(roomID string, uid int, username string, accessKey string) error {
	banned, until, reason, _ := isBanned(uid)
	if banned {
		if reason == "" {
			reason = "您的账号由于多次消极游戏已被封禁"
		}
		return fmt.Errorf("%s，直到 %s", reason, until.Format("2006-01-02 15:04:05"))
	}

	roomMutex.RLock()
	gameRoom, exists := rooms[roomID]
	roomMutex.RUnlock()

	if !exists {
		return errors.New("房间不存在")
	}

	gameRoom.mutex.Lock()
	defer gameRoom.mutex.Unlock()

	// 验证私密房间密钥
	if gameRoom.Room.IsPrivate && gameRoom.Room.AccessKey != "" {
		// 房主不需要验证密钥
		isCreator := len(gameRoom.Room.Players) > 0 && gameRoom.Room.Players[0] == uid
		// 已在房间的玩家不需要验证密钥
		alreadyInRoom := false
		for _, pid := range gameRoom.Room.Players {
			if pid == uid {
				alreadyInRoom = true
				break
			}
		}

		if !isCreator && !alreadyInRoom && accessKey != gameRoom.Room.AccessKey {
			return errors.New("访问密钥错误，无法加入私密房间")
		}
	}

	// 已经在房间里或试图重新加入
	for _, pid := range gameRoom.Room.Players {
		if pid == uid {
			// 如果在离线列表中，移除它
			wasOffline := false
			if _, exists := gameRoom.OfflineAt[uid]; exists {
				delete(gameRoom.OfflineAt, uid)
				wasOffline = true
			}
			// 只有在玩家之前处于离线状态时才广播更新和检查自动开始
			if wasOffline {
				gameRoom.checkAutoStart()
				gameRoom.broadcastRoomUpdate()
			}
			return nil
		}
	}
	for _, sid := range gameRoom.Room.Spectators {
		if sid == uid {
			return nil
		}
	}

	// 游戏已开始或房间已满，自动进入观战模式
	if gameRoom.Room.Status == "playing" || len(gameRoom.Room.Players) >= gameRoom.Room.MaxPlayers {
		// 如果是私密房间，非房间成员需要密钥才能观战
		if gameRoom.Room.IsPrivate && gameRoom.Room.AccessKey != "" {
			isCreator := len(gameRoom.Room.Players) > 0 && gameRoom.Room.Players[0] == uid
			alreadyInRoom := false
			for _, pid := range gameRoom.Room.Players {
				if pid == uid {
					alreadyInRoom = true
					break
				}
			}
			// 观战者也需要验证密钥（除非是房主或已在房间）
			if !isCreator && !alreadyInRoom && accessKey != gameRoom.Room.AccessKey {
				if gameRoom.Room.Status == "playing" {
					return errors.New("访问密钥错误，无法观战私密房间")
				} else {
					return errors.New("访问密钥错误，无法加入私密房间")
				}
			}
		}

		gameRoom.Room.Spectators = append(gameRoom.Room.Spectators, uid)
		if gameRoom.GameState != nil {
			gameRoom.GameState.Spectators = append(gameRoom.GameState.Spectators, uid)
		}
		gameRoom.broadcastRoomUpdate()
		return nil
	}

	// 检查是否已在房间中
	for _, pid := range gameRoom.Room.Players {
		if pid == uid {
			return errors.New("已在房间中")
		}
	}

	gameRoom.Room.Players = append(gameRoom.Room.Players, uid)
	repository.UserRepo.UpdateRoomReadyStatus(uint(uid), false)
	gameRoom.checkAutoStart()
	return nil
}

// 离开房间
func LeaveRoom(roomID string, uid int) error {
	roomMutex.RLock()
	gameRoom, exists := rooms[roomID]
	roomMutex.RUnlock()

	if !exists {
		return errors.New("房间不存在")
	}

	gameRoom.mutex.Lock()
	defer gameRoom.mutex.Unlock()

	// 移除玩家
	newPlayers := []int{}
	for _, pid := range gameRoom.Room.Players {
		if pid != uid {
			newPlayers = append(newPlayers, pid)
		}
	}
	gameRoom.Room.Players = newPlayers

	// 移除准备状态
	newReady := []int{}
	for _, rid := range gameRoom.Room.ReadyUIDs {
		if rid != uid {
			newReady = append(newReady, rid)
		}
	}
	gameRoom.Room.ReadyUIDs = newReady
	repository.UserRepo.UpdateRoomReadyStatus(uint(uid), false)

	// 移除观战者
	newSpectators := []int{}
	for _, sid := range gameRoom.Room.Spectators {
		if sid != uid {
			newSpectators = append(newSpectators, sid)
		}
	}
	gameRoom.Room.Spectators = newSpectators

	// 检查自动开始状态
	if gameRoom.Room.Status == "waiting" {
		gameRoom.checkAutoStart()
	}

	// 如果游戏正在进行，也从 GameState 中移除
	if gameRoom.GameState != nil {
		newPS := []*models.PlayerState{}
		leftIndex := -1
		isFinished := false
		for _, fuid := range gameRoom.GameState.FinishedPlayers {
			if fuid == uid {
				isFinished = true
				break
			}
		}

		for i, ps := range gameRoom.GameState.Players {
			if ps.UID != uid {
				newPS = append(newPS, ps)
			} else {
				leftIndex = i
			}
		}

		if !isFinished && leftIndex != -1 {
			gameRoom.GameState.QuittedCount++
		}

		gameRoom.GameState.Players = newPS

		// 调整当前玩家索引
		if leftIndex != -1 {
			if gameRoom.GameState.CurrentPlayer > leftIndex {
				gameRoom.GameState.CurrentPlayer--
			}
			if gameRoom.GameState.CurrentPlayer >= len(gameRoom.GameState.Players) {
				gameRoom.GameState.CurrentPlayer = 0
			}
		}
	}

	// 从 WebSocket Hub 中移除该用户的房间订阅
	if websocket.GlobalHub != nil {
		websocket.GlobalHub.LeaveRoomByUID(uid)
	}

	// 如果是正在游戏中且玩家少于2个，则解散
	if gameRoom.Room.Status == "playing" && len(gameRoom.Room.Players) < 2 {
		gameRoom.terminateRoom("由于实验样本不足（少于2人），本次反应宣告失败，实验室已关闭")
		return nil
	}

	// 如果所有玩家和观战者均已离开，销毁房间
	if len(gameRoom.Room.Players) == 0 && len(gameRoom.Room.Spectators) == 0 {
		gameRoom.cancelStartTimer()
		roomMutex.Lock()
		delete(rooms, roomID)
		roomMutex.Unlock()
		log.Printf("房间 %s 已空，已自动关闭并清理资源", roomID)
	}

	return nil
}

// 开始游戏
func StartGame(roomID string, uid int) error {
	roomMutex.RLock()
	gameRoom, exists := rooms[roomID]
	roomMutex.RUnlock()

	if !exists {
		return errors.New("房间不存在")
	}

	gameRoom.mutex.Lock()
	defer gameRoom.mutex.Unlock()

	if gameRoom.Room.Status != "waiting" {
		return errors.New("游戏已在进行中")
	}

	if len(gameRoom.Room.Players) < 2 {
		return errors.New("至少需要2名玩家")
	}

	if gameRoom.Room.Status != "waiting" {
		return errors.New("游戏已开始")
	}

	// 初始化游戏状态
	gameRoom.GameState = &models.GameState{
		RoomID:              roomID,
		Players:             []*models.PlayerState{},
		OriginalPlayerCount: len(gameRoom.Room.Players),
		QuittedCount:        0,
		CurrentPlayer:       0,
		Direction:           1,
		DrawPile:            []models.Card{},
		DiscardPile:         []models.PlayedCard{},
		Status:              "playing",
		TurnEndTime:         time.Now().Add(getPlayerActionTimeout()).UnixNano() / int64(time.Millisecond),
		PendingDrawCount:    0,
		PendingDrawTypes:    nil,
		AllowedAnyPlayer:    -1,
	}

	// 创建牌堆
	for cardType, count := range gameRoom.Room.DeckConfig.Cards {
		effect := getCardEffect(cardType)
		for i := 0; i < count; i++ {
			gameRoom.GameState.DrawPile = append(gameRoom.GameState.DrawPile, models.Card{
				Type:   cardType,
				Count:  1,
				Effect: effect,
			})
		}
	}

	// 洗牌
	rand.Shuffle(len(gameRoom.GameState.DrawPile), func(i, j int) {
		gameRoom.GameState.DrawPile[i], gameRoom.GameState.DrawPile[j] =
			gameRoom.GameState.DrawPile[j], gameRoom.GameState.DrawPile[i]
	})

	// 随机排序玩家顺序
	shuffledPlayers := make([]int, len(gameRoom.Room.Players))
	copy(shuffledPlayers, gameRoom.Room.Players)
	rand.Shuffle(len(shuffledPlayers), func(i, j int) {
		shuffledPlayers[i], shuffledPlayers[j] = shuffledPlayers[j], shuffledPlayers[i]
	})

	// 计算每个玩家应当获得相同的初始手牌数
	initialCardsCount := gameRoom.Room.DeckConfig.InitialCards
	if initialCardsCount <= 0 {
		initialCardsCount = 10 // 容错
	}

	numPlayers := len(shuffledPlayers)
	totalCardsNeeded := numPlayers * initialCardsCount

	// 检查牌堆是否足够
	if len(gameRoom.GameState.DrawPile) < totalCardsNeeded {
		log.Printf("[初始手牌] ⚠️  牌堆不足：需要 %d 张，实际 %d 张，将平均分配",
			totalCardsNeeded, len(gameRoom.GameState.DrawPile))
		initialCardsCount = len(gameRoom.GameState.DrawPile) / numPlayers
	}

	log.Printf("[初始手牌] 📊 牌堆统计：总计 %d 张，将为 %d 位玩家每人发 %d 张",
		len(gameRoom.GameState.DrawPile), numPlayers, initialCardsCount)

	// 统计牌堆中的卡牌类型分布（用于日志）
	cardTypeCount := make(map[string]int)
	for _, card := range gameRoom.GameState.DrawPile {
		cardTypeCount[card.Type]++
	}
	log.Printf("[初始手牌] 🎴 牌堆组成：%v", cardTypeCount)

	// 初始化玩家
	for _, pid := range shuffledPlayers {
		user, err := repository.UserRepo.FindByUID(uint(pid))
		username := ""
		avatar := ""
		if err != nil {
			username = fmt.Sprintf("研究员_%d", pid)
			avatar = "🧪"
		} else {
			username = user.Username
			avatar = user.Avatar
		}

		player := &models.PlayerState{
			UID:                   pid,
			Username:              username,
			Avatar:                avatar,
			HandCards:             []models.Card{},
			CardCount:             0,
			IsReady:               true,
			DoubleActionAvailable: false,
			ActionProgress:        0,
		}

		// 从洗好的牌堆顶部抽取初始手牌（按配置的比例随机分配）
		playerCardTypes := make(map[string]int)
		for i := 0; i < initialCardsCount && len(gameRoom.GameState.DrawPile) > 0; i++ {
			card := gameRoom.GameState.DrawPile[0]
			gameRoom.GameState.DrawPile = gameRoom.GameState.DrawPile[1:]
			player.HandCards = append(player.HandCards, card)
			player.CardCount++
			playerCardTypes[card.Type]++
		}

		log.Printf("[初始手牌] 👤 玩家 %s (UID:%d) 获得 %d 张手牌：%v",
			username, pid, player.CardCount, playerCardTypes)

		gameRoom.GameState.Players = append(gameRoom.GameState.Players, player)
	}

	// 抽出场上初始物质，从 reactions 表的 r1/r2 字段中随机选择
	var initialSubstance string
	var initialCard models.Card
	var foundBase bool

	// 从 reactions 表获取所有已批准的反应
	reactionRepo := repository.NewReactionRepository()
	approvedReactions, err := reactionRepo.FindApprovedReactions()

	if err == nil && len(approvedReactions) > 0 {
		// 提取所有 r1 和 r2 物质（去重）
		substanceSet := make(map[string]bool)
		for _, reaction := range approvedReactions {
			if reaction.R1 != "" {
				substanceSet[reaction.R1] = true
			}
			if reaction.R2 != "" {
				substanceSet[reaction.R2] = true
			}
		}

		// 转换为切片
		availableSubstances := make([]string, 0, len(substanceSet))
		for substance := range substanceSet {
			availableSubstances = append(availableSubstances, substance)
		}

		// 随机选择一个物质作为场上初始物质
		if len(availableSubstances) > 0 {
			randomIndex := rand.Intn(len(availableSubstances))
			initialSubstance = availableSubstances[randomIndex]

			log.Printf("[场上初始物质] 🎲 从 %d 个已批准反应中提取了 %d 种物质，随机选择: %s",
				len(approvedReactions), len(availableSubstances), initialSubstance)

			// 从牌堆中找出对应的卡牌作为"底座"
			// 优先找非功能牌，且类型匹配的卡牌
			for i, card := range gameRoom.GameState.DrawPile {
				// 检查是否为特殊功能牌
				isSpecial := false
				specialTypes := []string{"+2", "+4", "reverse", "Au", "He", "Ne", "Ar", "Kr"}
				for _, st := range specialTypes {
					if card.Type == st || card.Effect == st {
						isSpecial = true
						break
					}
				}

				// 优先选择与初始物质匹配的普通卡牌
				if !isSpecial && card.Type == initialSubstance {
					initialCard = card
					// 从牌堆移除该牌
					gameRoom.GameState.DrawPile = append(gameRoom.GameState.DrawPile[:i], gameRoom.GameState.DrawPile[i+1:]...)
					foundBase = true
					log.Printf("[场上初始物质] ✅ 找到匹配的卡牌: %s", card.Type)
					break
				}
			}

			// 如果没找到匹配的，选择任意非功能牌
			if !foundBase {
				for i, card := range gameRoom.GameState.DrawPile {
					isSpecial := false
					specialTypes := []string{"+2", "+4", "reverse", "Au", "He", "Ne", "Ar", "Kr"}
					for _, st := range specialTypes {
						if card.Type == st || card.Effect == st {
							isSpecial = true
							break
						}
					}

					if !isSpecial {
						initialCard = card
						// 从牌堆移除该牌
						gameRoom.GameState.DrawPile = append(gameRoom.GameState.DrawPile[:i], gameRoom.GameState.DrawPile[i+1:]...)
						foundBase = true
						log.Printf("[场上初始物质] ℹ️  未找到匹配卡牌，使用: %s (展示物质: %s)",
							card.Type, initialSubstance)
						break
					}
				}
			}
		}
	}

	if foundBase {
		playedCard := models.PlayedCard{
			Card:      initialCard,
			Substance: initialSubstance,
			PlayerUID: 0, // 表示系统出牌
		}
		gameRoom.GameState.DiscardPile = append(gameRoom.GameState.DiscardPile, playedCard)
		gameRoom.GameState.LastCard = &gameRoom.GameState.DiscardPile[0]
		log.Printf("[场上初始物质] 🎴 场上初始卡牌已设置: 卡牌类型=%s, 展示物质=%s",
			initialCard.Type, initialSubstance)
	} else {
		// 回退方案：如果 reactions 为空或没找到合适的底牌，从牌堆抽第一张非功能牌
		log.Println("[场上初始物质] ⚠️  未找到合适的初始物质，使用备用方案")
		for len(gameRoom.GameState.DrawPile) > 0 {
			firstCard := gameRoom.GameState.DrawPile[0]
			gameRoom.GameState.DrawPile = gameRoom.GameState.DrawPile[1:]

			specialTypes := []string{"+2", "+4", "reverse", "Au", "He", "Ne", "Ar", "Kr"}
			isSpecial := false
			for _, st := range specialTypes {
				if firstCard.Type == st || firstCard.Effect == st {
					isSpecial = true
					break
				}
			}

			if !isSpecial {
				playedCard := models.PlayedCard{
					Card:      firstCard,
					Substance: firstCard.Type,
					PlayerUID: 0, // 表示系统出牌
				}
				gameRoom.GameState.DiscardPile = append(gameRoom.GameState.DiscardPile, playedCard)
				gameRoom.GameState.LastCard = &gameRoom.GameState.DiscardPile[0]
				log.Printf("[场上初始物质] 🎴 备用方案：使用卡牌 %s", firstCard.Type)
				break
			} else {
				gameRoom.GameState.DrawPile = append(gameRoom.GameState.DrawPile, firstCard)
			}
		}
	}

	gameRoom.Room.Status = "playing"

	// 记录首个玩家的回合开始时间
	if len(gameRoom.GameState.Players) > 0 {
		firstUID := gameRoom.GameState.Players[0].UID
		repository.UserRepo.UpdateTurnStartedAt(uint(firstUID), time.Now())
	}

	// 添加详细日志
	log.Printf("[游戏开始] 房间 %s 游戏已开始，状态：%s，玩家数：%d",
		roomID, gameRoom.Room.Status, len(gameRoom.GameState.Players))
	for i, p := range gameRoom.GameState.Players {
		log.Printf("[游戏开始] 玩家 %d: UID=%d, 手牌数=%d", i, p.UID, p.CardCount)
	}
	log.Printf("[游戏开始] 牌堆剩余：%d张，弃牌堆：%d张，当前玩家索引：%d",
		len(gameRoom.GameState.DrawPile), len(gameRoom.GameState.DiscardPile), gameRoom.GameState.CurrentPlayer)

	return nil
}

// 获取房间状态（为当前玩家过滤信息）
func GetRoomState(roomID string, uid int) (map[string]interface{}, error) {
	roomMutex.RLock()
	gameRoom, exists := rooms[roomID]
	roomMutex.RUnlock()

	if !exists {
		return nil, errors.New("房间不存在")
	}

	gameRoom.mutex.RLock()
	defer gameRoom.mutex.RUnlock()

	// 检查玩家是否在房间中
	inRoom := false
	for _, pid := range gameRoom.Room.Players {
		if pid == uid {
			inRoom = true
			break
		}
	}
	if !inRoom {
		for _, sid := range gameRoom.Room.Spectators {
			if sid == uid {
				inRoom = true
				break
			}
		}
	}
	if !inRoom {
		return nil, errors.New("你不在该房间中")
	}

	// 获取玩家详细信息（用于准备页面）
	playersInfo := []map[string]interface{}{}
	for _, pid := range gameRoom.Room.Players {
		user, err := repository.UserRepo.FindByUID(uint(pid))
		username := fmt.Sprintf("研究员_%d", pid)
		avatar := "🧪"
		if err == nil {
			username = user.Username
			avatar = user.Avatar
		}
		offline := false
		if _, exists := gameRoom.OfflineAt[pid]; exists {
			offline = true
		}
		playersInfo = append(playersInfo, map[string]interface{}{
			"uid":        pid,
			"username":   username,
			"avatar":     avatar,
			"is_offline": offline,
		})
	}

	// 确保 ready_uids 永远不为 nil，避免 JSON 序列化为 null
	readyUIDs := gameRoom.Room.ReadyUIDs
	if readyUIDs == nil {
		readyUIDs = []int{}
	}

	result := map[string]interface{}{
		"id":             gameRoom.Room.ID,
		"name":           gameRoom.Room.Name,
		"players":        gameRoom.Room.Players,
		"ready_uids":     readyUIDs,
		"countdown":      gameRoom.Room.Countdown,
		"players_info":   playersInfo,
		"max_players":    gameRoom.Room.MaxPlayers,
		"status":         gameRoom.Room.Status,
		"is_points_mode": gameRoom.Room.IsPointsMode,
		"deck_config":    gameRoom.Room.DeckConfig,
		"is_private":     gameRoom.Room.IsPrivate,
		"access_key":     gameRoom.Room.AccessKey,
	}

	if gameRoom.GameState != nil {
		// 检查玩家是否由于已完成或中途加入而处于观战模式
		isSpectator := false
		for _, sid := range gameRoom.Room.Spectators {
			if sid == uid {
				isSpectator = true
				break
			}
		}
		for _, fuid := range gameRoom.GameState.FinishedPlayers {
			if fuid == uid {
				isSpectator = true
				break
			}
		}

		// 过滤其他玩家的手牌
		filteredPlayers := []*models.PlayerState{}
		for _, player := range gameRoom.GameState.Players {
			if player.UID == uid && !isSpectator {
				// 当前活跃玩家，显示全部信息
				filteredPlayers = append(filteredPlayers, player)
			} else {
				// 其他玩家或处于观战状态，隐藏手牌详情
				filteredPlayer := &models.PlayerState{
					UID:                   player.UID,
					Username:              player.Username,
					Avatar:                player.Avatar,
					HandCards:             nil, // 不显示具体手牌
					CardCount:             player.CardCount,
					IsReady:               player.IsReady,
					DoubleActionAvailable: player.DoubleActionAvailable,
					ActionProgress:        player.ActionProgress,
				}
				filteredPlayers = append(filteredPlayers, filteredPlayer)
			}
		}

		result["game_state"] = map[string]interface{}{
			"players":            filteredPlayers,
			"spectators":         gameRoom.Room.Spectators,
			"finished_players":   gameRoom.GameState.FinishedPlayers,
			"current_player":     gameRoom.GameState.CurrentPlayer,
			"direction":          gameRoom.GameState.Direction,
			"last_card":          gameRoom.GameState.LastCard,
			"deck_count":         len(gameRoom.GameState.DrawPile),
			"status":             gameRoom.GameState.Status,
			"turn_end_time":      gameRoom.GameState.TurnEndTime,
			"allowed_any_player": gameRoom.GameState.AllowedAnyPlayer,
			"pending_draw_count": gameRoom.GameState.PendingDrawCount,
			"is_spectator":       isSpectator,
		}
	}

	return result, nil
}

func getCardEffect(cardType string) string {
	effects := map[string]string{
		"+2": "+2",
		"+4": "+4",
		"He": "reverse",
		"Ne": "reverse",
		"Ar": "reverse",
		"Kr": "reverse",
		"Au": "Au",
		// "Choice": "wild", // 已移除
	}
	return effects[cardType]
}

// 出牌
func PlayCard(roomID string, uid int, card models.Card, substance string) error {
	roomMutex.RLock()
	gameRoom, exists := rooms[roomID]
	roomMutex.RUnlock()

	if !exists {
		return errors.New("房间不存在")
	}

	gameRoom.mutex.Lock()
	defer gameRoom.mutex.Unlock()

	if gameRoom.GameState == nil || gameRoom.GameState.Status != "playing" {
		return errors.New("游戏未开始")
	}

	// 验证玩家身份：必须是房间内的玩家，不能是观众
	isPlayer := false
	for _, pid := range gameRoom.Room.Players {
		if pid == uid {
			isPlayer = true
			break
		}
	}
	if !isPlayer {
		return errors.New("你不在游戏中")
	}

	// 检查是否已完成游戏（观众状态）
	for _, fuid := range gameRoom.GameState.FinishedPlayers {
		if fuid == uid {
			return errors.New("你已完成游戏，无法继续操作")
		}
	}

	// 检查是否轮到该玩家
	currentPlayer := gameRoom.GameState.Players[gameRoom.GameState.CurrentPlayer]
	// 预计算下家与下下家索引，便于处理 Au 跳过逻辑
	playersLen := len(gameRoom.GameState.Players)
	curIdx := gameRoom.GameState.CurrentPlayer
	dir := gameRoom.GameState.Direction
	next1 := curIdx + dir
	if next1 < 0 {
		next1 = playersLen - 1
	} else if next1 >= playersLen {
		next1 = 0
	}
	if currentPlayer.UID != uid {
		return errors.New("还没轮到你")
	}

	// 若未指定substance，自动用单质（如H->H2、O->O2等，或直接元素符号）
	if substance == "" {
		substance = card.Type
	}

	// 校验物质是否已录入
	if !IsValidSubstance(substance) {
		return errors.New("该物质未录入")
	}

	requiredElements := parseSubstance(substance)
	usedCards := []int{} // 记录将要从手牌中移除的索引
	for elemName := range requiredElements {
		// 普通反应时，仅考虑元素种类，不考虑元素系数
		count := 1
		foundCount := 0
		for c := 0; c < count; c++ {
			found := false
			for i, hCard := range currentPlayer.HandCards {
				// 检查该卡片是否已被标记为使用
				alreadyUsed := false
				for _, usedIdx := range usedCards {
					if usedIdx == i {
						alreadyUsed = true
						break
					}
				}
				if alreadyUsed {
					continue
				}
				if hCard.Type == elemName {
					usedCards = append(usedCards, i)
					found = true
					foundCount++
					break
				}
			}
			if !found {
				return errors.New("缺少元素牌: " + elemName + " (需要 " + fmt.Sprint(count) + " 张)")
			}
		}
	}

	// +2/4/Au/换向牌可随意打出，无需反应条件
	nobleGases := map[string]bool{"He": true, "Ne": true, "Ar": true, "Kr": true}
	specialTypes := map[string]bool{"+2": true, "+4": true, "Au": true}
	isSpecial := specialTypes[card.Type] || specialTypes[card.Effect] || nobleGases[card.Type]
	// 如果当前玩家被允许无视反应条件出牌，则跳过反应检查
	allowAny := gameRoom.GameState.AllowedAnyPlayer == gameRoom.GameState.CurrentPlayer
	if !isSpecial && gameRoom.GameState.LastCard != nil && !allowAny {
		canReact := false
		if len(gameRoom.GameState.LastCard.Reactants) > 0 {
			// 如果上一次是双联反应，则只需与其中任一物质反应即可
			for _, r := range gameRoom.GameState.LastCard.Reactants {
				if CanReact(r, substance) {
					canReact = true
					break
				}
			}
		} else if CanReact(gameRoom.GameState.LastCard.Substance, substance) {
			canReact = true
		}

		if !canReact {
			return errors.New("无法与上一张牌反应: " + substance)
		}
	}
	// nobleGases 作为换向牌

	// 检查选中的卡牌中是否有带效果的
	activeEffect := ""
	if card.Effect != "" {
		activeEffect = card.Effect
	}

	// 如果有累计加牌，本轮只能打出相同或更高数值的加牌进行叠加
	if gameRoom.GameState.PendingDrawCount > 0 {
		// 细化逻辑：必须打出加牌
		if activeEffect != "+2" && activeEffect != "+4" {
			return errors.New("当前累计加牌中，请打出加牌叠加或点摸牌结算")
		}
	}

	// 记录消耗的卡牌详情用于后续逻辑
	var consumedCards []models.Card
	sort.Ints(usedCards)
	for i := len(usedCards) - 1; i >= 0; i-- {
		idx := usedCards[i]
		consumedCard := currentPlayer.HandCards[idx]
		consumedCards = append(consumedCards, consumedCard)

		// 如果还没确定 activeEffect，且这张卡有效果，则使用它
		if activeEffect == "" && consumedCard.Effect != "" {
			activeEffect = consumedCard.Effect
		}

		currentPlayer.HandCards = append(
			currentPlayer.HandCards[:idx],
			currentPlayer.HandCards[idx+1:]...,
		)
		currentPlayer.CardCount--
	}
	// 将消耗的卡牌放入洗牌池
	gameRoom.GameState.AllUsedCards = append(gameRoom.GameState.AllUsedCards, consumedCards...)

	// 添加到弃牌堆
	// 使用第一张消耗的卡作为代表
	var displayCard models.Card
	if len(consumedCards) > 0 {
		displayCard = consumedCards[0]
	} else {
		displayCard = card
	}

	playedCard := models.PlayedCard{
		Card:      displayCard,
		Substance: substance,
		PlayerUID: uid,
	}
	// 如果是反转牌，不更新场上的物质（不更新 LastCard），使下家仍需与之前的物质反应
	if activeEffect != "reverse" {
		gameRoom.GameState.LastCard = &playedCard
	}
	gameRoom.GameState.DiscardPile = append(gameRoom.GameState.DiscardPile, playedCard)

	// 检查是否获胜
	if currentPlayer.CardCount == 0 {
		// 进入完成状态
		gameRoom.GameState.FinishedPlayers = append(gameRoom.GameState.FinishedPlayers, uid)

		// 检查是否场上只剩一人（或更少）未完成
		activeCount := 0
		var lastPlayerUID int
		for _, p := range gameRoom.GameState.Players {
			isFinished := false
			for _, fuid := range gameRoom.GameState.FinishedPlayers {
				if p.UID == fuid {
					isFinished = true
					break
				}
			}
			if !isFinished {
				activeCount++
				lastPlayerUID = p.UID
			}
		}

		if activeCount <= 1 {
			// 最后一名也将加入列表
			if activeCount == 1 {
				gameRoom.GameState.FinishedPlayers = append(gameRoom.GameState.FinishedPlayers, lastPlayerUID)
			}

			gameRoom.GameState.Status = "finished"
			gameRoom.Room.Status = "finished"
			winnerUID := gameRoom.GameState.FinishedPlayers[0]
			saveGameHistory(roomID, winnerUID, gameRoom.Room.Players, gameRoom.GameState.OriginalPlayerCount, gameRoom.GameState.QuittedCount)

			// 清理该房间的游戏邀请消息
			privateChatRepo := repository.NewPrivateChatRepository()
			if err := privateChatRepo.DeleteGameInvitesByRoom(roomID); err != nil {
				log.Printf("清理房间 %s 的游戏邀请失败: %v", roomID, err)
			}

			if gameRoom.Room.IsPointsMode {
				handlePointsCalculation(gameRoom)
			}
			return nil
		}

		gameRoom.GameState.CurrentPlayer = getNextPlayer(gameRoom.GameState)
		gameRoom.recordTurnStart()
		gameRoom.GameState.TurnEndTime = time.Now().Add(getPlayerActionTimeout()).UnixNano() / int64(time.Millisecond)
		return nil
	}

	// 加牌叠加规则
	if activeEffect == "+2" || activeEffect == "+4" {
		// 叠加pending
		gameRoom.GameState.PendingDrawCount += map[string]int{"+2": 2, "+4": 4}[activeEffect]
		gameRoom.GameState.PendingDrawTypes = append(gameRoom.GameState.PendingDrawTypes, activeEffect)
		// 传递到下家
		gameRoom.GameState.CurrentPlayer = getNextPlayer(gameRoom.GameState)
		gameRoom.recordTurnStart()
		gameRoom.GameState.TurnEndTime = time.Now().Add(getPlayerActionTimeout()).UnixNano() / int64(time.Millisecond)
		// 如果之前允许任意出牌的标记被消费（且未产生新转移），清除
		if gameRoom.GameState.AllowedAnyPlayer == curIdx {
			gameRoom.GameState.AllowedAnyPlayer = -1
		}
		return nil
	} else if gameRoom.GameState.PendingDrawCount > 0 {
		// 不能继续叠加时，强制摸pending
		drawCardsForPlayer(gameRoom, gameRoom.GameState.CurrentPlayer, gameRoom.GameState.PendingDrawCount)
		gameRoom.GameState.PendingDrawCount = 0
		gameRoom.GameState.PendingDrawTypes = nil
		// 回合传递
		gameRoom.GameState.CurrentPlayer = getNextPlayer(gameRoom.GameState)
		gameRoom.recordTurnStart()
		gameRoom.GameState.TurnEndTime = time.Now().Add(getPlayerActionTimeout()).UnixNano() / int64(time.Millisecond)

		// 结算罚牌后清空场面并允许下家随意出牌
		gameRoom.GameState.LastCard = nil
		gameRoom.GameState.AllowedAnyPlayer = gameRoom.GameState.CurrentPlayer

		return nil
	}
	// 其他效果
	switch activeEffect {
	case "reverse":
		gameRoom.GameState.Direction *= -1
	case "Au":
		// 使得跳过下一位玩家，并允许下下家任意出牌，同时清空场面
		gameRoom.GameState.LastCard = nil
		// 1. 先跳过第一个人 (考虑到已完成玩家)
		skippedIdx := getNextPlayer(gameRoom.GameState)
		gameRoom.GameState.CurrentPlayer = skippedIdx
		// 2. 找到真正该出牌的人
		targetIdx := getNextPlayer(gameRoom.GameState)
		gameRoom.GameState.AllowedAnyPlayer = targetIdx

		// 显式广播跳过信息
		if websocket.GlobalHub != nil {
			skippedPlayer := gameRoom.GameState.Players[skippedIdx].Username
			nextPlayer := gameRoom.GameState.Players[targetIdx].Username
			websocket.GlobalHub.BroadcastToRoom(gameRoom.Room.ID, websocket.Message{
				Type: "action_toast",
				Data: fmt.Sprintf("Au 金元素触发！跳过研究员 %s，等待 %s 出牌...", skippedPlayer, nextPlayer),
			})
		}
	}

	// 行动进度更新
	currentPlayer.ActionProgress++
	if currentPlayer.ActionProgress >= 2 {
		currentPlayer.DoubleActionAvailable = true
	}

	// 下一位玩家（大多数情况均走到这里）
	gameRoom.GameState.CurrentPlayer = getNextPlayer(gameRoom.GameState)
	gameRoom.recordTurnStart()
	// 如果之前允许任意出牌的标记被消费（且未在此处产生新的转移，如 Au 效果），清除
	if gameRoom.GameState.AllowedAnyPlayer == curIdx {
		gameRoom.GameState.AllowedAnyPlayer = -1
	}
	gameRoom.GameState.TurnEndTime = time.Now().Add(getPlayerActionTimeout()).UnixNano() / int64(time.Millisecond)
	return nil
}

func getNextPlayer(state *models.GameState) int {
	next := state.CurrentPlayer
	for {
		next = next + state.Direction
		if next < 0 {
			next = len(state.Players) - 1
		} else if next >= len(state.Players) {
			next = 0
		}

		// 检查该玩家是否已经出完牌
		uid := state.Players[next].UID
		isFinished := false
		for _, fuid := range state.FinishedPlayers {
			if uid == fuid {
				isFinished = true
				break
			}
		}

		if !isFinished {
			return next
		}
	}
}

func reshuffleDeck(gameRoom *GameRoom) {
	if len(gameRoom.GameState.AllUsedCards) == 0 {
		return
	}
	// 将池中卡牌放回摸牌堆
	gameRoom.GameState.DrawPile = append(gameRoom.GameState.DrawPile, gameRoom.GameState.AllUsedCards...)
	gameRoom.GameState.AllUsedCards = nil

	// 重新洗牌
	rand.Shuffle(len(gameRoom.GameState.DrawPile), func(i, j int) {
		gameRoom.GameState.DrawPile[i], gameRoom.GameState.DrawPile[j] =
			gameRoom.GameState.DrawPile[j], gameRoom.GameState.DrawPile[i]
	})
}

func drawCardsForPlayer(gameRoom *GameRoom, playerIndex int, count int) {
	player := gameRoom.GameState.Players[playerIndex]
	for i := 0; i < count; i++ {
		// 如果摸牌堆空了，尝试洗牌
		if len(gameRoom.GameState.DrawPile) == 0 {
			reshuffleDeck(gameRoom)
		}
		if len(gameRoom.GameState.DrawPile) == 0 {
			break // 彻底没牌了
		}
		card := gameRoom.GameState.DrawPile[0]
		gameRoom.GameState.DrawPile = gameRoom.GameState.DrawPile[1:]
		player.HandCards = append(player.HandCards, card)
		player.CardCount++
	}
}

// 摸牌
func DrawCard(roomID string, uid int, count int) error {
	roomMutex.RLock()
	gameRoom, exists := rooms[roomID]
	roomMutex.RUnlock()

	if !exists {
		return errors.New("房间不存在")
	}

	gameRoom.mutex.Lock()
	defer gameRoom.mutex.Unlock()

	if gameRoom.GameState == nil || gameRoom.GameState.Status != "playing" {
		return errors.New("游戏未开始")
	}

	// 验证玩家身份：必须是房间内的玩家
	isPlayer := false
	for _, pid := range gameRoom.Room.Players {
		if pid == uid {
			isPlayer = true
			break
		}
	}
	if !isPlayer {
		return errors.New("你不在游戏中")
	}

	// 检查是否已完成游戏
	for _, fuid := range gameRoom.GameState.FinishedPlayers {
		if fuid == uid {
			return errors.New("你已完成游戏，无法继续操作")
		}
	}

	currentPlayer := gameRoom.GameState.Players[gameRoom.GameState.CurrentPlayer]
	if currentPlayer.UID != uid {
		return errors.New("还没轮到你")
	}

	actualCount := count
	penaltyResolved := false
	// 如果有挂起的加牌，结算加牌
	if gameRoom.GameState.PendingDrawCount > 0 {
		actualCount = gameRoom.GameState.PendingDrawCount
		gameRoom.GameState.PendingDrawCount = 0
		gameRoom.GameState.PendingDrawTypes = nil
		penaltyResolved = true
	}

	drawCardsForPlayer(gameRoom, gameRoom.GameState.CurrentPlayer, actualCount)

	// 行动进度更新
	currentPlayer.ActionProgress++
	if currentPlayer.ActionProgress >= 2 {
		currentPlayer.DoubleActionAvailable = true
	}

	// 摸牌后跳过回合
	gameRoom.GameState.CurrentPlayer = getNextPlayer(gameRoom.GameState)
	gameRoom.recordTurnStart()
	gameRoom.GameState.TurnEndTime = time.Now().Add(getPlayerActionTimeout()).UnixNano() / int64(time.Millisecond)

	// 如果结算了罚牌，清空场面并允许下家随意出牌
	if penaltyResolved {
		gameRoom.GameState.LastCard = nil
		gameRoom.GameState.AllowedAnyPlayer = gameRoom.GameState.CurrentPlayer
	} else {
		// 普通摸牌清除可能存在的 allowAny 标记
		gameRoom.GameState.AllowedAnyPlayer = -1
	}

	return nil
}

// 获取可用物质
func GetAvailableSubstances(roomID string, uid int) ([]string, error) {
	roomMutex.RLock()
	gameRoom, exists := rooms[roomID]
	roomMutex.RUnlock()

	if !exists {
		return nil, errors.New("房间不存在")
	}

	gameRoom.mutex.RLock()
	defer gameRoom.mutex.RUnlock()

	if gameRoom.GameState == nil || gameRoom.GameState.Status != "playing" {
		return nil, errors.New("游戏未开始")
	}

	// 验证玩家身份
	isPlayer := false
	for _, pid := range gameRoom.Room.Players {
		if pid == uid {
			isPlayer = true
			break
		}
	}
	if !isPlayer {
		return nil, errors.New("你不在游戏中")
	}

	// 检查是否已完成游戏
	for _, fuid := range gameRoom.GameState.FinishedPlayers {
		if fuid == uid {
			return nil, errors.New("你已完成游戏，无法继续操作")
		}
	}

	currentPlayer := gameRoom.GameState.Players[gameRoom.GameState.CurrentPlayer]
	if currentPlayer.UID != uid {
		return nil, errors.New("还没轮到你")
	}

	// 如果有挂起的加牌，除非手牌有加牌，否则不能进行任何普通化学反应
	if gameRoom.GameState.PendingDrawCount > 0 {
		return []string{}, nil
	}

	// 获取手牌能组成的所有物质
	substances := GetSubstancesFromElements(currentPlayer.HandCards)

	// 如果有上一张牌，过滤出能反应的物质
	// 如果该玩家被允许无视条件出牌（如 Au 效果或罚牌结算后），则返回全部可用物质
	allowAny := gameRoom.GameState.AllowedAnyPlayer == gameRoom.GameState.CurrentPlayer
	if gameRoom.GameState.LastCard != nil && !allowAny {
		reactable := []string{}
		if len(gameRoom.GameState.LastCard.Reactants) > 0 {
			// 如果上一次是双联反应，则只需与其中任一物质参与反应即可
			for _, sub := range substances {
				canSubReact := false
				for _, r := range gameRoom.GameState.LastCard.Reactants {
					if CanReact(r, sub) {
						canSubReact = true
						break
					}
				}
				if canSubReact {
					reactable = append(reactable, sub)
				}
			}
		} else {
			lastSubstance := gameRoom.GameState.LastCard.Substance
			for _, sub := range substances {
				if CanReact(lastSubstance, sub) {
					reactable = append(reactable, sub)
				}
			}
		}
		return reactable, nil
	}

	return substances, nil
}

// GetReactionHints 根据场上物质查询反应表提供提示；若场上无卡牌，返回数据库中的物质列表
func GetReactionHints(roomID string, uid int) ([]map[string]string, error) {
	roomMutex.RLock()
	gameRoom, exists := rooms[roomID]
	roomMutex.RUnlock()

	if !exists {
		return nil, errors.New("房间不存在")
	}

	gameRoom.mutex.RLock()
	defer gameRoom.mutex.RUnlock()

	// 验证玩家身份
	isPlayer := false
	for _, pid := range gameRoom.Room.Players {
		if pid == uid {
			isPlayer = true
			break
		}
	}
	if !isPlayer {
		return nil, errors.New("你不在游戏中")
	}

	var hints []map[string]string

	// 游戏进行中且场上有卡牌：查询反应表
	if gameRoom.GameState != nil && gameRoom.GameState.Status == "playing" && gameRoom.GameState.LastCard != nil {
		var fieldSubstances []string
		if len(gameRoom.GameState.LastCard.Reactants) > 0 {
			fieldSubstances = gameRoom.GameState.LastCard.Reactants
		} else if gameRoom.GameState.LastCard.Substance != "" {
			fieldSubstances = []string{gameRoom.GameState.LastCard.Substance}
		}

		seen := make(map[string]bool)
		for _, fieldSub := range fieldSubstances {
			reactables := GetReactableSubstances(fieldSub)
			for _, r := range reactables {
				if !seen[r] {
					seen[r] = true
					hints = append(hints, map[string]string{
						"substance": r,
						"source":    fieldSub,
					})
				}
			}
		}
		return hints, nil
	}

	// 场上无卡牌或游戏未开始：从数据库查询已批准物质
	if database.DB != nil {
		substances, err := repository.SubstanceRepo.FindApproved()
		if err == nil {
			for _, sub := range substances {
				hints = append(hints, map[string]string{
					"substance": sub.Formula,
					"name":      sub.Name,
				})
			}
		}
	}

	return hints, nil
}

func saveGameHistory(roomID string, winnerUID int, players []int, originalPlayerCount int, quittedCount int) {
	// 创建游戏历史记录
	playersJSON, _ := json.Marshal(players)
	history := &database.GameHistory{
		RoomID:              roomID,
		Players:             playersJSON,
		OriginalPlayerCount: originalPlayerCount,
		QuittedCount:        quittedCount,
		FinishedAt:          time.Now(),
	}

	if winnerUID > 0 {
		wUID := uint(winnerUID)
		history.WinnerUID = &wUID
	}

	err := repository.GameRepo.Create(history)
	if err != nil {
		fmt.Printf("保存游戏历史失败: %v\n", err)
	}

	// 更新玩家的总场次
	for _, uid := range players {
		repository.UserRepo.IncrementTotalGames(uint(uid))
	}

	// 更新胜利者的胜利场数
	if winnerUID > 0 {
		repository.UserRepo.IncrementWinCount(uint(winnerUID))
	}

	fmt.Println("游戏历史已保存，玩家统计已更新")
}

func init() {
	// 启动超时检查协程
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			checkRoomsTimeout()
		}
	}()
}

// InitGameConfig 初始化游戏配置（需要在数据库初始化后调用）
func InitGameConfig() error {
	// 初始化配置仓库
	configRepo = repository.NewConfigRepository()
	// 初始化默认配置
	if err := configRepo.InitDefaultConfigs(); err != nil {
		return fmt.Errorf("初始化默认配置失败: %v", err)
	}
	log.Println("✅ 游戏时间配置初始化成功")

	// 初始化物质缓存
	RebuildSubstanceCache()

	return nil
}

// getPlayerKickTimeout 获取玩家离线踢出超时时间
func getPlayerKickTimeout() time.Duration {
	if configRepo == nil {
		return 30 * time.Second
	}
	return configRepo.GetDurationValue("player_kick_timeout", 30*time.Second)
}

// getPlayerActionTimeout 获取玩家操作超时时间
func getPlayerActionTimeout() time.Duration {
	if configRepo == nil {
		return 30 * time.Second
	}
	return configRepo.GetDurationValue("player_action_timeout", 30*time.Second)
}

// getAutoStartTimeout 获取满员全准备自动开始倒计时（秒）
func getAutoStartTimeout() int {
	if configRepo == nil {
		return 10
	}
	return configRepo.GetIntValue("auto_start_timeout", 10)
}

// getHalfReadyTimeout 获取半数准备自动开始倒计时（秒）
func getHalfReadyTimeout() int {
	if configRepo == nil {
		return 60
	}
	return configRepo.GetIntValue("half_ready_timeout", 60)
}

func checkRoomsTimeout() {
	roomMutex.RLock()
	activeRooms := make([]string, 0)
	for id, r := range rooms {
		if r.GameState != nil && r.GameState.Status == "playing" {
			activeRooms = append(activeRooms, id)
		}
	}
	roomMutex.RUnlock()

	for _, roomID := range activeRooms {
		processRoomTimeout(roomID)
	}
}

// 双联反应行动
func DoublePlay(roomID string, uid int, sub1 string, sub2 string) error {
	roomMutex.RLock()
	gameRoom, exists := rooms[roomID]
	roomMutex.RUnlock()

	if !exists {
		return errors.New("房间不存在")
	}

	gameRoom.mutex.Lock()
	defer gameRoom.mutex.Unlock()

	if gameRoom.GameState == nil || gameRoom.GameState.Status != "playing" {
		return errors.New("游戏未开始")
	}

	// 验证玩家身份：必须是房间内的玩家
	isPlayer := false
	for _, pid := range gameRoom.Room.Players {
		if pid == uid {
			isPlayer = true
			break
		}
	}
	if !isPlayer {
		return errors.New("你不在游戏中")
	}

	// 检查是否已完成游戏
	for _, fuid := range gameRoom.GameState.FinishedPlayers {
		if fuid == uid {
			return errors.New("你已完成游戏，无法继续操作")
		}
	}

	// 检查是否轮到该玩家
	curIdx := gameRoom.GameState.CurrentPlayer
	currentPlayer := gameRoom.GameState.Players[curIdx]
	if currentPlayer.UID != uid {
		return errors.New("还没轮到你")
	}

	// 检查冷却
	if !currentPlayer.DoubleActionAvailable {
		return errors.New("双联反应尚未就绪（每行动2次可使用1次）")
	}

	// 当玩家选择自身两物质反应时，不考虑先前出牌（即跳过与场上 LastCard 的连接检查）

	// 如果有挂起的加牌，禁止发动双联行动
	if gameRoom.GameState.PendingDrawCount > 0 {
		return errors.New("当前处于加牌结算状态，不可发动双联反应")
	}

	// 校验物质是否已录入
	if !IsValidSubstance(sub1) {
		return errors.New("该物质未录入: " + sub1)
	}
	if !IsValidSubstance(sub2) {
		return errors.New("该物质未录入: " + sub2)
	}

	// 校验两物质是否能反应
	if !CanReact(sub1, sub2) {
		return errors.New(sub1 + " 与 " + sub2 + " 之间无法产生反应，不可发动双联行动")
	}

	// 准备所需元素
	req1 := parseSubstance(sub1)
	req2 := parseSubstance(sub2)
	allReqs := make(map[string]int)
	// 仅考虑元素种类，分别计算元素，若两物质中有相同元素，计算两次
	for k := range req1 {
		allReqs[k]++
	}
	for k := range req2 {
		allReqs[k]++
	}

	// 检查手牌
	usedCards := []int{}
	for elemName, count := range allReqs {
		foundCount := 0
		for c := 0; c < count; c++ {
			found := false
			for i, hCard := range currentPlayer.HandCards {
				alreadyUsed := false
				for _, uIdx := range usedCards {
					if uIdx == i {
						alreadyUsed = true
						break
					}
				}
				if alreadyUsed {
					continue
				}
				if hCard.Type == elemName {
					usedCards = append(usedCards, i)
					found = true
					foundCount++
					break
				}
			}
			if !found {
				return errors.New("手牌对应元素牌不足: " + elemName + " (需要 " + fmt.Sprint(count) + " 张)")
			}
		}
	}

	// 消耗卡牌
	sort.Ints(usedCards)
	var representCard models.Card
	var consumedCards []models.Card
	for i := len(usedCards) - 1; i >= 0; i-- {
		idx := usedCards[i]
		c := currentPlayer.HandCards[idx]
		consumedCards = append(consumedCards, c)
		if i == len(usedCards)-1 {
			representCard = c
		}
		currentPlayer.HandCards = append(currentPlayer.HandCards[:idx], currentPlayer.HandCards[idx+1:]...)
		currentPlayer.CardCount--
	}
	// 将消耗的卡牌放入洗牌池
	gameRoom.GameState.AllUsedCards = append(gameRoom.GameState.AllUsedCards, consumedCards...)

	// 记录已出牌
	playedCard := models.PlayedCard{
		Card:      representCard,
		Substance: sub1 + " + " + sub2,
		PlayerUID: uid,
		Reactants: []string{sub1, sub2}, // 标记下家可接其中任一
	}
	gameRoom.GameState.LastCard = &playedCard
	gameRoom.GameState.DiscardPile = append(gameRoom.GameState.DiscardPile, playedCard)

	// 处理特殊效果（如果双联中包含功能牌）
	for _, c := range consumedCards {
		effect := c.Type
		if c.Effect != "" {
			effect = c.Effect
		}
		switch effect {
		case "+2":
			gameRoom.GameState.PendingDrawCount += 2
			gameRoom.GameState.PendingDrawTypes = append(gameRoom.GameState.PendingDrawTypes, "+2")
		case "+4":
			gameRoom.GameState.PendingDrawCount += 4
			gameRoom.GameState.PendingDrawTypes = append(gameRoom.GameState.PendingDrawTypes, "+4")
		case "reverse":
			gameRoom.GameState.Direction *= -1
		case "Au":
			// 双联中的 Au 效果：跳过下一位并清空场面
			gameRoom.GameState.LastCard = nil
			// 1. 先跳过第一个人 (考虑到已完成玩家)
			skippedIdx := getNextPlayer(gameRoom.GameState)
			gameRoom.GameState.CurrentPlayer = skippedIdx
			// 2. 找到真正该出牌的人
			targetIdx := getNextPlayer(gameRoom.GameState)
			gameRoom.GameState.AllowedAnyPlayer = targetIdx

			if websocket.GlobalHub != nil {
				skippedPlayer := gameRoom.GameState.Players[skippedIdx].Username
				nextPlayer := gameRoom.GameState.Players[targetIdx].Username
				websocket.GlobalHub.BroadcastToRoom(gameRoom.Room.ID, websocket.Message{
					Type: "action_toast",
					Data: fmt.Sprintf("Au 金元素双联触发！跳过研究员 %s，等待 %s 出牌...", skippedPlayer, nextPlayer),
				})
			}
		}
	}

	// 检查是否获胜
	if currentPlayer.CardCount == 0 {
		gameRoom.GameState.Status = "finished"
		gameRoom.Room.Status = "finished"
		// 记录游戏历史
		saveGameHistory(roomID, uid, gameRoom.Room.Players, gameRoom.GameState.OriginalPlayerCount, gameRoom.GameState.QuittedCount)

		// 清理该房间的游戏邀请消息
		privateChatRepo := repository.NewPrivateChatRepository()
		if err := privateChatRepo.DeleteGameInvitesByRoom(roomID); err != nil {
			log.Printf("清理房间 %s 的游戏邀请失败: %v", roomID, err)
		}

		return nil
	}

	// 重置冷却
	currentPlayer.ActionProgress = 0
	currentPlayer.DoubleActionAvailable = false

	// 下一位玩家
	gameRoom.GameState.CurrentPlayer = getNextPlayer(gameRoom.GameState)
	gameRoom.recordTurnStart()
	// 如果之前允许任意出牌的标记被消费（且未在此处产生新的转移，如 Au 效果），清除
	if gameRoom.GameState.AllowedAnyPlayer == curIdx {
		gameRoom.GameState.AllowedAnyPlayer = -1
	}
	gameRoom.GameState.TurnEndTime = time.Now().Add(getPlayerActionTimeout()).UnixNano() / int64(time.Millisecond)

	return nil
}

func processRoomTimeout(roomID string) {
	roomMutex.RLock()
	gameRoom, exists := rooms[roomID]
	roomMutex.RUnlock()

	if !exists || gameRoom.GameState == nil || gameRoom.GameState.Status != "playing" {
		return
	}

	gameRoom.mutex.Lock()
	defer gameRoom.mutex.Unlock()

	now := time.Now().UnixNano() / int64(time.Millisecond)
	if gameRoom.GameState.TurnEndTime > 0 && now > gameRoom.GameState.TurnEndTime {
		// 超时处理：强制摸牌并跳过
		drawCount := 2
		penaltyResolved := false
		if gameRoom.GameState.PendingDrawCount > 0 {
			drawCount = gameRoom.GameState.PendingDrawCount
			gameRoom.GameState.PendingDrawCount = 0
			gameRoom.GameState.PendingDrawTypes = nil
			penaltyResolved = true
		}
		drawCardsForPlayer(gameRoom, gameRoom.GameState.CurrentPlayer, drawCount)
		gameRoom.GameState.CurrentPlayer = getNextPlayer(gameRoom.GameState)
		gameRoom.recordTurnStart()
		gameRoom.GameState.TurnEndTime = time.Now().Add(getPlayerActionTimeout()).UnixNano() / int64(time.Millisecond)

		// 如果结算了罚牌，清空场面并允许下家随意出牌
		if penaltyResolved {
			gameRoom.GameState.LastCard = nil
			gameRoom.GameState.AllowedAnyPlayer = gameRoom.GameState.CurrentPlayer
		} else {
			gameRoom.GameState.AllowedAnyPlayer = -1
		}

		// 广播更新
		go func(id string) {
			websocket.GlobalHub.BroadcastToRoom(id, websocket.Message{
				Type: "game_update",
			})
		}(roomID)
	}
}

// AdminKickPlayer 管理员强制踢出玩家
func AdminKickPlayer(roomID string, targetUID int, reason string) error {
	roomMutex.RLock()
	gr, ok := rooms[roomID]
	roomMutex.RUnlock()

	if !ok {
		return errors.New("房间不存在")
	}

	gr.mutex.Lock()
	defer gr.mutex.Unlock()

	// 检查玩家是否在房间中
	found := false
	for _, puid := range gr.Room.Players {
		if puid == targetUID {
			found = true
			break
		}
	}
	if !found {
		return errors.New("玩家不在房间中")
	}

	if reason == "" {
		reason = "由于管理员操作，您已被踢出实验"
	}

	gr.mutex.Unlock()
	gr.kickPlayer(targetUID, reason)
	return nil
}
