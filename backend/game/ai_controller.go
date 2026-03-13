package game

import (
	"chemistryuno/backend/models"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/websocket"
	"fmt"
	"log"
	"math/rand"
	"time"
)

// TriggerAITurn 触发 AI 回合逻辑
func (gr *GameRoom) TriggerAITurn() {
	gameRoom := gr

	gr.mutex.RLock()
	if gameRoom.GameState == nil || gameRoom.GameState.Status != "playing" {
		gr.mutex.RUnlock()
		return
	}
	targetUID := gameRoom.GameState.Players[gameRoom.GameState.CurrentPlayer].UID
	gr.mutex.RUnlock()

	// 增加 1 秒模拟思考/等待时间，让玩家看清操作
	time.Sleep(1 * time.Second)

	gameRoom.mutex.Lock()

	// 再次检查是否仍是该同一个 AI 的回合（由于 Sleep 期间可能已经发生了变化）
	if gameRoom.GameState.Status != "playing" {
		gameRoom.mutex.Unlock()
		return
	}
	currentPlayer := gameRoom.GameState.Players[gameRoom.GameState.CurrentPlayer]
	if currentPlayer.UID != targetUID {
		gameRoom.mutex.Unlock()
		return
	}
	if !currentPlayer.IsAI {
		gameRoom.mutex.Unlock()
		return
	}

	log.Printf("[AI] 🤖 AI %d 立即行动...", currentPlayer.UID)
	gr.BroadcastSystemMessage(fmt.Sprintf("%s 正在实验台前审视战局...", currentPlayer.Nickname))

	// 教学脚本模式：按照脚本执行固定操作
	if gr.GameState.TutorialScriptMode {
		gameRoom.mutex.Unlock()
		gr.executeTutorialScript()
		return
	}

	// 0. 特殊情况：如果场上没牌 (开局或金卡)，强制打出最复杂的可能物质
	// 这不计入难度概率，是硬性指令，确保 AI 开局有气势
	isAllowedAny := gr.GameState.AllowedAnyPlayer == gr.GameState.CurrentPlayer
	if gr.GameState.LastCard == nil || isAllowedAny {
		log.Printf("[AI] 🏗️ 场面为空或拥有特权，AI %d 执行开局最优策略", currentPlayer.UID)
		if played := gr.aiTryPlayCard(currentPlayer); played {
			// aiTryPlayCard 已经释放锁
			return
		}
	}

	// 1. 难度判定：决定是否"尝试"寻找最优解
	difficulty := gameRoom.Room.PvEDifficulty
	// 如果不是 PvE 房间，尝试获取 AI 补位难度
	if !gameRoom.Room.IsPvE {
		if gameRoom.Room.EnableAIBackfill {
			difficulty = gameRoom.Room.AIBackfillDifficulty
		} else {
			// 兜底难度
			difficulty = 70
		}
	}

	if gameRoom.Room.IsPointsMode {
		difficulty += 10 // 积分模式更加激进
	}
	if difficulty > 100 {
		difficulty = 100
	}
	if difficulty < 20 {
		difficulty = 20 // 确保最低智力
	}

	shouldTryBest := rand.Intn(100) < difficulty

	if shouldTryBest {
		// 如果是积分模式且双联反应就绪，优先尝试双联
		if currentPlayer.DoubleActionAvailable {
			if played := gr.aiTryDoublePlay(currentPlayer); played {
				// aiTryDoublePlay 已经释放锁
				return
			}
		}

		// 尝试寻找出牌
		if played := gr.aiTryPlayCard(currentPlayer); played {
			// aiTryPlayCard 已经释放锁
			return
		}
	} else {
		log.Printf("[AI] 🎲 AI %d 判定为非最优决策，尝试随机出牌", currentPlayer.UID)
		// 非最优决策时：随机打出一张可出的牌，不考虑反应链条最长或攻击性
		if played := gr.aiTryPlayRandom(currentPlayer); played {
			// aiTryPlayRandom 已经释放锁
			return
		}
	}

	// 摸牌逻辑 (摸2张并过牌)
	log.Printf("[AI] 🛑 AI %s (%d) 摸牌", currentPlayer.Nickname, currentPlayer.UID)
	gr.BroadcastSystemMessage(fmt.Sprintf("%s 暂时无法合成目标，决定进入库房寻找灵感（摸 2 张牌）。", currentPlayer.Nickname))
	gr.aiDrawCard(currentPlayer.UID)
	// aiDrawCard 已经释放锁
}

// aiTryPlayRandom 随机尝试出牌（低难度逻辑）
func (gr *GameRoom) aiTryPlayRandom(player *models.PlayerState) bool {
	// 1. 打乱手牌顺序进行随机尝试
	hand := make([]models.Card, len(player.HandCards))
	copy(hand, player.HandCards)
	rand.Shuffle(len(hand), func(i, j int) { hand[i], hand[j] = hand[j], hand[i] })

	// 2. 依次尝试每一张牌
	for _, card := range hand {
		if isSpecialCard(card) {
			gr.aiExecutePlay(player.UID, card, "")
			return true
		}

		// 普通牌，尝试作为单质出牌（最简单逻辑）
		substance := card.Type
		diatomic := map[string]string{
			"H": "H2", "O": "O2", "N": "N2", "Cl": "Cl2", "F": "F2", "Br": "Br2", "I": "I2",
		}
		if s, ok := diatomic[substance]; ok {
			substance = s
		}

		// 校验是否能出
		if gr.canAIPlaySubstance(substance) {
			gr.aiExecutePlay(player.UID, card, substance)
			return true
		}
	}
	return false
}

func (gr *GameRoom) canAIPlaySubstance(substance string) bool {
	if gr.GameState.LastCard == nil || gr.GameState.AllowedAnyPlayer == gr.GameState.CurrentPlayer {
		return true
	}

	// 检查是否能与场上物质反应
	lastSub := gr.GameState.LastCard.Substance
	// 如果是双联，LastCard.Reactants 会有值
	if len(gr.GameState.LastCard.Reactants) > 0 {
		for _, r := range gr.GameState.LastCard.Reactants {
			if CanReact(r, substance) {
				return true
			}
		}
		return false
	}

	return CanReact(lastSub, substance)
}

// aiDrawCard AI 摸牌/过牌
// 注意：此函数假设调用者持有 gr.mutex 锁，函数内部会释放锁
func (gr *GameRoom) aiDrawCard(uid int) {
	// 获取 AI 昵称（在释放锁前获取，因为需要访问GameState）
	displayName := fmt.Sprintf("AI 研究员 %d", -uid)
	for _, p := range gr.GameState.Players {
		if p.UID == uid {
			displayName = p.Nickname
			break
		}
	}

	// 释放锁，调用 DrawCard (它会重新加锁)
	gr.mutex.Unlock()
	err := DrawCard(gr.Room.ID, uid, 2) // AI 不出牌时立即摸 2 张并自动过牌

	if err != nil {
		log.Printf("[AI] ❌ AI %s 摸牌失败: %v", displayName, err)
	} else {
		log.Printf("[AI] 🛑 AI %s 摸牌结束回合", displayName)
		// DrawCard 内部已经调用了 CheckNextTurnAI 和状态更新
		// 不需要在这里额外操作
	}
}

// aiTryDoublePlay 尝试发动双联反应
func (gr *GameRoom) aiTryDoublePlay(player *models.PlayerState) bool {
	// ... (content omitted for j loop fix)
	// 获取手牌能组成的所有物质
	availableSubstances := GetSubstancesFromElements(player.HandCards)
	if len(availableSubstances) < 2 {
		return false
	}

	// 准备手牌元素映射以便快速校验
	handElements := make(map[string]int)
	for _, c := range player.HandCards {
		if c.Effect == "" {
			handElements[c.Type]++
		}
	}

	// 尝试寻找两两反应的可能
	for i := 0; i < len(availableSubstances); i++ {
		for j := i + 1; j < len(availableSubstances); j++ {
			s1, s2 := availableSubstances[i], availableSubstances[j]

			// 排除特殊牌 (双联必须是普通物质化学反应，禁止功能牌和稀有气体)
			if isNobleGas(s1) || isNobleGas(s2) || s1 == "Au" || s1 == "+2" || s1 == "+4" || s1 == "reverse" || s1 == "skip" ||
				s2 == "Au" || s2 == "+2" || s2 == "+4" || s2 == "reverse" || s2 == "skip" {
				continue
			}

			// 校验手牌资源：现在 DoublePlay 只要求有对应元素种类，不计系数
			req1 := parseSubstance(s1)
			req2 := parseSubstance(s2)

			// 合并所有需要的元素种类
			allReqs := make(map[string]int)
			for k := range req1 {
				allReqs[k] = 1
			}
			for k := range req2 {
				allReqs[k] = 1
			}

			enough := true
			for k := range allReqs {
				if handElements[k] <= 0 {
					enough = false
					break
				}
			}
			if !enough {
				continue
			}

			// 查询这两个物质是否能反应
			canReact, err := repository.ReactionRepo.CheckReactionExists(s1, s2)
			if err == nil && canReact {
				// 发动双联！
				log.Printf("[AI] ⚡ AI %d 发动双联反应: %s + %s", player.UID, s1, s2)
				gr.BroadcastSystemMessage(fmt.Sprintf("%s 正在进行极端实验：联联动双向合成（%s + %s）！", player.Nickname, s1, s2))

				// 解锁，调用 DoublePlay（DoublePlay 会自己管理锁）
				gr.mutex.Unlock()
				err := DoublePlay(gr.Room.ID, player.UID, s1, s2)

				if err == nil {
					// 广播消息（不需要锁）
					if websocket.GlobalHub != nil {
						websocket.GlobalHub.BroadcastToRoom(gr.Room.ID, websocket.Message{
							Type: "action_toast",
							Data: fmt.Sprintf("%s 发动了双联反应: %s + %s！", player.Nickname, s1, s2),
						})
					}
					// 成功执行，已释放锁
					return true
				} else {
					log.Printf("[AI] ⚠️ AI 双联执行失败: %v", err)
					// 失败时重新获取锁，让调用者继续尝试其他策略
					gr.mutex.Lock()
					return false
				}
			}
		}
	}
	return false
}

// aiTryPlayCard 尝试打出一张牌
func (gr *GameRoom) aiTryPlayCard(player *models.PlayerState) bool {
	// 0. 威胁检测：是否有人类玩家快赢了（手牌数 <= 3）
	humanThreatIdx := -1
	minHumanCards := 999
	for i, ps := range gr.GameState.Players {
		if ps.UID > 0 { // 是真人玩家
			isFinished := false
			for _, fuid := range gr.GameState.FinishedPlayers {
				if ps.UID == fuid {
					isFinished = true
					break
				}
			}
			if !isFinished {
				if ps.CardCount < minHumanCards {
					minHumanCards = ps.CardCount
				}
				if ps.CardCount <= 3 {
					humanThreatIdx = i
					break
				}
			}
		}
	}

	// 1. 优先处理加牌堆叠 (+2/+4)
	if gr.GameState.PendingDrawCount > 0 {
		// 针对性策略：如果下家是人类，检查其手牌中的加牌防御力
		humanNextIdx := -1
		humanDrawDefense := 0
		nextIdx := getNextPlayer(gr.GameState)
		nextPlayer := gr.GameState.Players[nextIdx]
		if nextPlayer.UID > 0 {
			humanNextIdx = nextIdx
			for _, c := range nextPlayer.HandCards {
				if c.Effect == "+2" || c.Effect == "+4" {
					humanDrawDefense++
				}
			}
		}

		var bestDrawCard *models.Card
		for i, card := range player.HandCards {
			if card.Effect == "+2" || card.Effect == "+4" {
				// 如果必须接牌，且下家是已经没防御牌的人类，优先打出 +4 扩大战果
				if humanNextIdx != -1 && humanDrawDefense == 0 {
					if card.Effect == "+4" {
						bestDrawCard = &player.HandCards[i]
						break
					}
				}
				// 否则保留当前遍历到的最好的一张（优先打 +2 试探，高端玩家倾向于把 +4 留给绝杀）
				if bestDrawCard == nil || (card.Effect == "+4" && bestDrawCard.Effect == "+2") {
					bestDrawCard = &player.HandCards[i]
				}
			}
		}

		if bestDrawCard != nil {
			gr.aiExecutePlay(player.UID, *bestDrawCard, "")
			return true
		}
		// 没加牌，只能摸牌
		return false
	}

	// 获取难度参数
	difficulty := gr.Room.PvEDifficulty
	if !gr.Room.IsPvE {
		if gr.Room.EnableAIBackfill {
			difficulty = gr.Room.AIBackfillDifficulty
		} else {
			difficulty = 70 // 默认难度
		}
	}

	// 难度影响权重因子（难度越高，协作和卡位策略越强）
	difficultyFactor := float64(difficulty) / 100.0

	// 合作逻辑所需数据
	nextIdx := getNextPlayer(gr.GameState)
	nextPlayer := gr.GameState.Players[nextIdx]
	isNextAI := nextPlayer.UID < 0

	// 统计所有AI队友的信息
	aiTeammates := []struct {
		index          int
		player         *models.PlayerState
		availableSubst []string
		cardCount      int
	}{}
	for i, ps := range gr.GameState.Players {
		if ps.UID < 0 && ps.UID != player.UID {
			isFinished := false
			for _, fuid := range gr.GameState.FinishedPlayers {
				if ps.UID == fuid {
					isFinished = true
					break
				}
			}
			if !isFinished {
				aiTeammates = append(aiTeammates, struct {
					index          int
					player         *models.PlayerState
					availableSubst []string
					cardCount      int
				}{
					index:          i,
					player:         ps,
					availableSubst: GetSubstancesFromElements(ps.HandCards),
					cardCount:      ps.CardCount,
				})
			}
		}
	}

	// 2. 紧急防御逻辑：如果下一位是威胁中的人类玩家，优先打出功能牌进行拦截
	if humanThreatIdx == nextIdx {
		// 尝试打出功能牌拦截
		for _, card := range player.HandCards {
			// 优先级高的拦截牌
			if card.Effect == "+4" || card.Effect == "+2" || card.Effect == "Au" {
				log.Printf("[AI] 🛡️ AI %d 检测到人类威胁，执行顶级拦截: %s", player.UID, card.Effect)
				gr.aiExecutePlay(player.UID, card, "")
				return true
			}
		}
		// 备选拦截牌：稀有气体 (Skip) 或 转向 (Reverse)
		for _, card := range player.HandCards {
			if isNobleGas(card.Type) || card.Effect == "reverse" {
				log.Printf("[AI] 🛡️ AI %d 检测到人类威胁，执行次级拦截: %s", player.UID, card.Type)
				gr.aiExecutePlay(player.UID, card, card.Type)
				return true
			}
		}
	}

	// 3. 核心决策逻辑：优先寻找能组成的最优物质
	availableSubstances := GetSubstancesFromElements(player.HandCards)

	bestSub := ""
	bestScore := -1

	// 判断场上限制
	isAllowedAny := gr.GameState.AllowedAnyPlayer == gr.GameState.CurrentPlayer
	lastSubstances := []string{}
	if gr.GameState.LastCard != nil && !isAllowedAny {
		if len(gr.GameState.LastCard.Reactants) > 0 {
			lastSubstances = append(lastSubstances, gr.GameState.LastCard.Reactants...)
		} else if gr.GameState.LastCard.Substance != "" {
			lastSubstances = append(lastSubstances, gr.GameState.LastCard.Substance)
		}
	}

	// 增强评分函数：更智能的协作和卡位策略
	calculateScore := func(sub string) int {
		score := getComplexity(sub) * 10

		// === 队友协作逻辑（难度越高协作越强） ===
		teamCoopBonus := 0
		if len(aiTeammates) > 0 {
			// 检查所有AI队友能否接上
			canCoopCount := 0
			closestAllyCards := 999
			for _, ally := range aiTeammates {
				canReact := false
				for _, as := range ally.availableSubst {
					if CanReact(sub, as) {
						canReact = true
						break
					}
				}
				if canReact {
					canCoopCount++
					if ally.cardCount < closestAllyCards {
						closestAllyCards = ally.cardCount
					}
				}
			}

			// 协作加分：能与更多队友配合，分数越高
			teamCoopBonus = canCoopCount * int(50*difficultyFactor)

			// 如果有队友快赢了（手牌<=3），极大提升协作权重
			if closestAllyCards <= 3 && canCoopCount > 0 {
				teamCoopBonus += int(100 * difficultyFactor)
				log.Printf("[AI协作] AI %d 检测到队友快赢（%d张牌），提升协作权重: %s", player.UID, closestAllyCards, sub)
			}
		}

		// === 对手卡位逻辑（难度越高卡位越精准） ===
		enemyBlockBonus := 0
		if !isNextAI {
			allyAvailable := GetSubstancesFromElements(nextPlayer.HandCards)
			canHumanReact := false
			for _, as := range allyAvailable {
				if CanReact(sub, as) {
					canHumanReact = true
					break
				}
			}
			// 如果人类接不上，加分（鼓励卡位）
			if !canHumanReact {
				enemyBlockBonus = int(60 * difficultyFactor)
			}

			// 如果人类威胁在下家，且虽然接得上但 AI 想要强制重置（通过卡位或计谋），提升卡位权重
			if humanThreatIdx == nextIdx && !canHumanReact {
				enemyBlockBonus += int(150 * difficultyFactor) // 极力卡位
				log.Printf("[AI卡位] AI %d 极力卡位威胁人类(%d张牌): %s", player.UID, nextPlayer.CardCount, sub)
			}
		}

		// === 全局战术加分 ===
		// 如果人类整体威胁较大（最少手牌数<=3），AI应该更激进
		globalTacticBonus := 0
		if minHumanCards <= 3 {
			// 倾向于打出更复杂的物质（清理手牌）和功能牌
			globalTacticBonus = int(30 * difficultyFactor)
		}

		totalScore := score + teamCoopBonus + enemyBlockBonus + globalTacticBonus
		return totalScore
	}

	if len(lastSubstances) > 0 {
		// 有场上物质限制，必须能与其反应
		for _, ls := range lastSubstances {
			for _, sub := range availableSubstances {
				if CanReact(ls, sub) {
					score := calculateScore(sub)
					if score > bestScore {
						bestScore = score
						bestSub = sub
					}
				}
			}
		}
	} else {
		// 场上无物质 (开局/Au/AllowedAny)，指令要求：优先打出自己能打出的最复杂的物质
		for _, sub := range availableSubstances {
			// 在无物质限制时，极大提高复杂度的权重，确保 AI 优先清理多张手牌
			score := getComplexity(sub)*100 + calculateScore(sub)
			if score > bestScore {
				bestScore = score
				bestSub = sub
			}
		}
	}

	// 如果找到了最佳出牌物质
	if bestSub != "" {
		log.Printf("[AI] 🧠 AI %d 采用智能协作策略: %s (综合得分: %d, 难度: %d)", player.UID, bestSub, bestScore, difficulty)
		gr.aiExecutePlay(player.UID, models.Card{Type: "AI_BEST_CHOICE_CARD"}, bestSub)
		return true
	}

	// 4. 特殊功能牌优先级逻辑
	// 在配合逻辑下，对攻击牌的打出时机进行微调
	specialPriority := []string{"reverse", "Au"}

	// 稀有气体也作为特殊优先级
	nobleGases := []string{"He", "Ne", "Ar", "Kr", "Xe", "Rn"}

	// 获取人类玩家的加牌防御信息
	humanDrawDefense := 0
	if !isNextAI {
		for _, c := range nextPlayer.HandCards {
			if c.Effect == "+2" || c.Effect == "+4" {
				humanDrawDefense++
			}
		}
	}

	// 如果下家是人类，或者全场有人类威胁，将拦截/攻击牌加入优先级
	if !isNextAI || humanThreatIdx != -1 {
		// 针对性策略：如果下家人类手里没有加牌了，且我手里有加牌，极大提升加牌优先级
		if !isNextAI && humanDrawDefense == 0 {
			specialPriority = append([]string{"+4", "+2"}, specialPriority...)
		} else {
			specialPriority = append(specialPriority, "+2", "+4")
		}
		// 稀有气体随后
		specialPriority = append(specialPriority, nobleGases...)
	} else {
		// 如果下家是 AI 队友
		// 情况 A：我快赢了（手牌少），为了出完牌清空手牌，必须打出功能牌
		if player.CardCount <= 3 {
			specialPriority = append(specialPriority, "+2", "+4")
			specialPriority = append(specialPriority, nobleGases...)
		} else {
			// 情况 B：我手牌还多，将攻击性牌（+2, +4）优先级调到最低，避免误伤队友
			// 但仍需包含在内作为最后的出牌手段，否则 AI 会因"不想误伤"而被迫摸牌
			specialPriority = append(specialPriority, nobleGases...)
			specialPriority = append(specialPriority, "+2", "+4")
		}
	}

	for _, effect := range specialPriority {
		for _, card := range player.HandCards {
			if card.Effect == effect || card.Type == effect {
				sub := ""
				if isNobleGas(card.Type) {
					sub = card.Type
				}
				gr.aiExecutePlay(player.UID, card, sub)
				return true
			}
		}
	}

	// 兜底：如果存在人类威胁且手中有转向牌，直接打出以自保/拦截
	if humanThreatIdx != -1 {
		for _, card := range player.HandCards {
			if card.Effect == "reverse" {
				gr.aiExecutePlay(player.UID, card, "")
				return true
			}
		}
	}

	return false
}

// aiTryPlaySpecialOnly 只尝试打出功能牌
func (gr *GameRoom) aiTryPlaySpecialOnly(player *models.PlayerState) bool {
	for _, card := range player.HandCards {
		if isSpecialCard(card) {
			gr.aiExecutePlay(player.UID, card, "")
			return true
		}
	}
	return false
}

// aiExecutePlay 执行出牌 (封装 PlayCard 调用)
// 注意：此函数假设调用者持有 gr.mutex 锁，函数内部会释放锁
func (gr *GameRoom) aiExecutePlay(uid int, card models.Card, substance string) {
	cardName := card.Type
	if card.Effect != "" {
		cardName = card.Effect
	}

	// 如果没有提供 substance 且是功能牌，则保持为空让 PlayCard 自动处理
	// 但如果提供了 (比如 Noble Gas)，则保留它
	if substance == "" && isSpecialCard(card) {
		substance = ""
	}

	// 解锁，调用 PlayCard（PlayCard 会自己管理锁）
	gr.mutex.Unlock()
	err := PlayCard(gr.Room.ID, uid, card, substance)

	if err != nil {
		log.Printf("[AI] ⚠️ AI %d 出牌失败 (%s): %v", uid, substance, err)
		// 失败时重新获取锁，然后摸牌
		gr.mutex.Lock()
		gr.aiDrawCard(uid)
		// aiDrawCard 会释放锁
	} else {
		log.Printf("[AI] ✅ AI %d 成功出牌: %s (物质: %s)", uid, cardName, substance)
		// 发送广播
		nickname := "AI"
		gr.mutex.RLock()
		for _, p := range gr.GameState.Players {
			if p.UID == uid {
				nickname = p.Nickname
				break
			}
		}
		gr.mutex.RUnlock()
		gr.BroadcastSystemMessage(fmt.Sprintf("%s 熟练地在反应皿中投入了 %s (%s)。", nickname, cardName, substance))
		// PlayCard 内部已经调用了 CheckNextTurnAI，不需要重新获取锁
	}
}

// 辅助函数

func isSpecialCard(card models.Card) bool {
	return card.Effect != "" || isNobleGas(card.Type)
}

func isNobleGas(t string) bool {
	return t == "He" || t == "Ne" || t == "Ar" || t == "Kr" || t == "Xe" || t == "Rn"
}

// getComplexity 计算物质的复杂度 (综合考量原子总数与元素多样性)
func getComplexity(substance string) int {
	req := parseSubstance(substance)
	totalAtoms := 0
	distinctElements := len(req)

	for _, count := range req {
		totalAtoms += count
	}

	// 复杂度公式：总原子数 + 元素种类 * 5
	// 这样 C12H22O11 (45+3*5=60) 会远高于 O2 (2+1*5=7)
	return totalAtoms + (distinctElements * 5)
}

// hasCards 检查手牌是否满足需求 (只需包含相关元素即可，不限制数量)
func hasCards(hand []models.Card, needed map[string]int) bool {
	// 统计手牌中的元素种类
	handElements := make(map[string]bool)
	for _, c := range hand {
		handElements[c.Type] = true
	}

	// 检查所需的所有元素是否都在手牌中
	for elem := range needed {
		if !handElements[elem] {
			return false
		}
	}
	return true
}

// executeTutorialScript 执行教学脚本中的AI行动
func (gr *GameRoom) executeTutorialScript() {
	log.Printf("[教学脚本] 🎯 AI触发教学脚本执行")

	gr.mutex.RLock()
	if gr.GameState == nil || gr.GameState.Status != "playing" {
		gr.mutex.RUnlock()
		return
	}

	currentStep := gr.GameState.TutorialCurrentStep
	currentIndex := gr.GameState.CurrentPlayer
	if currentIndex < 0 || currentIndex >= len(gr.GameState.Players) {
		gr.mutex.RUnlock()
		return
	}
	currentPlayer := gr.GameState.Players[currentIndex]
	currentUID := currentPlayer.UID
	currentNickname := currentPlayer.Nickname
	currentCardCount := currentPlayer.CardCount
	roomID := gr.Room.ID
	gr.mutex.RUnlock()

	log.Printf("[教学脚本] 📊 当前步骤: %d", currentStep)
	currentScriptStep := getTutorialScriptStep(currentStep)

	if currentScriptStep == nil {
		log.Printf("[教学脚本] ⚠️  步骤 %d 不存在，跳过AI行动", currentStep)
		return
	}

	if currentScriptStep.Player != "ai" {
		log.Printf("[教学脚本] 步骤 %d 不是AI回合，跳过", currentStep)
		return
	}

	log.Printf("[教学脚本] 🤖 执行步骤 %d: AI %s %s", currentStep, currentScriptStep.Action, currentScriptStep.Substance)
	log.Printf("[教学脚本] 当前AI玩家: %s (UID:%d), 手牌数: %d", currentNickname, currentUID, currentCardCount)

	if currentScriptStep.Action == "play" {
		substance := currentScriptStep.Substance
		// 🎓 教学脚本模式：AI 强制从虚空中打出指定物质，不消耗实际手牌以防万一
		// 构造一张临时虚拟卡牌进行出牌
		virtualCard := models.Card{Type: substance, Effect: getCardEffect(substance)}
		if err := PlayCard(roomID, currentUID, virtualCard, substance); err != nil {
			log.Printf("[教学脚本] ⚠️  AI脚本强制出牌失败(%s): %v", substance, err)
			return
		}
		log.Printf("[教学脚本] ✅ AI强制打出了 %s", substance)
		return
	}

	if currentScriptStep.Action == "draw" {
		// AI 摸一张牌
		if err := DrawCard(roomID, currentUID, 1); err != nil {
			log.Printf("[教学脚本] ⚠️  AI脚本摸牌失败: %v", err)
			return
		}
		log.Printf("[教学脚本] ✅ AI执行了摸牌")
		return
	}

	log.Printf("[教学脚本] ⚠️  当前步骤动作 %s 不受支持", currentScriptStep.Action)
}
