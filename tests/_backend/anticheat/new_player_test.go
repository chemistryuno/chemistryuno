package anticheat

import (
	"testing"
	"time"
)

func newPlayerEngine(t *testing.T, enabled bool, relax float64) *RiskScoringEngine {
	t.Helper()
	cfg := NewDefaultConfig()
	cfg.Optimization.NewPlayer.Enabled = enabled
	cfg.Optimization.NewPlayer.MinGames = 30
	cfg.Optimization.NewPlayer.MinAgeDays = 7
	cfg.Optimization.NewPlayer.RelaxationFactor = relax
	engine := NewRiskScoringEngine(cfg)
	NewBuiltInStrategies().RegisterAll(engine)
	return engine
}

// highRiskContext builds a context that scores high without optimization features.
func highRiskContext(roomID string, isNew bool) *DetectionContext {
	now := time.Now()
	return &DetectionContext{
		PlayerUID:      1,
		RoomID:         roomID,
		ResponseTimes:  []int64{20, 22, 21, 19, 20, 23, 18},
		OperationCount: 30,
		WinCount:       98,
		TotalGames:     100,
		AccountAgeDays: 60,
		IsNewPlayer:    isNew,
		OperationTimes: []time.Time{
			now, now.Add(30 * time.Millisecond), now.Add(60 * time.Millisecond),
			now.Add(90 * time.Millisecond), now.Add(120 * time.Millisecond),
		},
	}
}

// Observation period relaxes the score and flags the result.
func TestNewPlayer_RelaxesScore(t *testing.T) {
	engine := newPlayerEngine(t, true, 0.5)

	full, _ := engine.CalculateRiskScore(highRiskContext("room_full", false))
	relaxed, _ := engine.CalculateRiskScore(highRiskContext("room_new", true))

	if !relaxed.NewPlayerObserve {
		t.Fatalf("expected NewPlayerObserve flag set for new player")
	}
	if relaxed.RiskScore >= full.RiskScore {
		t.Fatalf("expected relaxed score (%.2f) below full score (%.2f)", relaxed.RiskScore, full.RiskScore)
	}
}

// Disabled feature leaves new players scored normally.
func TestNewPlayer_DisabledNoEffect(t *testing.T) {
	engine := newPlayerEngine(t, false, 0.5)
	result, _ := engine.CalculateRiskScore(highRiskContext("room_x", true))
	if result.NewPlayerObserve {
		t.Fatalf("NewPlayerObserve must be false when feature disabled")
	}
}

// Multi-account contribution bypasses relaxation: a new player with a strong
// multi_account signal keeps that portion of the score un-relaxed.
func TestNewPlayer_MultiAccountNotShielded(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.Optimization.NewPlayer.Enabled = true
	cfg.Optimization.NewPlayer.RelaxationFactor = 0.0 // fully relax the relaxable portion
	cfg.Dimensions["multi_account"] = DimensionConfig{Weight: 0.5}

	engine := NewRiskScoringEngine(cfg)
	NewBuiltInStrategies().RegisterAll(engine)
	engine.RegisterStrategy(&fixedScoreDetector{name: "multi_account", score: 100})
	// Ensure multi_account is in the enabled list.
	cfg.EnabledStrategies = append(cfg.EnabledStrategies, "multi_account")
	engine.UpdateConfig(cfg)

	ctx := highRiskContext("room_smurf", true)
	result, err := engine.CalculateRiskScore(ctx)
	if err != nil {
		t.Fatalf("calculate: %v", err)
	}
	// Even with relaxation factor 0, the protected multi-account contribution must
	// keep the score above zero.
	if result.RiskScore <= 0 {
		t.Fatalf("expected multi-account contribution to survive relaxation, score=%.2f", result.RiskScore)
	}
}

// fixedScoreDetector is a test detector returning a constant score.
type fixedScoreDetector struct {
	name  string
	score float64
}

func (f *fixedScoreDetector) Name() string { return f.name }
func (f *fixedScoreDetector) Detect(_ *DetectionContext) (float64, error) {
	return f.score, nil
}
