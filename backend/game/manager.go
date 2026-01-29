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
		Room: room,
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

	// 如果房主离开，转移房主或删除房间
	if gameRoom.Room.HostUID == uid {
		if len(newPlayers) > 0 {
			gameRoom.Room.HostUID = newPlayers[0]
		} else {
			roomMutex.Lock()
			delete(rooms, roomID)
			roomMutex.Unlock()
		}
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
		RoomID:        roomID,
		Players:       []*models.PlayerState{},
		CurrentPlayer: 0,
		Direction:     1,
		DrawPile:      []models.Card{},
		DiscardPile:   []models.PlayedCard{},
		Status:        "playing",
		TurnEndTime:   time.Now().Add(30*time.Second).UnixNano() / int64(time.Millisecond),
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
			UID:       pid,
			Username:  username,
			Avatar:    avatar,
			HandCards: []models.Card{},
			CardCount: 0,
			IsReady:   true,
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

	result := map[string]interface{}{
		"id":          gameRoom.Room.ID,
		"name":        gameRoom.Room.Name,
		"host_uid":    gameRoom.Room.HostUID,
		"players":     gameRoom.Room.Players,
		"max_players": gameRoom.Room.MaxPlayers,
		"status":      gameRoom.Room.Status,
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
					UID:       player.UID,
					Username:  player.Username,
					Avatar:    player.Avatar,
					HandCards: nil, // 不显示具体手牌
					CardCount: player.CardCount,
					IsReady:   player.IsReady,
				}
				filteredPlayers = append(filteredPlayers, filteredPlayer)
			}
		}

		result["game_state"] = map[string]interface{}{
			"players":        filteredPlayers,
			"current_player": gameRoom.GameState.CurrentPlayer,
			"direction":      gameRoom.GameState.Direction,
			"last_card":      gameRoom.GameState.LastCard,
			"deck_count":     len(gameRoom.GameState.DrawPile),
			"status":         gameRoom.GameState.Status,
			"turn_end_time":  gameRoom.GameState.TurnEndTime,
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

	// +2/+4/Au/换向牌可随意打出，无需反应条件
	nobleGases := map[string]bool{"He": true, "Ne": true, "Ar": true, "Kr": true}
	specialTypes := map[string]bool{"+2": true, "+4": true, "Au": true}
	isSpecial := specialTypes[card.Type] || specialTypes[card.Effect] || nobleGases[card.Type]
	if !isSpecial && gameRoom.GameState.LastCard != nil {
		if !CanReact(gameRoom.GameState.LastCard.Substance, substance) {
			return errors.New("无法与上一张牌反应: " + substance + " 不与 " + gameRoom.GameState.LastCard.Substance + " 反应")
		}
	}
	// nobleGases 作为换向牌

	// 检查选中的卡牌中是否有带效果的
	activeEffect := ""
	if card.Effect != "" {
		activeEffect = card.Effect
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
	gameRoom.GameState.LastCard = &playedCard
	gameRoom.GameState.DiscardPile = append(gameRoom.GameState.DiscardPile, playedCard)

	// 加牌叠加规则
	if activeEffect == "+2" || activeEffect == "+4" {
		// 叠加pending
		gameRoom.GameState.PendingDrawCount += map[string]int{"+2": 2, "+4": 4}[activeEffect]
		gameRoom.GameState.PendingDrawTypes = append(gameRoom.GameState.PendingDrawTypes, activeEffect)
		// 传递到下家
		gameRoom.GameState.CurrentPlayer = getNextPlayer(gameRoom.GameState)
		gameRoom.GameState.TurnEndTime = time.Now().Add(30*time.Second).UnixNano() / int64(time.Millisecond)
		return nil
	} else if gameRoom.GameState.PendingDrawCount > 0 {
		// 不能继续叠加时，强制摸pending
		drawCardsForPlayer(gameRoom, gameRoom.GameState.CurrentPlayer, gameRoom.GameState.PendingDrawCount)
		gameRoom.GameState.PendingDrawCount = 0
		gameRoom.GameState.PendingDrawTypes = nil
		// 回合传递
		gameRoom.GameState.CurrentPlayer = getNextPlayer(gameRoom.GameState)
		gameRoom.GameState.TurnEndTime = time.Now().Add(30*time.Second).UnixNano() / int64(time.Millisecond)
		return nil
	}
	// 其他效果
	switch activeEffect {
	case "reverse":
		gameRoom.GameState.Direction *= -1
	case "Au":
		gameRoom.GameState.CurrentPlayer = getNextPlayer(gameRoom.GameState)
	}

	// 检查是否获胜
	if currentPlayer.CardCount == 0 {
		gameRoom.GameState.Status = "finished"
		gameRoom.Room.Status = "finished"
		// 记录游戏历史
		saveGameHistory(roomID, uid, gameRoom.Room.Players)
		return nil
	}

	// 下一位玩家
	gameRoom.GameState.CurrentPlayer = getNextPlayer(gameRoom.GameState)
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

func drawCardsForPlayer(gameRoom *GameRoom, playerIndex int, count int) {
	player := gameRoom.GameState.Players[playerIndex]
	for i := 0; i < count && len(gameRoom.GameState.DrawPile) > 0; i++ {
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

	drawCardsForPlayer(gameRoom, gameRoom.GameState.CurrentPlayer, count)

	// 摸牌后跳过回合
	gameRoom.GameState.CurrentPlayer = getNextPlayer(gameRoom.GameState)
	gameRoom.GameState.TurnEndTime = time.Now().Add(30*time.Second).UnixNano() / int64(time.Millisecond)

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
		drawCardsForPlayer(gameRoom, gameRoom.GameState.CurrentPlayer, 2)
		gameRoom.GameState.CurrentPlayer = getNextPlayer(gameRoom.GameState)
		gameRoom.GameState.TurnEndTime = time.Now().Add(30*time.Second).UnixNano() / int64(time.Millisecond)

		// 广播更新
		go func(id string) {
			websocket.GlobalHub.BroadcastToRoom(id, websocket.Message{
				Type: "game_update",
			})
		}(roomID)
	}
}
