package repository

import (
	"chemistryuno/backend/database"
	"encoding/json"
	"errors"
	"strings"
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

func (cr *CheatRepository) GetUserBanStatus(uid uint) (bannedUntil *time.Time, frozenUntil *time.Time, banReason string, err error) {
	if !cr.db.Migrator().HasTable(&database.User{}) {
		return nil, nil, "", nil
	}
	var user database.User
	err = cr.db.Select("banned_until, frozen_until, ban_reason").First(&user, uid).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, "", nil
	}
	if err != nil {
		return nil, nil, "", err
	}
	return user.BannedUntil, user.FrozenUntil, user.BanReason, nil
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

func (cr *CheatRepository) UpdateRiskScoreReview(id uint, reviewStatus string, punishmentDecision string) error {
	updates := map[string]interface{}{
		"review_status": reviewStatus,
	}
	if punishmentDecision != "" {
		updates["punishment_decision"] = punishmentDecision
	}
	return cr.db.Model(&database.CheatRiskScore{}).Where("id = ?", id).Updates(updates).Error
}

func (cr *CheatRepository) GetLatestRiskScoreByPlayer(playerUID uint) (*database.CheatRiskScore, error) {
	var score database.CheatRiskScore
	if err := cr.db.Where("player_uid = ?", playerUID).
		Order("detection_time DESC, created_at DESC, id DESC").
		First(&score).Error; err != nil {
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

func (cr *CheatRepository) GetGameHistoriesForPlayerSince(playerUID uint, since time.Time) ([]database.GameHistory, error) {
	if !cr.db.Migrator().HasTable(&database.GameHistory{}) {
		return nil, nil
	}
	var histories []database.GameHistory
	query := cr.db.Where("(finished_at >= ? OR created_at >= ?)", since, since).
		Order("finished_at ASC, created_at ASC, id ASC")
	if err := query.Find(&histories).Error; err != nil {
		return nil, err
	}

	result := make([]database.GameHistory, 0, len(histories))
	for _, history := range histories {
		if gameHistoryIncludesPlayer(history, playerUID) {
			result = append(result, history)
		}
	}
	return result, nil
}

func gameHistoryIncludesPlayer(history database.GameHistory, playerUID uint) bool {
	if len(history.Players) == 0 {
		return false
	}
	var players []uint
	if err := json.Unmarshal(history.Players, &players); err == nil {
		for _, uid := range players {
			if uid == playerUID {
				return true
			}
		}
	}
	var ints []int
	if err := json.Unmarshal(history.Players, &ints); err == nil {
		for _, uid := range ints {
			if uid > 0 && uint(uid) == playerUID {
				return true
			}
		}
	}
	return false
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

func (cr *CheatRepository) IsReplayProtected(gameHistoryID uint, roomID string, replayID string) (bool, []string, error) {
	reasons := []string{}
	if cr == nil || cr.db == nil {
		return false, reasons, nil
	}
	check := func(model interface{}, label string) error {
		query := cr.db.Model(model)
		conditions := []string{}
		args := []interface{}{}
		if gameHistoryID > 0 {
			conditions = append(conditions, "game_history_id = ?")
			args = append(args, gameHistoryID)
		}
		if replayID != "" {
			conditions = append(conditions, "replay_id = ?")
			args = append(args, replayID)
		}
		if roomID != "" {
			conditions = append(conditions, "room_id = ?")
			args = append(args, roomID)
		}
		if len(conditions) == 0 {
			return nil
		}
		var count int64
		if err := query.Where(strings.Join(conditions, " OR "), args...).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			reasons = append(reasons, label)
		}
		return nil
	}
	if err := check(&database.CheatRiskScore{}, "risk_score"); err != nil {
		return false, reasons, err
	}
	if err := check(&database.CheatSanction{}, "sanction"); err != nil {
		return false, reasons, err
	}
	if err := check(&database.CheatAppeal{}, "appeal"); err != nil {
		return false, reasons, err
	}
	if err := check(&database.CheatAuditLog{}, "audit"); err != nil {
		return false, reasons, err
	}
	return len(reasons) > 0, reasons, nil
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

func (cr *CheatRepository) UpdateSanctionDecision(sanctionID uint, sanctionType string, reason string, duration *int, effectiveUntil *time.Time) error {
	updates := map[string]interface{}{
		"sanction_type":   sanctionType,
		"reason":          reason,
		"duration":        duration,
		"effective_until": effectiveUntil,
		"status":          "active",
	}
	return cr.db.Model(&database.CheatSanction{}).Where("id = ?", sanctionID).Updates(updates).Error
}

// GetActiveSanctionsByPlayer 获取玩家的活跃处罚
func (cr *CheatRepository) GetActiveSanctionsByPlayer(playerUID uint) ([]database.CheatSanction, error) {
	var sanctions []database.CheatSanction
	if err := cr.db.Where("player_uid = ? AND (status = ? OR status = ? OR status IS NULL)", playerUID, "active", "").
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
		Where("player_uid = ? AND sanction_type = ? AND (status = ? OR status = ? OR status IS NULL)", playerUID, "ban", "active", "").
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

// ===== Player behavior baselines (adaptive threshold) =====

// GetPlayerBaseline returns the baseline row for a player/indicator pair, or nil if absent.
func (cr *CheatRepository) GetPlayerBaseline(playerUID uint, indicator string) (*database.PlayerBehaviorBaseline, error) {
	if !cr.db.Migrator().HasTable(&database.PlayerBehaviorBaseline{}) {
		return nil, nil
	}
	var baseline database.PlayerBehaviorBaseline
	err := cr.db.Where("player_uid = ? AND indicator = ?", playerUID, indicator).First(&baseline).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &baseline, nil
}

// GetPlayerBaselines returns all baseline rows for a player keyed by indicator.
func (cr *CheatRepository) GetPlayerBaselines(playerUID uint) (map[string]database.PlayerBehaviorBaseline, error) {
	result := make(map[string]database.PlayerBehaviorBaseline)
	if !cr.db.Migrator().HasTable(&database.PlayerBehaviorBaseline{}) {
		return result, nil
	}
	var rows []database.PlayerBehaviorBaseline
	if err := cr.db.Where("player_uid = ?", playerUID).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.Indicator] = row
	}
	return result, nil
}

// GetPlayerBaselinesMulti returns baselines for multiple players in a single query.
// Returns a map of playerUID -> (indicator -> baseline).
func (cr *CheatRepository) GetPlayerBaselinesMulti(playerUIDs []uint) (map[uint]map[string]database.PlayerBehaviorBaseline, error) {
	result := make(map[uint]map[string]database.PlayerBehaviorBaseline)

	if len(playerUIDs) == 0 {
		return result, nil
	}

	if !cr.db.Migrator().HasTable(&database.PlayerBehaviorBaseline{}) {
		return result, nil
	}

	// Single query with IN clause
	var rows []database.PlayerBehaviorBaseline
	if err := cr.db.Where("player_uid IN ?", playerUIDs).Find(&rows).Error; err != nil {
		return nil, err
	}

	// Group by player UID
	for _, row := range rows {
		if result[row.PlayerUID] == nil {
			result[row.PlayerUID] = make(map[string]database.PlayerBehaviorBaseline)
		}
		result[row.PlayerUID][row.Indicator] = row
	}

	return result, nil
}


// UpsertPlayerBaseline inserts or updates a player's baseline for an indicator.
func (cr *CheatRepository) UpsertPlayerBaseline(baseline *database.PlayerBehaviorBaseline) error {
	if baseline == nil {
		return nil
	}
	existing, err := cr.GetPlayerBaseline(baseline.PlayerUID, baseline.Indicator)
	if err != nil {
		return err
	}
	if existing == nil {
		return cr.db.Create(baseline).Error
	}
	return cr.db.Model(&database.PlayerBehaviorBaseline{}).Where("id = ?", existing.ID).Updates(map[string]interface{}{
		"mean":            baseline.Mean,
		"variance":        baseline.Variance,
		"sample_count":    baseline.SampleCount,
		"window_size":     baseline.WindowSize,
		"window_kind":     baseline.WindowKind,
		"last_sampled_at": baseline.LastSampledAt,
	}).Error
}

// ===== Rule test records (offline sandbox) =====

// SaveRuleTest persists a rule-test run summary.
func (cr *CheatRepository) SaveRuleTest(test *database.AnticheatRuleTest) error {
	return cr.db.Create(test).Error
}

// GetRecentRuleTests returns the most recent rule-test runs.
func (cr *CheatRepository) GetRecentRuleTests(limit int) ([]database.AnticheatRuleTest, error) {
	if limit <= 0 {
		limit = 20
	}
	var tests []database.AnticheatRuleTest
	if err := cr.db.Order("created_at DESC").Limit(limit).Find(&tests).Error; err != nil {
		return nil, err
	}
	return tests, nil
}

// GetRecentRiskScoresForSampling returns recent risk score records usable as rule-test samples.
func (cr *CheatRepository) GetRecentRiskScoresForSampling(limit int) ([]database.CheatRiskScore, error) {
	if limit <= 0 {
		limit = 100
	}
	var scores []database.CheatRiskScore
	if err := cr.db.Order("detection_time DESC, id DESC").Limit(limit).Find(&scores).Error; err != nil {
		return nil, err
	}
	return scores, nil
}

// ===== Player risk profile (new-player protection + risk decay inputs) =====

// PlayerRiskProfile summarizes the data needed to drive the optimization features
// for one player at detection time.
type PlayerRiskProfile struct {
	TotalGames                int
	AccountAgeDays            int
	HistoricalRisk            float64
	NormalGamesSinceViolation int
	LastViolationAt           *time.Time
}

// GetPlayerRiskProfile builds the optimization profile for a player. It excludes
// confirmed-sanction/ban records from the decaying historical risk (only
// un-escalated risk decays) and counts normal games since the last violation.
func (cr *CheatRepository) GetPlayerRiskProfile(playerUID uint) (*PlayerRiskProfile, error) {
	profile := &PlayerRiskProfile{}

	// Account age + total games from the user/game-history tables.
	if cr.db.Migrator().HasTable(&database.User{}) {
		var user database.User
		if err := cr.db.Select("created_at").First(&user, playerUID).Error; err == nil {
			if !user.CreatedAt.IsZero() {
				profile.AccountAgeDays = int(time.Since(user.CreatedAt).Hours() / 24)
			}
		}
	}

	// Find the most recent confirmed violation (a sanction that was applied, i.e.
	// escalated). Confirmed records do not decay and anchor the time floor.
	if cr.db.Migrator().HasTable(&database.CheatSanction{}) {
		var sanction database.CheatSanction
		err := cr.db.Where("player_uid = ? AND sanction_type IN ?", playerUID, []string{"mute", "ban"}).
			Order("applied_at DESC, id DESC").First(&sanction).Error
		if err == nil {
			t := sanction.AppliedAt
			profile.LastViolationAt = &t
		}
	}

	// Sum un-escalated historical risk: risk scores that never became an active
	// mute/ban sanction. We approximate by summing risk from scores whose
	// punishment_decision is observe/warning/none (not mute/ban).
	if cr.db.Migrator().HasTable(&database.CheatRiskScore{}) {
		var scores []database.CheatRiskScore
		q := cr.db.Where("player_uid = ?", playerUID).
			Where("punishment_decision NOT IN ?", []string{"mute", "ban"}).
			Order("detection_time DESC, id DESC").Limit(50)
		if err := q.Find(&scores).Error; err == nil {
			var maxRisk float64
			for _, s := range scores {
				if s.RiskScore > maxRisk {
					maxRisk = s.RiskScore
				}
			}
			profile.HistoricalRisk = maxRisk
		}

		// Count normal games (risk below observe threshold) since the last violation.
		since := time.Time{}
		if profile.LastViolationAt != nil {
			since = *profile.LastViolationAt
		}
		var normalCount int64
		nq := cr.db.Model(&database.CheatRiskScore{}).
			Where("player_uid = ? AND punishment_decision IN ?", playerUID, []string{"none", "observe"})
		if !since.IsZero() {
			nq = nq.Where("detection_time > ?", since)
		}
		if err := nq.Count(&normalCount).Error; err == nil {
			profile.NormalGamesSinceViolation = int(normalCount)
		}
	}

	return profile, nil
}

// GetPlayerRiskProfilesMulti returns risk profiles for multiple players in optimized batch queries.
// Returns a map of playerUID -> profile.
func (cr *CheatRepository) GetPlayerRiskProfilesMulti(playerUIDs []uint) (map[uint]*PlayerRiskProfile, error) {
	result := make(map[uint]*PlayerRiskProfile)

	if len(playerUIDs) == 0 {
		return result, nil
	}

	// Initialize profiles
	for _, uid := range playerUIDs {
		result[uid] = &PlayerRiskProfile{}
	}

	// Batch: Account age from users table
	if cr.db.Migrator().HasTable(&database.User{}) {
		var users []database.User
		if err := cr.db.Select("uid, created_at").Where("uid IN ?", playerUIDs).Find(&users).Error; err == nil {
			for _, user := range users {
				if profile, ok := result[user.UID]; ok && !user.CreatedAt.IsZero() {
					profile.AccountAgeDays = int(time.Since(user.CreatedAt).Hours() / 24)
				}
			}
		}
	}

	// Batch: Most recent violations
	if cr.db.Migrator().HasTable(&database.CheatSanction{}) {
		var sanctions []database.CheatSanction
		// Get the most recent mute/ban for each player
		subquery := cr.db.Table("cheat_sanctions").
			Select("player_uid, MAX(id) as max_id").
			Where("player_uid IN ? AND sanction_type IN ?", playerUIDs, []string{"mute", "ban"}).
			Group("player_uid")

		if err := cr.db.Table("cheat_sanctions").
			Joins("INNER JOIN (?) as latest ON cheat_sanctions.id = latest.max_id", subquery).
			Find(&sanctions).Error; err == nil {
			for _, sanction := range sanctions {
				if profile, ok := result[sanction.PlayerUID]; ok {
					t := sanction.AppliedAt
					profile.LastViolationAt = &t
				}
			}
		}
	}

	// Batch: Historical risk scores
	if cr.db.Migrator().HasTable(&database.CheatRiskScore{}) {
		// Get max risk score for each player (non-escalated only)
		type MaxRisk struct {
			PlayerUID uint
			MaxScore  float64
		}
		var maxRisks []MaxRisk
		if err := cr.db.Table("cheat_risk_scores").
			Select("player_uid, MAX(risk_score) as max_score").
			Where("player_uid IN ? AND punishment_decision NOT IN ?", playerUIDs, []string{"mute", "ban"}).
			Group("player_uid").
			Scan(&maxRisks).Error; err == nil {
			for _, mr := range maxRisks {
				if profile, ok := result[mr.PlayerUID]; ok {
					profile.HistoricalRisk = mr.MaxScore
				}
			}
		}

		// Batch: Count normal games since violation for each player
		for uid, profile := range result {
			since := time.Time{}
			if profile.LastViolationAt != nil {
				since = *profile.LastViolationAt
			}

			var normalCount int64
			nq := cr.db.Model(&database.CheatRiskScore{}).
				Where("player_uid = ? AND punishment_decision IN ?", uid, []string{"none", "observe"})
			if !since.IsZero() {
				nq = nq.Where("detection_time > ?", since)
			}
			if err := nq.Count(&normalCount).Error; err == nil {
				profile.NormalGamesSinceViolation = int(normalCount)
			}
		}
	}

	return result, nil
}

