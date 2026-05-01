package anticheat

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"encoding/json"
	"log"
	"time"
)

// AuditLogger 审计日志记录器
type AuditLogger struct {
	repository *repository.CheatRepository
}

// NewAuditLogger 创建审计日志记录器
func NewAuditLogger(repo *repository.CheatRepository) *AuditLogger {
	return &AuditLogger{
		repository: repo,
	}
}

// LogDetection 记录风险检测日志
func (al *AuditLogger) LogDetection(roomID string, playerUID uint, riskScoreID uint, riskScore float64, dimensions map[string]float64) error {
	details, _ := json.Marshal(dimensions)

	auditLog := &database.CheatAuditLog{
		EventType:   "detection",
		RoomID:      roomID,
		PlayerUID:   playerUID,
		RiskScoreID: &riskScoreID,
		RiskScore:   &riskScore,
		Details:     details,
		Remark:      "自动检测",
	}

	if err := al.repository.SaveAuditLog(auditLog); err != nil {
		errorLog := &database.CheatAuditLog{
			EventType: "detection_error",
			RoomID:    roomID,
			PlayerUID: playerUID,
			Remark:    "检测日志记录失败: " + err.Error(),
		}
		al.repository.SaveAuditLog(errorLog)
		return err
	}

	log.Printf("[审计] 记录检测日志: 房间 %s, 玩家 %d, 风险分数: %.1f", roomID, playerUID, riskScore)
	return nil
}

// LogSanction 记录处罚日志
func (al *AuditLogger) LogSanction(roomID string, playerUID uint, riskScoreID, sanctionID uint, sanctionType string, reason string) error {
	auditLog := &database.CheatAuditLog{
		EventType:   "sanction",
		RoomID:      roomID,
		PlayerUID:   playerUID,
		RiskScoreID: &riskScoreID,
		SanctionID:  &sanctionID,
		SanctionType: sanctionType,
		NewStatus:   "active",
		Remark:      reason,
	}

	if err := al.repository.SaveAuditLog(auditLog); err != nil {
		return err
	}

	log.Printf("[审计] 记录处罚日志: 房间 %s, 玩家 %d, 处罚类型: %s", roomID, playerUID, sanctionType)
	return nil
}

// LogAppealSubmitted 记录申诉提交日志
func (al *AuditLogger) LogAppealSubmitted(roomID string, playerUID uint, appealID uint, reason string) error {
	auditLog := &database.CheatAuditLog{
		EventType:   "appeal",
		RoomID:      roomID,
		PlayerUID:   playerUID,
		AppealID:    &appealID,
		NewStatus:   "pending",
		Remark:      reason,
	}

	if err := al.repository.SaveAuditLog(auditLog); err != nil {
		return err
	}

	log.Printf("[审计] 记录申诉日志: 房间 %s, 玩家 %d, 申诉 ID: %d", roomID, playerUID, appealID)
	return nil
}

// LogAppealReview 记录申诉审核日志
func (al *AuditLogger) LogAppealReview(playerUID uint, appealID uint, reviewerUID uint, oldStatus, newStatus string, remark string) error {
	auditLog := &database.CheatAuditLog{
		EventType:   "review",
		PlayerUID:   playerUID,
		OperatorUID: &reviewerUID,
		AppealID:    &appealID,
		OldStatus:   oldStatus,
		NewStatus:   newStatus,
		Remark:      remark,
	}

	if err := al.repository.SaveAuditLog(auditLog); err != nil {
		return err
	}

	log.Printf("[审计] 记录申诉审核日志: 申诉 %d, 审核人 %d, 结果: %s", appealID, reviewerUID, newStatus)
	return nil
}

// LogSanctionRevoked 记录处罚撤销日志
func (al *AuditLogger) LogSanctionRevoked(playerUID uint, sanctionID uint, reviewerUID uint, reason string) error {
	auditLog := &database.CheatAuditLog{
		EventType:   "revoke",
		PlayerUID:   playerUID,
		OperatorUID: &reviewerUID,
		SanctionID:  &sanctionID,
		OldStatus:   "active",
		NewStatus:   "revoked",
		Remark:      reason,
	}

	if err := al.repository.SaveAuditLog(auditLog); err != nil {
		return err
	}

	log.Printf("[审计] 记录处罚撤销日志: 处罚 %d, 撤销人 %d", sanctionID, reviewerUID)
	return nil
}

// GetPlayerAuditLogs 获取玩家的审计日志
func (al *AuditLogger) GetPlayerAuditLogs(playerUID uint, limit int) ([]database.CheatAuditLog, error) {
	logs, err := al.repository.GetAuditLogsByPlayer(playerUID, limit)
	if err != nil {
		log.Printf("[审计] 查询玩家审计日志失败: %v", err)
		return nil, err
	}
	return logs, nil
}

// GetRoomAuditLogs 获取房间的审计日志
func (al *AuditLogger) GetRoomAuditLogs(roomID string) ([]database.CheatAuditLog, error) {
	logs, err := al.repository.GetAuditLogsByRoom(roomID)
	if err != nil {
		log.Printf("[审计] 查询房间审计日志失败: %v", err)
		return nil, err
	}
	return logs, nil
}

// GetAuditLogsByTimeRange 按时间范围获取审计日志
func (al *AuditLogger) GetAuditLogsByTimeRange(startTime, endTime time.Time, limit int) ([]database.CheatAuditLog, error) {
	logs, err := al.repository.GetAuditLogsByTimeRange(startTime, endTime, limit)
	if err != nil {
		log.Printf("[审计] 按时间范围查询审计日志失败: %v", err)
		return nil, err
	}
	return logs, nil
}

// GetAuditLogsByEventType 按事件类型获取审计日志
func (al *AuditLogger) GetAuditLogsByEventType(eventType string, limit int) ([]database.CheatAuditLog, error) {
	logs, err := al.repository.GetAuditLogsByEventType(eventType, limit)
	if err != nil {
		log.Printf("[审计] 按事件类型查询审计日志失败: %v", err)
		return nil, err
	}
	return logs, nil
}

// ExportAuditLogs 导出审计日志
func (al *AuditLogger) ExportAuditLogs(startTime, endTime time.Time) ([]database.CheatAuditLog, error) {
	return al.GetAuditLogsByTimeRange(startTime, endTime, 10000) // 导出最多10000条
}

// GetAuditStatistics 获取审计统计信息
func (al *AuditLogger) GetAuditStatistics(startTime, endTime time.Time) (map[string]interface{}, error) {
	detectionLogs, _ := al.GetAuditLogsByEventType("detection", 10000)
	sanctionLogs, _ := al.GetAuditLogsByEventType("sanction", 10000)
	appealLogs, _ := al.GetAuditLogsByEventType("appeal", 10000)
	reviewLogs, _ := al.GetAuditLogsByEventType("review", 10000)

	// 统计处罚类型
	sanctionTypeCount := make(map[string]int)
	for _, log := range sanctionLogs {
		if log.SanctionType != "" {
			sanctionTypeCount[log.SanctionType]++
		}
	}

	// 统计申诉结果
	appealResultCount := make(map[string]int)
	for _, log := range reviewLogs {
		if log.NewStatus != "" {
			appealResultCount[log.NewStatus]++
		}
	}

	stats := map[string]interface{}{
		"period_start":        startTime,
		"period_end":          endTime,
		"total_detections":    len(detectionLogs),
		"total_sanctions":     len(sanctionLogs),
		"total_appeals":       len(appealLogs),
		"total_reviews":       len(reviewLogs),
		"sanction_types":      sanctionTypeCount,
		"appeal_results":      appealResultCount,
	}

	return stats, nil
}
