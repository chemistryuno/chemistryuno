package game

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/models"
	"chemistryuno/backend/plugins"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/websocket"
	crand "crypto/rand"
	"encoding/binary"
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

// scientistNames AI 玩家姓名库
var scientistNames = []string{
	"门捷列夫", "拉瓦锡", "居里夫人", "波义耳", "诺贝尔",
	"道尔顿", "阿伏伽德罗", "法拉第", "海森堡", "薛定谔",
	"波尔", "欧本海默", "屠呦呦", "钱学森", "袁隆平",
	"爱因斯坦", "牛顿", "达尔文", "特斯拉", "霍金",
	"费曼", "普朗克", "玻尔兹曼", "盖-吕萨克", "舍勒",
	"贝切里乌斯", "凯库勒", "范霍夫", "维勒", "李比希",
	"路易斯", "鲍林", "伍德沃德", "桑格", "霍奇金",
}

// cryptoRandUint64 使用 crypto/rand 生成密码学安全的随机 uint64
func cryptoRandUint64() uint64 {
	var b [8]byte
	if _, err := crand.Read(b[:]); err != nil {
		// 极少发生：降级到 math/rand
		return uint64(rand.Int63())
	}
	return binary.LittleEndian.Uint64(b[:])
}

type GameRoom struct {
	Room       *models.Room
	GameState  *models.GameState
	mutex      sync.RWMutex
	OfflineAt  map[int]time.Time // UID -> 离线起始时间
	StartTimer *time.Timer       // 游戏开始倒计时器
}

// BroadcastSystemMessage 广播一条系统级信息到房间聊天室
func (gr *GameRoom) BroadcastSystemMessage(msg string) {
	websocket.GlobalHub.BroadcastToRoom(gr.Room.ID, websocket.Message{
		Type:    "chat",
		UID:     0, // 0 表示系统
		Message: msg,
		Data: map[string]string{
			"username":  "系统",
			"nickname":  "量子记录仪",
			"avatar":    "🤖",
			"is_system": "true",
		},
	})
}

func (gr *GameRoom) cancelStartTimer() {
	if gr.StartTimer != nil {
		gr.StartTimer.Stop()
		gr.StartTimer = nil
	}
	gr.Room.Countdown = 0
}

func (gr *GameRoom) shouldTerminateRoom() bool {
	if gr.Room.Status != "playing" {
		return false
	}

	// 统计人类玩家数量（场上的）
	numHumans := 0
	for _, pid := range gr.Room.Players {
		if pid >= 0 {
			numHumans++
		}
	}

	// 即使场上没有人类玩家，只要有观战者，就不关闭房间
	// 这样允许人们观战 AI 之间的对局
	numSpectators := len(gr.Room.Spectators)

	if gr.Room.IsPvE {
		// 人机模式：场上没有人类且没有观战者，才关闭
		return numHumans == 0 && numSpectators == 0
	}

	// 多人模式：
	// 1. 如果完全没人类（场上+观战），关闭
	// 2. 如果人类离开了（中途退出），导致场上只剩1个或更少人（含AI），且没有任何人完成比赛，
	//    且没有观战者，关闭
	if numHumans == 0 && numSpectators == 0 {
		return true
	}

	// 如果人类玩家少于2个，且没有任何人（人类或AI）已经完成比赛
	// 且没有观战者正在观看
	if numHumans < 2 && numSpectators == 0 {
		if gr.GameState == nil || len(gr.GameState.FinishedPlayers) == 0 {
			return true
		}
	}
	return false
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
				// AI 玩家或已准备的玩家保留
				if uid < 0 || readyMap[uid] {
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
					gr.BroadcastSystemMessage(fmt.Sprintf("研究员 %d 因未及时佩戴防护装备（未准备）被自动请出实验室。", uid))
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

func emitPluginEvent(event string, payload map[string]interface{}) {
	cloned := make(map[string]interface{}, len(payload))
	for k, v := range payload {
		cloned[k] = v
	}
	go plugins.Emit(event, cloned)
}

func emitRoomCreated(room *models.Room, source string) {
	if room == nil {
		return
	}
	emitPluginEvent(plugins.EventRoomCreated, map[string]interface{}{
		"room_id":        room.ID,
		"name":           room.Name,
		"created_by_uid": room.CreatedByUID,
		"max_players":    room.MaxPlayers,
		"is_private":     room.IsPrivate,
		"is_pve":         room.IsPvE,
		"is_points_mode": room.IsPointsMode,
		"source":         source,
	})
}

func emitRoomClosed(roomID string, reason string, source string, remainingPlayers int) {
	emitPluginEvent(plugins.EventRoomClosed, map[string]interface{}{
		"room_id":            roomID,
		"reason":             reason,
		"source":             source,
		"remaining_players":  remainingPlayers,
		"closed_at_unix_mil": time.Now().UnixMilli(),
	})
}

func emitPlayerJoin(roomID string, uid int, username string, playerCount int, source string) {
	emitPluginEvent(plugins.EventPlayerJoin, map[string]interface{}{
		"room_id":      roomID,
		"uid":          uid,
		"username":     username,
		"player_count": playerCount,
		"source":       source,
	})
}

func emitPlayerLeave(roomID string, uid int, playerCount int, source string) {
	emitPluginEvent(plugins.EventPlayerLeave, map[string]interface{}{
		"room_id":      roomID,
		"uid":          uid,
		"player_count": playerCount,
		"source":       source,
	})
}

func emitGameStart(roomID string, state *models.GameState) {
	if state == nil || len(state.Players) == 0 {
		return
	}
	uids := make([]int, 0, len(state.Players))
	for _, player := range state.Players {
		uids = append(uids, player.UID)
	}
	currentIdx := state.CurrentPlayer
	if currentIdx < 0 || currentIdx >= len(state.Players) {
		currentIdx = 0
	}
	emitPluginEvent(plugins.EventGameStart, map[string]interface{}{
		"room_id":               roomID,
		"player_uids":           uids,
		"current_player_index":  currentIdx,
		"current_player_uid":    state.Players[currentIdx].UID,
		"original_player_count": state.OriginalPlayerCount,
		"is_tutorial":           state.TutorialScriptMode,
	})
}

func emitTurnChanged(roomID string, state *models.GameState, previousIndex int, source string) {
	if state == nil || state.Status != "playing" || len(state.Players) == 0 {
		return
	}
	currentIndex := state.CurrentPlayer
	if currentIndex < 0 || currentIndex >= len(state.Players) {
		return
	}
	if previousIndex == currentIndex {
		return
	}
	payload := map[string]interface{}{
		"room_id":              roomID,
		"source":               source,
		"current_player_index": currentIndex,
		"current_player_uid":   state.Players[currentIndex].UID,
		"turn_end_time":        state.TurnEndTime,
		"pending_draw_count":   state.PendingDrawCount,
		"allowed_any_player":   state.AllowedAnyPlayer,
		"pending_forced_plays": state.PendingForcedPlays,
	}
	if previousIndex >= 0 && previousIndex < len(state.Players) {
		payload["previous_player_index"] = previousIndex
		payload["previous_player_uid"] = state.Players[previousIndex].UID
	}
	emitPluginEvent(plugins.EventTurnChanged, payload)
}

func emitCardPlayed(roomID string, state *models.GameState, uid int, cardType string, substance string, effect string, isDouble bool) {
	payload := map[string]interface{}{
		"room_id":            roomID,
		"uid":                uid,
		"card_type":          cardType,
		"substance":          substance,
		"effect":             effect,
		"is_double":          isDouble,
		"played_at_unix_mil": time.Now().UnixMilli(),
	}
	if state != nil {
		payload["pending_draw_count"] = state.PendingDrawCount
		payload["discard_count"] = len(state.DiscardPile)
	}
	emitPluginEvent(plugins.EventCardPlayed, payload)
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

// GetUserRoomID 获取玩家所在的房间ID，如果不在任何房间则返回空字符串
func GetUserRoomID(uid int) string {
	roomMutex.RLock()
	defer roomMutex.RUnlock()

	for roomID, gr := range rooms {
		if gr.Room.Status == "playing" {
			for _, puid := range gr.Room.Players {
				if puid == uid {
					return roomID
				}
			}
		}
	}
	return ""
}

// 初始化默认牌组配置
func getDefaultDeckConfig() map[string]int {
	return BuiltinDeckDefaults()
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
	normalizedCards, _ := NormalizeBuiltinDeckCards(cards)
	if len(normalizedCards) == 0 {
		normalizedCards = getDefaultDeckConfig()
	}
	initialCards := deck.InitialCards
	if initialCards <= 0 {
		initialCards = 10
	}
	return normalizedCards, deck.Name, initialCards
}

// 创建房间
func CreateRoom(name string, creatorUID int, maxPlayers int, deckID int, isPointsMode bool, isPrivate bool) (*models.Room, error) {
	return CreateRoomWithKey(name, creatorUID, maxPlayers, deckID, isPointsMode, isPrivate, "", false, 0, 0, false, 0, false, 5, false)
}

// 创建房间（支持自定义访问密钥）
func CreateRoomWithKey(name string, creatorUID int, maxPlayers int, deckID int, isPointsMode bool, isPrivate bool, customKey string, isPvE bool, difficulty int, aiCount int, enableAIBackfill bool, aiBackfillDifficulty int, isRanked bool, levelRange int, tutorialScript bool) (*models.Room, error) {
	if isPointsMode && isPrivate && !isPvE {
		return nil, errors.New("积分模式下不可创建私密房间")
	}
	if isPvE {
		if difficulty < 1 || difficulty > 100 {
			return nil, errors.New("AI难度必须在1-100之间")
		}
		if aiCount < 1 || aiCount > 7 {
			return nil, errors.New("AI数量必须在1-7之间")
		}
		// PvE 模式下，MaxPlayers 由 AI 数量决定 (1位玩家 + N位AI)
		maxPlayers = 1 + aiCount
	}
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

	// 生成密码学安全的房间 ID（128 位随机，抵御穷举）
	roomID := fmt.Sprintf("room_%x%x", cryptoRandUint64(), cryptoRandUint64())

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
				normalizedCards, _ := NormalizeBuiltinDeckCards(cards)
				if len(normalizedCards) == 0 {
					cards, dname, initialCards := getGlobalDeckConfigFromDB()
					deckConfig.Cards = cards
					deckConfig.Name = dname
					deckConfig.InitialCards = initialCards
					deckConfig.IsGlobal = true
					deckConfig.ID = 1
				} else {
					deckConfig.Cards = normalizedCards
				}
			}
		}
	}

	// 生成私密房间访问密钥
	if deckConfig.InitialCards <= 0 {
		deckConfig.InitialCards = 10
	}

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

	// 计算等级匹配范围
	var minLevel, maxLevel int
	if !isPrivate && (isRanked || levelRange > 0) {
		// 获取房主等级
		user, err := repository.UserRepo.FindByUID(uint(creatorUID))
		if err == nil {
			minLevel = user.Level - levelRange
			maxLevel = user.Level + levelRange
			if minLevel < 1 {
				minLevel = 1
			}
			if maxLevel > 100 {
				maxLevel = 100
			}
		}
	}

	room := &models.Room{
		ID:                   roomID,
		Name:                 name,
		Players:              []int{creatorUID},
		ReadyUIDs:            []int{},
		Countdown:            0,
		Spectators:           []int{},
		MaxPlayers:           maxPlayers,
		DeckConfig:           &deckConfig,
		Status:               "waiting",
		IsPointsMode:         isPointsMode,
		IsPrivate:            isPrivate,
		AccessKey:            accessKey,
		IsPvE:                isPvE,
		PvEDifficulty:        difficulty,
		AICount:              aiCount,
		EnableAIBackfill:     enableAIBackfill,
		AIBackfillDifficulty: aiBackfillDifficulty,
		BackfilledAIUIDs:     []int{},
		IsRanked:             isRanked,
		LevelRange:           levelRange,
		MinLevel:             minLevel,
		MaxLevel:             maxLevel,
		CreatedByUID:         creatorUID,
		CreatedAt:            time.Now(),
		TutorialScript:       tutorialScript,
	}

	gameRoom := &GameRoom{
		Room:      room,
		OfflineAt: make(map[int]time.Time),
	}

	// 如果是 PvE 模式，立即初始化 AI 玩家
	if isPvE {
		// AI 玩家 UID 使用负数: -1, -2, -3...
		for i := 1; i <= aiCount; i++ {
			aiUID := -i
			room.Players = append(room.Players, aiUID)
			room.ReadyUIDs = append(room.ReadyUIDs, aiUID) // AI 默认已准备
		}
		// PvE 模式下，人类玩家也自动准备（无需等待）
		room.ReadyUIDs = append(room.ReadyUIDs, creatorUID)
	}

	roomMutex.Lock()
	rooms[roomID] = gameRoom
	roomMutex.Unlock()

	emitRoomCreated(room, "create_room")
	return room, nil
}

// generateBackfillAIUID 为补位AI生成唯一的负数UID
// 扫描房间中现有的所有玩家UID，找到最小的负数UID并递减1
func generateBackfillAIUID(room *models.Room) int {
	minUID := 0 // 从 0 开始，确保任何已存在的负数 UID（含 -1 的 PvE AI）都能被正确检测

	// 检查现有玩家中的AI UID
	for _, uid := range room.Players {
		if uid < minUID {
			minUID = uid
		}
	}

	// 检查已记录的补位AI UID
	for _, uid := range room.BackfilledAIUIDs {
		if uid < minUID {
			minUID = uid
		}
	}

	return minUID - 1
}

// addBackfillAI 向房间添加指定数量的补位AI
// count: 需要添加的AI数量
func addBackfillAI(room *models.Room, count int) {
	for i := 0; i < count; i++ {
		// 生成唯一的AI UID
		aiUID := generateBackfillAIUID(room)

		// 添加到玩家列表
		room.Players = append(room.Players, aiUID)

		// AI默认处于准备状态
		room.ReadyUIDs = append(room.ReadyUIDs, aiUID)

		// 记录为补位AI
		room.BackfilledAIUIDs = append(room.BackfilledAIUIDs, aiUID)
	}
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

	roomID := fmt.Sprintf("duel_%x%x", cryptoRandUint64(), cryptoRandUint64())
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

	emitRoomCreated(room, "start_duel")
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

// GetAllRoomsAdmin 获取所有活跃房间（含私密房间，管理员专用）
func GetAllRoomsAdmin() []*models.Room {
	roomMutex.RLock()
	defer roomMutex.RUnlock()

	result := []*models.Room{}
	for _, gr := range rooms {
		if gr.Room.Status != "finished" {
			result = append(result, gr.Room)
		}
	}

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

// BroadcastRoomsUpdate 广播房间列表更新到所有在线玩家（实时推送）
func BroadcastRoomsUpdate() {
	ws := websocket.GlobalHub
	if ws != nil {
		message := websocket.Message{
			Type: "rooms_update",
			Data: GetAllRooms(0), // 获取所有公开房间
		}
		ws.BroadcastToAll(message)
	}
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

	// 确保所有玩家都在排名列表中（即使没打完）
	allUIDs := []int{}
	for _, p := range gr.GameState.Players {
		allUIDs = append(allUIDs, p.UID)
	}

	fullRanking := append([]int{}, finished...)
	for _, uid := range allUIDs {
		exists := false
		for _, fuid := range finished {
			if fuid == uid {
				exists = true
				break
			}
		}
		if !exists {
			fullRanking = append(fullRanking, uid)
		}
	}

	// 1. 计算总人数（包括已退出的玩家，反映原本的对局规模）
	totalPlayers := gr.GameState.OriginalPlayerCount
	if totalPlayers <= 0 {
		totalPlayers = len(fullRanking)
	}

	// 2. 如果 PointsChanges 还没初始化，初始化它
	if gr.GameState.PointsChanges == nil {
		gr.GameState.PointsChanges = make(map[int]int)
	}
	if gr.GameState.XPChanges == nil {
		gr.GameState.XPChanges = make(map[int]int)
	}

	for i, uid := range fullRanking {
		rank := i + 1

		// 3. 处理积分
		points := 0

		// AI 玩家（UID < 0）和已退出玩家不获得额外积分
		if uid > 0 {
			// 如果已经在中间实时结算过了，不再重复增加数据库积分，仅确保 PointsChanges 中有值
			if prevPoints, ok := gr.GameState.PointsChanges[uid]; ok && prevPoints > 0 {
				points = prevPoints
			} else {
				// 4. 计算应得积分
				if gr.Room.IsPvE {
					// 仅在 AI 模式（PvE）下使用动态人数加成规则
					baseBonus := (totalPlayers - rank + 1) * 10
					if rank == 1 {
						baseBonus += totalPlayers * 25 // 获胜者额外加成
					}
					points = baseBonus

					// PvE 模式积分难度系数修正
					if gr.Room.PvEDifficulty < 50 {
						points = 0
					} else {
						points = int(float64(points) * float64(gr.Room.PvEDifficulty) / 100.0)
					}
				} else {
					// PvP 模式（常规对局）
					// 基础规则：100 / 排名
					points = 100 / rank

					// 如果是最后一名，只获得 5 分参与奖
					if rank == len(fullRanking) && len(fullRanking) > 1 {
						points = 5
					}

					// 结算倍率（只有常规对局启用）：每有一个玩家离开，结算减少 1/总人数
					multiplier := 1.0
					if configRepo.GetBoolValue("points_scaling_enabled", true) && totalPlayers > 0 {
						multiplier = 1.0 - (float64(gr.GameState.QuittedCount) / float64(totalPlayers))
						if multiplier < 0 {
							multiplier = 0
						}
					}
					points = int(float64(points) * multiplier)
				}

				if points < 1 && (!gr.Room.IsPvE || gr.Room.PvEDifficulty >= 50) {
					points = 1
				}

				// 仅对尚未结算过的玩家执行数据库增加
				if points > 0 {
					repository.UserRepo.IncrementPoints(uint(uid), points)
					repository.UserRepo.IncrementMonthlyPoints(uint(uid), points)
				}
			}
		}

		gr.GameState.PointsChanges[uid] = points

		// 2. 处理 XP
		if uid > 0 {
			// 如果 XPChanges 还没算过（通常最后结算时大家都没算过）
			if _, ok := gr.GameState.XPChanges[uid]; !ok {
				xp := CalculateXPReward(gr, uid, rank)
				gr.GameState.XPChanges[uid] = xp
				if xp > 0 {
					go AwardXP(uid, xp)
				}
			}
		} else {
			gr.GameState.XPChanges[uid] = 0
		}
	}

	// 悬赏逻辑处理（仅奖励第一名）
	if len(finished) > 0 {
		winnerUID := finished[0]
		totalBountyForWinner := 0
		for _, targetUID := range allUIDs {
			bounties, err := repository.BountyRepo.FindActiveByTarget(uint(targetUID))
			if err != nil {
				continue
			}
			for _, bounty := range bounties {
				if gr.Room.IsDuel && targetUID == gr.Room.TargetUID {
					if winnerUID == gr.Room.ChallengerUID {
						totalBountyForWinner += bounty.Amount
						repository.BountyRepo.UpdateStatus(bounty.ID, "claimed")
					} else if winnerUID == gr.Room.TargetUID {
						reward := bounty.Amount / 2
						totalBountyForWinner += reward
						repository.BountyRepo.UpdateStatus(bounty.ID, "claimed")
					}
				} else {
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
			gr.GameState.PointsChanges[winnerUID] += totalBountyForWinner
		}
	}
}

// handleXPCalculation 处理经验值计算（用于非积分模式）
func handleXPCalculation(gr *GameRoom) {
	finished := gr.GameState.FinishedPlayers

	// 确保所有玩家都在排名列表中（即使没打完）
	allUIDs := []int{}
	for _, p := range gr.GameState.Players {
		allUIDs = append(allUIDs, p.UID)
	}

	fullRanking := append([]int{}, finished...)
	for _, uid := range allUIDs {
		exists := false
		for _, fuid := range finished {
			if fuid == uid {
				exists = true
				break
			}
		}
		if !exists {
			fullRanking = append(fullRanking, uid)
		}
	}

	// 初始化 XPChanges 和 PointsChanges map
	if gr.GameState.XPChanges == nil {
		gr.GameState.XPChanges = make(map[int]int)
	}
	if gr.GameState.PointsChanges == nil {
		gr.GameState.PointsChanges = make(map[int]int)
	}

	// 计算并存储每个玩家的 XP 变化（非积分模式积分为0，仅用于前端显示排名）
	for i, uid := range fullRanking {
		gr.GameState.PointsChanges[uid] = 0 // 非积分模式不发放积分
		if uid < 0 {
			gr.GameState.XPChanges[uid] = 0
			continue
		}
		rank := i + 1
		xp := CalculateXPReward(gr, uid, rank)
		gr.GameState.XPChanges[uid] = xp

		if xp > 0 {
			go AwardXP(uid, xp) // 异步授予经验，避免阻塞
		}
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
		if uid < 0 { // AI 玩家永远视为在线
			continue
		}
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
	} else if gr.shouldTerminateRoom() {
		gr.mutex.Unlock()
		gr.terminateRoom("由于玩家人数不足，房间已被关闭")
		return
	}
	gr.mutex.Unlock()
}

// CheckNextTurnAI 检查并触发 AI 回合
func (gr *GameRoom) CheckNextTurnAI() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[AI] ❌ CheckNextTurnAI panic recovered: %v", r)
		}
	}()

	gr.mutex.RLock()
	// 如果房间已结束或未开始，忽略
	if gr.Room.Status != "playing" || gr.GameState == nil {
		gr.mutex.RUnlock()
		return
	}

	currentPlayerIdx := gr.GameState.CurrentPlayer
	if currentPlayerIdx < 0 || currentPlayerIdx >= len(gr.GameState.Players) {
		gr.mutex.RUnlock()
		return
	}

	currentPlayer := gr.GameState.Players[currentPlayerIdx]
	shouldTrigger := currentPlayer.UID < 0
	gr.mutex.RUnlock()

	// 如果是 AI，触发思考逻辑
	if shouldTrigger {
		gr.TriggerAITurn()
	}
}

func (gr *GameRoom) kickPlayer(uid int, reason string) {
	if uid < 0 {
		return // 不能踢出 AI 玩家
	}
	gr.mutex.Lock()
	roomID := gr.Room.ID

	// 通知被踢出的玩家
	if websocket.GlobalHub != nil {
		websocket.GlobalHub.SendToUID(uid, websocket.Message{
			Type:    "player_kicked",
			Message: reason,
		})
	}

	gr.BroadcastSystemMessage(fmt.Sprintf("警报：研究员 %d 因违反实验室安全准则被强制带离。", uid))

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
	if gr.shouldTerminateRoom() {
		gr.mutex.Unlock()
		gr.terminateRoom("由于实验样本不足（少于2人），本次反应宣告失败，实验室已关闭")
		return
	}

	remainingPlayers := len(gr.Room.Players)
	gr.mutex.Unlock()
	emitPlayerLeave(roomID, uid, remainingPlayers, "kick_player")

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

	// 确保状态更新，防止 AI 继续行动
	gr.mutex.Lock()
	if gr.GameState != nil {
		gr.GameState.Status = "terminated"
	}
	gr.Room.Status = "terminated"
	gr.mutex.Unlock()

	emitRoomClosed(roomID, reason, "terminate_room", len(gr.Room.Players))
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
		if uid > 0 {
			repository.UserRepo.UpdateRoomReadyStatus(uint(uid), false)
		}
	} else {
		// 准备
		gr.Room.ReadyUIDs = append(gr.Room.ReadyUIDs, uid)
		if uid > 0 {
			repository.UserRepo.UpdateRoomReadyStatus(uint(uid), true)
		}
	}

	gr.checkAutoStart()
	gr.broadcastRoomUpdate()
	return nil
}

// 加入房间
func JoinRoom(roomID string, uid int) error {
	return JoinRoomWithKey(roomID, uid, "")
}

// JoinRoomWithKey 携带密钥加入房间
func JoinRoomWithKey(roomID string, uid int, accessKey string) error {
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

	// 等级验证（私密房间豁免）
	if !gameRoom.Room.IsPrivate && (gameRoom.Room.IsRanked || gameRoom.Room.LevelRange > 0) && gameRoom.Room.MinLevel > 0 {
		user, err := repository.UserRepo.FindByUID(uint(uid))
		if err != nil {
			return errors.New("无法获取玩家等级信息")
		}

		if user.Level < gameRoom.Room.MinLevel || user.Level > gameRoom.Room.MaxLevel {
			return fmt.Errorf("等级不符合要求（需要等级 %d-%d，当前等级 %d）",
				gameRoom.Room.MinLevel, gameRoom.Room.MaxLevel, user.Level)
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
			ensurePvEReady := gameRoom.Room.IsPvE && gameRoom.Room.Status == "waiting"
			if ensurePvEReady {
				isReady := false
				for _, readyUID := range gameRoom.Room.ReadyUIDs {
					if readyUID == uid {
						isReady = true
						break
					}
				}
				if !isReady {
					gameRoom.Room.ReadyUIDs = append(gameRoom.Room.ReadyUIDs, uid)
				}
				if uid > 0 {
					repository.UserRepo.UpdateRoomReadyStatus(uint(uid), true)
				}
			}

			// 离线重连或 PvE 重进等待房时都需要重新检查自动开局
			if wasOffline || ensurePvEReady {
				gameRoom.checkAutoStart()
				gameRoom.broadcastRoomUpdate()
				if wasOffline {
					emitPlayerJoin(roomID, uid, "", len(gameRoom.Room.Players), "reconnect")
				}
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
		// 如果是公开房间，无需密钥即可观战
		if !gameRoom.Room.IsPrivate {
			accessKey = ""
		}

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

		// 禁止 AI 玩家观战
		if uid < 0 {
			return errors.New("AI 玩家无法观战")
		}

		// 检查是否已在观战者列表中
		for _, sid := range gameRoom.Room.Spectators {
			if sid == uid {
				return errors.New("已在观战者列表中")
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
	if gameRoom.Room.IsPvE {
		isReady := false
		for _, readyUID := range gameRoom.Room.ReadyUIDs {
			if readyUID == uid {
				isReady = true
				break
			}
		}
		if !isReady {
			gameRoom.Room.ReadyUIDs = append(gameRoom.Room.ReadyUIDs, uid)
		}
		if uid > 0 {
			repository.UserRepo.UpdateRoomReadyStatus(uint(uid), true)
		}
	} else if uid > 0 {
		repository.UserRepo.UpdateRoomReadyStatus(uint(uid), false)
	}
	gameRoom.checkAutoStart()
	gameRoom.broadcastRoomUpdate()
	emitPlayerJoin(roomID, uid, "", len(gameRoom.Room.Players), "join_room")
	return nil
}

// JoinRoomWithKeyAsSpectator 以观战者身份加入房间
func JoinRoomWithKeyAsSpectator(roomID string, uid int, accessKey string, asSpectator bool) error {
	// 如果不要求作为观战者加入，使用原有逻辑
	if !asSpectator {
		return JoinRoomWithKey(roomID, uid, accessKey)
	}

	// 以下是观战模式的逻辑
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

	// 检查用户是否已在房间中（玩家或观战者）
	for _, pid := range gameRoom.Room.Players {
		if pid == uid {
			return nil
		}
	}
	for _, sid := range gameRoom.Room.Spectators {
		if sid == uid {
			return nil
		}
	}

	// 私密房间需要验证密钥
	if gameRoom.Room.IsPrivate && gameRoom.Room.AccessKey != "" {
		isCreator := len(gameRoom.Room.Players) > 0 && gameRoom.Room.Players[0] == uid
		alreadyInRoom := false
		for _, pid := range gameRoom.Room.Players {
			if pid == uid {
				alreadyInRoom = true
				break
			}
		}
		// 观战者需要验证密钥（除非是房主或已在房间）
		if !isCreator && !alreadyInRoom && accessKey != gameRoom.Room.AccessKey {
			return errors.New("访问密钥错误，无法观战私密房间")
		}
	}

	// 禁止 AI 玩家观战
	if uid < 0 {
		return errors.New("AI 玩家无法观战")
	}

	// 添加为观战者
	gameRoom.Room.Spectators = append(gameRoom.Room.Spectators, uid)
	if gameRoom.GameState != nil {
		gameRoom.GameState.Spectators = append(gameRoom.GameState.Spectators, uid)
	}
	gameRoom.broadcastRoomUpdate()
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
	removedPlayer := false
	for _, pid := range gameRoom.Room.Players {
		if pid != uid {
			newPlayers = append(newPlayers, pid)
		} else {
			removedPlayer = true
		}
	}
	gameRoom.Room.Players = newPlayers

	// 清除离线记录，防止旧的OfflineAt条目影响后续重连
	delete(gameRoom.OfflineAt, uid)

	// 移除准备状态
	newReady := []int{}
	for _, rid := range gameRoom.Room.ReadyUIDs {
		if rid != uid {
			newReady = append(newReady, rid)
		}
	}
	gameRoom.Room.ReadyUIDs = newReady
	if uid > 0 {
		repository.UserRepo.UpdateRoomReadyStatus(uint(uid), false)
	}

	// 移除观战者
	newSpectators := []int{}
	for _, sid := range gameRoom.Room.Spectators {
		if sid != uid {
			newSpectators = append(newSpectators, sid)
		}
	}
	gameRoom.Room.Spectators = newSpectators

	// 同时从 GameState.Spectators 中移除观战者
	if gameRoom.GameState != nil {
		newGameStateSpectators := []int{}
		for _, sid := range gameRoom.GameState.Spectators {
			if sid != uid {
				newGameStateSpectators = append(newGameStateSpectators, sid)
			}
		}
		gameRoom.GameState.Spectators = newGameStateSpectators
	}

	// 如果游戏还未开始，尝试将观战者提升为玩家
	if gameRoom.Room.Status == "waiting" && len(gameRoom.Room.Spectators) > 0 && len(gameRoom.Room.Players) < gameRoom.Room.MaxPlayers {
		promotedUID := gameRoom.Room.Spectators[0]
		// 检查观战者是否已在Players中（防止重复）
		alreadyPlayer := false
		for _, pid := range gameRoom.Room.Players {
			if pid == promotedUID {
				alreadyPlayer = true
				break
			}
		}
		if !alreadyPlayer {
			gameRoom.Room.Spectators = gameRoom.Room.Spectators[1:]
			gameRoom.Room.Players = append(gameRoom.Room.Players, promotedUID)
			// 同时从GameState.Spectators中移除升级的观战者
			if gameRoom.GameState != nil {
				newGameStateSpectators := []int{}
				for _, sid := range gameRoom.GameState.Spectators {
					if sid != promotedUID {
						newGameStateSpectators = append(newGameStateSpectators, sid)
					}
				}
				gameRoom.GameState.Spectators = newGameStateSpectators
			}
			log.Printf("[房间] 玩家 %d 离开，观战者 %d 自动晋升为玩家", uid, promotedUID)
			gameRoom.BroadcastSystemMessage(fmt.Sprintf("观战员已自动加入游戏，现在房间有 %d/%d 位研究员", len(gameRoom.Room.Players), gameRoom.Room.MaxPlayers))
		}
	}

	// 检查自动开始状态
	if gameRoom.Room.Status == "waiting" {
		gameRoom.checkAutoStart()
	}

	// 如果游戏正在进行，也从 GameState 中移除
	if gameRoom.GameState != nil {
		isFinished := false
		for _, fuid := range gameRoom.GameState.FinishedPlayers {
			if fuid == uid {
				isFinished = true
				break
			}
		}

		newPS := []*models.PlayerState{}
		leftIndex := -1
		for i, ps := range gameRoom.GameState.Players {
			if ps.UID != uid {
				newPS = append(newPS, ps)
			} else {
				leftIndex = i
			}
		}

		// 只有未完成比赛的玩家离开才算作中途退出
		if !isFinished && leftIndex != -1 {
			gameRoom.GameState.QuittedCount++
		}

		gameRoom.GameState.Players = newPS

		// 调整当前玩家索引
		if leftIndex != -1 && len(gameRoom.GameState.Players) > 0 {
			if gameRoom.GameState.CurrentPlayer > leftIndex {
				gameRoom.GameState.CurrentPlayer--
			}
			// 使用模运算确保CurrentPlayer始终在有效范围内
			gameRoom.GameState.CurrentPlayer = gameRoom.GameState.CurrentPlayer % len(gameRoom.GameState.Players)
		} else if len(gameRoom.GameState.Players) == 0 {
			gameRoom.GameState.CurrentPlayer = 0
		}
	}

	// GameState.Spectators已在上面的代码中移除，这里做额外安全检查
	if gameRoom.GameState != nil && len(gameRoom.GameState.Spectators) > 0 {
		// 验证一致性：如果用户仍在GameState观战者中，移除（防御性编程）
		newGameStateSpectators := []int{}
		for _, sid := range gameRoom.GameState.Spectators {
			if sid != uid {
				newGameStateSpectators = append(newGameStateSpectators, sid)
			}
		}
		if len(newGameStateSpectators) < len(gameRoom.GameState.Spectators) {
			gameRoom.GameState.Spectators = newGameStateSpectators
		}
	}

	// 从 WebSocket Hub 中移除该用户的房间订阅
	if websocket.GlobalHub != nil {
		websocket.GlobalHub.LeaveRoomByUID(uid)
	}
	if removedPlayer {
		emitPlayerLeave(roomID, uid, len(gameRoom.Room.Players), "leave_room")
	}

	// 如果是正在游戏中且玩家少于2个，则解散
	if gameRoom.shouldTerminateRoom() {
		gameRoom.terminateRoom("由于实验样本不足（少于2人），本次反应宣告失败，实验室已关闭")
		return nil
	}

	// 检查游戏是否应该自动结算
	// 当活跃玩家数 <= 1 时，自动结算游戏
	if gameRoom.GameState != nil && gameRoom.GameState.Status == "playing" {
		activeCount := 0
		lastPlayerUID := 0
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

		// 如果只剩 ≤1 个活跃玩家，自动结算游戏
		if activeCount <= 1 {
			if activeCount == 1 {
				gameRoom.GameState.FinishedPlayers = append(gameRoom.GameState.FinishedPlayers, lastPlayerUID)
			}
			finalizeGame(gameRoom)
			return nil
		}
	}

	// 如果所有玩家和观战者均已离开，销毁房间
	if len(gameRoom.Room.Players) == 0 && len(gameRoom.Room.Spectators) == 0 {
		gameRoom.cancelStartTimer()
		emitRoomClosed(roomID, "房间无人，自动关闭", "empty_room", 0)
		roomMutex.Lock()
		delete(rooms, roomID)
		roomMutex.Unlock()
		log.Printf("房间 %s 已空，已自动关闭并清理资源", roomID)
	} else {
		// 检查下一位是否是 AI
		go gameRoom.CheckNextTurnAI()
		// 广播更新
		gameRoom.broadcastRoomUpdate()
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

	gameRoom.BroadcastSystemMessage("正在校准对撞机... 游戏即将开始！")

	if gameRoom.Room.Status != "waiting" {
		return errors.New("游戏已在进行中")
	}

	if len(gameRoom.Room.Players) < 2 {
		return errors.New("至少需要2名玩家")
	}

	// 刷新插件卡牌 registry，确保使用最新数据
	LoadPluginCards()

	// ===== AI补位逻辑 =====
	if gameRoom.Room.EnableAIBackfill && !gameRoom.Room.IsPvE {
		currentPlayerCount := len(gameRoom.Room.Players)
		maxPlayers := gameRoom.Room.MaxPlayers

		if currentPlayerCount < maxPlayers {
			vacancies := maxPlayers - currentPlayerCount
			log.Printf("[AI补位] 房间 %s 有 %d 个空位，添加AI玩家...", roomID, vacancies)

			addBackfillAI(gameRoom.Room, vacancies)

			log.Printf("[AI补位] 成功添加 %d 个AI玩家，当前玩家列表: %v", vacancies, gameRoom.Room.Players)
		} else {
			log.Printf("[AI补位] 房间 %s 已满员，无需补位", roomID)
		}
	}
	// ===== AI补位逻辑结束 =====

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
		TurnEndTime:         0,
		PendingDrawCount:    0,
		PendingDrawTypes:    nil,
		AllowedAnyPlayer:    -1,
		TutorialScriptMode:  gameRoom.Room.TutorialScript,
		TutorialCurrentStep: 1, // 从第一步开始
	}
	setTurnEndTimeByMode(gameRoom.GameState)

	// 脚本化教学模式：使用固定配置
	if gameRoom.Room.TutorialScript {
		gameRoom.BroadcastSystemMessage("实验脚本已加载。正在按照教学规程引导研究员进行初次反应。")
		log.Printf("[教学脚本] 启用脚本化教学模式，使用固定手牌和初始配置")
		if err := initTutorialGame(gameRoom, roomID); err != nil {
			return err
		}
		emitGameStart(roomID, gameRoom.GameState)
		emitTurnChanged(roomID, gameRoom.GameState, -1, "game_start")
		return nil
	}

	gameRoom.BroadcastSystemMessage("对撞机已启动！实验开始。")

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
	// 插件牌由插件系统管理，不进入全局卡组配置；开局按插件默认数量自动注入。
	pluginCardCount := make(map[string]int)
	for _, pluginCard := range GetAllPluginCards() {
		if pluginCard == nil || pluginCard.DefaultCount <= 0 {
			continue
		}
		for i := 0; i < pluginCard.DefaultCount; i++ {
			gameRoom.GameState.DrawPile = append(gameRoom.GameState.DrawPile, models.Card{
				Type:  pluginCard.Symbol,
				Count: 1,
			})
		}
		pluginCardCount[pluginCard.Symbol] += pluginCard.DefaultCount
	}
	if len(pluginCardCount) > 0 {
		log.Printf("[插件牌] 已注入插件牌到牌堆: %v", pluginCardCount)
	}

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
		initialCardsCount = len(gameRoom.GameState.DrawPile) / numPlayers
	}

	// 手牌分配准备完成
	gameRoom.BroadcastSystemMessage(fmt.Sprintf("正在分发实验素材... 每位研究员已领取 %d 份基础试剂。", initialCardsCount))

	// 牌堆已准备好

	// 批量预加载所有人类玩家信息，避免 N+1 查询
	humanUIDs := make([]uint, 0)
	for _, pid := range shuffledPlayers {
		if pid >= 0 {
			humanUIDs = append(humanUIDs, uint(pid))
		}
	}
	userMap, err := repository.UserRepo.FindByUIDs(humanUIDs)
	if err != nil {
		log.Printf("[StartGame] ⚠️  批量加载玩家信息失败: %v，降级为逐个查询", err)
		userMap = map[uint]*database.User{}
	}

	// 初始化玩家
	usedScientistNames := make(map[string]bool)
	shuffledScientistNames := make([]string, len(scientistNames))
	copy(shuffledScientistNames, scientistNames)
	rand.Shuffle(len(shuffledScientistNames), func(i, j int) {
		shuffledScientistNames[i], shuffledScientistNames[j] = shuffledScientistNames[j], shuffledScientistNames[i]
	})
	nextScientistIdx := 0

	for _, pid := range shuffledPlayers {
		username := ""
		nickname := ""
		avatar := ""
		if pid < 0 {
			// AI 玩家 - 从姓名库中随机挑选科学家姓名，确保不重复
			sName := fmt.Sprintf("AI_%d", -pid)
			for nextScientistIdx < len(shuffledScientistNames) {
				potentialName := shuffledScientistNames[nextScientistIdx]
				nextScientistIdx++
				if !usedScientistNames[potentialName] {
					sName = potentialName
					usedScientistNames[sName] = true
					break
				}
			}
			username = fmt.Sprintf("AI_%d", -pid)
			nickname = sName
			avatar = "🤖"
		} else {
			user := userMap[uint(pid)]
			if user == nil {
				username = fmt.Sprintf("研究员_%d", pid)
				nickname = username
				avatar = "🧪"
			} else {
				username = user.Username
				nickname = user.Nickname
				if nickname == "" {
					nickname = username
				}
				avatar = user.Avatar
			}
			// 记录人类玩家已使用的昵称，防止 AI 撞名（虽然概率极低，但仍需处理）
			usedScientistNames[nickname] = true
		}

		player := &models.PlayerState{
			UID:                   pid,
			Username:              username,
			Nickname:              nickname,
			Avatar:                avatar,
			HandCards:             []models.Card{},
			CardCount:             0,
			IsReady:               true,
			DoubleActionAvailable: false,
			ActionProgress:        0,
			IsAI:                  pid < 0,
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

		// 玩家初始手牌已分配

		gameRoom.GameState.Players = append(gameRoom.GameState.Players, player)
	}

	// 抽出场上初始物质，从 reactions 表的 r1/r2 字段中随机选择
	var initialSubstance string
	var initialCard models.Card
	var foundBase bool

	// 从 reactions 表获取所有已批准物质（仅读 r1/r2 列，避免全表加载）
	reactionRepo := repository.NewReactionRepository()
	availableSubstances, err := reactionRepo.FindDistinctSubstances()

	if err == nil && len(availableSubstances) > 0 {
		// 随机选择一个物质作为场上初始物质
		if len(availableSubstances) > 0 {
			randomIndex := rand.Intn(len(availableSubstances))
			initialSubstance = availableSubstances[randomIndex]

			log.Printf("[场上初始物质] 随机选择: %s", initialSubstance)

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
				if IsPluginCard(card.Type) {
					isSpecial = true
				}

				// 优先选择与初始物质匹配的普通卡牌
				if !isSpecial && card.Type == initialSubstance {
					initialCard = card
					// 从牌堆移除该牌
					gameRoom.GameState.DrawPile = append(gameRoom.GameState.DrawPile[:i], gameRoom.GameState.DrawPile[i+1:]...)
					foundBase = true
					log.Printf("[场上初始物质] 找到匹配的卡牌: %s", card.Type)
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
					if IsPluginCard(card.Type) {
						isSpecial = true
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
		gameRoom.BroadcastSystemMessage(fmt.Sprintf("对撞机底座已安装完成，观测到初始物质：[ %s ]。", initialSubstance))
		log.Printf("[场上初始卡牌] 卡牌类型=%s, 展示物质=%s", initialCard.Type, initialSubstance)
	} else {
		// 回退方案：如果 reactions 为空或没找到合适的底牌，从牌堆抽第一张非功能牌
		log.Println("[场上初始物质] 使用备用方案")
		maxAttempts := len(gameRoom.GameState.DrawPile)
		for attempts := 0; attempts < maxAttempts && len(gameRoom.GameState.DrawPile) > 0; attempts++ {
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
			if IsPluginCard(firstCard.Type) {
				isSpecial = true
			}

			if !isSpecial {
				playedCard := models.PlayedCard{
					Card:      firstCard,
					Substance: firstCard.Type,
					PlayerUID: 0, // 表示系统出牌
				}
				gameRoom.GameState.DiscardPile = append(gameRoom.GameState.DiscardPile, playedCard)
				gameRoom.GameState.LastCard = &gameRoom.GameState.DiscardPile[0]
				log.Printf("[场上初始物质] 备用方案：使用卡牌 %s", firstCard.Type)
				break
			} else {
				gameRoom.GameState.DrawPile = append(gameRoom.GameState.DrawPile, firstCard)
			}
		}
		if gameRoom.GameState.LastCard == nil {
			log.Printf("[场上初始物质] ⚠️  牌堆中未找到可作为底牌的普通牌，跳过初始底牌设置（draw_pile=%d）", len(gameRoom.GameState.DrawPile))
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

	// 检查第一位是否是 AI
	emitGameStart(roomID, gameRoom.GameState)
	emitTurnChanged(roomID, gameRoom.GameState, -1, "game_start")
	go gameRoom.CheckNextTurnAI()

	return nil
}

// initTutorialGame 初始化教学脚本游戏
func initTutorialGame(gameRoom *GameRoom, roomID string) error {
	log.Printf("[教学脚本] 开始初始化教学关卡...")

	// 固定手牌配置
	humanHand := []string{"Mg", "Na", "O", "H", "Ar", "Au", "+2"}
	aiHand := []string{"H", "Cl", "Br", "O"}
	initialDiscard := "Cl2"

	// 确定玩家顺序（人类玩家先手）
	var humanUID, aiUID int
	for _, pid := range gameRoom.Room.Players {
		if pid >= 0 {
			humanUID = pid
		} else {
			aiUID = pid
		}
	}

	// 准备AI昵称
	aiNames := []string{"AI"}
	aiName := aiNames[rand.Intn(len(aiNames))]

	// 创建人类玩家状态
	humanUser, err := repository.UserRepo.FindByUID(uint(humanUID))
	humanUsername := fmt.Sprintf("Player_%d", humanUID)
	humanNickname := humanUsername
	humanAvatar := "🧪"
	if err == nil {
		humanUsername = humanUser.Username
		humanNickname = humanUser.Nickname
		if humanNickname == "" {
			humanNickname = humanUsername
		}
		humanAvatar = humanUser.Avatar
	}

	humanPlayer := &models.PlayerState{
		UID:                   humanUID,
		Username:              humanUsername,
		Nickname:              humanNickname,
		Avatar:                humanAvatar,
		HandCards:             []models.Card{},
		CardCount:             len(humanHand),
		IsReady:               true,
		DoubleActionAvailable: false,
		ActionProgress:        0,
		IsAI:                  false,
	}

	// 添加人类玩家手牌
	for _, cardType := range humanHand {
		effect := getCardEffect(cardType)
		humanPlayer.HandCards = append(humanPlayer.HandCards, models.Card{
			Type:   cardType,
			Count:  1,
			Effect: effect,
		})
	}

	// 创建AI玩家状态
	aiPlayer := &models.PlayerState{
		UID:                   aiUID,
		Username:              fmt.Sprintf("AI_%d", -aiUID),
		Nickname:              aiName,
		Avatar:                "🤖",
		HandCards:             []models.Card{},
		CardCount:             len(aiHand),
		IsReady:               true,
		DoubleActionAvailable: false,
		ActionProgress:        0,
		IsAI:                  true,
	}

	// 添加AI手牌
	for _, cardType := range aiHand {
		effect := getCardEffect(cardType)
		aiPlayer.HandCards = append(aiPlayer.HandCards, models.Card{
			Type:   cardType,
			Count:  1,
			Effect: effect,
		})
	}

	// 设置玩家顺序（人类先手）
	gameRoom.GameState.Players = []*models.PlayerState{humanPlayer, aiPlayer}

	// 设置初始场上牌
	initialCard := models.Card{
		Type:   initialDiscard,
		Count:  1,
		Effect: "",
	}
	playedCard := models.PlayedCard{
		Card:      initialCard,
		Substance: initialDiscard,
		PlayerUID: 0, // 系统出牌
	}
	gameRoom.GameState.DiscardPile = append(gameRoom.GameState.DiscardPile, playedCard)
	gameRoom.GameState.LastCard = &gameRoom.GameState.DiscardPile[0]

	// 创建基础牌堆（用于摸牌，包含常见物质）
	basicDeck := []string{
		"H2O", "CO2", "NaCl", "H2SO4", "HNO3", "NH3", "CH4",
		"O2", "N2", "H2", "Cl2", "Br2", "I2",
		"Na", "K", "Ca", "Mg", "Al", "Fe", "Cu", "Zn", "Ag", "Au",
		"NaOH", "KOH", "Ca(OH)2", "HCl", "HBr", "HI",
		"+2", "+2", "reverse",
	}
	for _, cardType := range basicDeck {
		effect := getCardEffect(cardType)
		gameRoom.GameState.DrawPile = append(gameRoom.GameState.DrawPile, models.Card{
			Type:   cardType,
			Count:  1,
			Effect: effect,
		})
	}

	// 洗牌
	rand.Shuffle(len(gameRoom.GameState.DrawPile), func(i, j int) {
		gameRoom.GameState.DrawPile[i], gameRoom.GameState.DrawPile[j] =
			gameRoom.GameState.DrawPile[j], gameRoom.GameState.DrawPile[i]
	})

	// 设置房间状态为进行中
	gameRoom.Room.Status = "playing"

	// 记录人类玩家的回合开始时间
	repository.UserRepo.UpdateTurnStartedAt(uint(humanUID), time.Now())

	log.Printf("[教学脚本] ✅ 教学关卡初始化完成")
	log.Printf("[教学脚本] 👤 人类玩家: %s (UID:%d), 手牌: %v", humanUsername, humanUID, humanHand)
	log.Printf("[教学脚本] 🤖 AI玩家: %s (UID:%d), 手牌: %v", aiName, aiUID, aiHand)
	log.Printf("[教学脚本] 🎴 初始场上牌: %s", initialDiscard)
	log.Printf("[教学脚本] 📊 牌堆剩余: %d张", len(gameRoom.GameState.DrawPile))

	// 检查第一位玩家（应该是人类）
	go gameRoom.CheckNextTurnAI()

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
		username := ""
		nickname := ""
		avatar := ""

		// 先从 GameState.Players 中寻找（如果游戏已开始）
		foundInState := false
		if gameRoom.GameState != nil {
			for _, ps := range gameRoom.GameState.Players {
				if ps.UID == pid {
					username = ps.Username
					nickname = ps.Nickname
					avatar = ps.Avatar
					foundInState = true
					break
				}
			}
		}

		if !foundInState {
			user, err := repository.UserRepo.FindByUID(uint(pid))
			if pid < 0 {
				// AI 玩家（游戏未开始时的默认信息）
				username = fmt.Sprintf("AI_%d", -pid)
				nickname = "AI 研究员"
				avatar = "🤖"
			} else if err != nil {
				username = fmt.Sprintf("研究员_%d", pid)
				nickname = username
				avatar = "🧪"
			} else {
				username = user.Username
				nickname = user.Nickname
				if nickname == "" {
					nickname = username
				}
				avatar = user.Avatar
			}
		}

		offline := false
		if _, exists := gameRoom.OfflineAt[pid]; exists {
			offline = true
		}
		// AI 玩家（UID < 0）永远被视为在线
		if pid < 0 {
			offline = false
		}
		playersInfo = append(playersInfo, map[string]interface{}{
			"uid":        pid,
			"username":   username,
			"nickname":   nickname,
			"avatar":     avatar,
			"is_offline": offline,
		})
	}

	// 确保 ready_uids 永远不为 nil，避免 JSON 序列化为 null
	readyUIDs := gameRoom.Room.ReadyUIDs
	if readyUIDs == nil {
		readyUIDs = []int{}
	}

	// 确保 spectators 列表始终返回
	spectators := gameRoom.Room.Spectators
	if spectators == nil {
		spectators = []int{}
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
		"is_pve":         gameRoom.Room.IsPvE, // 添加 is_pve 字段
		"spectators":     spectators,          // 即使GameState为nil也返回观战者列表
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
					Nickname:              player.Nickname,
					Avatar:                player.Avatar,
					HandCards:             nil, // 不显示具体手牌
					CardCount:             player.CardCount,
					IsReady:               player.IsReady,
					DoubleActionAvailable: player.DoubleActionAvailable,
					ActionProgress:        player.ActionProgress,
					IsAI:                  player.IsAI,
				}
				filteredPlayers = append(filteredPlayers, filteredPlayer)
			}
		}

		result["game_state"] = map[string]interface{}{
			"players":               filteredPlayers,
			"spectators":            gameRoom.Room.Spectators,
			"finished_players":      gameRoom.GameState.FinishedPlayers,
			"current_player":        gameRoom.GameState.CurrentPlayer,
			"direction":             gameRoom.GameState.Direction,
			"last_card":             gameRoom.GameState.LastCard,
			"deck_count":            len(gameRoom.GameState.DrawPile),
			"status":                gameRoom.GameState.Status,
			"turn_end_time":         gameRoom.GameState.TurnEndTime,
			"allowed_any_player":    gameRoom.GameState.AllowedAnyPlayer,
			"pending_draw_count":    gameRoom.GameState.PendingDrawCount,
			"is_spectator":          isSpectator,
			"points_changes":        gameRoom.GameState.PointsChanges,
			"xp_changes":            gameRoom.GameState.XPChanges,
			"original_player_count": gameRoom.GameState.OriginalPlayerCount,
			"quitted_count":         gameRoom.GameState.QuittedCount,
			"tutorial_script_mode":  gameRoom.GameState.TutorialScriptMode,
			"tutorial_current_step": gameRoom.GameState.TutorialCurrentStep,
		}
	}

	return result, nil
}

func getCardEffect(cardType string) string {
	effects := map[string]string{
		"+2": "+2",
		"+4": "+4",
		"He": "skip",
		"Ne": "skip",
		"Ar": "skip",
		"Kr": "skip",
		"Xe": "skip",
		"Rn": "skip",
		"Au": "Au",
	}
	return effects[cardType]
}

// 出牌
func PlayCard(roomID string, uid int, card models.Card, substance string) error {
	substance = NormalizeSubscripts(substance)
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

	// 若未指定substance，说明玩家单击了元素手牌
	// 直接使用元素符号，不进行单质转换
	if substance == "" {
		substance = card.Type // 直接使用元素符号（如 H）
	}

	// 🎓 教学脚本模式：严格验证出牌是否符合当前步骤
	if gameRoom.GameState.TutorialScriptMode {
		currentStep := gameRoom.GameState.TutorialCurrentStep
		currentScriptStep := getTutorialScriptStep(currentStep)
		if currentScriptStep != nil {
			// 仅允许脚本中定义的出牌动作
			if currentScriptStep.Action != "play" {
				return errors.New("当前步骤不允许出牌，请按教学提示操作")
			}

			// 人类玩家只能在 human 步骤行动
			if uid >= 0 && currentScriptStep.Player != "human" {
				return errors.New("当前是 AI 演示步骤，请等待 AI 操作")
			}
			// AI 只能在 ai 步骤行动
			if uid < 0 && currentScriptStep.Player != "ai" {
				return errors.New("当前不是 AI 出牌步骤")
			}

			if substance != currentScriptStep.Substance {
				log.Printf("[教学脚本] ❌ 玩家尝试打出 %s，但当前步骤 %d 要求打出 %s",
					substance, currentStep, currentScriptStep.Substance)
				return fmt.Errorf("请按照教程出牌，当前步骤应打出 %s", currentScriptStep.Substance)
			}

			log.Printf("[教学脚本] ✅ 玩家正确打出 %s (步骤 %d)", substance, currentStep)
		}
	}

	// +2/4/Au/换向牌/插件卡 可随意打出，无需反应条件
	nobleGases := map[string]bool{"He": true, "Ne": true, "Ar": true, "Kr": true, "Xe": true, "Rn": true}
	specialTypes := map[string]bool{"+2": true, "+4": true, "Au": true, "reverse": true, "skip": true}
	isSpecial := specialTypes[card.Type] || specialTypes[card.Effect] || nobleGases[card.Type] || nobleGases[substance] || specialTypes[substance] || IsPluginCard(card.Type)

	// 无论是否为特殊牌，所有出牌物质均需经过 substances 表校验（纯功能牌在 IsValidSubstance 中有白名单放行）
	// 插件卡无需物质校验
	if !IsPluginCard(card.Type) && !IsValidSubstance(substance) {
		return errors.New("该物质非法，请先在百科中录入: " + substance)
	}

	requiredElements := parseSubstance(substance)
	usedCards := []int{} // 记录将要从手牌中移除的索引

	// 如果是功能牌且 substance 无法解析出元素（即不是反应），则通过 card.Type 直接定位手牌
	if isSpecial && len(requiredElements) == 0 {
		for i, hCard := range currentPlayer.HandCards {
			if hCard.Type == card.Type {
				usedCards = append(usedCards, i)
				break
			}
		}
	}

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

	// 如果当前玩家被允许无视反应条件出牌，则跳过反应检查
	allowAny := gameRoom.GameState.AllowedAnyPlayer == gameRoom.GameState.CurrentPlayer
	if !isSpecial && gameRoom.GameState.LastCard != nil && !allowAny {
		canReact := false
		reactionDisplay := ""

		if len(gameRoom.GameState.LastCard.Reactants) > 0 {
			// 如果上一次是双联反应，则只需与其中任一物质反应即可
			for _, r := range gameRoom.GameState.LastCard.Reactants {
				react, err := repository.ReactionRepo.GetReaction(r, substance)
				if err == nil && react != nil {
					canReact = true
					reactionDisplay = react.Display
					break
				}
			}
		} else {
			react, err := repository.ReactionRepo.GetReaction(gameRoom.GameState.LastCard.Substance, substance)
			if err == nil && react != nil {
				canReact = true
				reactionDisplay = react.Display
			}
		}

		if !canReact {
			return errors.New("无法与上一张牌反应: " + substance)
		}

		// 更新当前反应方程式
		if reactionDisplay != "" {
			gameRoom.GameState.CurrentReaction = reactionDisplay
		} else {
			// 如果不是反应（如无视条件出牌），清空显示
			gameRoom.GameState.CurrentReaction = ""
		}
	} else if isSpecial {
		// 特殊卡牌操作结束后，清空反应显示
		gameRoom.GameState.CurrentReaction = ""
	}
	// nobleGases 作为换向牌

	gameRoom.BroadcastSystemMessage(fmt.Sprintf("反应成功！%s 合成了物质：[ %s ]。", currentPlayer.Nickname, substance))

	// 检查选中的卡牌中是否有带效果的
	activeEffect := ""
	if specialTypes[card.Type] {
		activeEffect = card.Type
	}
	if card.Effect != "" {
		activeEffect = card.Effect
	}

	// 补回：如果 AI 使用其合成的物质作为卡牌类型，则从 substance 中提取潜在特权效果
	// 特殊卡牌类型在 specialTypes 中已定义，如 "Au", "+2" 等
	if activeEffect == "" && specialTypes[substance] {
		activeEffect = substance
	}

	// 稀有气体具有稳定性，能够逆转实验方向
	if nobleGases[substance] || nobleGases[card.Type] {
		activeEffect = "reverse"
	}

	// +2/4/Au/换向牌可随意打出，无需反应条件
	if gameRoom.GameState.PendingDrawCount > 0 {
		// 细化逻辑：必须打出加牌，Au 也无法在加牌挑战中生效 (因为它不是加牌防御)
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
		Substance: substance, // 使用元素符号保存
		PlayerUID: uid,
	}

	displayName := currentPlayer.Nickname
	if displayName == "" {
		displayName = currentPlayer.Username
	}
	gameRoom.BroadcastSystemMessage(fmt.Sprintf("%s 成功合成了 %s。", displayName, substance))

	// 1. 更新场面状态
	// 转向牌、跳过牌、Au、+2、+4、插件卡 不更新场上的物质（不更新 LastCard），使下家仍需与之前的物质反应
	// 其中 Au 会在后续逻辑中显式清空 LastCard
	isPluginEffect := IsPluginCard(card.Type)
	isActionCard := isPluginEffect || activeEffect == "reverse" || activeEffect == "skip" || activeEffect == "Au" || activeEffect == "+2" || activeEffect == "+4"
	if !isActionCard {
		gameRoom.GameState.LastCard = &playedCard
	} else if activeEffect == "+2" || activeEffect == "+4" {
		// 当打出加牌时，将场上显示物质改为具体的加牌张数
		displayPlayedCard := playedCard
		displayPlayedCard.Substance = activeEffect
		gameRoom.GameState.LastCard = &displayPlayedCard
	}
	gameRoom.GameState.DiscardPile = append(gameRoom.GameState.DiscardPile, playedCard)
	emitCardPlayed(roomID, gameRoom.GameState, uid, card.Type, substance, activeEffect, false)

	// 2. 检查并注册获胜（先注册，以便 getNextPlayer 能正确跳过已完成玩家）
	isWinner := false
	if currentPlayer.CardCount == 0 {
		gameRoom.GameState.FinishedPlayers = append(gameRoom.GameState.FinishedPlayers, uid)
		isWinner = true

		// 实时结算积分：玩家完成即可获得部分积分并可随时离开
		if gameRoom.Room.IsPointsMode && uid > 0 {
			rank := len(gameRoom.GameState.FinishedPlayers)
			earnedPoints := 0
			if rank == 1 {
				earnedPoints = 100
			} else if rank == 2 {
				earnedPoints = 50
			} else if rank == 3 {
				earnedPoints = 33
			} else {
				earnedPoints = 25
			}

			// PvE 难度修正
			if gameRoom.Room.IsPvE && gameRoom.Room.PvEDifficulty >= 50 {
				earnedPoints = earnedPoints * gameRoom.Room.PvEDifficulty / 100
			}

			if earnedPoints > 0 {
				_ = repository.UserRepo.AddPoints(uint(uid), earnedPoints)
				log.Printf("[实时结算] 玩家 %d (第 %d 名) 获得积分: %d", uid, rank, earnedPoints)

				// 更新 PointsChanges，确保前端能显示
				if gameRoom.GameState.PointsChanges == nil {
					gameRoom.GameState.PointsChanges = make(map[int]int)
				}
				gameRoom.GameState.PointsChanges[uid] = earnedPoints

				if websocket.GlobalHub != nil {
					websocket.GlobalHub.SendToUID(uid, websocket.Message{
						Type: "action_toast",
						Data: fmt.Sprintf("恭喜完成实验！你是第 %d 名，获得 %d 积分。你可以继续观战或随时离开实验室。", rank, earnedPoints),
					})
				}
			}
		} else if uid < 0 {
			// AI 玩家完成时也要记录到 PointsChanges（积分为0）
			if gameRoom.GameState.PointsChanges == nil {
				gameRoom.GameState.PointsChanges = make(map[int]int)
			}
			gameRoom.GameState.PointsChanges[uid] = 0
		}
	}

	// 3. 处理卡牌效果及回合转移
	if activeEffect == "+2" || activeEffect == "+4" {
		// 叠加加牌累计
		gameRoom.GameState.PendingDrawCount += map[string]int{"+2": 2, "+4": 4}[activeEffect]
		gameRoom.GameState.PendingDrawTypes = append(gameRoom.GameState.PendingDrawTypes, activeEffect)
		// 传递至下家
		gameRoom.GameState.CurrentPlayer = getNextPlayer(gameRoom.GameState)
	} else if activeEffect == "Au" {
		// Au 效果：清空场面且跳过一人
		gameRoom.GameState.LastCard = nil
		skippedIdx := getNextPlayer(gameRoom.GameState)

		// 暂时移动到被跳过的玩家，以便下一步 getNextPlayer 能找到正确的人
		gameRoom.GameState.CurrentPlayer = skippedIdx
		targetIdx := getNextPlayer(gameRoom.GameState)

		gameRoom.GameState.CurrentPlayer = targetIdx
		gameRoom.GameState.AllowedAnyPlayer = targetIdx

		if websocket.GlobalHub != nil {
			skippedPlayer := gameRoom.GameState.Players[skippedIdx].Nickname
			nextPlayer := gameRoom.GameState.Players[targetIdx].Nickname
			websocket.GlobalHub.BroadcastToRoom(gameRoom.Room.ID, websocket.Message{
				Type: "action_toast",
				Data: fmt.Sprintf("Au 金元素触发！跳过研究员 %s，等待 %s 出牌...", skippedPlayer, nextPlayer),
			})
		}
	} else if activeEffect == "skip" {
		// 稀有气体跳过效果
		skippedIdx := getNextPlayer(gameRoom.GameState)
		gameRoom.GameState.CurrentPlayer = skippedIdx
		targetIdx := getNextPlayer(gameRoom.GameState)
		gameRoom.GameState.CurrentPlayer = targetIdx

		if websocket.GlobalHub != nil {
			skippedPlayer := gameRoom.GameState.Players[skippedIdx].Nickname
			nextPlayer := gameRoom.GameState.Players[targetIdx].Nickname
			websocket.GlobalHub.BroadcastToRoom(gameRoom.Room.ID, websocket.Message{
				Type: "action_toast",
				Data: fmt.Sprintf("稀有气体具有稳定性！跳过研究员 %s，轮到 %s 出牌...", skippedPlayer, nextPlayer),
			})
		}
	} else {
		// 转向效果及常规回合转移
		if activeEffect == "reverse" {
			gameRoom.GameState.Direction *= -1

			if websocket.GlobalHub != nil {
				// 获取下一位玩家名称
				nextIdx := getNextPlayer(gameRoom.GameState)
				nextPlayer := gameRoom.GameState.Players[nextIdx].Nickname

				msg := fmt.Sprintf("⚛️ 元素稳定性触发！实验方向发生逆转，现在轮到 %s 进行研究！", nextPlayer)
				websocket.GlobalHub.BroadcastToRoom(gameRoom.Room.ID, websocket.Message{
					Type: "action_toast",
					Data: msg,
				})
			}
		}

		// 在执行插件效果之前快照：当前玩家是否已有 force_play 强制出牌任务
		hadForcedPlays := gameRoom.GameState.PendingForcedPlays > 0

		// 插件卡效果执行（在回合转移前）
		if isPluginEffect {
			if err := ExecutePluginEffect(gameRoom, curIdx, card.Type); err != nil {
				return err
			}
		}

		// force_play 机制：若当前玩家本轮有强制出牌任务，消耗一次后决定是否推进回合
		if hadForcedPlays {
			gameRoom.GameState.PendingForcedPlays--
			if gameRoom.GameState.PendingForcedPlays > 0 {
				// 仍有强制出牌次数，保持当前玩家
				if websocket.GlobalHub != nil {
					websocket.GlobalHub.BroadcastToRoom(gameRoom.Room.ID, websocket.Message{
						Type: "action_toast",
						Data: fmt.Sprintf("⚡ %s 还需强制打出 %d 张牌！", currentPlayer.Nickname, gameRoom.GameState.PendingForcedPlays),
					})
				}
				gameRoom.GameState.AllowedAnyPlayer = -1
				gameRoom.GameState.CurrentReaction = ""
				gameRoom.recordTurnStart()
				setTurnEndTimeByMode(gameRoom.GameState)
				gameRoom.broadcastRoomUpdate()
				go gameRoom.CheckNextTurnAI()
				return nil
			}
		}

		gameRoom.GameState.CurrentPlayer = getNextPlayer(gameRoom.GameState)
	}

	// 4. 更新行动进度与通用状态
	currentPlayer.ActionProgress++
	if currentPlayer.ActionProgress >= 2 {
		currentPlayer.DoubleActionAvailable = true
	}

	gameRoom.recordTurnStart()
	setTurnEndTimeByMode(gameRoom.GameState)

	// 清除消费掉的 AllowedAnyPlayer 标记
	if gameRoom.GameState.AllowedAnyPlayer == curIdx {
		gameRoom.GameState.AllowedAnyPlayer = -1
	}

	// 5. 若玩家已获胜，检查整体游戏是否结束
	// 5. 若玩家已获胜，检查整体游戏是否结束
	if isWinner {
		activeCount := 0
		var lastPlayerUID int

		// 重新计算剩余活跃玩家
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

		// 判断游戏是否结束
		// PvE 模式和其他模式：统计所有玩家，当剩余 ≤1 个玩家时结束
		shouldEndGame := activeCount <= 1

		if shouldEndGame {
			if activeCount == 1 {
				gameRoom.GameState.FinishedPlayers = append(gameRoom.GameState.FinishedPlayers, lastPlayerUID)
			}
			finalizeGame(gameRoom)
		} else {

			// 游戏继续，广播更新包含 finished_players
			// 前端需要根据 finished_players 列表展示"已完成"状态
			log.Printf("玩家 %d 完成游戏，剩余活跃玩家: %d", uid, activeCount)
		}
	}

	// 🎓 教学脚本模式：仅在脚本步骤存在时递增
	if gameRoom.GameState.TutorialScriptMode {
		currentScriptStep := getTutorialScriptStep(gameRoom.GameState.TutorialCurrentStep)
		if currentScriptStep != nil && (currentScriptStep.Action == "play" || currentScriptStep.Action == "draw" || currentScriptStep.Action == "double") {
			gameRoom.GameState.TutorialCurrentStep++
			log.Printf("[教学脚本] 📈 步骤递增至 %d", gameRoom.GameState.TutorialCurrentStep)
		}
	}

	// 检查下一位是否是 AI
	emitTurnChanged(roomID, gameRoom.GameState, curIdx, "play_card")
	go gameRoom.CheckNextTurnAI()

	// 广播状态更新 (确保 AI 回合等非 HTTP 触发的消息能到达前端)
	gameRoom.broadcastRoomUpdate()

	return nil
}

func getNextPlayer(state *models.GameState) int {
	playersLen := len(state.Players)
	if playersLen == 0 {
		return 0
	}

	next := state.CurrentPlayer
	// 最多尝试寻找玩家总次数，防止所有玩家都已完成导致的死循环
	for i := 0; i < playersLen; i++ {
		next = next + state.Direction
		if next < 0 {
			next = playersLen - 1
		} else if next >= playersLen {
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

	// 极端兜底：如果找不到未完成的玩家，返回当前索引
	return state.CurrentPlayer
}

func buildFreshDrawPileFromConfig(gameRoom *GameRoom) []models.Card {
	baseCards := gameRoom.Room.DeckConfig.Cards
	if len(baseCards) == 0 {
		baseCards = getDefaultDeckConfig()
	}

	freshPile := make([]models.Card, 0)
	for cardType, count := range baseCards {
		if count <= 0 {
			continue
		}
		effect := getCardEffect(cardType)
		for i := 0; i < count; i++ {
			freshPile = append(freshPile, models.Card{
				Type:   cardType,
				Count:  1,
				Effect: effect,
			})
		}
	}

	// 插件牌按默认数量注入新牌堆
	for _, pluginCard := range GetAllPluginCards() {
		if pluginCard == nil || pluginCard.DefaultCount <= 0 {
			continue
		}
		for i := 0; i < pluginCard.DefaultCount; i++ {
			freshPile = append(freshPile, models.Card{
				Type:  pluginCard.Symbol,
				Count: 1,
			})
		}
	}

	return freshPile
}

func shuffleCardsWithFreshSeed(cards []models.Card) int64 {
	seed := int64(cryptoRandUint64() & 0x7fffffffffffffff)
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})
	return seed
}

func reshuffleDeck(gameRoom *GameRoom) {
	if len(gameRoom.GameState.AllUsedCards) == 0 {
		// 彻底没有可回收牌时，重建一副新牌堆继续游戏。
		freshPile := buildFreshDrawPileFromConfig(gameRoom)
		if len(freshPile) == 0 {
			return
		}
		seed := shuffleCardsWithFreshSeed(freshPile)
		gameRoom.GameState.DrawPile = freshPile
		log.Printf("[牌堆重建] ♻️ 牌堆已耗尽，已按配置重建新牌堆 %d 张（seed=%d）", len(freshPile), seed)
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
	previousTurnIndex := gameRoom.GameState.CurrentPlayer

	// 教学脚本关卡禁用摸牌，仅在脚本要求时开放
	if gameRoom.GameState.TutorialScriptMode {
		currentStep := gameRoom.GameState.TutorialCurrentStep
		currentScriptStep := getTutorialScriptStep(currentStep)
		if currentScriptStep == nil || currentScriptStep.Action != "draw" {
			return errors.New("教学关卡禁止在此步摸牌，请按底部提示操作")
		}

		// AI 摸牌校验（如果 AI 通过此函数摸牌）
		if uid < 0 && currentScriptStep.Player != "ai" {
			return errors.New("当前不是 AI 摸牌步骤")
		}
		// 人类玩家摸牌校验
		if uid >= 0 && currentScriptStep.Player != "human" {
			return errors.New("当前是 AI 演示步骤，请等待 AI 操作")
		}

		log.Printf("[教学脚本] ✅ 玩家执行摸牌 (步骤 %d)", currentStep)
	}

	actualCount := count
	penaltyResolved := false
	// 如果有强制出牌任务，摸牌意味着放弃，清零后继续
	if gameRoom.GameState.PendingForcedPlays > 0 {
		gameRoom.GameState.PendingForcedPlays = 0
	}
	// 如果有挂起的加牌，结算加牌
	if gameRoom.GameState.PendingDrawCount > 0 {
		actualCount = gameRoom.GameState.PendingDrawCount
		gameRoom.GameState.PendingDrawCount = 0
		gameRoom.GameState.PendingDrawTypes = nil
		penaltyResolved = true
	}

	drawCardsForPlayer(gameRoom, gameRoom.GameState.CurrentPlayer, actualCount)

	displayName := currentPlayer.Nickname
	if displayName == "" {
		displayName = currentPlayer.Username
	}

	if penaltyResolved {
		gameRoom.BroadcastSystemMessage(fmt.Sprintf("反应失控！%s 已被强制送往洗眼台（处理了 %d 张累计罚牌）。", displayName, actualCount))
	} else {
		gameRoom.BroadcastSystemMessage(fmt.Sprintf("%s 进入库房寻找灵感（摸了 %d 张牌）。", displayName, actualCount))
	}

	// 行动进度更新
	currentPlayer.ActionProgress++
	if currentPlayer.ActionProgress >= 2 {
		currentPlayer.DoubleActionAvailable = true
	}

	// 🎓 教学脚本模式：仅在脚本步骤存在时递增
	if gameRoom.GameState.TutorialScriptMode {
		currentScriptStep := getTutorialScriptStep(gameRoom.GameState.TutorialCurrentStep)
		if currentScriptStep != nil && (currentScriptStep.Action == "play" || currentScriptStep.Action == "draw" || currentScriptStep.Action == "double") {
			gameRoom.GameState.TutorialCurrentStep++
			log.Printf("[教学脚本] 📈 步骤递增至 %d", gameRoom.GameState.TutorialCurrentStep)
		}
	}

	// 摸牌后跳过回合
	gameRoom.GameState.CurrentPlayer = getNextPlayer(gameRoom.GameState)
	// 摸牌操作结束，更新反应显示为无
	gameRoom.GameState.CurrentReaction = ""

	gameRoom.recordTurnStart()
	setTurnEndTimeByMode(gameRoom.GameState)

	// 如果结算了罚牌，清空场面并允许下家随意出牌
	if penaltyResolved {
		gameRoom.GameState.LastCard = nil
		gameRoom.GameState.AllowedAnyPlayer = gameRoom.GameState.CurrentPlayer
	} else {
		// 普通摸牌清除可能存在的 allowAny 标记
		gameRoom.GameState.AllowedAnyPlayer = -1
	}

	// 检查下一位是否是 AI
	emitTurnChanged(roomID, gameRoom.GameState, previousTurnIndex, "draw_card")
	go gameRoom.CheckNextTurnAI()

	// 广播状态更新
	gameRoom.broadcastRoomUpdate()

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

func saveGameHistory(roomID string, winnerUID int, players []int, originalPlayerCount int, quittedCount int, finishedPlayers []int) {
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
		if uid > 0 {
			repository.UserRepo.IncrementTotalGames(uint(uid))
		}
	}

	// 更新胜利者的胜利场数
	if winnerUID > 0 {
		repository.UserRepo.IncrementWinCount(uint(winnerUID))
	}

	// 排名后50%的真实玩家记为 loser（递增 negative_play_count）
	total := len(finishedPlayers)
	if total >= 2 {
		loserStart := (total + 1) / 2 // 后50%起始索引（向上取整）
		for i := loserStart; i < total; i++ {
			uid := finishedPlayers[i]
			if uid > 0 {
				repository.UserRepo.IncrementNegativePlayCount(uint(uid))
			}
		}
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

func setTurnEndTimeByMode(state *models.GameState) {
	if state == nil {
		return
	}
	// 教学脚本关卡不启用回合倒计时
	if state.TutorialScriptMode {
		state.TurnEndTime = 0
		return
	}
	state.TurnEndTime = time.Now().Add(getPlayerActionTimeout()).UnixNano() / int64(time.Millisecond)
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
	sub1 = NormalizeSubscripts(sub1)
	sub2 = NormalizeSubscripts(sub2)
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

	isAllowedAnyPlayer := gameRoom.GameState.AllowedAnyPlayer != -1 && gameRoom.GameState.Players[gameRoom.GameState.AllowedAnyPlayer].UID == uid
	if !isAllowedAnyPlayer && currentPlayer.UID != uid {
		return errors.New("还没轮到你")
	}

	// 如果是 AllowedAnyPlayer 跳出顺序出牌逻辑
	if isAllowedAnyPlayer {
		curIdx = gameRoom.GameState.AllowedAnyPlayer
		currentPlayer = gameRoom.GameState.Players[curIdx]
	}

	// 🎓 教学脚本模式：当前脚本不允许双联反应
	if gameRoom.GameState.TutorialScriptMode {
		currentScriptStep := getTutorialScriptStep(gameRoom.GameState.TutorialCurrentStep)
		if currentScriptStep != nil {
			return errors.New("当前步骤不允许双联反应，请按教学提示操作")
		}
	}

	// 检查冷却
	if !currentPlayer.DoubleActionAvailable {
		// 如果是有权出牌的人（比如金卡触发），则无视此判定
		if !isAllowedAnyPlayer {
			return errors.New("双联反应尚未就绪（每行动2次可使用1次）")
		}
	}

	// 当玩家选择自身两物质反应时，不考虑先前出牌（即跳过与场上 LastCard 的连接检查）

	// 如果有挂起的加牌，禁止发动双联行动
	if gameRoom.GameState.PendingDrawCount > 0 {
		return errors.New("当前处于加牌结算状态，不可发动双联反应")
	}

	// 校验物质是否已录入
	if !IsValidSubstance(sub1) {
		return errors.New("该物质非法，请先在百科中录入: " + sub1)
	}
	if !IsValidSubstance(sub2) {
		return errors.New("该物质非法，请先在百科中录入: " + sub2)
	}

	// 准备所需元素和特殊卡牌识别
	specialTypes := map[string]bool{"+2": true, "+4": true, "Au": true, "reverse": true, "skip": true}
	nobleGases := map[string]bool{"He": true, "Ne": true, "Ar": true, "Kr": true, "Xe": true, "Rn": true}

	// 禁止功能牌参与双联反应
	if specialTypes[sub1] || nobleGases[sub1] || specialTypes[sub2] || nobleGases[sub2] {
		return errors.New("功能牌或稀有气体性质稳定，无法参与双联反应")
	}

	// 校验两物质是否能反应
	reaction, err := repository.ReactionRepo.GetReaction(sub1, sub2)
	if err != nil || reaction == nil {
		return errors.New(sub1 + " 与 " + sub2 + " 之间无法产生反应，不可发动双联行动")
	}

	req1 := parseSubstance(sub1)
	req2 := parseSubstance(sub2)

	usedCards := []int{} // 记录将要从手牌中移除的索引

	allReqs := make(map[string]int)
	// 仅考虑元素种类，不考虑系数
	for k := range req1 {
		allReqs[k] = 1
	}
	for k := range req2 {
		allReqs[k] = 1
	}

	// 检查并记录元素消耗
	for elemName, count := range allReqs {
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
					break
				}
			}
			if !found {
				return errors.New("手牌对应元素牌不足: " + elemName + " (需要 1 张)")
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
	gameRoom.GameState.CurrentReaction = reaction.Display
	gameRoom.GameState.LastCard = &playedCard
	gameRoom.GameState.DiscardPile = append(gameRoom.GameState.DiscardPile, playedCard)
	emitCardPlayed(roomID, gameRoom.GameState, uid, representCard.Type, playedCard.Substance, "", true)

	// 处理特殊效果（如果双联中包含功能牌）
	for _, c := range consumedCards {
		effect := c.Effect
		if effect == "" {
			if nobleGases[c.Type] {
				effect = "skip"
			} else {
				effect = c.Type
			}
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
			if websocket.GlobalHub != nil {
				nextIdx := getNextPlayer(gameRoom.GameState)
				nextPlayer := gameRoom.GameState.Players[nextIdx].Nickname
				websocket.GlobalHub.BroadcastToRoom(gameRoom.Room.ID, websocket.Message{
					Type: "action_toast",
					Data: fmt.Sprintf("⚛️ 元素稳定性触发！实验方向发生逆转，现在轮到 %s 进行研究！", nextPlayer),
				})
			}
		case "skip":
			// 跳过下一位
			skippedIdx := getNextPlayer(gameRoom.GameState)
			gameRoom.GameState.CurrentPlayer = skippedIdx
			if websocket.GlobalHub != nil {
				skippedPlayer := gameRoom.GameState.Players[skippedIdx].Nickname
				websocket.GlobalHub.BroadcastToRoom(gameRoom.Room.ID, websocket.Message{
					Type: "action_toast",
					Data: fmt.Sprintf("⚠️ 能量激增！禁制场域使研究员 %s 被迫暂离实验桌！", skippedPlayer),
				})
			}
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
				skippedPlayer := gameRoom.GameState.Players[skippedIdx].Nickname
				nextPlayer := gameRoom.GameState.Players[targetIdx].Nickname
				websocket.GlobalHub.BroadcastToRoom(gameRoom.Room.ID, websocket.Message{
					Type: "action_toast",
					Data: fmt.Sprintf("Au 金元素双联触发！跳过研究员 %s，等待 %s 出牌...", skippedPlayer, nextPlayer),
				})
			}
		}
	}

	// 检查获胜
	if currentPlayer.CardCount == 0 {
		// 注册获胜玩家
		alreadyFinished := false
		for _, fuid := range gameRoom.GameState.FinishedPlayers {
			if fuid == uid {
				alreadyFinished = true
				break
			}
		}
		if !alreadyFinished {
			gameRoom.GameState.FinishedPlayers = append(gameRoom.GameState.FinishedPlayers, uid)
		}

		// 计算活跃玩家
		activeCount := 0
		var lastPlayerUID int
		for _, p := range gameRoom.GameState.Players {
			isF := false
			for _, fuid := range gameRoom.GameState.FinishedPlayers {
				if p.UID == fuid {
					isF = true
					break
				}
			}
			if !isF {
				activeCount++
				lastPlayerUID = p.UID
			}
		}

		if activeCount <= 1 {
			if activeCount == 1 {
				gameRoom.GameState.FinishedPlayers = append(gameRoom.GameState.FinishedPlayers, lastPlayerUID)
			}
			finalizeGame(gameRoom)
			return nil
		}

		log.Printf("[双联反应] 玩家 %d 完成游戏，剩余活跃玩家: %d", uid, activeCount)
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
	setTurnEndTimeByMode(gameRoom.GameState)

	// 检查下一位是否是 AI
	emitTurnChanged(roomID, gameRoom.GameState, curIdx, "double_play")
	go gameRoom.CheckNextTurnAI()

	// 广播状态更新
	gameRoom.broadcastRoomUpdate()

	return nil
}

// finalizeGame 统一处理游戏结束逻辑
func finalizeGame(gr *GameRoom) {
	gr.GameState.Status = "finished"
	gr.Room.Status = "finished"

	winnerUID := gr.GameState.FinishedPlayers[0]
	saveGameHistory(gr.Room.ID, winnerUID, gr.Room.Players, gr.GameState.OriginalPlayerCount, gr.GameState.QuittedCount, gr.GameState.FinishedPlayers)

	privateChatRepo := repository.NewPrivateChatRepository()
	if err := privateChatRepo.DeleteGameInvitesByRoom(gr.Room.ID); err != nil {
		log.Printf("清理房间 %s 的游戏邀请失败: %v", gr.Room.ID, err)
	}

	if gr.Room.IsPointsMode || gr.Room.IsPvE {
		handlePointsCalculation(gr)
	} else {
		handleXPCalculation(gr)
	}

	gr.BroadcastSystemMessage("实验完成！反应已达平衡，正在生成结果报告。")

	log.Printf("[游戏结束] 房间 %s 已正式结算，冠军 UID: %d", gr.Room.ID, winnerUID)
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
	if gameRoom.GameState.TutorialScriptMode {
		return
	}

	now := time.Now().UnixNano() / int64(time.Millisecond)
	if gameRoom.GameState.TurnEndTime > 0 && now > gameRoom.GameState.TurnEndTime {
		// 超时处理：强制摸牌并跳过
		previousTurnIndex := gameRoom.GameState.CurrentPlayer
		currentPlayer := gameRoom.GameState.Players[gameRoom.GameState.CurrentPlayer]
		if !currentPlayer.IsAI {
			log.Printf("[Game] ⚡ 玩家 %s (%d) 超时，自动摸牌并跳过本回合", currentPlayer.Nickname, currentPlayer.UID)
			gameRoom.BroadcastSystemMessage(fmt.Sprintf("研究员 %s 操作超时。出于实验室安全考虑（因为你太慢了），已自动摸牌并跳过回合。", currentPlayer.Nickname))
		}

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
		setTurnEndTimeByMode(gameRoom.GameState)

		// 如果结算了罚牌，清空场面并允许下家随意出牌
		if penaltyResolved {
			gameRoom.GameState.LastCard = nil
			gameRoom.GameState.AllowedAnyPlayer = gameRoom.GameState.CurrentPlayer
		} else {
			gameRoom.GameState.AllowedAnyPlayer = -1
		}

		// 检查下一位是否是 AI
		emitTurnChanged(roomID, gameRoom.GameState, previousTurnIndex, "timeout_auto_draw")
		go gameRoom.CheckNextTurnAI()

		// 广播更新
		gameRoom.broadcastRoomUpdate()
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
