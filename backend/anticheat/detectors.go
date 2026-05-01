package anticheat

import (
	"fmt"
	"log"
	"math"
	"sort"
	"time"
)

// ResponseTimeDetector 响应时间检测器
type ResponseTimeDetector struct {
	thresholdMs int64
	percentile  float64
}

// NewResponseTimeDetector 创建响应时间检测器
func NewResponseTimeDetector(thresholdMs int64, percentile float64) *ResponseTimeDetector {
	return &ResponseTimeDetector{
		thresholdMs: thresholdMs,
		percentile:  percentile,
	}
}

func (rtd *ResponseTimeDetector) Name() string {
	return "response_time"
}

func (rtd *ResponseTimeDetector) Detect(context *DetectionContext) (float64, error) {
	if len(context.ResponseTimes) == 0 {
		return 0, nil
	}

	// 计算平均响应时间
	var totalTime int64
	for _, rt := range context.ResponseTimes {
		totalTime += rt
	}
	avgTime := totalTime / int64(len(context.ResponseTimes))

	// 计算异常响应时间的比例
	aboveThresholdCount := 0
	for _, rt := range context.ResponseTimes {
		if rt < rtd.thresholdMs {
			aboveThresholdCount++
		}
	}

	anomalyRatio := float64(aboveThresholdCount) / float64(len(context.ResponseTimes))

	// 归一化到0-100
	score := anomalyRatio * 100

	// 额外惩罚极快的响应时间
	if avgTime < 50 && len(context.ResponseTimes) > 5 {
		score = math.Min(100, score*1.5)
	}

	return score, nil
}

// FrequencyDetector 频率检测器
type FrequencyDetector struct {
	maxActionsPerWindow int
	windowSize          time.Duration
}

// NewFrequencyDetector 创建频率检测器
func NewFrequencyDetector(maxActionsPerWindow int, windowSize time.Duration) *FrequencyDetector {
	return &FrequencyDetector{
		maxActionsPerWindow: maxActionsPerWindow,
		windowSize:          windowSize,
	}
}

func (fd *FrequencyDetector) Name() string {
	return "frequency"
}

func (fd *FrequencyDetector) Detect(context *DetectionContext) (float64, error) {
	if len(context.OperationTimes) == 0 {
		return 0, nil
	}

	// 按时间排序
	times := make([]time.Time, len(context.OperationTimes))
	copy(times, context.OperationTimes)
	sort.Slice(times, func(i, j int) bool {
		return times[i].Before(times[j])
	})

	// 检查滑动窗口中的最大操作数
	maxInWindow := 0
	for i := 0; i < len(times); i++ {
		windowEnd := times[i].Add(fd.windowSize)
		count := 1
		for j := i + 1; j < len(times) && times[j].Before(windowEnd); j++ {
			count++
		}
		if count > maxInWindow {
			maxInWindow = count
		}
	}

	// 计算超过阈值的程度
	if maxInWindow <= fd.maxActionsPerWindow {
		return 0, nil
	}

	excessRatio := float64(maxInWindow-fd.maxActionsPerWindow) / float64(fd.maxActionsPerWindow)
	score := math.Min(100, excessRatio*50) // 超过5倍则达到100分

	return score, nil
}

// WinRateDetector 胜率检测器
type WinRateDetector struct {
	thresholdRate float64 // 0.0-1.0
}

// NewWinRateDetector 创建胜率检测器
func NewWinRateDetector(thresholdRate float64) *WinRateDetector {
	return &WinRateDetector{
		thresholdRate: thresholdRate,
	}
}

func (wrd *WinRateDetector) Name() string {
	return "win_rate"
}

func (wrd *WinRateDetector) Detect(context *DetectionContext) (float64, error) {
	if context.TotalGames == 0 {
		return 0, nil
	}

	winRate := float64(context.WinCount) / float64(context.TotalGames)

	// 仅在胜率明显异常时才计分
	if winRate <= wrd.thresholdRate {
		return 0, nil
	}

	// 计算超过阈值的程度
	excess := winRate - wrd.thresholdRate
	score := (excess / (1.0 - wrd.thresholdRate)) * 80 // 最多80分

	// 需要足够的样本量
	if context.TotalGames < 10 {
		score *= float64(context.TotalGames) / 10.0
	}

	return score, nil
}

// PatternDetector 操作模式检测器
type PatternDetector struct {
	minIntervalMs int64
}

// NewPatternDetector 创建操作模式检测器
func NewPatternDetector(minIntervalMs int64) *PatternDetector {
	return &PatternDetector{
		minIntervalMs: minIntervalMs,
	}
}

func (pd *PatternDetector) Name() string {
	return "pattern"
}

func (pd *PatternDetector) Detect(context *DetectionContext) (float64, error) {
	if len(context.OperationTimes) < 2 {
		return 0, nil
	}

	// 按时间排序
	times := make([]time.Time, len(context.OperationTimes))
	copy(times, context.OperationTimes)
	sort.Slice(times, func(i, j int) bool {
		return times[i].Before(times[j])
	})

	// 检查相邻操作的时间间隔
	tooCloseCount := 0
	intervalVariance := 0.0

	intervals := make([]int64, 0)
	for i := 1; i < len(times); i++ {
		intervalMs := times[i].Sub(times[i-1]).Milliseconds()
		intervals = append(intervals, intervalMs)

		if intervalMs < pd.minIntervalMs {
			tooCloseCount++
		}
	}

	// 计算间隔的方差（规律性检测）
	if len(intervals) > 1 {
		// 计算平均值
		var avgInterval int64
		var totalInterval int64
		for _, interval := range intervals {
			totalInterval += interval
		}
		avgInterval = totalInterval / int64(len(intervals))

		// 计算方差
		var sumSquaredDiff int64
		for _, interval := range intervals {
			diff := interval - avgInterval
			sumSquaredDiff += diff * diff
		}
		variance := float64(sumSquaredDiff) / float64(len(intervals))
		stdDev := math.Sqrt(variance)

		// 方差过小说明操作过于规律
		if stdDev < 10 && avgInterval < pd.minIntervalMs*2 {
			intervalVariance = 50
		}
	}

	// 计算得分
	closeRatio := float64(tooCloseCount) / float64(len(intervals))
	score := closeRatio*50 + intervalVariance

	return math.Min(100, score), nil
}

// AccountAgeDetector 账号年龄检测器
type AccountAgeDetector struct {
	youngAccountDays int
}

// NewAccountAgeDetector 创建账号年龄检测器
func NewAccountAgeDetector(youngAccountDays int) *AccountAgeDetector {
	return &AccountAgeDetector{
		youngAccountDays: youngAccountDays,
	}
}

func (aad *AccountAgeDetector) Name() string {
	return "account_age"
}

func (aad *AccountAgeDetector) Detect(context *DetectionContext) (float64, error) {
	// 账号年龄本身不是直接的作弊指标，但在加权时会被考虑
	// 返回0表示账号本身没有异常，权重增加由引擎在加权时处理
	if context.AccountAgeDays < aad.youngAccountDays {
		// 极其新的账号（例如 < 1小时）给予额外怀疑
		if context.AccountAgeDays < 1 {
			return 20, nil // 给予基础怀疑分
		}
		return 0, nil // 权重增加由引擎处理
	}
	return 0, nil
}

// BuiltInStrategies 预定义的内置策略集合
type BuiltInStrategies struct {
	strategies map[string]DetectionStrategy
}

// NewBuiltInStrategies 创建内置策略集合
func NewBuiltInStrategies() *BuiltInStrategies {
	return &BuiltInStrategies{
		strategies: map[string]DetectionStrategy{
			"response_time": NewResponseTimeDetector(100, 0.05),
			"frequency":     NewFrequencyDetector(20, 10*time.Second),
			"win_rate":      NewWinRateDetector(0.85),
			"pattern":       NewPatternDetector(50),
			"account_age":   NewAccountAgeDetector(7),
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
