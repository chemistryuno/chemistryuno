package anticheat

import (
	"chemistryuno/backend/database"
	"encoding/json"
	"log"
	"sync"
	"time"
)

// DetectionStrategy 检测策略接口
type DetectionStrategy interface {
	Name() string
	Detect(context *DetectionContext) (float64, error)
}

// DetectionContext 检测上下文，包含检测所需的所有数据
type DetectionContext struct {
	PlayerUID       int
	RoomID          string
	ReplayID        string
	GameHistoryID   uint
	OperationIndex  int
	PrimaryEvidence ReplayEvidenceAnchor
	RelatedEvidence []ReplayEvidenceAnchor
	ReportEvidence  []ReplayEvidenceAnchor
	ReportCount     int
	ReportSummary   string
	ResponseTimes   []int64       // 响应时间列表(毫秒)
	OperationCount  int           // 操作总数
	TimestampOffset time.Duration // 时间窗口大小
	WinCount        int           // 赢的次数
	TotalGames      int           // 总游戏次数
	AccountAgeDays  int           // 账号年龄(天)
	OperationTimes  []time.Time   // 操作时间列表
}

// RiskScoringEngine 风险评分引擎
type RiskScoringEngine struct {
	strategies   map[string]DetectionStrategy
	config       *RiskScoringConfig
	strategyLock sync.RWMutex
	configLock   sync.RWMutex
}

// RiskScoringConfig 风险评分配置
type RiskScoringConfig struct {
	Dimensions         map[string]DimensionConfig `json:"dimensions" yaml:"dimensions"`
	SanctionThresholds SanctionThresholds         `json:"sanction_thresholds" yaml:"sanction_thresholds"`
	EnabledStrategies  []string                   `json:"enabled_strategies" yaml:"enabled_strategies"`
	UnbanConfig        UnbanConfig                `json:"unban" yaml:"unban"`
}

// UnbanConfig 解封补偿配置
type UnbanConfig struct {
	Enabled            bool   `json:"enabled" yaml:"enabled"`                         // 是否启用补偿功能
	CompensationAmount int    `json:"compensation_amount" yaml:"compensation_amount"` // 默认补偿燃素数量
	DefaultMessage     string `json:"default_message" yaml:"default_message"`         // 默认解封消息文案
	MessageMaxLength   int    `json:"message_max_length" yaml:"message_max_length"`   // 补偿消息字符限制
	MinAmount          int    `json:"min_amount" yaml:"min_amount"`                   // 补偿金额最小值
	MaxAmount          int    `json:"max_amount" yaml:"max_amount"`                   // 补偿金额最大值
	IdempotencyTTL     int    `json:"idempotency_ttl" yaml:"idempotency_ttl"`         // 幂等性缓存TTL（分钟）
}

// DimensionConfig 维度配置
type DimensionConfig struct {
	Weight     float64 `json:"weight"`
	Threshold  int64   `json:"threshold"`  // 毫秒或其他单位
	Percentile float64 `json:"percentile"` // 用于异常检测的百分位数
}

// SanctionThresholds 处罚阈值
type SanctionThresholds struct {
	ObserveMin float64 `json:"observe_min"`
	ObserveMax float64 `json:"observe_max"`
	WarningMin float64 `json:"warning_min"`
	WarningMax float64 `json:"warning_max"`
	MuteMin    float64 `json:"mute_min"`
	MuteMax    float64 `json:"mute_max"`
	BanMin     float64 `json:"ban_min"`
	BanMax     float64 `json:"ban_max"`
}

// RiskScoringResult 风险评分结果
type RiskScoringResult struct {
	RiskScore          float64
	Dimensions         map[string]float64
	IndicatorDetails   []RiskIndicatorDetail
	ReportContribution ReportContribution
	SanctionType       string // "none", "observe", "warning", "mute", "ban"
	SuggestedAction    string
	SuggestionReason   string
	ReplayID           string
	GameHistoryID      uint
	OperationIndex     int
	OperationTimestamp *time.Time
	PrimaryEvidence    ReplayEvidenceAnchor
	RelatedEvidence    []ReplayEvidenceAnchor
	Timestamp          time.Time
}

// RiskIndicatorDetail snapshots one scoring signal for later review.
type RiskIndicatorDetail struct {
	Name            string  `json:"name"`
	RawValue        float64 `json:"raw_value"`
	NormalizedScore float64 `json:"normalized_score"`
	Weight          float64 `json:"weight"`
	Contribution    float64 `json:"contribution"`
	Explanation     string  `json:"explanation"`
	EvidenceAnchors []ReplayEvidenceAnchor `json:"evidence_anchors,omitempty"`
}

// ReportContribution captures the report signal used in the risk score.
type ReportContribution struct {
	DeduplicatedCount int     `json:"deduplicated_count"`
	Weight            float64 `json:"weight"`
	Contribution      float64 `json:"contribution"`
	SourceSummary     string  `json:"source_summary"`
	EvidenceAnchors   []ReplayEvidenceAnchor `json:"evidence_anchors,omitempty"`
}

// NewRiskScoringEngine 创建风险评分引擎
func NewRiskScoringEngine(config *RiskScoringConfig) *RiskScoringEngine {
	return &RiskScoringEngine{
		strategies: make(map[string]DetectionStrategy),
		config:     config,
	}
}

// RegisterStrategy 注册检测策略
func (rse *RiskScoringEngine) RegisterStrategy(strategy DetectionStrategy) {
	rse.strategyLock.Lock()
	defer rse.strategyLock.Unlock()
	rse.strategies[strategy.Name()] = strategy
}

// UnregisterStrategy 注销检测策略
func (rse *RiskScoringEngine) UnregisterStrategy(name string) {
	rse.strategyLock.Lock()
	defer rse.strategyLock.Unlock()
	delete(rse.strategies, name)
}

// UpdateConfig 更新配置
func (rse *RiskScoringEngine) UpdateConfig(config *RiskScoringConfig) {
	rse.configLock.Lock()
	defer rse.configLock.Unlock()
	rse.config = config
}

// CalculateRiskScore 计算风险分数
func (rse *RiskScoringEngine) CalculateRiskScore(context *DetectionContext) (*RiskScoringResult, error) {
	rse.strategyLock.RLock()
	strategies := make([]DetectionStrategy, 0)
	for _, strategyName := range rse.getEnabledStrategies() {
		if strategy, exists := rse.strategies[strategyName]; exists {
			strategies = append(strategies, strategy)
		}
	}
	rse.strategyLock.RUnlock()

	result := &RiskScoringResult{
		RiskScore:       0,
		Dimensions:      make(map[string]float64),
		PrimaryEvidence: EvidenceAnchorFromContext(context),
		RelatedEvidence: append([]ReplayEvidenceAnchor(nil), context.RelatedEvidence...),
		Timestamp:       time.Now(),
	}

	rse.configLock.RLock()
	defer rse.configLock.RUnlock()

	totalWeight := 0.0
	weightedTotal := 0.0
	for _, strategy := range strategies {
		score, err := strategy.Detect(context)
		if err != nil {
			log.Printf("[风险评分] 策略 %s 执行失败: %v", strategy.Name(), err)
			continue
		}

		// 获取策略的权重
		weight := 1.0
		if dimConfig, exists := rse.config.Dimensions[strategy.Name()]; exists {
			weight = dimConfig.Weight
		}

		// 考虑账号新旧程度
		if context.AccountAgeDays < 7 && weight > 0 {
			weight *= 1.5 // 新账号权重增加50%
		}

		result.Dimensions[strategy.Name()] = score
		contribution := score * weight
		weightedTotal += contribution
		totalWeight += weight
		result.IndicatorDetails = append(result.IndicatorDetails, RiskIndicatorDetail{
			Name:            strategy.Name(),
			RawValue:        score,
			NormalizedScore: clampRiskScore(score),
			Weight:          weight,
			Contribution:    contribution,
			Explanation:     "in-game collected signal",
			EvidenceAnchors: []ReplayEvidenceAnchor{result.PrimaryEvidence},
		})
	}

	if context.ReportCount > 0 {
		reportWeight := 0.10
		reportScore := clampRiskScore(float64(context.ReportCount) * 12.5)
		contribution := reportScore * reportWeight
		weightedTotal += contribution
		totalWeight += reportWeight
		summary := context.ReportSummary
		if summary == "" {
			summary = "deduplicated player reports"
		}
		result.ReportContribution = ReportContribution{
			DeduplicatedCount: context.ReportCount,
			Weight:            reportWeight,
			Contribution:      contribution,
			SourceSummary:     summary,
			EvidenceAnchors:   context.ReportEvidence,
		}
		reportAnchors := context.ReportEvidence
		if len(reportAnchors) == 0 {
			reportAnchors = []ReplayEvidenceAnchor{result.PrimaryEvidence}
		}
		result.IndicatorDetails = append(result.IndicatorDetails, RiskIndicatorDetail{
			Name:            "player_reports",
			RawValue:        float64(context.ReportCount),
			NormalizedScore: reportScore,
			Weight:          reportWeight,
			Contribution:    contribution,
			Explanation:     summary,
			EvidenceAnchors: reportAnchors,
		})
	}

	// 归一化到 0-100
	if totalWeight > 0 {
		result.RiskScore = weightedTotal / totalWeight
	}

	result.RiskScore = clampRiskScore(result.RiskScore)

	// 确定处罚类型
	result.SanctionType = rse.determineSanctionType(result.RiskScore)
	result.SuggestedAction = result.SanctionType
	result.SuggestionReason = rse.suggestionReason(result.RiskScore, result.SanctionType)
	result.ReplayID = context.ReplayID
	if result.ReplayID == "" {
		result.ReplayID = context.RoomID
	}
	result.GameHistoryID = context.GameHistoryID
	result.OperationIndex = context.OperationIndex
	if result.OperationIndex <= 0 && len(context.OperationTimes) > 0 {
		result.OperationIndex = len(context.OperationTimes)
	}
	if len(context.OperationTimes) > 0 {
		ts := context.OperationTimes[len(context.OperationTimes)-1]
		result.OperationTimestamp = &ts
	}
	if result.PrimaryEvidence.ReplayID == "" {
		result.PrimaryEvidence.ReplayID = result.ReplayID
	}
	if result.PrimaryEvidence.EventIndex == 0 {
		result.PrimaryEvidence.EventIndex = result.OperationIndex
	}
	if result.PrimaryEvidence.GameHistoryID == 0 {
		result.PrimaryEvidence.GameHistoryID = result.GameHistoryID
	}
	result.PrimaryEvidence = NormalizeReplayEvidenceAnchor(result.PrimaryEvidence)
	if len(result.RelatedEvidence) == 0 {
		result.RelatedEvidence = []ReplayEvidenceAnchor{result.PrimaryEvidence}
	}

	return result, nil
}

func clampRiskScore(score float64) float64 {
	if score > 100 {
		return 100
	}
	if score < 0 {
		return 0
	}
	return score
}

// determineSanctionType 根据风险分数确定处罚类型
func (rse *RiskScoringEngine) determineSanctionType(riskScore float64) string {
	rse.configLock.RLock()
	defer rse.configLock.RUnlock()

	thresholds := rse.config.SanctionThresholds

	switch {
	case riskScore >= thresholds.BanMin && riskScore <= thresholds.BanMax:
		return "ban"
	case riskScore >= thresholds.MuteMin && riskScore <= thresholds.MuteMax:
		return "mute"
	case riskScore >= thresholds.WarningMin && riskScore <= thresholds.WarningMax:
		return "warning"
	case riskScore >= thresholds.ObserveMin && riskScore <= thresholds.ObserveMax:
		return "observe"
	default:
		return "none"
	}
}

func (rse *RiskScoringEngine) suggestionReason(riskScore float64, sanctionType string) string {
	rse.configLock.RLock()
	defer rse.configLock.RUnlock()
	switch sanctionType {
	case "ban":
		return "risk score reached ban threshold"
	case "mute":
		return "risk score reached mute threshold"
	case "warning":
		return "risk score reached warning threshold"
	case "observe":
		return "risk score reached observe threshold"
	default:
		return "risk score below observe threshold"
	}
}

func MarshalIndicatorDetails(details []RiskIndicatorDetail) database.JSON {
	if len(details) == 0 {
		return nil
	}
	data, _ := json.Marshal(details)
	return data
}

func MarshalReportContribution(contribution ReportContribution) database.JSON {
	if contribution.DeduplicatedCount == 0 && contribution.Contribution == 0 {
		return nil
	}
	data, _ := json.Marshal(contribution)
	return data
}

// getEnabledStrategies 获取启用的策略列表
func (rse *RiskScoringEngine) getEnabledStrategies() []string {
	if len(rse.config.EnabledStrategies) > 0 {
		return rse.config.EnabledStrategies
	}

	// 默认启用所有已注册的策略
	rse.strategyLock.RLock()
	defer rse.strategyLock.RUnlock()

	strategies := make([]string, 0, len(rse.strategies))
	for name := range rse.strategies {
		strategies = append(strategies, name)
	}
	return strategies
}

// NewDefaultConfig 创建默认配置
func NewDefaultConfig() *RiskScoringConfig {
	return &RiskScoringConfig{
		Dimensions: map[string]DimensionConfig{
			"response_time": {
				Weight:     0.25,
				Threshold:  100,  // 100ms
				Percentile: 0.05, // 5%百分位
			},
			"frequency": {
				Weight:     0.25,
				Threshold:  20,   // 每10秒20个操作
				Percentile: 0.95, // 95%百分位
			},
			"win_rate": {
				Weight:     0.20,
				Threshold:  85, // 85%胜率
				Percentile: 0.99,
			},
			"pattern": {
				Weight:     0.15,
				Threshold:  50, // 50ms最小操作间隔
				Percentile: 0.10,
			},
			"account_age": {
				Weight:     0.15,
				Threshold:  7, // 7天新账号
				Percentile: 1.0,
			},
		},
		SanctionThresholds: SanctionThresholds{
			ObserveMin: 20,
			ObserveMax: 40,
			WarningMin: 40,
			WarningMax: 60,
			MuteMin:    60,
			MuteMax:    80,
			BanMin:     80,
			BanMax:     100,
		},
		EnabledStrategies: []string{
			"response_time",
			"frequency",
			"win_rate",
			"pattern",
			"account_age",
		},
		UnbanConfig: UnbanConfig{
			Enabled:            true,
			CompensationAmount: 100,
			DefaultMessage:     "由于反作弊系统将您误封，在此，ChemistryUNO开发组向受到影响的研究员提供燃素补偿，感谢研究员对维护纯净游戏环境做出的贡献",
			MessageMaxLength:   500,
			MinAmount:          1,
			MaxAmount:          10000,
			IdempotencyTTL:     60,
		},
	}
}
