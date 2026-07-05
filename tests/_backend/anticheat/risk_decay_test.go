package anticheat

import (
	"testing"
	"time"
)

func decayEngine(t *testing.T, enabled bool, factor float64, floorHours int) *RiskScoringEngine {
	t.Helper()
	cfg := NewDefaultConfig()
	cfg.Optimization.RiskDecay.Enabled = enabled
	cfg.Optimization.RiskDecay.DecayFactor = factor
	cfg.Optimization.RiskDecay.MinFloorHours = floorHours
	engine := NewRiskScoringEngine(cfg)
	NewBuiltInStrategies().RegisterAll(engine)
	return engine
}

func lowRiskContext(roomID string) *DetectionContext {
	return &DetectionContext{
		PlayerUID:      1,
		RoomID:         roomID,
		ResponseTimes:  []int64{1500, 1600, 1550}, // slow, human
		AccountAgeDays: 60,
		TotalGames:     100,
		WinCount:       40,
	}
}

// Historical risk decays toward zero as normal games accumulate.
func TestRiskDecay_DecaysWithNormalGames(t *testing.T) {
	engine := decayEngine(t, true, 0.85, 0) // no time floor
	past := time.Now().Add(-1000 * time.Hour)

	fewGames := lowRiskContext("room_few")
	fewGames.HistoricalRisk = 80
	fewGames.NormalGamesSinceViolation = 1
	fewGames.LastViolationAt = &past

	manyGames := lowRiskContext("room_many")
	manyGames.HistoricalRisk = 80
	manyGames.NormalGamesSinceViolation = 40
	manyGames.LastViolationAt = &past

	rFew, _ := engine.CalculateRiskScore(fewGames)
	rMany, _ := engine.CalculateRiskScore(manyGames)

	if rMany.RiskScore >= rFew.RiskScore {
		t.Fatalf("expected more normal games to decay risk further: many=%.2f few=%.2f", rMany.RiskScore, rFew.RiskScore)
	}
	if rMany.DecayFactorApplied == nil || *rMany.DecayFactorApplied >= *rFew.DecayFactorApplied {
		t.Fatalf("expected smaller decay factor after many games")
	}
}

// The time floor prevents rapid games from fully shedding historical risk.
func TestRiskDecay_TimeFloorBlocksRapidGames(t *testing.T) {
	engine := decayEngine(t, true, 0.5, 48) // 48h floor
	justNow := time.Now().Add(-1 * time.Hour) // only 1h since violation

	ctx := lowRiskContext("room_floor")
	ctx.HistoricalRisk = 80
	ctx.NormalGamesSinceViolation = 50 // many games, but within floor window
	ctx.LastViolationAt = &justNow
	ctx.Now = time.Now()

	result, _ := engine.CalculateRiskScore(ctx)
	if result.DecayFactorApplied == nil {
		t.Fatalf("expected decay factor recorded")
	}
	// Within the floor window the factor must stay well above the pure game-driven
	// value (0.5^50 ≈ 0), so historical risk is not fully shed.
	if *result.DecayFactorApplied < 0.5 {
		t.Fatalf("expected time floor to keep decay factor high, got %.4f", *result.DecayFactorApplied)
	}
}

// Disabled decay leaves historical risk untouched (no folding).
func TestRiskDecay_DisabledNoEffect(t *testing.T) {
	engine := decayEngine(t, false, 0.85, 24)
	ctx := lowRiskContext("room_off")
	ctx.HistoricalRisk = 80
	ctx.NormalGamesSinceViolation = 1

	result, _ := engine.CalculateRiskScore(ctx)
	if result.DecayFactorApplied != nil {
		t.Fatalf("expected no decay factor when disabled")
	}
}
