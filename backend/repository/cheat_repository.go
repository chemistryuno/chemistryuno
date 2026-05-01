package repository

import (
	"chemistryuno/backend/database"
	"time"

	"gorm.io/gorm"
)

// CheatRepository 反作弊数据库操作仓库
type CheatRepository struct {
	db *gorm.DB
}

// NewCheatRepository 创建反作弊仓库
func NewCheatRepository(db *gorm.DB) *CheatRepository {
	return &CheatRepository{db: db}
}

// SaveRiskScore 保存风险评分
func (cr *CheatRepository) SaveRiskScore(score *database.CheatRiskScore) error {
	return cr.db.Create(score).Error
}

// GetRiskScoreByID 根据ID获取风险评分
func (cr *CheatRepository) GetRiskScoreByID(id uint) (*database.CheatRiskScore, error) {
	var score database.CheatRiskScore
	if err := cr.db.First(&score, id).Error; err != nil {
		return nil, err
	}
	return &score, nil
}

// GetRiskScoresByPlayer 获取玩家的风险评分历史
func (cr *CheatRepository) GetRiskScoresByPlayer(playerUID uint, limit int) ([]database.CheatRiskScore, error) {
	var scores []database.CheatRiskScore
	if err := cr.db.Where("player_uid = ?", playerUID).
		Order("detection_time DESC").
		Limit(limit).
		Find(&scores).Error; err != nil {
		return nil, err
	}
	return scores, nil
}

// GetRiskScoresByRoom 获取房间的所有风险评分
func (cr *CheatRepository) GetRiskScoresByRoom(roomID string) ([]database.CheatRiskScore, error) {
	var scores []database.CheatRiskScore
	if err := cr.db.Where("room_id = ?", roomID).
		Order("detection_time DESC").
		Find(&scores).Error; err != nil {
		return nil, err
	}
	return scores, nil
}

// SaveSanction 保存处罚记录
func (cr *CheatRepository) SaveSanction(sanction *database.CheatSanction) error {
	return cr.db.Create(sanction).Error
}

// GetSanctionByID 根据ID获取处罚
func (cr *CheatRepository) GetSanctionByID(id uint) (*database.CheatSanction, error) {
	var sanction database.CheatSanction
	if err := cr.db.First(&sanction, id).Error; err != nil {
		return nil, err
	}
	return &sanction, nil
}

// GetActiveSanctionsByPlayer 获取玩家的活跃处罚
func (cr *CheatRepository) GetActiveSanctionsByPlayer(playerUID uint) ([]database.CheatSanction, error) {
	var sanctions []database.CheatSanction
	if err := cr.db.Where("player_uid = ? AND status = ?", playerUID, "active").
		Order("applied_at DESC").
		Find(&sanctions).Error; err != nil {
		return nil, err
	}
	return sanctions, nil
}

// UpdateSanctionStatus 更新处罚状态
func (cr *CheatRepository) UpdateSanctionStatus(sanctionID uint, status string) error {
	return cr.db.Model(&database.CheatSanction{}).
		Where("id = ?", sanctionID).
		Update("status", status).Error
}

// SaveAppeal 保存申诉
func (cr *CheatRepository) SaveAppeal(appeal *database.CheatAppeal) error {
	return cr.db.Create(appeal).Error
}

// GetAppealByID 根据ID获取申诉
func (cr *CheatRepository) GetAppealByID(id uint) (*database.CheatAppeal, error) {
	var appeal database.CheatAppeal
	if err := cr.db.First(&appeal, id).Error; err != nil {
		return nil, err
	}
	return &appeal, nil
}

// GetPendingAppeals 获取待审核的申诉
func (cr *CheatRepository) GetPendingAppeals(limit int) ([]database.CheatAppeal, error) {
	var appeals []database.CheatAppeal
	if err := cr.db.Where("status IN ?", []string{"pending", "under_review"}).
		Order("submitted_at ASC").
		Limit(limit).
		Find(&appeals).Error; err != nil {
		return nil, err
	}
	return appeals, nil
}

// GetAppealsByPlayer 获取玩家的申诉历史
func (cr *CheatRepository) GetAppealsByPlayer(playerUID uint) ([]database.CheatAppeal, error) {
	var appeals []database.CheatAppeal
	if err := cr.db.Where("player_uid = ?", playerUID).
		Order("submitted_at DESC").
		Find(&appeals).Error; err != nil {
		return nil, err
	}
	return appeals, nil
}

// UpdateAppealStatus 更新申诉状态
func (cr *CheatRepository) UpdateAppealStatus(appealID uint, status string, reviewerUID *uint, remark string) error {
	updates := map[string]interface{}{
		"status":       status,
		"reviewed_at":  time.Now(),
		"review_remark": remark,
	}
	if reviewerUID != nil {
		updates["reviewer_uid"] = *reviewerUID
	}
	return cr.db.Model(&database.CheatAppeal{}).
		Where("id = ?", appealID).
		Updates(updates).Error
}

// SaveAuditLog 保存审计日志
func (cr *CheatRepository) SaveAuditLog(log *database.CheatAuditLog) error {
	return cr.db.Create(log).Error
}

// GetAuditLogsByPlayer 获取玩家的审计日志
func (cr *CheatRepository) GetAuditLogsByPlayer(playerUID uint, limit int) ([]database.CheatAuditLog, error) {
	var logs []database.CheatAuditLog
	if err := cr.db.Where("player_uid = ?", playerUID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// GetAuditLogsByRoom 获取房间的审计日志
func (cr *CheatRepository) GetAuditLogsByRoom(roomID string) ([]database.CheatAuditLog, error) {
	var logs []database.CheatAuditLog
	if err := cr.db.Where("room_id = ?", roomID).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// GetAuditLogsByTimeRange 按时间范围获取审计日志
func (cr *CheatRepository) GetAuditLogsByTimeRange(startTime, endTime time.Time, limit int) ([]database.CheatAuditLog, error) {
	var logs []database.CheatAuditLog
	if err := cr.db.Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// GetAuditLogsByEventType 按事件类型获取审计日志
func (cr *CheatRepository) GetAuditLogsByEventType(eventType string, limit int) ([]database.CheatAuditLog, error) {
	var logs []database.CheatAuditLog
	if err := cr.db.Where("event_type = ?", eventType).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// CountRiskScoresInRange 统计时间范围内的风险评分数
func (cr *CheatRepository) CountRiskScoresInRange(startTime, endTime time.Time) (int64, error) {
	var count int64
	if err := cr.db.Model(&database.CheatRiskScore{}).
		Where("detection_time BETWEEN ? AND ?", startTime, endTime).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetRiskScoreDistribution 获取风险分数分布统计
func (cr *CheatRepository) GetRiskScoreDistribution(startTime, endTime time.Time) (map[string]int64, error) {
	distribution := make(map[string]int64)

	// 统计不同范围的风险分数数量
	ranges := []struct {
		name string
		min  float64
		max  float64
	}{
		{"observe_20_40", 20, 40},
		{"warning_40_60", 40, 60},
		{"mute_60_80", 60, 80},
		{"ban_80_100", 80, 100},
	}

	for _, r := range ranges {
		var count int64
		if err := cr.db.Model(&database.CheatRiskScore{}).
			Where("risk_score >= ? AND risk_score < ? AND detection_time BETWEEN ? AND ?",
				r.min, r.max, startTime, endTime).
			Count(&count).Error; err != nil {
			return nil, err
		}
		distribution[r.name] = count
	}

	return distribution, nil
}

// UpdateAppealCompensation 更新申诉的补偿状态
func (cr *CheatRepository) UpdateAppealCompensation(appealID uint, status string, amount int, note string) error {
	return cr.db.Model(&database.CheatAppeal{}).Where("id = ?", appealID).Updates(map[string]interface{}{
		"compensation_status": status,
		"compensation_amount": amount,
		"compensation_note":   note,
	}).Error
}
