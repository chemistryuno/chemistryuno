package anticheat

// 本文件提供集成示例和最佳实践

import (
	"log"
	"time"
)

// IntegrationExample 演示如何在游戏系统中使用反作弊模块
type IntegrationExample struct {
	FastReactionChecker *FastReactionChecker
	ActivityDetector    *SuspiciousActivityDetector
	ReportManager       *CheatReportManager
}

// NewIntegrationExample 创建集成示例
func NewIntegrationExample() *IntegrationExample {
	return &IntegrationExample{
		FastReactionChecker: NewFastReactionChecker(),
		ActivityDetector:    NewSuspiciousActivityDetector(10*time.Second, 20), // 10秒内最多20个动作
		ReportManager:       NewCheatReportManager(),
	}
}

// OnPlayerAction 玩家执行动作时调用
// 这是主要的防作弊检测点
func (ie *IntegrationExample) OnPlayerAction(uid int, actionType string, responseTime int64) {
	// 1. 检测快速反应（响应时间通常 < 100ms 被视为异常快速）
	isFastReaction := responseTime < 100
	ie.FastReactionChecker.RecordFastReaction(uid, isFastReaction)

	// 2. 记录活动用于频率检测
	ie.ActivityDetector.RecordAction(uid)

	// 3. 检查是否有可疑活动
	if ie.ActivityDetector.IsSuspicious(uid) {
		log.Printf("[反作弊] 用户 %d 活动频繁，可能存在自动化作弊 (动作: %s, 响应时间: %dms)",
			uid, actionType, responseTime)
	}

	// 4. 如果是快速反应，额外记录
	if isFastReaction {
		log.Printf("[反作弊] 用户 %d 发现快速反应 (动作: %s, 响应时间: %dms)",
			uid, actionType, responseTime)
	}
}

// OnGameEnd 游戏结束时调用
func (ie *IntegrationExample) OnGameEnd(roomID int64) CheatDetectionResult {
	// 清理过期的活动日志
	ie.ActivityDetector.ClearOldLogs()

	// 检测快速反应作弊
	result := ie.FastReactionChecker.DetectCheat()

	if result.CheatDetected {
		log.Printf("[反作弊] 游戏 %d 检测到作弊行为，涉及用户: %v", roomID, result.CheatUIDs)

		// 可选：自动创建举报
		for _, uid := range result.CheatUIDs {
			report := &CheatReport{
				RoomID:      roomID,
				ReportedUID: uid,
				ReporterUID: 0, // 系统自动检测
				Reason:      "Fast reaction detected by system",
				Evidence:    "Multiple fast reactions detected during gameplay",
			}
			ie.ReportManager.SubmitReport(report)
		}
	}

	return result
}

// OnPlayerReport 玩家举报其他玩家时调用
func (ie *IntegrationExample) OnPlayerReport(roomID int64, reportedUID int, reporterUID int, reason string) {
	report := &CheatReport{
		RoomID:      roomID,
		ReportedUID: reportedUID,
		ReporterUID: reporterUID,
		Reason:      reason,
		CreatedAt:   time.Now(),
		Status:      "pending",
	}

	ie.ReportManager.SubmitReport(report)
	log.Printf("[反作弊] 收到举报: 用户 %d 举报了用户 %d (原因: %s)", reporterUID, reportedUID, reason)
}

// BestPractices 反作弊的最佳实践说明
/*

## 1. 数据验证 (服务端验证至关重要)
- 永远不信任客户端数据
- 服务端验证所有玩家动作是否合法
- 检查动作是否符合当前游戏状态
- 验证响应时间是否物理上可能

## 2. 检测策略
- 快速反应检测：异常快速的响应时间（< 100ms）
- 频率检测：短时间内的动作过多（如10秒内>20个动作）
- 模式检测：不自然的游戏模式（完美的最优操作序列）
- 统计检测：与其他玩家的胜率明显不符

## 3. 日志记录
- 记录所有可疑行为
- 保存完整的游戏重放数据
- 包含精确的时间戳
- 便于人工审核

## 4. 举报管理
- 允许玩家举报可疑行为
- 实现举报历史跟踪
- 对持续作弊者进行惩罚
- 定期审查和验证举报

## 5. 参数调优
根据实际游戏特性调整参数：
- 快速反应阈值：根据游戏类型调整（卡牌游戏通常100ms，反应快的游戏可能50ms）
- 频率限制：根据游戏的正常操作频率调整
- 时间窗口：较大的窗口检测长期模式，较小的窗口检测瞬时异常

## 6. 用户体验
- 不要过度惩罚合法用户
- 为真正的快速反应者留有余地
- 实现申诉机制
- 保持透明的反作弊政策

## 7. 性能优化
- 定期清理过期数据
- 使用合适的数据结构
- 批量处理举报
- 异步处理重型检测操作
*/
