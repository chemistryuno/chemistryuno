package anticheat

import (
	"testing"
	"time"
)

// With all optimization features disabled (the default), the engine must produce
// exactly the same score, dimensions, and sanction as before the optimization work
// — i.e. no optimization dimension, no provenance, and no behavioral change. This
// is the regression guard for the "off by default" rollout requirement.
func TestRollout_DisabledMatchesLegacyBehavior(t *testing.T) {
	cfg := NewDefaultConfig()
	// Sanity: every optimization feature is off by default.
	if cfg.Optimization.AdaptiveThreshold.Enabled ||
		cfg.Optimization.ZScore.Enabled ||
		cfg.Optimization.NewPlayer.Enabled ||
		cfg.Optimization.RiskDecay.Enabled {
		t.Fatalf("optimization features must default to disabled")
	}

	engine := NewRiskScoringEngine(cfg)
	NewBuiltInStrategies().RegisterAll(engine)

	now := time.Now()
	ctx := &DetectionContext{
		PlayerUID:      1,
		RoomID:         "room_legacy",
		ResponseTimes:  []int64{30, 35, 40, 45, 50},
		OperationCount: 15,
		WinCount:       95,
		TotalGames:     100,
		AccountAgeDays: 2,
		// Provide optimization inputs that MUST be ignored when features are off.
		PersonalBaselines: establishedBaseline(baselineIndicatorResponseMean, 1000, 10000, 30),
		GlobalBaselines: map[string]GlobalBaselineStat{
			baselineIndicatorResponseMean: {Mean: 1200, StdDev: 200},
		},
		HistoricalRisk:            90,
		NormalGamesSinceViolation: 1,
		IsNewPlayer:               true,
		OperationTimes: []time.Time{
			now, now.Add(100 * time.Millisecond), now.Add(200 * time.Millisecond),
		},
	}

	result, err := engine.CalculateRiskScore(ctx)
	if err != nil {
		t.Fatalf("calculate: %v", err)
	}

	// No optimization dimensions present.
	for _, dim := range []string{"adaptive_threshold", "zscore_anomaly", "historical_risk_decayed"} {
		if _, ok := result.Dimensions[dim]; ok {
			t.Fatalf("dimension %q must not exist when optimization disabled", dim)
		}
	}
	// No provenance recorded.
	if result.ThresholdSource != "" || len(result.AdaptiveDeviations) != 0 {
		t.Fatalf("no adaptive provenance expected when disabled")
	}
	if result.DecayFactorApplied != nil {
		t.Fatalf("no decay factor expected when disabled")
	}
	if result.NewPlayerObserve {
		t.Fatalf("new-player observe flag must be false when disabled")
	}

	// Only the legacy five dimensions are present.
	legacy := map[string]bool{"response_time": true, "frequency": true, "win_rate": true, "pattern": true, "account_age": true}
	for name := range result.Dimensions {
		if !legacy[name] {
			t.Fatalf("unexpected non-legacy dimension %q present when disabled", name)
		}
	}
}
