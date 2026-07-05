package anticheat

import (
	"testing"
)

// Lowering the ban threshold should escalate a borderline sample's tier without
// touching any live state.
func TestRuleTest_ThresholdChangeEscalates(t *testing.T) {
	draft := NewDefaultConfig()
	// Make the ban band start lower so a 70-score sample becomes a ban.
	draft.SanctionThresholds.MuteMin = 60
	draft.SanctionThresholds.MuteMax = 69
	draft.SanctionThresholds.BanMin = 70
	draft.SanctionThresholds.BanMax = 100

	samples := []RuleTestSample{
		{
			RiskScoreID:     1,
			PlayerUID:       42,
			LiveScore:       70,
			LiveTier:        "mute",
			DimensionScores: map[string]float64{"response_time": 70, "frequency": 70, "win_rate": 70, "pattern": 70, "account_age": 70},
		},
	}

	result, err := RunRuleTest(draft, samples)
	if err != nil {
		t.Fatalf("rule test: %v", err)
	}
	if result.SampleCount != 1 {
		t.Fatalf("expected 1 sample, got %d", result.SampleCount)
	}
	out := result.Outcomes[0]
	if out.DraftTier != "ban" {
		t.Fatalf("expected draft tier ban, got %q (score %.2f)", out.DraftTier, out.DraftScore)
	}
	if !out.TierChanged || result.Escalations != 1 {
		t.Fatalf("expected one escalation, got changed=%v escalations=%d", out.TierChanged, result.Escalations)
	}
	if result.HitDistribution["ban"] != 1 {
		t.Fatalf("expected ban hit distribution 1, got %+v", result.HitDistribution)
	}
}

// An identical draft config should produce no tier changes.
func TestRuleTest_NoChangeWhenConfigEqual(t *testing.T) {
	draft := NewDefaultConfig()
	samples := []RuleTestSample{
		{
			RiskScoreID:     1,
			LiveScore:       30,
			LiveTier:        "observe",
			DimensionScores: map[string]float64{"response_time": 30, "frequency": 30, "win_rate": 30, "pattern": 30, "account_age": 30},
		},
	}
	result, err := RunRuleTest(draft, samples)
	if err != nil {
		t.Fatalf("rule test: %v", err)
	}
	if result.Escalations != 0 || result.Deescalations != 0 {
		t.Fatalf("expected no tier changes, got esc=%d deesc=%d", result.Escalations, result.Deescalations)
	}
}

// An invalid draft config must be rejected before any evaluation.
func TestRuleTest_RejectsInvalidConfig(t *testing.T) {
	draft := NewDefaultConfig()
	draft.SanctionThresholds.BanMin = 200 // out of range
	if _, err := RunRuleTest(draft, nil); err == nil {
		t.Fatalf("expected invalid config to be rejected")
	}
}
