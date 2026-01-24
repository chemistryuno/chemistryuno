package game

import (
	"chemistryuno/database"
	"chemistryuno/models"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
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
	}
}

// 创建房间
func CreateRoom(name string, hostID int, hostName string, maxPlayers int, deckID int) (*models.Room, error) {
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
		HostID:     hostID,
		Players:    []int{hostID},
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
func JoinRoom(roomID string, userID int, username string) error {
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
		if pid == userID {
			return errors.New("已在房间中")
		}
	}

	gameRoom.Room.Players = append(gameRoom.Room.Players, userID)
	return nil
}

// 离开房间
func LeaveRoom(roomID string, userID int) error {
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
		if pid != userID {
			newPlayers = append(newPlayers, pid)
		}
	}
	gameRoom.Room.Players = newPlayers

	// 如果房主离开，转移房主或删除房间
	if gameRoom.Room.HostID == userID {
		if len(newPlayers) > 0 {
			gameRoom.Room.HostID = newPlayers[0]
		} else {
			roomMutex.Lock()
			delete(rooms, roomID)
			roomMutex.Unlock()
		}
	}

	return nil
}

// 开始游戏
func StartGame(roomID string, userID int) error {
	roomMutex.RLock()
	gameRoom, exists := rooms[roomID]
	roomMutex.RUnlock()

	if !exists {
		return errors.New("房间不存在")
	}

	gameRoom.mutex.Lock()
	defer gameRoom.mutex.Unlock()

	if gameRoom.Room.HostID != userID {
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

	// 初始化玩家
	for _, pid := range gameRoom.Room.Players {
		var username, avatar string
		database.DB.QueryRow("SELECT username, avatar FROM users WHERE id = ?", pid).
			Scan(&username, &avatar)

		player := &models.PlayerState{
			UserID:    pid,
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
func GetRoomState(roomID string, userID int) (map[string]interface{}, error) {
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
		if pid == userID {
			inRoom = true
			break
		}
	}
	if !inRoom {
		return nil, errors.New("你不在该房间中")
	}

	result := map[string]interface{}{
		"room":   gameRoom.Room,
		"status": gameRoom.Room.Status,
	}

	if gameRoom.GameState != nil {
		// 过滤其他玩家的手牌
		filteredPlayers := []*models.PlayerState{}
		for _, player := range gameRoom.GameState.Players {
			if player.UserID == userID {
				// 当前玩家，显示全部信息
				filteredPlayers = append(filteredPlayers, player)
			} else {
				// 其他玩家，隐藏手牌详情
				filteredPlayer := &models.PlayerState{
					UserID:    player.UserID,
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
		}
	}

	return result, nil
}

func getCardEffect(cardType string) string {
	effects := map[string]string{
		"+2":  "+2",
		"+4":  "+4",
		"He":  "reverse",
		"Ne":  "reverse",
		"Ar":  "reverse",
		"Kr":  "reverse",
		"Au":  "skip",
	}
	return effects[cardType]
}

// 出牌
func PlayCard(roomID string, userID int, card models.Card, substance string) error {
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
	if currentPlayer.UserID != userID {
		return errors.New("还没轮到你")
	}

	// 检查是否有这张牌
	cardIndex := -1
	for i, c := range currentPlayer.HandCards {
		if c.Type == card.Type {
			cardIndex = i
			break
		}
	}
	if cardIndex == -1 {
		return errors.New("你没有这张牌")
	}

	// 检查是否能出这张牌（与上一张牌能否反应）
	if gameRoom.GameState.LastCard != nil {
		if !CanReact(gameRoom.GameState.LastCard.Substance, substance) {
			return errors.New("无法与上一张牌反应")
		}
	}

	// 移除手牌
	currentPlayer.HandCards = append(
		currentPlayer.HandCards[:cardIndex],
		currentPlayer.HandCards[cardIndex+1:]...,
	)
	currentPlayer.CardCount--

	// 添加到弃牌堆
	playedCard := models.PlayedCard{
		Card:      card,
		Substance: substance,
		PlayerID:  userID,
	}
	gameRoom.GameState.LastCard = &playedCard
	gameRoom.GameState.DiscardPile = append(gameRoom.GameState.DiscardPile, playedCard)

	// 处理特殊效果
	switch card.Effect {
	case "reverse":
		gameRoom.GameState.Direction *= -1
	case "skip":
		gameRoom.GameState.CurrentPlayer = getNextPlayer(gameRoom.GameState)
	case "+2":
		nextPlayerIndex := getNextPlayer(gameRoom.GameState)
		drawCardsForPlayer(gameRoom, nextPlayerIndex, 2)
	case "+4":
		nextPlayerIndex := getNextPlayer(gameRoom.GameState)
		drawCardsForPlayer(gameRoom, nextPlayerIndex, 4)
	}

	// 检查是否获胜
	if currentPlayer.CardCount == 0 {
		gameRoom.GameState.Status = "finished"
		gameRoom.Room.Status = "finished"
		// 记录游戏历史
		saveGameHistory(roomID, userID, gameRoom.Room.Players)
		return nil
	}

	// 下一位玩家
	gameRoom.GameState.CurrentPlayer = getNextPlayer(gameRoom.GameState)
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
func DrawCard(roomID string, userID int, count int) error {
	roomMutex.RLock()
	gameRoom, exists := rooms[roomID]
	roomMutex.RUnlock()

	if !exists {
		return errors.New("房间不存在")
	}

	gameRoom.mutex.Lock()
	defer gameRoom.mutex.Unlock()

	currentPlayer := gameRoom.GameState.Players[gameRoom.GameState.CurrentPlayer]
	if currentPlayer.UserID != userID {
		return errors.New("还没轮到你")
	}

	drawCardsForPlayer(gameRoom, gameRoom.GameState.CurrentPlayer, count)
	
	// 摸牌后跳过回合
	gameRoom.GameState.CurrentPlayer = getNextPlayer(gameRoom.GameState)
	
	return nil
}

// 获取可用物质
func GetAvailableSubstances(roomID string, userID int) ([]string, error) {
	roomMutex.RLock()
	gameRoom, exists := rooms[roomID]
	roomMutex.RUnlock()

	if !exists {
		return nil, errors.New("房间不存在")
	}

	gameRoom.mutex.RLock()
	defer gameRoom.mutex.RUnlock()

	currentPlayer := gameRoom.GameState.Players[gameRoom.GameState.CurrentPlayer]
	if currentPlayer.UserID != userID {
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

func saveGameHistory(roomID string, winnerID int, players []int) {
	playersJSON, _ := json.Marshal(players)
	database.DB.Exec(
		"INSERT INTO game_history (room_id, winner_id, players, started_at, finished_at) VALUES (?, ?, ?, datetime('now', '-1 hour'), datetime('now'))",
		roomID, winnerID, string(playersJSON),
	)
	fmt.Println("游戏历史已保存")
}
