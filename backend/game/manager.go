package game

import (
	"chemistryuno/database"
	"chemistryuno/models"
	"chemistryuno/websocket"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

var (
	rooms     = make(map[string]*GameRoom)
	roomMutex sync.RWMutex
)

type GameRoom struct {
	Room      *models.Room
	GameState *models.GameState
	mutex     sync.RWMutex
	OfflineAt map[int]time.Time // UID -> 离线起始时间
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

// 创建房间
func CreateRoom(name string, hostUID int, hostName string, maxPlayers int, deckID int) (*models.Room, error) {
	if name == "" {
		const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		rand.Seed(time.Now().UnixNano())
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
	if deckID == 0 {
		deckConfig.Cards = getDefaultDeckConfig()
	} else {
		var cardsJSON string
		err := database.DB.QueryRow(
			"SELECT id, name, cards FROM deck_configs WHERE id = ?",
			deckID,
		).Scan(&deckConfig.ID, &deckConfig.Name, &cardsJSON)

		if err != nil {
			deckConfig.Cards = getDefaultDeckConfig()
		} else {
			json.Unmarshal([]byte(cardsJSON), &deckConfig.Cards)
		}
	}

	room := &models.Room{
		ID:         roomID,
		Name:       name,
		HostUID:    hostUID,
		Players:    []int{hostUID},
		MaxPlayers: maxPlayers,
		DeckConfig: &deckConfig,
		Status:     "waiting",
		CreatedAt:  time.Now(),
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

// 获取所有房间
func GetAllRooms() []*models.Room {
	roomMutex.RLock()
	defer roomMutex.RUnlock()

	result := []*models.Room{}
	for _, gr := range rooms {
		result = append(result, gr.Room)
	}
	return result
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
	if gr.Room.Status != "playing" {
		gr.mutex.Unlock()
		return
	}

	roomID := gr.Room.ID
	now := time.Now()
	playersToKick := []int{}

	// 1. 检测离线超过2分钟的玩家
	for _, uid := range gr.Room.Players {
		isOnline := websocket.GlobalHub.IsUIDInRoom(roomID, uid)
		if !isOnline {
			if _, exists := gr.OfflineAt[uid]; !exists {
				gr.OfflineAt[uid] = now
			} else if now.Sub(gr.OfflineAt[uid]) > 2*time.Minute {
				playersToKick = append(playersToKick, uid)
			}
		} else {
			delete(gr.OfflineAt, uid)
		}
	}
	gr.mutex.Unlock()

	// 2. 执行踢出操作
	for _, uid := range playersToKick {
		gr.kickPlayer(uid, "由于消极游戏，您已被踢出")
	}

	// 3. 检测玩家人数是否不足
	gr.mutex.Lock()
	if gr.Room.Status == "playing" && len(gr.Room.Players) < 2 {
		gr.mutex.Unlock()
		gr.terminateRoom("由于玩家人数不足，房间已被关闭")
		return
	}
	gr.mutex.Unlock()
}

func (gr *GameRoom) kickPlayer(uid int, reason string) {
	gr.mutex.Lock()
	roomID := gr.Room.ID
	isHost := gr.Room.HostUID == uid

	// 通知被踢出的玩家
	if websocket.GlobalHub != nil {
		websocket.GlobalHub.SendToUID(uid, websocket.Message{
			Type:    "player_kicked",
			Message: reason,
		})
	}

	if isHost {
		gr.mutex.Unlock()
		gr.terminateRoom("由于房主消极游戏，房间已被关闭")
		return
	}

	// 移除玩家
	newPlayers := []int{}
	for _, pid := range gr.Room.Players {
		if pid != uid {
			newPlayers = append(newPlayers, pid)
		}
	}
	gr.Room.Players = newPlayers

	// 如果游戏正在进行，也从 GameState 中移除
	if gr.GameState != nil {
		newPS := []*models.PlayerState{}
		kickedIndex := -1
		for i, ps := range gr.GameState.Players {
			if ps.UID != uid {
				newPS = append(newPS, ps)
			} else {
				kickedIndex = i
			}
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
	gr.mutex.Unlock()

	// 广播玩家离开消息
	if websocket.GlobalHub != nil {
		websocket.GlobalHub.BroadcastToRoom(roomID, websocket.Message{
			Type: "player_left",
			UID:  uid,
			Data: fmt.Sprintf("玩家 %d 已被系统踢出", uid),
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

// 加入房间
func JoinRoom(roomID string, uid int, username string) error {
	roomMutex.RLock()
	gameRoom, exists := rooms[roomID]
	roomMutex.RUnlock()

	if !exists {
		return errors.New("房间不存在")
	}

	gameRoom.mutex.Lock()
	defer gameRoom.mutex.Unlock()

	if gameRoom.Room.Status != "waiting" {
		return errors.New("游戏已开始")
	}

	if len(gameRoom.Room.Players) >= gameRoom.Room.MaxPlayers {
		return errors.New("房间已满")
	}

	// 检查是否已在房间中
	for _, pid := range gameRoom.Room.Players {
		if pid == uid {
			return errors.New("已在房间中")
		}
	}

	gameRoom.Room.Players = append(gameRoom.Room.Players, uid)
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

	// 如果房主离开，不管有没有其他玩家，直接解散房间（游戏终止）
	if gameRoom.Room.HostUID == uid {
		roomMutex.Lock()
		delete(rooms, roomID)
		roomMutex.Unlock()

		// 通知房间内所有玩家游戏已终止
		if websocket.GlobalHub != nil {
			websocket.GlobalHub.BroadcastToRoom(roomID, websocket.Message{
				Type: "room_terminated",
				Data: "房主已离开房间，游戏终止",
			})
		}
		return nil
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

	if gameRoom.Room.HostUID != uid {
		return errors.New("只有房主可以开始游戏")
	}

	if len(gameRoom.Room.Players) < 2 {
		return errors.New("至少需要2名玩家")
	}

	if gameRoom.Room.Status != "waiting" {
		return errors.New("游戏已开始")
	}

	// 初始化游戏状态
	gameRoom.GameState = &models.GameState{
		RoomID:           roomID,
		Players:          []*models.PlayerState{},
		CurrentPlayer:    0,
		Direction:        1,
		DrawPile:         []models.Card{},
		DiscardPile:      []models.PlayedCard{},
		Status:           "playing",
		TurnEndTime:      time.Now().Add(30*time.Second).UnixNano() / int64(time.Millisecond),
		PendingDrawCount: 0,
		PendingDrawTypes: nil,
		AllowedAnyPlayer: -1,
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
	rand.Seed(time.Now().UnixNano())
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

	// 初始化玩家
	for _, pid := range shuffledPlayers {
		var username, avatar string
		database.DB.QueryRow("SELECT username, avatar FROM users WHERE UID = ?", pid).
			Scan(&username, &avatar)

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

		// 发10张初始手牌
		for i := 0; i < 10 && len(gameRoom.GameState.DrawPile) > 0; i++ {
			card := gameRoom.GameState.DrawPile[0]
			gameRoom.GameState.DrawPile = gameRoom.GameState.DrawPile[1:]
			player.HandCards = append(player.HandCards, card)
			player.CardCount++
		}

		gameRoom.GameState.Players = append(gameRoom.GameState.Players, player)
	}

	gameRoom.Room.Status = "playing"
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
		return nil, errors.New("你不在该房间中")
	}

	// 获取玩家详细信息（用于准备页面）
	playersInfo := []map[string]interface{}{}
	for _, pid := range gameRoom.Room.Players {
		var username, avatar string
		err := database.DB.QueryRow("SELECT username, avatar FROM users WHERE UID = ?", pid).
			Scan(&username, &avatar)
		if err != nil {
			// 如果数据库查询失败，提供默认值
			username = fmt.Sprintf("研究员_%d", pid)
			avatar = "🧪"
		}
		playersInfo = append(playersInfo, map[string]interface{}{
			"uid":      pid,
			"username": username,
			"avatar":   avatar,
			"is_host":  pid == gameRoom.Room.HostUID,
		})
	}

	result := map[string]interface{}{
		"id":           gameRoom.Room.ID,
		"name":         gameRoom.Room.Name,
		"host_uid":     gameRoom.Room.HostUID,
		"players":      gameRoom.Room.Players,
		"players_info": playersInfo,
		"max_players":  gameRoom.Room.MaxPlayers,
		"status":       gameRoom.Room.Status,
	}

	if gameRoom.GameState != nil {
		// 过滤其他玩家的手牌
		filteredPlayers := []*models.PlayerState{}
		for _, player := range gameRoom.GameState.Players {
			if player.UID == uid {
				// 当前玩家，显示全部信息
				filteredPlayers = append(filteredPlayers, player)
			} else {
				// 其他玩家，隐藏手牌详情
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
			"current_player":     gameRoom.GameState.CurrentPlayer,
			"direction":          gameRoom.GameState.Direction,
			"last_card":          gameRoom.GameState.LastCard,
			"deck_count":         len(gameRoom.GameState.DrawPile),
			"status":             gameRoom.GameState.Status,
			"turn_end_time":      gameRoom.GameState.TurnEndTime,
			"allowed_any_player": gameRoom.GameState.AllowedAnyPlayer,
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
	next2 := next1 + dir
	if next2 < 0 {
		next2 = playersLen - 1
	} else if next2 >= playersLen {
		next2 = 0
	}
	if currentPlayer.UID != uid {
		return errors.New("还没轮到你")
	}

	// 若未指定substance，自动用单质（如H->H2、O->O2等，或直接元素符号）
	if substance == "" {
		substance = card.Type
	}
	requiredElements := parseSubstance(substance)
	usedCards := []int{} // 记录将要从手牌中移除的索引
	for elemName := range requiredElements {
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
				break
			}
		}
		if !found {
			return errors.New("缺少元素牌: " + elemName)
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
		gameRoom.GameState.Status = "finished"
		gameRoom.Room.Status = "finished"
		// 记录游戏历史
		saveGameHistory(roomID, uid, gameRoom.Room.Players)
		return nil
	}

	// 加牌叠加规则
	if activeEffect == "+2" || activeEffect == "+4" {
		// 叠加pending
		gameRoom.GameState.PendingDrawCount += map[string]int{"+2": 2, "+4": 4}[activeEffect]
		gameRoom.GameState.PendingDrawTypes = append(gameRoom.GameState.PendingDrawTypes, activeEffect)
		// 传递到下家
		gameRoom.GameState.CurrentPlayer = getNextPlayer(gameRoom.GameState)
		gameRoom.GameState.TurnEndTime = time.Now().Add(30*time.Second).UnixNano() / int64(time.Millisecond)
		// 如果之前允许任意出牌的标记被消费，清除
		if allowAny {
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
		gameRoom.GameState.TurnEndTime = time.Now().Add(30*time.Second).UnixNano() / int64(time.Millisecond)

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
		// 使得跳过下一位玩家，并允许下下家任意出牌
		gameRoom.GameState.CurrentPlayer = next1
		gameRoom.GameState.AllowedAnyPlayer = next2
	}

	// 行动进度更新
	currentPlayer.ActionProgress++
	if currentPlayer.ActionProgress >= 2 {
		currentPlayer.DoubleActionAvailable = true
	}

	// 下一位玩家（大多数情况均走到这里）
	gameRoom.GameState.CurrentPlayer = getNextPlayer(gameRoom.GameState)
	// 如果之前允许任意出牌的标记被消费，清除（保险）
	if allowAny {
		gameRoom.GameState.AllowedAnyPlayer = -1
	}
	gameRoom.GameState.TurnEndTime = time.Now().Add(30*time.Second).UnixNano() / int64(time.Millisecond)
	return nil
}

func getNextPlayer(state *models.GameState) int {
	next := state.CurrentPlayer + state.Direction
	if next < 0 {
		next = len(state.Players) - 1
	} else if next >= len(state.Players) {
		next = 0
	}
	return next
}

func reshuffleDeck(gameRoom *GameRoom) {
	if len(gameRoom.GameState.AllUsedCards) == 0 {
		return
	}
	// 将池中卡牌放回摸牌堆
	gameRoom.GameState.DrawPile = append(gameRoom.GameState.DrawPile, gameRoom.GameState.AllUsedCards...)
	gameRoom.GameState.AllUsedCards = nil

	// 重新洗牌
	rand.Seed(time.Now().UnixNano())
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
	gameRoom.GameState.TurnEndTime = time.Now().Add(30*time.Second).UnixNano() / int64(time.Millisecond)

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
	if gameRoom.GameState.LastCard != nil {
		reactable := []string{}
		lastSubstance := gameRoom.GameState.LastCard.Substance
		for _, sub := range substances {
			if CanReact(lastSubstance, sub) {
				reactable = append(reactable, sub)
			}
		}
		return reactable, nil
	}

	return substances, nil
}

func saveGameHistory(roomID string, winnerUID int, players []int) {
	playersJSON, _ := json.Marshal(players)
	database.DB.Exec(
		"INSERT INTO game_history (room_id, winner_uid, players, started_at, finished_at) VALUES (?, ?, ?, datetime('now', '-1 hour'), datetime('now'))",
		roomID, winnerUID, string(playersJSON),
	)
	fmt.Println("游戏历史已保存")
}

func init() {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			checkRoomsTimeout()
		}
	}()
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

	// 检查是否轮到该玩家
	currentPlayer := gameRoom.GameState.Players[gameRoom.GameState.CurrentPlayer]
	if currentPlayer.UID != uid {
		return errors.New("还没轮到你")
	}

	// 检查冷却
	if !currentPlayer.DoubleActionAvailable {
		return errors.New("双联反应尚未就绪（每行动2次可使用1次）")
	}

	// 如果有挂起的加牌，禁止发动双联行动
	if gameRoom.GameState.PendingDrawCount > 0 {
		return errors.New("当前处于加牌结算状态，不可发动双联反应")
	}

	// 校验两物质是否能反应
	if !CanReact(sub1, sub2) {
		return errors.New(sub1 + " 与 " + sub2 + " 之间无法产生反应，不可发动双联行动")
	}

	// 准备所需元素
	req1 := parseSubstance(sub1)
	req2 := parseSubstance(sub2)
	allReqs := make(map[string]int)
	for k, v := range req1 {
		allReqs[k] += v
	}
	for k, v := range req2 {
		allReqs[k] += v
	}

	// 检查手牌
	usedCards := []int{}
	for elemName, count := range allReqs {
		foundCount := 0
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
				foundCount++
				if foundCount == count {
					break
				}
			}
		}
		if foundCount < count {
			return errors.New("手牌元素不足以组成以上两种物质")
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

	// 检查是否获胜
	if currentPlayer.CardCount == 0 {
		gameRoom.GameState.Status = "finished"
		gameRoom.Room.Status = "finished"
		// 记录游戏历史
		saveGameHistory(roomID, uid, gameRoom.Room.Players)
		return nil
	}

	// 重置冷却
	currentPlayer.ActionProgress = 0
	currentPlayer.DoubleActionAvailable = false

	// 下一位玩家
	gameRoom.GameState.CurrentPlayer = getNextPlayer(gameRoom.GameState)
	gameRoom.GameState.TurnEndTime = time.Now().Add(30*time.Second).UnixNano() / int64(time.Millisecond)

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
		gameRoom.GameState.TurnEndTime = time.Now().Add(30*time.Second).UnixNano() / int64(time.Millisecond)

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
