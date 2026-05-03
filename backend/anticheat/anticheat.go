package anticheat

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"
)

var (
	ErrDimensionNotFound   = fmt.Errorf("维度不存在")
	ErrInvalidSanctionType = fmt.Errorf("无效的处罚类型")
	ErrStrategyNotFound    = fmt.Errorf("策略不存在")
	ErrInvalidConfig       = fmt.Errorf("无效的反作弊配置")
)

// CheatDetectionResult 作弊检测结果
type CheatDetectionResult struct {
	CheatDetected bool  `json:"cheat_detected"` // 是否检测到作弊
	CheatUIDs     []int `json:"cheat_uids"`     // 作弊用户ID列表
}

// FastReactionChecker 快速反应检查器
type FastReactionChecker struct {
	FastReactionUIDs map[int]int // uid -> count 快速反应计数
}

// NewFastReactionChecker 创建快速反应检查器
func NewFastReactionChecker() *FastReactionChecker {
	return &FastReactionChecker{
		FastReactionUIDs: make(map[int]int),
	}
}

// RecordFastReaction 记录快速反应
// uid: 用户ID
// isFastReaction: 是否是快速反应
func (f *FastReactionChecker) RecordFastReaction(uid int, isFastReaction bool) {
	if isFastReaction {
		f.FastReactionUIDs[uid]++
	}
}

// GetFastReactionCount 获取用户的快速反应计数
func (f *FastReactionChecker) GetFastReactionCount(uid int) int {
	return f.FastReactionUIDs[uid]
}

// DetectCheat 检测是否有快速反应作弊行为
func (f *FastReactionChecker) DetectCheat() CheatDetectionResult {
	cheatUIDs := make([]int, 0)
	for uid, count := range f.FastReactionUIDs {
		if count > 0 {
			cheatUIDs = append(cheatUIDs, uid)
		}
	}
	sort.Ints(cheatUIDs)

	return CheatDetectionResult{
		CheatDetected: len(cheatUIDs) > 0,
		CheatUIDs:     cheatUIDs,
	}
}

// ReplaySnapshot 回放快照（用于记录游戏过程）
type ReplaySnapshot struct {
	Version             int                      `json:"version"`
	RoomID              int64                    `json:"room_id"`
	GeneratedAt         string                   `json:"generated_at"`
	Participants        json.RawMessage          `json:"participants"`
	Events              []map[string]interface{} `json:"events"`
	CheatDetected       bool                     `json:"cheat_detected"`
	CheatUIDs           []int                    `json:"cheat_uids"`
	Reason              string                   `json:"reason,omitempty"`
	StartedAt           string                   `json:"started_at,omitempty"`
	Status              string                   `json:"status,omitempty"`
	FinishedPlayers     []int                    `json:"finished_players,omitempty"`
	OriginalPlayerCount int                      `json:"original_player_count,omitempty"`
	QuittedCount        int                      `json:"quitted_count,omitempty"`
}

// SnapshotBuilder 回放快照构建器
type SnapshotBuilder struct {
	snapshot ReplaySnapshot
}

// NewSnapshotBuilder 创建快照构建器
func NewSnapshotBuilder(roomID int64) *SnapshotBuilder {
	return &SnapshotBuilder{
		snapshot: ReplaySnapshot{
			Version:     1,
			RoomID:      roomID,
			GeneratedAt: time.Now().Format(time.RFC3339),
			Events:      make([]map[string]interface{}, 0),
			CheatUIDs:   make([]int, 0),
		},
	}
}

// WithParticipants 添加参与者信息
func (sb *SnapshotBuilder) WithParticipants(participants json.RawMessage) *SnapshotBuilder {
	sb.snapshot.Participants = participants
	return sb
}

// WithEvents 添加事件列表
func (sb *SnapshotBuilder) WithEvents(events []map[string]interface{}) *SnapshotBuilder {
	sb.snapshot.Events = events
	return sb
}

// WithCheatDetection 添加作弊检测结果
func (sb *SnapshotBuilder) WithCheatDetection(result CheatDetectionResult) *SnapshotBuilder {
	sb.snapshot.CheatDetected = result.CheatDetected
	sb.snapshot.CheatUIDs = result.CheatUIDs
	return sb
}

// WithReason 添加原因
func (sb *SnapshotBuilder) WithReason(reason string) *SnapshotBuilder {
	sb.snapshot.Reason = reason
	return sb
}

// WithStartedAt 添加开始时间
func (sb *SnapshotBuilder) WithStartedAt(startedAt time.Time) *SnapshotBuilder {
	if !startedAt.IsZero() {
		sb.snapshot.StartedAt = startedAt.Format(time.RFC3339)
	}
	return sb
}

// WithGameStatus 添加游戏状态信息
func (sb *SnapshotBuilder) WithGameStatus(status string, finishedPlayers []int, originalPlayerCount, quittedCount int) *SnapshotBuilder {
	sb.snapshot.Status = status
	sb.snapshot.FinishedPlayers = finishedPlayers
	sb.snapshot.OriginalPlayerCount = originalPlayerCount
	sb.snapshot.QuittedCount = quittedCount
	return sb
}

// Build 构建快照并返回JSON字符串
func (sb *SnapshotBuilder) Build() (string, error) {
	encoded, err := json.Marshal(sb.snapshot)
	if err != nil {
		log.Printf("[反作弊] 生成回放快照失败: %v", err)
		return "", err
	}
	return string(encoded), nil
}

// CheatReport 作弊举报
type CheatReport struct {
	RoomID      int64     `json:"room_id"`
	ReportedUID int       `json:"reported_uid"`
	ReporterUID int       `json:"reporter_uid"`
	Reason      string    `json:"reason"`
	Evidence    string    `json:"evidence,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	Status      string    `json:"status"` // "pending", "investigating", "confirmed", "dismissed"
}

// CheatReportManager 作弊举报管理器
type CheatReportManager struct {
	reports map[int64]*CheatReport // reportID -> CheatReport
}

// NewCheatReportManager 创建作弊举报管理器
func NewCheatReportManager() *CheatReportManager {
	return &CheatReportManager{
		reports: make(map[int64]*CheatReport),
	}
}

// SubmitReport 提交举报
func (crm *CheatReportManager) SubmitReport(report *CheatReport) {
	report.CreatedAt = time.Now()
	report.Status = "pending"
	// 使用时间戳作为举报ID
	reportID := int64(time.Now().UnixNano())
	crm.reports[reportID] = report
}

// GetReport 获取举报
func (crm *CheatReportManager) GetReport(reportID int64) *CheatReport {
	return crm.reports[reportID]
}

// UpdateReportStatus 更新举报状态
func (crm *CheatReportManager) UpdateReportStatus(reportID int64, status string) bool {
	if report, exists := crm.reports[reportID]; exists {
		report.Status = status
		return true
	}
	return false
}

// GetReportsByUID 获取某个用户相关的所有举报
func (crm *CheatReportManager) GetReportsByUID(uid int) []*CheatReport {
	var reports []*CheatReport
	for _, report := range crm.reports {
		if report.ReportedUID == uid {
			reports = append(reports, report)
		}
	}
	return reports
}

// SuspiciousActivityDetector 可疑活动检测器
type SuspiciousActivityDetector struct {
	windowSize         time.Duration       // 时间窗口大小
	maxActionsInWindow int                 // 时间窗口内的最大动作数
	userActivityLog    map[int][]time.Time // uid -> 动作时间列表
}

// NewSuspiciousActivityDetector 创建可疑活动检测器
func NewSuspiciousActivityDetector(windowSize time.Duration, maxActionsInWindow int) *SuspiciousActivityDetector {
	return &SuspiciousActivityDetector{
		windowSize:         windowSize,
		maxActionsInWindow: maxActionsInWindow,
		userActivityLog:    make(map[int][]time.Time),
	}
}

// RecordAction 记录用户动作
func (sad *SuspiciousActivityDetector) RecordAction(uid int) {
	sad.userActivityLog[uid] = append(sad.userActivityLog[uid], time.Now())
}

// IsSuspicious 检查用户是否有可疑活动
func (sad *SuspiciousActivityDetector) IsSuspicious(uid int) bool {
	if actions, exists := sad.userActivityLog[uid]; exists {
		now := time.Now()
		windowStart := now.Add(-sad.windowSize)

		// 统计时间窗口内的动作数
		count := 0
		for _, actionTime := range actions {
			if actionTime.After(windowStart) {
				count++
			}
		}

		return count > sad.maxActionsInWindow
	}
	return false
}

// GetSuspiciousUsers 获取所有可疑用户
func (sad *SuspiciousActivityDetector) GetSuspiciousUsers() []int {
	var suspiciousUIDs []int
	for uid := range sad.userActivityLog {
		if sad.IsSuspicious(uid) {
			suspiciousUIDs = append(suspiciousUIDs, uid)
		}
	}
	sort.Ints(suspiciousUIDs)
	return suspiciousUIDs
}

// ClearOldLogs 清除过期的活动日志（避免内存泄漏）
func (sad *SuspiciousActivityDetector) ClearOldLogs() {
	now := time.Now()
	windowStart := now.Add(-sad.windowSize * 2) // 清除2倍窗口外的日志

	for uid, actions := range sad.userActivityLog {
		var validActions []time.Time
		for _, actionTime := range actions {
			if actionTime.After(windowStart) {
				validActions = append(validActions, actionTime)
			}
		}

		if len(validActions) == 0 {
			delete(sad.userActivityLog, uid)
		} else {
			sad.userActivityLog[uid] = validActions
		}
	}
}
