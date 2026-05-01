package anticheat

import (
	"encoding/json"
	"testing"
	"time"
)

// TestFastReactionChecker_RecordAndDetect 测试快速反应记录和检测
func TestFastReactionChecker_RecordAndDetect(t *testing.T) {
	checker := NewFastReactionChecker()

	// 模拟用户快速反应
	checker.RecordFastReaction(1, true)
	checker.RecordFastReaction(1, true)
	checker.RecordFastReaction(2, false)
	checker.RecordFastReaction(3, true)

	result := checker.DetectCheat()

	if !result.CheatDetected {
		t.Error("Expected cheat detected, but got false")
	}

	if len(result.CheatUIDs) != 2 {
		t.Errorf("Expected 2 cheat UIDs, but got %d", len(result.CheatUIDs))
	}

	if result.CheatUIDs[0] != 1 || result.CheatUIDs[1] != 3 {
		t.Errorf("Expected [1, 3], but got %v", result.CheatUIDs)
	}
}

// TestFastReactionChecker_NoCheat 测试没有作弊时的检测
func TestFastReactionChecker_NoCheat(t *testing.T) {
	checker := NewFastReactionChecker()

	// 所有用户都没有快速反应
	checker.RecordFastReaction(1, false)
	checker.RecordFastReaction(2, false)

	result := checker.DetectCheat()

	if result.CheatDetected {
		t.Error("Expected no cheat detected, but got true")
	}

	if len(result.CheatUIDs) != 0 {
		t.Errorf("Expected 0 cheat UIDs, but got %d", len(result.CheatUIDs))
	}
}

// TestSnapshotBuilder 测试回放快照构建器
func TestSnapshotBuilder(t *testing.T) {
	builder := NewSnapshotBuilder(12345)
	now := time.Now()

	participants, _ := json.Marshal([]int{1, 2, 3})
	events := []map[string]interface{}{
		{"type": "move", "player": 1},
		{"type": "skip", "player": 2},
	}
	cheatResult := CheatDetectionResult{
		CheatDetected: true,
		CheatUIDs:     []int{1},
	}

	snapshot, err := builder.
		WithParticipants(participants).
		WithEvents(events).
		WithCheatDetection(cheatResult).
		WithReason("game_finished").
		WithStartedAt(now).
		WithGameStatus("finished", []int{2}, 3, 0).
		Build()

	if err != nil {
		t.Errorf("Failed to build snapshot: %v", err)
	}

	var result ReplaySnapshot
	err = json.Unmarshal([]byte(snapshot), &result)
	if err != nil {
		t.Errorf("Failed to unmarshal snapshot: %v", err)
	}

	if result.RoomID != 12345 {
		t.Errorf("Expected room_id 12345, but got %d", result.RoomID)
	}

	if !result.CheatDetected {
		t.Error("Expected cheat_detected to be true")
	}

	if len(result.CheatUIDs) != 1 || result.CheatUIDs[0] != 1 {
		t.Errorf("Expected cheat_uids [1], but got %v", result.CheatUIDs)
	}
}

// TestSuspiciousActivityDetector_IsSuspicious 测试可疑活动检测
func TestSuspiciousActivityDetector_IsSuspicious(t *testing.T) {
	windowSize := 10 * time.Second
	maxActions := 5
	detector := NewSuspiciousActivityDetector(windowSize, maxActions)

	uid := 1
	// 记录6个动作在短时间内（超过限制）
	for i := 0; i < 6; i++ {
		detector.RecordAction(uid)
	}

	if !detector.IsSuspicious(uid) {
		t.Error("Expected user to be suspicious")
	}

	// 测试不可疑的用户
	uid2 := 2
	for i := 0; i < 3; i++ {
		detector.RecordAction(uid2)
	}

	if detector.IsSuspicious(uid2) {
		t.Error("Expected user 2 to not be suspicious")
	}
}

// TestSuspiciousActivityDetector_GetSuspiciousUsers 测试获取可疑用户列表
func TestSuspiciousActivityDetector_GetSuspiciousUsers(t *testing.T) {
	windowSize := 10 * time.Second
	maxActions := 3
	detector := NewSuspiciousActivityDetector(windowSize, maxActions)

	// 用户1：4个动作（可疑）
	for i := 0; i < 4; i++ {
		detector.RecordAction(1)
	}

	// 用户2：2个动作（正常）
	for i := 0; i < 2; i++ {
		detector.RecordAction(2)
	}

	// 用户3：5个动作（可疑）
	for i := 0; i < 5; i++ {
		detector.RecordAction(3)
	}

	suspiciousUIDs := detector.GetSuspiciousUsers()

	if len(suspiciousUIDs) != 2 {
		t.Errorf("Expected 2 suspicious users, but got %d", len(suspiciousUIDs))
	}

	if suspiciousUIDs[0] != 1 || suspiciousUIDs[1] != 3 {
		t.Errorf("Expected [1, 3], but got %v", suspiciousUIDs)
	}
}

// TestCheatReportManager 测试作弊举报管理
func TestCheatReportManager(t *testing.T) {
	manager := NewCheatReportManager()

	report := &CheatReport{
		RoomID:      12345,
		ReportedUID: 1,
		ReporterUID: 2,
		Reason:      "Fast reaction",
	}

	manager.SubmitReport(report)

	// 获取举报
	reports := manager.GetReportsByUID(1)
	if len(reports) != 1 {
		t.Errorf("Expected 1 report, but got %d", len(reports))
	}

	if reports[0].Reason != "Fast reaction" {
		t.Errorf("Expected reason 'Fast reaction', but got %s", reports[0].Reason)
	}
}

// TestUnbanConfig_Defaults 测试解封补偿配置默认值
func TestUnbanConfig_Defaults(t *testing.T) {
	config := NewDefaultConfig()

	if config.UnbanConfig.CompensationAmount != 100 {
		t.Errorf("Expected default compensation amount 100, got %d", config.UnbanConfig.CompensationAmount)
	}

	if config.UnbanConfig.DefaultMessage == "" {
		t.Error("Expected non-empty default unban message")
	}
}

// TestUnbanConfig_Serialization 测试解封配置序列化
func TestUnbanConfig_Serialization(t *testing.T) {
	config := NewDefaultConfig()
	config.UnbanConfig.CompensationAmount = 200

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	var restored RiskScoringConfig
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if restored.UnbanConfig.CompensationAmount != 200 {
		t.Errorf("Expected compensation amount 200 after round-trip, got %d", restored.UnbanConfig.CompensationAmount)
	}
}
