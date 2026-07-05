package anticheat

import (
	"testing"
)

func zscoreEngine(t *testing.T, enabled bool, threshold, weight float64) *RiskScoringEngine {
	t.Helper()
	cfg := NewDefaultConfig()
	cfg.Optimization.ZScore.Enabled = enabled
	cfg.Optimization.ZScore.Threshold = threshold
	cfg.Optimization.ZScore.Weight = weight
	engine := NewRiskScoringEngine(cfg)
	NewBuiltInStrategies().RegisterAll(engine)
	return engine
}

// A population outlier beyond the threshold contributes a z-score dimension.
func TestZScore_OutlierTriggers(t *testing.T) {
	engine := zscoreEngine(t, true, 3.0, 0.2)
	ctx := &DetectionContext{
		PlayerUID:      1,
		RoomID:         "room_z",
		ResponseTimes:  []int64{50, 52, 48, 51}, // mean ~50ms
		AccountAgeDays: 30,
		GlobalBaselines: map[string]GlobalBaselineStat{
			baselineIndicatorResponseMean: {Mean: 1200, StdDev: 300}, // ~3.8 std away
		},
	}
	result, err := engine.CalculateRiskScore(ctx)
	if err != nil {
		t.Fatalf("calculate: %v", err)
	}
	if _, ok := result.Dimensions["zscore_anomaly"]; !ok {
		t.Fatalf("expected zscore_anomaly dimension to be present")
	}
	if result.Dimensions["zscore_anomaly"] <= 0 {
		t.Fatalf("expected positive zscore anomaly, got %.2f", result.Dimensions["zscore_anomaly"])
	}
}

// A value within the threshold adds no z-score contribution.
func TestZScore_InRangeNoContribution(t *testing.T) {
	engine := zscoreEngine(t, true, 3.0, 0.2)
	ctx := &DetectionContext{
		PlayerUID:      1,
		RoomID:         "room_z2",
		ResponseTimes:  []int64{1180, 1200, 1220}, // mean ~1200, matches population
		AccountAgeDays: 30,
		GlobalBaselines: map[string]GlobalBaselineStat{
			baselineIndicatorResponseMean: {Mean: 1200, StdDev: 300},
		},
	}
	result, err := engine.CalculateRiskScore(ctx)
	if err != nil {
		t.Fatalf("calculate: %v", err)
	}
	if _, ok := result.Dimensions["zscore_anomaly"]; ok {
		t.Fatalf("expected no zscore_anomaly dimension for in-range value")
	}
}

// The z-score dimension is independent and coexists with the variance-based
// pattern analysis; disabling z-score leaves existing dimensions intact.
func TestZScore_DisabledLeavesBaseDimensions(t *testing.T) {
	engine := zscoreEngine(t, false, 3.0, 0.2)
	ctx := &DetectionContext{
		PlayerUID:      1,
		RoomID:         "room_z3",
		ResponseTimes:  []int64{50, 52, 48},
		AccountAgeDays: 30,
		GlobalBaselines: map[string]GlobalBaselineStat{
			baselineIndicatorResponseMean: {Mean: 1200, StdDev: 300},
		},
	}
	result, err := engine.CalculateRiskScore(ctx)
	if err != nil {
		t.Fatalf("calculate: %v", err)
	}
	if _, ok := result.Dimensions["zscore_anomaly"]; ok {
		t.Fatalf("zscore_anomaly must not exist when disabled")
	}
	if _, ok := result.Dimensions["response_time"]; !ok {
		t.Fatalf("base response_time dimension should still be evaluated")
	}
}
