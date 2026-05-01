package game

import (
	"chemistryuno/backend/models"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/websocket"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
)

// pluginCardRegistry 插件卡牌内存注册表，symbol -> PluginCardDef
var pluginCardRegistry = make(map[string]*models.PluginCardDef)
var pluginCardsMutex sync.RWMutex

// LoadPluginCards 从数据库加载所有激活插件的卡牌到内存 registry
func LoadPluginCards() {
	cards, err := repository.PluginRepo.GetAllActiveCards()
	if err != nil {
		log.Printf("[Plugin] ⚠️  加载插件卡牌失败: %v", err)
		return
	}

	newRegistry := make(map[string]*models.PluginCardDef)
	for _, c := range cards {
		def := &models.PluginCardDef{
			ID:           c.ID,
			PluginID:     c.PluginID,
			Symbol:       c.Symbol,
			DisplayName:  c.DisplayName,
			EffectType:   c.EffectType,
			EffectConfig: c.EffectConfig,
			DefaultCount: c.DefaultCount,
			Color:        c.Color,
		}
		newRegistry[c.Symbol] = def
	}

	pluginCardsMutex.Lock()
	pluginCardRegistry = newRegistry
	pluginCardsMutex.Unlock()

	log.Printf("[Plugin] ✅ 已加载 %d 个插件卡牌到 registry", len(newRegistry))
}

// IsPluginCard 判断卡牌符号是否为插件卡牌
func IsPluginCard(symbol string) bool {
	pluginCardsMutex.RLock()
	defer pluginCardsMutex.RUnlock()
	_, ok := pluginCardRegistry[symbol]
	return ok
}

// GetPluginCard 获取插件卡牌定义
func GetPluginCard(symbol string) *models.PluginCardDef {
	pluginCardsMutex.RLock()
	defer pluginCardsMutex.RUnlock()
	return pluginCardRegistry[symbol]
}

// GetAllPluginCards 获取所有插件卡牌（用于 deck builder）
func GetAllPluginCards() []*models.PluginCardDef {
	pluginCardsMutex.RLock()
	defer pluginCardsMutex.RUnlock()
	result := make([]*models.PluginCardDef, 0, len(pluginCardRegistry))
	for _, def := range pluginCardRegistry {
		result = append(result, def)
	}
	return result
}

// ExecutePluginEffect 执行插件卡牌效果（在 GameRoom 锁内调用，无需再加锁）
// playerIdx 为当前玩家在 GameState.Players 中的索引
func ExecutePluginEffect(gameRoom *GameRoom, playerIdx int, symbol string) error {
	def := GetPluginCard(symbol)
	if def == nil {
		return errors.New("未知的插件卡牌: " + symbol)
	}

	player := gameRoom.GameState.Players[playerIdx]
	cardName := def.DisplayName
	if cardName == "" {
		cardName = def.Symbol
	}

	var toastMsg string

	switch def.EffectType {
	case "swap":
		var cfg models.SwapConfig
		if err := json.Unmarshal([]byte(def.EffectConfig), &cfg); err != nil {
			return errors.New("swap 配置解析失败: " + err.Error())
		}
		if err := executeSwap(gameRoom, playerIdx, cfg.Count); err != nil {
			return err
		}
		toastMsg = fmt.Sprintf("🔄 %s 使用了「%s」，随机交换了 %d 张手牌！", player.Nickname, cardName, cfg.Count)

	case "force_play":
		var cfg models.ForcePlayConfig
		if err := json.Unmarshal([]byte(def.EffectConfig), &cfg); err != nil {
			return errors.New("force_play 配置解析失败: " + err.Error())
		}
		executeForcePlay(gameRoom.GameState, cfg.Count)
		// 确定下一位玩家
		nextIdx := getNextPlayer(gameRoom.GameState)
		nextPlayer := gameRoom.GameState.Players[nextIdx]
		toastMsg = fmt.Sprintf("⚡ %s 使用了「%s」，%s 必须额外打出 %d 张牌！", player.Nickname, cardName, nextPlayer.Nickname, cfg.Count)

	case "convert":
		var cfg models.ConvertConfig
		if err := json.Unmarshal([]byte(def.EffectConfig), &cfg); err != nil {
			return errors.New("convert 配置解析失败: " + err.Error())
		}
		if err := executeConvert(gameRoom, playerIdx, symbol, cfg.SourceCount, cfg.TargetCount); err != nil {
			return err
		}
		toastMsg = fmt.Sprintf("🔁 %s 使用了「%s」，消耗 %d 张换取 %d 张新牌！", player.Nickname, cardName, cfg.SourceCount, cfg.TargetCount)

	default:
		return errors.New("不支持的效果类型: " + def.EffectType)
	}

	// 广播效果提示到房间
	if toastMsg != "" && websocket.GlobalHub != nil {
		websocket.GlobalHub.BroadcastToRoom(gameRoom.Room.ID, websocket.Message{
			Type: "action_toast",
			Data: toastMsg,
		})
	}

	return nil
}

// executeSwap 随机将 count 张手牌放回摸牌堆，再摸 count 张新牌
func executeSwap(gameRoom *GameRoom, playerIdx int, count int) error {
	gs := gameRoom.GameState
	player := gs.Players[playerIdx]

	handLen := len(player.HandCards)
	if handLen == 0 {
		return errors.New("手牌为空，无法交换")
	}

	// 实际交换数量不超过手牌数
	actualCount := count
	if actualCount > handLen {
		actualCount = handLen
	}

	// 随机选 actualCount 张手牌放回摸牌堆
	indices := rand.Perm(handLen)[:actualCount]
	// 从大到小排序，方便倒序删除
	for i := 0; i < len(indices)-1; i++ {
		for j := i + 1; j < len(indices); j++ {
			if indices[i] < indices[j] {
				indices[i], indices[j] = indices[j], indices[i]
			}
		}
	}

	var swappedCards []models.Card
	for _, idx := range indices {
		swappedCards = append(swappedCards, player.HandCards[idx])
		player.HandCards = append(player.HandCards[:idx], player.HandCards[idx+1:]...)
	}
	player.CardCount = len(player.HandCards)

	// 放回摸牌堆（洗入）
	gs.DrawPile = append(gs.DrawPile, swappedCards...)
	rand.Shuffle(len(gs.DrawPile), func(i, j int) {
		gs.DrawPile[i], gs.DrawPile[j] = gs.DrawPile[j], gs.DrawPile[i]
	})

	// 摸 actualCount 张新牌
	drawCardsForPlayer(gameRoom, playerIdx, actualCount)

	log.Printf("[Plugin] 🔄 swap: 玩家 %s 交换了 %d 张手牌", player.Nickname, actualCount)

	return nil
}

// executeForcePlay 设置下一位玩家必须额外出 count 张牌
func executeForcePlay(gs *models.GameState, count int) {
	gs.PendingForcedPlays = count
	log.Printf("[Plugin] ⚡ force_play: 下一位玩家须额外打出 %d 张牌", count)
}

// executeConvert 消耗手牌中 sourceCount 张该卡牌，摸取 targetCount 张新牌
// 注意：出牌时已经消耗了 1 张，这里额外再消耗 sourceCount-1 张
func executeConvert(gameRoom *GameRoom, playerIdx int, symbol string, sourceCount int, targetCount int) error {
	gs := gameRoom.GameState
	player := gs.Players[playerIdx]

	// 需要额外消耗的数量（出牌时已消耗 1 张）
	extraNeeded := sourceCount - 1
	if extraNeeded < 0 {
		extraNeeded = 0
	}

	// 统计手牌中该符号的数量
	var matchIndices []int
	for i, c := range player.HandCards {
		if c.Type == symbol {
			matchIndices = append(matchIndices, i)
		}
	}

	if len(matchIndices) < extraNeeded {
		return errors.New("手牌中该卡牌数量不足以完成转换")
	}

	// 从大到小倒序删除额外的卡牌
	for i := extraNeeded - 1; i >= 0; i-- {
		idx := matchIndices[i]
		gs.AllUsedCards = append(gs.AllUsedCards, player.HandCards[idx])
		player.HandCards = append(player.HandCards[:idx], player.HandCards[idx+1:]...)
	}
	player.CardCount = len(player.HandCards)

	// 摸取 targetCount 张新牌
	drawCardsForPlayer(gameRoom, playerIdx, targetCount)

	log.Printf("[Plugin] 🔁 convert: 玩家 %s 消耗 %d 张 %s，摸取 %d 张新牌",
		player.Nickname, sourceCount, symbol, targetCount)

	return nil
}
