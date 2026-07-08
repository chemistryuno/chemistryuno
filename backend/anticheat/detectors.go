package anticheat

import (
	"fmt"
	"log"
	"math"
)

// 指标重设计（docs/anticheat/METRICS_REDESIGN.md）：
// 旧的 response_time / frequency / pattern 检测器依赖编造或不可得的信号，已移除。
// 新指标只使用真实可测的信号，针对回合制化学游戏的作弊向量。

// DecisionOptimalityDetector 决策最优度检测器（核心指标）。
// 复用化学引擎在出牌当下算出的「最优出法匹配」，长期接近 100% 最优且低经验
// 账号是强作弊指纹。评分相对人群基线（均值/标准差），无基线时退化为线性映射。
type DecisionOptimalityDetector struct {
	minDecisions int     // 生效所需最小决策数
	populationMean float64 // 人群 optimalityRate 均值（冷启动缺省）
	populationStd  float64 // 人群 optimalityRate 标准差（冷启动缺省）
}

func NewDecisionOptimalityDetector(minDecisions int, popMean, popStd float64) *DecisionOptimalityDetector {
	if minDecisions <= 0 {
		minDecisions = 15
	}
	if popStd <= 0 {
		popStd = 0.15
	}
	return &DecisionOptimalityDetector{
		minDecisions:   minDecisions,
		populationMean: popMean,
		populationStd:  popStd,
	}
}

func (d *DecisionOptimalityDetector) Name() string { return "decision_optimality" }

func (d *DecisionOptimalityDetector) Detect(context *DetectionContext) (float64, error) {
	if context.TotalDecisions <= 0 {
		return 0, nil
	}
	rate := float64(context.OptimalDecisions) / float64(context.TotalDecisions)

	// 偏离人群均值多少个标准差 → sigmoid 映射到 0-100。
	// 只有「显著高于」均值才计分（低于均值不是作弊信号）。
	z := (rate - d.populationMean) / d.populationStd
	if z <= 0 {
		return 0, nil
	}
	score := sigmoid01(z) * 100

	// 样本不足时按比例衰减，避免几手就定性。
	if context.TotalDecisions < d.minDecisions {
		score *= float64(context.TotalDecisions) / float64(d.minDecisions)
	}
	return math.Min(100, score), nil
}

// ThinkTimeDetector 思考时长异常检测器。
// 只对「复杂局面」计分：复杂局面却几乎零思考（客户端上报、服务端已做上界校验）
// 是脚本/工具作弊的信号。简单局面秒出属正常，不计分。
type ThinkTimeDetector struct {
	minComplexDecisions int
}

func NewThinkTimeDetector(minComplexDecisions int) *ThinkTimeDetector {
	if minComplexDecisions <= 0 {
		minComplexDecisions = 5
	}
	return &ThinkTimeDetector{minComplexDecisions: minComplexDecisions}
}

func (t *ThinkTimeDetector) Name() string { return "think_time" }

func (t *ThinkTimeDetector) Detect(context *DetectionContext) (float64, error) {
	if context.ComplexDecisionCount <= 0 {
		return 0, nil
	}
	superhumanRatio := float64(context.SuperhumanDecisionCount) / float64(context.ComplexDecisionCount)
	score := superhumanRatio * 100

	if context.ComplexDecisionCount < t.minComplexDecisions {
		score *= float64(context.ComplexDecisionCount) / float64(t.minComplexDecisions)
	}
	return math.Min(100, score), nil
}

// RecentPerformanceDetector 近期战绩检测器（替换旧 win_rate）。
// 使用近 N 局滑动窗口胜率 + 对手强度加权，避免长期高手被历史累计胜率反复误判。
type RecentPerformanceDetector struct {
	thresholdRate float64 // 触发计分的窗口胜率下限
	minGames      int     // 生效所需最小窗口样本
}

func NewRecentPerformanceDetector(thresholdRate float64, minGames int) *RecentPerformanceDetector {
	if thresholdRate <= 0 {
		thresholdRate = 0.85
	}
	if minGames <= 0 {
		minGames = 10
	}
	return &RecentPerformanceDetector{thresholdRate: thresholdRate, minGames: minGames}
}

func (r *RecentPerformanceDetector) Name() string { return "recent_performance" }

func (r *RecentPerformanceDetector) Detect(context *DetectionContext) (float64, error) {
	if !context.HasRecentPerf || context.RecentGames == 0 {
		return 0, nil
	}
	if context.RecentWinRate <= r.thresholdRate {
		return 0, nil
	}
	excess := context.RecentWinRate - r.thresholdRate
	score := (excess / (1.0 - r.thresholdRate)) * 80 // 最多 80 分

	// 对手强度加权：战胜更强对手更可疑（>1 放大，<1 收敛）。
	if context.OpponentStrength > 0 {
		score *= math.Min(1.5, math.Max(0.5, context.OpponentStrength))
	}

	// 样本不足按比例衰减。
	if context.RecentGames < r.minGames {
		score *= float64(context.RecentGames) / float64(r.minGames)
	}
	return math.Min(100, score), nil
}

// MultiAccountDetector 多开/小号检测器。
// 分数由登录侧信号（同 IP/设备指纹聚集）直接填充，本检测器仅透传。
// 该维度不受新手保护放宽（见 optimization_scoring.go multiAccountSignals）。
type MultiAccountDetector struct{}

func NewMultiAccountDetector() *MultiAccountDetector { return &MultiAccountDetector{} }

func (m *MultiAccountDetector) Name() string { return "multi_account" }

func (m *MultiAccountDetector) Detect(context *DetectionContext) (float64, error) {
	if !context.HasMultiAccount {
		return 0, nil
	}
	return math.Min(100, math.Max(0, context.MultiAccountScore)), nil
}

// BuiltInStrategies 预定义的内置策略集合
type BuiltInStrategies struct {
	strategies map[string]DetectionStrategy
}

// NewBuiltInStrategies 创建内置策略集合（新指标体系）。
func NewBuiltInStrategies() *BuiltInStrategies {
	return &BuiltInStrategies{
		strategies: map[string]DetectionStrategy{
			"decision_optimality": NewDecisionOptimalityDetector(15, 0.6, 0.15),
			"think_time":          NewThinkTimeDetector(5),
			"recent_performance":  NewRecentPerformanceDetector(0.85, 10),
			"multi_account":       NewMultiAccountDetector(),
		},
	}
}

// RegisterAll 将所有内置策略注册到引擎
func (bis *BuiltInStrategies) RegisterAll(engine *RiskScoringEngine) error {
	for name, strategy := range bis.strategies {
		if strategy == nil {
			return fmt.Errorf("策略 %s 为空", name)
		}
		engine.RegisterStrategy(strategy)
		log.Printf("[风险评分] 已注册策略: %s", name)
	}
	return nil
}

// sigmoid01 将 z≥0 映射到 (0.5,1) 再拉伸到 (0,1)，用于把「偏离标准差数」
// 平滑映射为 0-1 的异常强度。z=0→0，z 越大越接近 1。
func sigmoid01(z float64) float64 {
	// 标准 logistic 在 z=0 时为 0.5；减 0.5 再 ×2 使 z=0→0、z→∞→1。
	s := 1.0/(1.0+math.Exp(-z)) - 0.5
	return math.Max(0, math.Min(1, s*2))
}
