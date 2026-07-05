package anticheat

import (
	"testing"
	"time"
)

// End-to-end: with adaptive threshold, z-score, and decay all enabled, the engine
// produces a coherent score, records effective weights for every contributing
// dimension, and the score maps to the expected sanction tier.
func TestEnsemble_AllFeaturesCoherent(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.Optimization.AdaptiveThreshold.Enabled = true
	cfg.Optimization.AdaptiveThreshold.MinSamples = 5
	cfg.Optimization.AdaptiveThreshold.ContributionWeight = 0.25
	cfg.Optimization.ZScore.Enabled = true
	cfg.Optimization.ZScore.Threshold = 3.0
	cfg.Optimization.ZScore.Weight = 0.2
	cfg.Optimization.RiskDecay.Enabled = true
	cfg.Optimization.RiskDecay.DecayFactor = 0.85
	cfg.Optimization.RiskDecay.MinFloorHours = 0

	engine := NewRiskScoringEngine(cfg)
	NewBuiltInStrategies().RegisterAll(engine)

	past := time.Now().Add(-1000 * time.Hour)
	now := time.Now()
	ctx := &DetectionContext{
		PlayerUID:      1,
		RoomID:         "room_ensemble",
		ResponseTimes:  []int64{30, 32, 31, 29, 33, 30, 28},
		AccountAgeDays: 60,
		WinCount:       97,
		TotalGames:     100,
		WinRate:        0.97,
		HasWinRate:     true,
		PersonalBaselines: establishedBaseline(baselineIndicatorResponseMean, 1100, 10000, 30),
		GlobalBaselines: map[string]GlobalBaselineStat{
			baselineIndicatorResponseMean: {Mean: 1200, StdDev: 250},
		},
		HistoricalRisk:            40,
		NormalGamesSinceViolation: 2,
		LastViolationAt:           &past,
		Now:                       now,
		OperationTimes: []time.Time{
			now, now.Add(30 * time.Millisecond), now.Add(60 * time.Millisecond),
		},
	}

	result, err := engine.CalculateRiskScore(ctx)
	if err != nil {
		t.Fatalf("calculate: %v", err)
	}

	// Effective weights recorded for base + optimization dimensions.
	for _, dim := range []string{"response_time", "adaptive_threshold", "zscore_anomaly"} {
		if _, ok := result.EffectiveWeights[dim]; !ok {
			t.Fatalf("expected effective weight recorded for %q; got %+v", dim, result.EffectiveWeights)
		}
	}

	// Decay factor was applied (historical risk folded in).
	if result.DecayFactorApplied == nil {
		t.Fatalf("expected decay factor applied")
	}

	// Score maps to a sanction tier consistent with thresholds.
	tier := result.SanctionType
	switch tier {
	case "none", "observe", "warning", "mute", "ban":
		// ok
	default:
		t.Fatalf("unexpected sanction tier %q", tier)
	}

	// The high-suspicion fixture should land at warning or above.
	if result.RiskScore < cfg.SanctionThresholds.WarningMin {
		t.Fatalf("expected elevated risk for suspicious fixture, got %.2f (tier %s)", result.RiskScore, tier)
	}
}

// Confirmed sanctions are excluded from the decaying historical contribution: the
// engine only decays the HistoricalRisk passed in, which the System computes from
// non-mute/ban records. This test asserts the engine treats a zero historical risk
// (i.e. all prior risk was confirmed/escalated) as no decay contribution.
func TestEnsemble_NoHistoricalRiskNoDecayContribution(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.Optimization.RiskDecay.Enabled = true
	engine := NewRiskScoringEngine(cfg)
	NewBuiltInStrategies().RegisterAll(engine)

	ctx := lowRiskContext("room_no_hist")
	ctx.HistoricalRisk = 0 // all prior risk was confirmed -> excluded by System
	ctx.NormalGamesSinceViolation = 5

	result, err := engine.CalculateRiskScore(ctx)
	if err != nil {
		t.Fatalf("calculate: %v", err)
	}
	if result.DecayFactorApplied != nil {
		t.Fatalf("expected no decay contribution when historical risk is zero")
	}
}
