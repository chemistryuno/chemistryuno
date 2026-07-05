package anticheat

import (
	"chemistryuno/backend/database"
	"encoding/json"
)

// RuleTestSample is one input to the rule-test sandbox. It is derived from a
// historical detection record or constructed by the admin. Because raw detection
// context (response times, etc.) is not persisted, the sandbox recomputes the
// weighted ensemble from the stored per-dimension scores under the draft config.
type RuleTestSample struct {
	RiskScoreID     uint               `json:"risk_score_id,omitempty"`
	PlayerUID       uint               `json:"player_uid,omitempty"`
	LiveScore       float64            `json:"live_score"` // score under the live config
	LiveTier        string             `json:"live_tier"`  // tier under the live config
	DimensionScores map[string]float64 `json:"dimension_scores"`
	AccountAgeDays  int                `json:"account_age_days,omitempty"`
}

// RuleTestOutcome is the per-sample result of re-running the draft config.
type RuleTestOutcome struct {
	RiskScoreID uint    `json:"risk_score_id,omitempty"`
	PlayerUID   uint    `json:"player_uid,omitempty"`
	LiveScore   float64 `json:"live_score"`
	DraftScore  float64 `json:"draft_score"`
	LiveTier    string  `json:"live_tier"`
	DraftTier   string  `json:"draft_tier"`
	TierChanged bool    `json:"tier_changed"`
}

// RuleTestResult summarizes a rule-test run.
type RuleTestResult struct {
	SampleCount      int                  `json:"sample_count"`
	HitDistribution  map[string]int       `json:"hit_distribution"`   // draft tier -> count
	TierChangeCounts map[string]int       `json:"tier_change_counts"` // "live->draft" -> count
	Escalations      int                  `json:"escalations"`        // samples whose tier got stricter
	Deescalations    int                  `json:"deescalations"`      // samples whose tier got looser
	Outcomes         []RuleTestOutcome    `json:"outcomes"`
}

// tierRank orders sanction tiers from least to most severe for change comparison.
var tierRank = map[string]int{
	"none":    0,
	"observe": 1,
	"warning": 2,
	"mute":    3,
	"ban":     4,
}

// RunRuleTest evaluates the draft config against the provided samples in an
// isolated context. It performs NO persistence and mutates no live state: it
// recomputes the weighted ensemble for each sample's stored per-dimension scores
// under the draft weights/thresholds and compares the resulting tier to the live
// tier. Raw in-game context is not replayed, so optimization dimensions that need
// live context (adaptive/z-score) are not re-derived here; the test focuses on the
// effect of weight and threshold changes on existing dimension scores.
func RunRuleTest(draft *RiskScoringConfig, samples []RuleTestSample) (*RuleTestResult, error) {
	if draft == nil {
		return nil, ErrInvalidConfig
	}
	if err := ValidateConfig(draft); err != nil {
		return nil, err
	}

	cfg := copyConfig(draft)

	result := &RuleTestResult{
		HitDistribution:  make(map[string]int),
		TierChangeCounts: make(map[string]int),
		Outcomes:         make([]RuleTestOutcome, 0, len(samples)),
	}

	for _, sample := range samples {
		draftScore := recomputeEnsemble(cfg, sample.DimensionScores, sample.AccountAgeDays)
		draftTier := tierForScore(cfg.SanctionThresholds, draftScore)
		liveTier := sample.LiveTier
		if liveTier == "" {
			liveTier = tierForScore(cfg.SanctionThresholds, sample.LiveScore)
		}

		changed := tierRank[draftTier] != tierRank[liveTier]
		outcome := RuleTestOutcome{
			RiskScoreID: sample.RiskScoreID,
			PlayerUID:   sample.PlayerUID,
			LiveScore:   sample.LiveScore,
			DraftScore:  draftScore,
			LiveTier:    liveTier,
			DraftTier:   draftTier,
			TierChanged: changed,
		}
		result.Outcomes = append(result.Outcomes, outcome)
		result.HitDistribution[draftTier]++
		if changed {
			result.TierChangeCounts[liveTier+"->"+draftTier]++
			if tierRank[draftTier] > tierRank[liveTier] {
				result.Escalations++
			} else {
				result.Deescalations++
			}
		}
	}
	result.SampleCount = len(result.Outcomes)
	return result, nil
}

// recomputeEnsemble re-derives the weighted-average risk score from stored
// per-dimension scores under the draft dimension weights, mirroring the engine's
// normalization (including the new-account weight boost).
func recomputeEnsemble(cfg *RiskScoringConfig, dims map[string]float64, accountAgeDays int) float64 {
	if len(dims) == 0 {
		return 0
	}
	totalWeight := 0.0
	weightedTotal := 0.0
	for name, score := range dims {
		weight := 1.0
		if dc, ok := cfg.Dimensions[name]; ok {
			weight = dc.Weight
		}
		if accountAgeDays > 0 && accountAgeDays < 7 && weight > 0 {
			weight *= 1.5
		}
		weightedTotal += score * weight
		totalWeight += weight
	}
	if totalWeight <= 0 {
		return 0
	}
	return clampRiskScore(weightedTotal / totalWeight)
}

// tierForScore maps a score to a sanction tier using the given thresholds.
func tierForScore(t SanctionThresholds, score float64) string {
	switch {
	case score >= t.BanMin && score <= t.BanMax:
		return "ban"
	case score >= t.MuteMin && score <= t.MuteMax:
		return "mute"
	case score >= t.WarningMin && score <= t.WarningMax:
		return "warning"
	case score >= t.ObserveMin && score <= t.ObserveMax:
		return "observe"
	default:
		return "none"
	}
}

// copyConfig returns a deep copy of a risk scoring config for isolated use.
func copyConfig(src *RiskScoringConfig) *RiskScoringConfig {
	dst := &RiskScoringConfig{
		Dimensions:         make(map[string]DimensionConfig, len(src.Dimensions)),
		SanctionThresholds: src.SanctionThresholds,
		EnabledStrategies:  append([]string(nil), src.EnabledStrategies...),
		UnbanConfig:        src.UnbanConfig,
		Optimization:       src.Optimization,
	}
	for k, v := range src.Dimensions {
		dst.Dimensions[k] = v
	}
	return dst
}

// SampleFromRiskScore builds a rule-test sample from a persisted detection record,
// reconstructing the per-dimension scores stored on the record.
func SampleFromRiskScore(score database.CheatRiskScore) RuleTestSample {
	dims := map[string]float64{
		"response_time": score.ResponseTimeDim,
		"frequency":     score.FrequencyDim,
		"win_rate":      score.WinRateDim,
		"pattern":       score.PatternDim,
		"account_age":   score.AccountAgeDim,
	}
	return RuleTestSample{
		RiskScoreID:     score.ID,
		PlayerUID:       score.PlayerUID,
		LiveScore:       score.RiskScore,
		LiveTier:        score.PunishmentDecision,
		DimensionScores: dims,
	}
}

// MarshalRuleTestResult serializes a rule-test result for persistence.
func MarshalRuleTestResult(result *RuleTestResult) database.JSON {
	if result == nil {
		return nil
	}
	data, err := json.Marshal(result)
	if err != nil {
		return nil
	}
	return data
}
