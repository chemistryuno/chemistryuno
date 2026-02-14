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

	// 延时执行，模拟思考时间 (1.5秒)
	time.AfterFunc(1500*time.Millisecond, func() {
		gameRoom.mutex.Lock()
		defer gameRoom.mutex.Unlock()

		// 再次检查是否仍是该 AI 的回合 (防止状态变更)
		if gameRoom.GameState.Status != "playing" {
			return
		}
		currentPlayer := gameRoom.GameState.Players[gameRoom.GameState.CurrentPlayer]
		if currentPlayer.UID >= 0 { // 只有 UID < 0 才是 AI
			return
		}

		log.Printf("[AI] 🤖 AI %d 正在思考... (难度: %d)", currentPlayer.UID, gameRoom.Room.PvEDifficulty)

		// 1. 难度判定：决定是否"尝试"寻找最优解
		// 难度 hitRate = PvEDifficulty (1-100)
		// r <= Difficulty: AI 尝试寻找可出的反应牌或功能牌
		// r > Difficulty: AI "犯错"，只尝试打出无脑牌(功能牌)或者直接摸牌，即使手里有能反应的牌也不出
		shouldTryBest := rand.Intn(100) < gameRoom.Room.PvEDifficulty

		if shouldTryBest {
			// 尝试寻找出牌
			if played := gr.aiTryPlayCard(currentPlayer); played {
				return
			}
		} else {
			log.Printf("[AI] 🎲 AI %d 判定为失误/消极操作", currentPlayer.UID)
			// 即使失误，如果有 +2/+4/Au 这种无脑牌，也有一定概率打出 (50%)
			if rand.Intn(100) < 50 {
				if played := gr.aiTryPlaySpecialOnly(currentPlayer); played {
					return
				}
			}
		}

		// 摸牌逻辑
		log.Printf("[AI] 🛑 AI %d 决定摸/无法出牌", currentPlayer.UID)
		gr.aiDrawCard(currentPlayer.UID)
	})
}

// aiDrawCard AI 摸牌/过牌
func (gr *GameRoom) aiDrawCard(uid int) {
	// 释放锁，调用 DrawCard (它会重新加锁)
	gr.mutex.Unlock()
	err := DrawCard(gr.Room.ID, uid, 0) // 0 表示按规则默认数量
	gr.mutex.Lock()

	if err != nil {
		log.Printf("[AI] ❌ AI 摸牌失败: %v", err)
	} else {
		// DrawCard 内部会处理 NextTurn，这里不需要额外操作
		// 注意：DrawCard 可能会导致下家也是 AI，会递归调用 TriggerAITurn (通过 broadcastUpdate -> 前端? 不，后端自循环最好)
		// 我们需要在 DrawCard 结束时检查下家
		// 由于 DrawCard 是公共方法，最好在 DrawCard 逻辑末尾统一处理，或者依靠 broadcast
		// 这里我们假设 DrawCard 后如果不切换回合（比如摸牌后），AI 还需要继续操作？
		// 通常 UNO 规则：摸牌后如果能出可以出，不能出则过。
		// 简化版：AI 摸牌后直接过回合。
	}
}

// aiTryPlayCard 尝试打出一张牌
func (gr *GameRoom) aiTryPlayCard(player *models.PlayerState) bool {
	// 1. 优先处理加牌堆叠 (+2/+4)
	if gr.GameState.PendingDrawCount > 0 {
		for _, card := range player.HandCards {
			if card.Effect == "+2" || card.Effect == "+4" {
				// 有加牌必跟
				gr.aiExecutePlay(player.UID, card, "")
				return true
			}
		}
		// 没加牌，只能摸牌（在 TriggerAITurn 的后续逻辑会处理）
		return false
	}

	// 2. 也是优先打出加牌/功能牌 (攻击性策略)
	// 只有当 AI 手牌数较多时才倾向于保留功能牌，否则尽快打出
	for _, card := range player.HandCards {
		if isSpecialCard(card) {
			gr.aiExecutePlay(player.UID, card, "")
			return true
		}
	}

	// 3. 寻找普通反应牌
	// 获取场上物质
	lastSubstance := ""
	if gr.GameState.LastCard != nil {
		lastSubstance = gr.GameState.LastCard.Substance
	}

	// 如果没有场上物质（例如刚开始或被 Au 清空），任意非功能牌都可（通常规则）
	// 但我们的 PlayCard 逻辑需要 ValidReaction
	// 如果 LastCard 为 nil，PlayCard 允许任意牌吗？看代码逻辑：allowAny logic.
	// 假设 LastCard != nil

	if lastSubstance != "" {
		// 遍历手牌，寻找能与 lastSubstance 反应的组合
		// 简单起见，我们只能出一张牌 + lastSubstance = NewSubstance
		// 复杂逻辑：Query DB

		// 优化：先筛选出手牌中可能的元素
		for _, card := range player.HandCards {
			if card.Effect != "" || isNobleGas(card.Type) {
				continue
			}

			// 尝试查询数据库：(lastSubstance) + (card.Type) = ?
			// 或者是 (lastSubstance) + (card.Type * 2) = ? (虽然我们一次只出一张，在此简化模型下)
			// 注意：PlayCard 逻辑是：Card + LastSubstance -> NewSubstance
			// 我们需要找到一个 Target Substance，它由 LastSubstance 和 Card 组成

			// 这里有个难点：PlayCard 需要传入 `substance` (结果物质)
			// 我们不知道结果物质是什么，只知道反应物。
			// 所以我们需要查询 repository: FindReactionByReactants(r1, r2)

			// 构造查询：R1 = lastSubstance AND R2 = card.Type (或者反过来)
			// 且 status = approved
			// 如果能找到，说明可以反应

			// 处理化学式下标，数据库存的是带下标的吗？repository.CheckReactionExists 做了双向查
			// 但是我们需要知道生成的 substance (Display?) 还是 R1+R2 组合？
			// PlayCard 逻辑：
			// requiredElements := parseSubstance(substance)
			// 检查手牌是否有 requiredElements
			// 检查 substance 是否能与 LastCard 反应 (CanReact)

			// 这里的 AI 思路反了：
			// AI 应该遍历所有已知的 "Approved Reactions"，看看哪一个能用 "LastSubstance" 和 "手牌" 凑出来。

			// 方案：
			// 1. 获取包含 LastSubstance 的所有反应： SELECT * FROM reactions WHERE r1=Last OR r2=Last
			// 2. 遍历这些反应，看另一半 (NewR) 是否在手牌里

			reactionRepo := repository.NewReactionRepository()
			possibleReactions, err := reactionRepo.FindReactionsBySubstance(lastSubstance)
			if err != nil {
				continue
			}

			for _, reaction := range possibleReactions {
				var neededComponent string
				if reaction.R1 == lastSubstance {
					neededComponent = reaction.R2
				} else {
					neededComponent = reaction.R1
				}

				// 检查 neededComponent 是否在手牌里
				// neededComponent 可能是 "O2", "H2", "Cl" 等
				// 手牌是 "O", "H", "Cl"
				// 需要解析 neededComponent 需要几张什么牌
				neededElements := parseSubstance(neededComponent)

				if hasCards(player.HandCards, neededElements) {
					// 找到了！
					// 结果物质应该是 reaction.Display 吗？
					// 不，PlayCard 的 substance 参数是指 "打出的牌组成的物质" 还是 "反应生成的物质"？
					// 回看 PlayCard:
					// substance = NormalizeSubscripts(substance)
					// Verify ValidSubstance(substance)
					// parseSubstance(substance) -> check hand cards
					// CanReact(LastCard.Substance, substance)

					// 所以 `substance` 参数是 AI 要打出的牌组成的物质（例如 "H2"）
					// 它必须能与 LastCard 反应。
					// 也就是说 reaction.R1/R2 中的 neededComponent 就是我们要打出的 substance。

					targetSubstance := neededComponent
					// 执行出牌
					log.Printf("[AI] 💡 AI %d 发现反应: %s + %s -> %s", player.UID, lastSubstance, targetSubstance, reaction.Display)
					gr.aiExecutePlay(player.UID, models.Card{Type: "AI_PlaceHolder"}, targetSubstance) // Card 参数在 aiExecutePlay 会自动找
					return true
				}
			}
		}
	} else {
		// 场上无物质（Au 后），随便出一张普通牌
		for _, card := range player.HandCards {
			if !isSpecialCard(card) {
				gr.aiExecutePlay(player.UID, card, card.Type) // 单质
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
func (gr *GameRoom) aiExecutePlay(uid int, card models.Card, substance string) {
	// 需要找到真实的手牌对象（PlayCard 逻辑需要精确匹配吗？PlayCard 会根据 substance 自动找牌，或者根据 card.Type）
	// PlayCard 逻辑：
	// if substance == "" -> substance = card.Type
	// parseSubstance(substance) -> 找手牌
	// 所以如果我们传了 substance (比如 "H2")，PlayCard 会自动去手牌找 2 张 H。
	// 这里的 card 参数主要是为了 passed if substance is empty, or for logging.

	// 解锁，调用 PlayCard
	gr.mutex.Unlock()

	// 如果是功能牌，substance 为空
	if isSpecialCard(card) {
		substance = ""
	}

	err := PlayCard(gr.Room.ID, uid, card, substance)
	gr.mutex.Lock()

	if err != nil {
		log.Printf("[AI] ⚠️ AI %d 出牌失败 (%s): %v", uid, substance, err)
		// 失败回退：摸牌
		gr.aiDrawCard(uid)
	} else {
		log.Printf("[AI] ✅ AI %d 成功出牌: %s / %s", uid, card.Type, substance)
		if websocket.GlobalHub != nil {
			websocket.GlobalHub.BroadcastToRoom(gr.Room.ID, websocket.Message{
				Type: "action_toast",
				Data: fmt.Sprintf("AI 研究员 %d 打出了 %s", uid, substance),
			})
		}
	}
}

// 辅助函数

func isSpecialCard(card models.Card) bool {
	return card.Effect != "" || isNobleGas(card.Type)
}

func isNobleGas(t string) bool {
	return t == "He" || t == "Ne" || t == "Ar" || t == "Kr"
}

// hasCards 检查手牌是否满足需求 (简单的计数检查)
func hasCards(hand []models.Card, needed map[string]int) bool {
	// 统计手牌
	counts := make(map[string]int)
	for _, c := range hand {
		counts[c.Type]++
	}

	// 比较
	for elem, count := range needed {
		if counts[elem] < count {
			return false
		}
	}
	return true
}
