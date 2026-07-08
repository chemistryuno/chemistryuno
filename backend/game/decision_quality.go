package game

import (
	"chemistryuno/backend/models"
)

// 决策质量评估（反作弊指标重设计，见 docs/anticheat/METRICS_REDESIGN.md）。
//
// 在出牌当下（已持有完整游戏状态与化学引擎）顺带评估一次决策的质量，
// 供反作弊的 decision_optimality / think_time 指标使用。避免反作弊阶段重放整局。

// DecisionQuality 单次出牌决策的质量快照。
type DecisionQuality struct {
	HadReactingOption bool // 当时手牌是否存在可与场面反应的出法
	PlayedOptimal     bool // 玩家本次是否打出了「最优」出法（存在反应机会时选择了反应）
	OptionCount       int  // 当时可组成的合法物质数（局面复杂度近似）
	Complex           bool // 是否为复杂局面（需要真实思考，用于 think_time）
	Evaluated         bool // 本次决策是否纳入评估（简单/无意义局面不计入）
}

// complexOptionThreshold 认定「复杂局面」的最小可选物质数。
// 低于此值的局面几乎无需思考（唯一或极少出法），秒出属正常，不计入 think_time。
const complexOptionThreshold = 3

// EvaluatePlayDecision 评估一次普通出牌的决策质量。
//
// handCards: 出牌前的手牌快照；fieldSubstance: 场上待反应物质（LastCard.Substance，
// 空表示自由出牌）；playedSubstance: 玩家实际打出的物质。
func EvaluatePlayDecision(handCards []models.Card, fieldSubstance, playedSubstance string) DecisionQuality {
	q := DecisionQuality{}

	options := GetSubstancesFromElements(handCards)
	// 过滤掉功能牌/稀有气体，只统计参与化学反应的普通物质，作为复杂度近似。
	normalOptions := make([]string, 0, len(options))
	for _, s := range options {
		if isSpecialOrNoble(s) {
			continue
		}
		normalOptions = append(normalOptions, s)
	}
	q.OptionCount = len(normalOptions)

	// 自由出牌（场上无待反应物质）：没有「最优反应」概念，不纳入 decision_optimality。
	if fieldSubstance == "" {
		q.Complex = q.OptionCount >= complexOptionThreshold
		q.Evaluated = false
		return q
	}

	// 是否存在可与场面反应的出法。
	for _, s := range normalOptions {
		if JudgeReaction(s, fieldSubstance) {
			q.HadReactingOption = true
			break
		}
	}

	// 玩家本次是否打出了可反应的出法（= 最优选择：抓住了反应机会）。
	if playedSubstance != "" && !isSpecialOrNoble(playedSubstance) {
		q.PlayedOptimal = JudgeReaction(playedSubstance, fieldSubstance)
	}

	// 纳入评估的前提：存在反应机会时，玩家是否找到才有区分度。
	q.Evaluated = q.HadReactingOption
	// 复杂局面：可选物质多，且确实存在需要判断的反应机会。
	q.Complex = q.HadReactingOption && q.OptionCount >= complexOptionThreshold
	return q
}

// isSpecialOrNoble 判断是否为功能牌或稀有气体（不参与常规化学反应最优性评估）。
func isSpecialOrNoble(s string) bool {
	switch s {
	case "Au", "+2", "+4", "reverse", "skip":
		return true
	}
	return isNobleGas(s)
}
