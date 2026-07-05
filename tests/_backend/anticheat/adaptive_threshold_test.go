package anticheat

import (
	"chemistryuno/backend/database"
	"testing"
)

func establishedBaseline(indicator string, mean, variance float64, count int) map[string]database.PlayerBehaviorBaseline {
	return map[string]database.PlayerBehaviorBaseline{
		indicator: {
			Indicator:   indicator,
			Mean:        mean,
			Variance:    variance,
			SampleCount: count,
		},
	}
}

func adaptiveTestCfg() AdaptiveThresholdConfig {
	return AdaptiveThresholdConfig{
		Enabled:           true,
		MinSamples:        10,
		PersonalWeight:    0.5,
		GlobalSuperhumanZ: 3.0,
	}
}

// A sudden drop from the player's own baseline should be detected via the
// personal track.
func TestAdaptive_PersonalBaselineSuddenDeviation(t *testing.T) {
	personal := establishedBaseline(baselineIndicatorResponseMean, 1000, 10000, 30) // mean 1000ms, std 100ms
	eval := NewAdaptiveThresholdEvaluator(personal, nil, adaptiveTestCfg())

	// Observed 500ms = 5 std below personal mean.
	dev := eval.Evaluate(baselineIndicatorResponseMean, 500)
	if dev.ThresholdSource != "personal" {
		t.Fatalf("expected personal source, got %q", dev.ThresholdSource)
	}
	if dev.Score <= 0 {
		t.Fatalf("expected positive score for sudden personal deviation, got %.2f", dev.Score)
	}
	if dev.PersonalZ < 4 {
		t.Fatalf("expected personal z >= 4, got %.2f", dev.PersonalZ)
	}
}

// An absolute super-human value must be detected via the global track even when
// it matches the player's own (abnormal) baseline.
func TestAdaptive_GlobalSuperhumanNotSuppressed(t *testing.T) {
	// Player's personal baseline is itself abnormal (always ~50ms): a long-term
	// cheater. Personal track would see no deviation.
	personal := establishedBaseline(baselineIndicatorResponseMean, 50, 25, 30) // mean 50ms, std 5ms
	global := map[string]GlobalBaselineStat{
		baselineIndicatorResponseMean: {Mean: 1200, StdDev: 300}, // population ~1200ms
	}
	eval := NewAdaptiveThresholdEvaluator(personal, global, adaptiveTestCfg())

	// Observed 50ms matches personal baseline but is ~3.8 std below population.
	dev := eval.Evaluate(baselineIndicatorResponseMean, 50)
	if dev.ThresholdSource != "global" {
		t.Fatalf("expected global source to dominate, got %q", dev.ThresholdSource)
	}
	if dev.Score <= 0 {
		t.Fatalf("expected positive score from global super-human track, got %.2f", dev.Score)
	}
}

// With an unestablished personal baseline (too few samples) and an in-range global
// value, no contribution should be produced.
func TestAdaptive_NoBaselineNoScore(t *testing.T) {
	personal := establishedBaseline(baselineIndicatorResponseMean, 1000, 10000, 3) // below MinSamples
	global := map[string]GlobalBaselineStat{
		baselineIndicatorResponseMean: {Mean: 1000, StdDev: 100},
	}
	eval := NewAdaptiveThresholdEvaluator(personal, global, adaptiveTestCfg())

	dev := eval.Evaluate(baselineIndicatorResponseMean, 1010) // within global range, personal not ready
	if dev.Score != 0 {
		t.Fatalf("expected zero score, got %.2f (source %q)", dev.Score, dev.ThresholdSource)
	}
}

// Provenance fields must be populated on the engine result when adaptive
// thresholding is enabled.
func TestAdaptive_EngineRecordsProvenance(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.Optimization.AdaptiveThreshold.Enabled = true
	cfg.Optimization.AdaptiveThreshold.MinSamples = 5
	cfg.Optimization.AdaptiveThreshold.ContributionWeight = 0.3

	engine := NewRiskScoringEngine(cfg)
	NewBuiltInStrategies().RegisterAll(engine)

	ctx := &DetectionContext{
		PlayerUID:     1,
		RoomID:        "room_adaptive",
		ResponseTimes: []int64{40, 45, 42, 41, 43},
		AccountAgeDays: 30,
		PersonalBaselines: establishedBaseline(baselineIndicatorResponseMean, 1000, 10000, 30),
	}

	result, err := engine.CalculateRiskScore(ctx)
	if err != nil {
		t.Fatalf("calculate: %v", err)
	}
	if result.ThresholdSource == "" {
		t.Fatalf("expected threshold source to be recorded")
	}
	if len(result.AdaptiveDeviations) == 0 {
		t.Fatalf("expected adaptive deviations recorded for provenance")
	}
	if _, ok := result.EffectiveWeights["adaptive_threshold"]; !ok {
		t.Fatalf("expected adaptive_threshold effective weight recorded")
	}
}

// When the feature is disabled the engine must produce no adaptive provenance and
// behave exactly as before.
func TestAdaptive_DisabledNoEffect(t *testing.T) {
	cfg := NewDefaultConfig() // adaptive disabled by default
	engine := NewRiskScoringEngine(cfg)
	NewBuiltInStrategies().RegisterAll(engine)

	ctx := &DetectionContext{
		PlayerUID:         1,
		RoomID:            "room_off",
		ResponseTimes:     []int64{40, 45, 42},
		AccountAgeDays:    30,
		PersonalBaselines: establishedBaseline(baselineIndicatorResponseMean, 1000, 10000, 30),
	}

	result, err := engine.CalculateRiskScore(ctx)
	if err != nil {
		t.Fatalf("calculate: %v", err)
	}
	if result.ThresholdSource != "" || len(result.AdaptiveDeviations) != 0 {
		t.Fatalf("expected no adaptive provenance when disabled, got source=%q devs=%d", result.ThresholdSource, len(result.AdaptiveDeviations))
	}
	if _, ok := result.Dimensions["adaptive_threshold"]; ok {
		t.Fatalf("adaptive_threshold dimension must not exist when disabled")
	}
}
