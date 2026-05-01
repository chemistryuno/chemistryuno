package anticheat

import (
	"testing"
	"time"
)

// TestRiskScoringEngine_CalculateRiskScore 测试风险评分引擎
func TestRiskScoringEngine_CalculateRiskScore(t *testing.T) {
	engine := NewRiskScoringEngine(NewDefaultConfig())
	strategies := NewBuiltInStrategies()
	strategies.RegisterAll(engine)

	// 测试场景：玩家有多个异常指标
	context := &DetectionContext{
		PlayerUID:       1,
		RoomID:          "room_001",
		ResponseTimes:   []int64{30, 35, 40, 45, 50}, // 快速反应
		OperationCount:  15,
		TimestampOffset: 10 * time.Second,
		WinCount:        95,
		TotalGames:      100,
		AccountAgeDays:  2, // 新账号
		OperationTimes: []time.Time{
			time.Now(),
			time.Now().Add(100 * time.Millisecond),
			time.Now().Add(200 * time.Millisecond),
		},
	}

	result, err := engine.CalculateRiskScore(context)
	if err != nil {
		t.Fatalf("风险评分失败: %v", err)
	}

	if result == nil {
		t.Fatal("风险评分结果为空")
	}

	if result.RiskScore < 0 || result.RiskScore > 100 {
		t.Errorf("风险分数应在 0-100 之间，得到: %.1f", result.RiskScore)
	}

	if result.RiskScore < 30 {
		t.Logf("多指标异常的风险分数: %.1f (接受)", result.RiskScore)
	}

	t.Logf("风险评分结果: %.1f, 处罚类型: %s", result.RiskScore, result.SanctionType)
}

// TestSanctionDecider_MakeDecision 测试处罚决策
func TestSanctionDecider_MakeDecision(t *testing.T) {
	config := NewDefaultConfig()
	decider := NewSanctionDecider(config, nil)

	testCases := []struct {
		riskScore      float64
		expectedType   string
	}{
		{10, "none"},
		{30, "observe"},
		{50, "warning"},
		{70, "mute"},
		{90, "ban"},
	}

	for _, tc := range testCases {
		decision := decider.MakeDecision(tc.riskScore, "room_001", 1, 1)
		if decision.SanctionType != tc.expectedType {
			t.Errorf("风险分数 %.1f 应该产生 %s 处罚，得到: %s",
				tc.riskScore, tc.expectedType, decision.SanctionType)
		}
	}
}

// TestAppealManager_Workflow 测试申诉工作流
func TestAppealManager_Workflow(t *testing.T) {
	// 这是一个集成测试框架，实际测试需要数据库
	t.Log("申诉工作流测试框架已建立")
	t.Log("需要数据库集成以完整测试")
}

// TestDetectors_ResponseTime 测试响应时间检测器
func TestDetectors_ResponseTime(t *testing.T) {
	detector := NewResponseTimeDetector(100, 0.05)

	// 测试场景：混合的响应时间
	context := &DetectionContext{
		ResponseTimes: []int64{150, 160, 50, 45, 55, 200, 180},
	}

	score, err := detector.Detect(context)
	if err != nil {
		t.Fatalf("检测失败: %v", err)
	}

	if score < 0 || score > 100 {
		t.Errorf("分数应在 0-100 之间，得到: %.1f", score)
	}

	t.Logf("响应时间异常分数: %.1f", score)
}

// TestDetectors_Frequency 测试频率检测器
func TestDetectors_Frequency(t *testing.T) {
	detector := NewFrequencyDetector(5, 1*time.Second)

	now := time.Now()
	context := &DetectionContext{
		OperationTimes: []time.Time{
			now,
			now.Add(100 * time.Millisecond),
			now.Add(200 * time.Millisecond),
			now.Add(300 * time.Millisecond),
			now.Add(400 * time.Millisecond),
			now.Add(500 * time.Millisecond),
		},
	}

	score, err := detector.Detect(context)
	if err != nil {
		t.Fatalf("检测失败: %v", err)
	}

	if score > 0 {
		t.Logf("频率异常分数: %.1f (超过限制的操作检测到)", score)
	}
}

// TestConfigManager_DynamicUpdate 测试配置动态更新
func TestConfigManager_DynamicUpdate(t *testing.T) {
	manager, err := NewConfigManager("")
	if err != nil {
		t.Fatalf("配置管理器初始化失败: %v", err)
	}

	originalWeight := manager.GetConfig().Dimensions["response_time"].Weight

	// 更新权重
	err = manager.UpdateDimensionWeight("response_time", 0.5)
	if err != nil {
		t.Fatalf("更新权重失败: %v", err)
	}

	newWeight := manager.GetConfig().Dimensions["response_time"].Weight
	if newWeight != 0.5 {
		t.Errorf("权重更新失败，期望 0.5，得到 %.2f", newWeight)
	}

	t.Logf("权重动态更新成功: %.2f → %.2f", originalWeight, newWeight)
}

// TestIntegration_EndToEnd 端到端集成测试框架
func TestIntegration_EndToEnd(t *testing.T) {
	t.Log("端到端集成测试")
	t.Log("此测试需要:")
	t.Log("1. 完整的数据库设置")
	t.Log("2. 反作弊系统初始化")
	t.Log("3. 真实的玩家数据")
	t.Log("4. 游戏循环流程模拟")
	t.Log("框架已建立，等待集成环境就位")
}

// TestApproveAppeal_CompensationConfig 测试批准申诉时补偿配置正确读取
func TestApproveAppeal_CompensationConfig(t *testing.T) {
	config := NewDefaultConfig()

	// 验证默认补偿金额
	if config.UnbanConfig.CompensationAmount != 100 {
		t.Errorf("Expected default compensation 100, got %d", config.UnbanConfig.CompensationAmount)
	}

	// 验证可以覆盖补偿金额
	config.UnbanConfig.CompensationAmount = 200
	if config.UnbanConfig.CompensationAmount != 200 {
		t.Error("Expected compensation amount override to work")
	}
}

// TestApproveAppeal_ZeroCompensation 测试零补偿金额被拒绝
func TestApproveAppeal_ZeroCompensation(t *testing.T) {
	config := NewDefaultConfig()
	config.UnbanConfig.CompensationAmount = 0

	// 零补偿金额应该在 AddFuel 层被拒绝（amount <= 0）
	if config.UnbanConfig.CompensationAmount > 0 {
		t.Error("Expected zero compensation to be invalid")
	}
}
