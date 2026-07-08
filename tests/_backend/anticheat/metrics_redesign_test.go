package anticheat

import (
	"testing"
)

// 指标重设计的检测器与评分单测（docs/anticheat/METRICS_REDESIGN.md）。

func TestDecisionOptimalityDetector_HighOptimalityScores(t *testing.T) {
	d := NewDecisionOptimalityDetector(15, 0.6, 0.15)
	// 20 次决策全部最优（rate=1.0），显著高于人群均值 0.6 → 高分。
	ctx := &DetectionContext{TotalDecisions: 20, OptimalDecisions: 20}
	score, err := d.Detect(ctx)
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	if score < 50 {
		t.Errorf("expected high score for 100%% optimality, got %.1f", score)
	}
}

func TestDecisionOptimalityDetector_NormalNoScore(t *testing.T) {
	d := NewDecisionOptimalityDetector(15, 0.6, 0.15)
	// 60% 最优 = 人群均值，不应计分。
	ctx := &DetectionContext{TotalDecisions: 20, OptimalDecisions: 12}
	score, _ := d.Detect(ctx)
	if score != 0 {
		t.Errorf("expected 0 for at-mean optimality, got %.1f", score)
	}
}

func TestDecisionOptimalityDetector_SmallSampleDecays(t *testing.T) {
	d := NewDecisionOptimalityDetector(15, 0.6, 0.15)
	full := &DetectionContext{TotalDecisions: 20, OptimalDecisions: 20}
	small := &DetectionContext{TotalDecisions: 5, OptimalDecisions: 5}
	fullScore, _ := d.Detect(full)
	smallScore, _ := d.Detect(small)
	if smallScore >= fullScore {
		t.Errorf("small sample should decay: small=%.1f full=%.1f", smallScore, fullScore)
	}
}

func TestThinkTimeDetector_SuperhumanScores(t *testing.T) {
	d := NewThinkTimeDetector(5)
	// 10 个复杂决策全部超人（零思考）→ 满分。
	ctx := &DetectionContext{ComplexDecisionCount: 10, SuperhumanDecisionCount: 10}
	score, _ := d.Detect(ctx)
	if score < 90 {
		t.Errorf("expected near-max for all-superhuman, got %.1f", score)
	}
}

func TestThinkTimeDetector_NoComplexNoScore(t *testing.T) {
	d := NewThinkTimeDetector(5)
	ctx := &DetectionContext{ComplexDecisionCount: 0, SuperhumanDecisionCount: 0}
	score, _ := d.Detect(ctx)
	if score != 0 {
		t.Errorf("expected 0 when no complex decisions, got %.1f", score)
	}
}

func TestRecentPerformanceDetector_HighWinRateScores(t *testing.T) {
	d := NewRecentPerformanceDetector(0.85, 10)
	ctx := &DetectionContext{HasRecentPerf: true, RecentGames: 20, RecentWinRate: 0.98, OpponentStrength: 1.0}
	score, _ := d.Detect(ctx)
	if score <= 0 {
		t.Errorf("expected positive score for 98%% recent win rate, got %.1f", score)
	}
}

func TestRecentPerformanceDetector_BelowThresholdNoScore(t *testing.T) {
	d := NewRecentPerformanceDetector(0.85, 10)
	ctx := &DetectionContext{HasRecentPerf: true, RecentGames: 20, RecentWinRate: 0.70, OpponentStrength: 1.0}
	score, _ := d.Detect(ctx)
	if score != 0 {
		t.Errorf("expected 0 below threshold, got %.1f", score)
	}
}

func TestMultiAccountDetector_PassThrough(t *testing.T) {
	d := NewMultiAccountDetector()
	on := &DetectionContext{HasMultiAccount: true, MultiAccountScore: 90}
	off := &DetectionContext{HasMultiAccount: false, MultiAccountScore: 90}
	if s, _ := d.Detect(on); s != 90 {
		t.Errorf("expected 90 pass-through, got %.1f", s)
	}
	if s, _ := d.Detect(off); s != 0 {
		t.Errorf("expected 0 when no multi-account signal, got %.1f", s)
	}
}

// TestStrongEvidenceFloor 验证强证据下限：单一决定性指标不被弱维度稀释。
func TestStrongEvidenceFloor(t *testing.T) {
	if f := strongEvidenceFloor("decision_optimality", 90); f != 60 {
		t.Errorf("expected floor 60 for strong decision_optimality, got %.1f", f)
	}
	if f := strongEvidenceFloor("decision_optimality", 50); f != 0 {
		t.Errorf("expected no floor below trigger, got %.1f", f)
	}
	if f := strongEvidenceFloor("player_reports", 100); f != 0 {
		t.Errorf("non-core metric should have no floor, got %.1f", f)
	}
}

// TestCalculateRiskScore_FloorAppliesUnderDilution 端到端：仅 decision_optimality
// 极端异常、其他维度为 0 时，加权平均会很低，但强证据 floor 应把风险分抬到 ≥60。
func TestCalculateRiskScore_FloorAppliesUnderDilution(t *testing.T) {
	engine := NewRiskScoringEngine(NewDefaultConfig())
	if err := NewBuiltInStrategies().RegisterAll(engine); err != nil {
		t.Fatalf("register: %v", err)
	}
	ctx := &DetectionContext{
		PlayerUID:      1,
		RoomID:         "room-1",
		TotalDecisions: 20, OptimalDecisions: 20, // decision_optimality 满
		// think_time / recent_performance / multi_account 全部无信号 → 0
	}
	result, err := engine.CalculateRiskScore(ctx)
	if err != nil {
		t.Fatalf("calculate: %v", err)
	}
	if result.RiskScore < 60 {
		t.Errorf("strong evidence floor should lift diluted score to >=60, got %.1f", result.RiskScore)
	}
}
