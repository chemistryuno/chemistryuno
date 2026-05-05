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

// AuditLogFilter describes admin audit search criteria.
type AuditLogFilter struct {
	PlayerUID            *uint
	StartTime            *time.Time
	EndTime              *time.Time
	ActionType           string
	CompensationStatuses []string
	Limit                int
	Offset               int
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
	query := cr.db.Model(&database.CheatRiskScore{})
	if playerUID > 0 {
		query = query.Where("player_uid = ?", playerUID)
	}
	if err := query.Order("detection_time DESC").
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
func (cr *CheatRepository) GetCheatDetectedGameHistories(limit int) ([]database.GameHistory, error) {
	if !cr.db.Migrator().HasTable(&database.GameHistory{}) {
		return nil, nil
	}
	if limit <= 0 {
		limit = 50
	}
	var histories []database.GameHistory
	if err := cr.db.Where("cheat_detected = ?", true).
		Order("finished_at DESC, created_at DESC").
		Limit(limit).
		Find(&histories).Error; err != nil {
		return nil, err
	}
	return histories, nil
}

func (cr *CheatRepository) GetGameHistoryByID(id uint) (*database.GameHistory, error) {
	if !cr.db.Migrator().HasTable(&database.GameHistory{}) {
		return nil, gorm.ErrRecordNotFound
	}
	var history database.GameHistory
	if err := cr.db.First(&history, id).Error; err != nil {
		return nil, err
	}
	return &history, nil
}

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

// GetSanctionsByRiskScoreIDs returns sanctions keyed by their risk score id.
func (cr *CheatRepository) GetSanctionsByRiskScoreIDs(riskScoreIDs []uint) (map[uint][]database.CheatSanction, error) {
	result := make(map[uint][]database.CheatSanction)
	if len(riskScoreIDs) == 0 {
		return result, nil
	}

	var sanctions []database.CheatSanction
	if err := cr.db.Where("risk_score_id IN ?", riskScoreIDs).
		Order("applied_at DESC").
		Find(&sanctions).Error; err != nil {
		return nil, err
	}
	for _, sanction := range sanctions {
		result[sanction.RiskScoreID] = append(result[sanction.RiskScoreID], sanction)
	}
	return result, nil
}

// UpdateSanctionStatus 更新处罚状态
func (cr *CheatRepository) UpdateSanctionStatus(sanctionID uint, status string) error {
	return cr.db.Model(&database.CheatSanction{}).
		Where("id = ?", sanctionID).
		Update("status", status).Error
}

func (cr *CheatRepository) RevokeActiveBanSanctionsByPlayer(playerUID uint) error {
	return cr.db.Model(&database.CheatSanction{}).
		Where("player_uid = ? AND sanction_type = ? AND status = ?", playerUID, "ban", "active").
		Update("status", "revoked").Error
}

func (cr *CheatRepository) ClearPlayerAccountBan(playerUID uint) error {
	if !cr.db.Migrator().HasTable(&database.User{}) {
		return nil
	}
	return cr.db.Model(&database.User{}).Where("uid = ?", playerUID).Updates(map[string]interface{}{
		"banned_until": nil,
		"ban_reason":   "",
	}).Error
}

// CountBansInRange counts ban sanctions issued during a time range.
func (cr *CheatRepository) CountBansInRange(startTime, endTime time.Time) (int64, error) {
	var count int64
	if err := cr.db.Model(&database.CheatSanction{}).
		Where("sanction_type = ? AND applied_at BETWEEN ? AND ?", "ban", startTime, endTime).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
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
	query := cr.db.Model(&database.CheatAppeal{})
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Order("submitted_at DESC").Find(&appeals).Error; err != nil {
		return nil, err
	}
	return appeals, nil
}

func (cr *CheatRepository) HasPendingAppealForContext(playerUID uint, riskScoreID uint, sanctionID *uint) (bool, error) {
	query := cr.db.Model(&database.CheatAppeal{}).
		Where("player_uid = ? AND status IN ?", playerUID, []string{"pending", "under_review"})
	if sanctionID != nil && *sanctionID > 0 {
		query = query.Where("sanction_id = ?", *sanctionID)
	} else if riskScoreID > 0 {
		query = query.Where("risk_score_id = ?", riskScoreID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
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
		"status":        status,
		"reviewed_at":   time.Now(),
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

func (cr *CheatRepository) auditLogQuery(filter AuditLogFilter) *gorm.DB {
	query := cr.db.Model(&database.CheatAuditLog{})
	if filter.PlayerUID != nil {
		query = query.Where("player_uid = ?", *filter.PlayerUID)
	}
	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", *filter.EndTime)
	}
	if filter.ActionType != "" {
		query = query.Where("event_type = ? OR sanction_type = ?", filter.ActionType, filter.ActionType)
	}
	if len(filter.CompensationStatuses) > 0 {
		query = query.Where("compensation_status IN ?", filter.CompensationStatuses)
	}
	return query
}

// QueryAuditLogs returns filtered audit logs and a total count before pagination.
func (cr *CheatRepository) QueryAuditLogs(filter AuditLogFilter) ([]database.CheatAuditLog, int64, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 1000 {
		filter.Limit = 1000
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	var total int64
	countQuery := cr.auditLogQuery(filter)
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []database.CheatAuditLog
	if err := cr.auditLogQuery(filter).
		Order("created_at DESC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

// ExportAuditLogs returns filtered audit logs for CSV export.
func (cr *CheatRepository) ExportAuditLogs(filter AuditLogFilter, limit int) ([]database.CheatAuditLog, error) {
	if limit <= 0 || limit > 10000 {
		limit = 10000
	}

	var logs []database.CheatAuditLog
	if err := cr.auditLogQuery(filter).
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

// UpdateAuditCompensation records the final compensation status on an audit row.
func (cr *CheatRepository) UpdateAuditCompensation(auditLogID uint, status string, note string) error {
	now := time.Now()
	return cr.db.Model(&database.CheatAuditLog{}).Where("id = ?", auditLogID).Updates(map[string]interface{}{
		"compensation_status": status,
		"compensation_note":   note,
		"compensation_date":   now,
	}).Error
}
