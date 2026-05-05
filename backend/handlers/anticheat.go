package handlers

import (
	"chemistryuno/backend/anticheat"
	"chemistryuno/backend/cache"
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// AnticheatHandler 反作弊处理程序
type AnticheatHandler struct {
	acSystem *anticheat.System
}

type playerAnticheatStatsResponse struct {
	BansToday        int64 `json:"bans_today"`
	SystemUptimeDays int   `json:"system_uptime_days"`
}

var playerStatsFallbackLimiter = struct {
	mu       sync.Mutex
	lastSeen map[uint]time.Time
}{
	lastSeen: make(map[uint]time.Time),
}

// 全局反作弊系统实例 (由main.go初始化)
var globalAnticheatSystem *anticheat.System

// SetGlobalAnticheatSystem 设置全局反作弊系统
func SetGlobalAnticheatSystem(system *anticheat.System) {
	globalAnticheatSystem = system
}

// 全局反作弊处理程序
var globalAnticheatHandler *AnticheatHandler

// InitializeAnticheatHandler 初始化全局反作弊处理程序
func InitializeAnticheatHandler(system *anticheat.System) {
	if system != nil {
		globalAnticheatHandler = &AnticheatHandler{acSystem: system}
		globalAnticheatHandler.recordStartupTime(context.Background())
	}
}

// 处理程序包装器函数
func GetAnticheatStats(c *gin.Context) {
	if globalAnticheatHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat handler not initialized"})
		return
	}
	globalAnticheatHandler.GetAnticheatStats(c)
}

func GetPlayerAnticheatStats(c *gin.Context) {
	if globalAnticheatHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat handler not initialized"})
		return
	}
	globalAnticheatHandler.GetPlayerAnticheatStats(c)
}

func GetConfig(c *gin.Context) {
	if globalAnticheatHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat handler not initialized"})
		return
	}
	globalAnticheatHandler.GetConfig(c)
}

func UpdateConfig(c *gin.Context) {
	if globalAnticheatHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat handler not initialized"})
		return
	}
	globalAnticheatHandler.UpdateConfig(c)
}

func GetDetectionList(c *gin.Context) {
	if globalAnticheatHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat handler not initialized"})
		return
	}
	globalAnticheatHandler.GetDetectionList(c)
}

func GetDetectionDetail(c *gin.Context) {
	if globalAnticheatHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat handler not initialized"})
		return
	}
	globalAnticheatHandler.GetDetectionDetail(c)
}

func ReviewDetection(c *gin.Context) {
	if globalAnticheatHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat handler not initialized"})
		return
	}
	globalAnticheatHandler.ReviewDetection(c)
}

func GetAppealsList(c *gin.Context) {
	if globalAnticheatHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat handler not initialized"})
		return
	}
	globalAnticheatHandler.GetAppealsList(c)
}

func ApproveAppeal(c *gin.Context) {
	if globalAnticheatHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat handler not initialized"})
		return
	}
	globalAnticheatHandler.ApproveAppeal(c)
}

func RejectAppeal(c *gin.Context) {
	if globalAnticheatHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat handler not initialized"})
		return
	}
	globalAnticheatHandler.RejectAppeal(c)
}

func GetAuditLog(c *gin.Context) {
	if globalAnticheatHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat handler not initialized"})
		return
	}
	globalAnticheatHandler.GetAuditLog(c)
}

func ExportAuditLog(c *gin.Context) {
	if globalAnticheatHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat handler not initialized"})
		return
	}
	globalAnticheatHandler.ExportAuditLog(c)
}

func SubmitAppeal(c *gin.Context) {
	if globalAnticheatHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat handler not initialized"})
		return
	}
	globalAnticheatHandler.SubmitAppeal(c)
}

func GetPlayerAppeals(c *gin.Context) {
	if globalAnticheatHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat handler not initialized"})
		return
	}
	globalAnticheatHandler.GetPlayerAppeals(c)
}

func ClaimAppealCompensation(c *gin.Context) {
	if globalAnticheatHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat handler not initialized"})
		return
	}
	globalAnticheatHandler.ClaimAppealCompensation(c)
}

func GetPlayerSanctions(c *gin.Context) {
	if globalAnticheatHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat handler not initialized"})
		return
	}
	globalAnticheatHandler.GetPlayerSanctions(c)
}

func BanFromAnticheatPanel(c *gin.Context) {
	if globalAnticheatHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat handler not initialized"})
		return
	}
	globalAnticheatHandler.BanFromAnticheatPanel(c)
}

func UnbanFromAnticheatPanel(c *gin.Context) {
	if globalAnticheatHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat handler not initialized"})
		return
	}
	globalAnticheatHandler.UnbanFromAnticheatPanel(c)
}

// NewAnticheatHandler 创建反作弊处理程序
func NewAnticheatHandler(acSystem *anticheat.System) *AnticheatHandler {
	handler := &AnticheatHandler{
		acSystem: acSystem,
	}
	handler.recordStartupTime(context.Background())
	return handler
}

func (h *AnticheatHandler) recordStartupTime(ctx context.Context) {
	if h == nil || h.acSystem == nil {
		return
	}
	if h.acSystem.StartedAt.IsZero() {
		h.acSystem.StartedAt = time.Now()
	}
	if err := cache.SetAnticheatStartupTime(ctx, h.acSystem.StartedAt); err != nil {
		log.Printf("⚠️  Anticheat startup timestamp cache unavailable: %v", err)
	}
}

func getAuthenticatedUID(c *gin.Context) (uint, bool) {
	uidValue, exists := c.Get("uid")
	if !exists {
		return 0, false
	}
	switch uid := uidValue.(type) {
	case int:
		if uid <= 0 {
			return 0, false
		}
		return uint(uid), true
	case uint:
		if uid == 0 {
			return 0, false
		}
		return uid, true
	case int64:
		if uid <= 0 {
			return 0, false
		}
		return uint(uid), true
	case float64:
		if uid <= 0 {
			return 0, false
		}
		return uint(uid), true
	default:
		return 0, false
	}
}

func allowPlayerStatsRequestFallback(uid uint, now time.Time) bool {
	if uid == 0 {
		return true
	}

	playerStatsFallbackLimiter.mu.Lock()
	defer playerStatsFallbackLimiter.mu.Unlock()

	if last, ok := playerStatsFallbackLimiter.lastSeen[uid]; ok && now.Sub(last) < time.Second {
		return false
	}
	playerStatsFallbackLimiter.lastSeen[uid] = now

	for seenUID, last := range playerStatsFallbackLimiter.lastSeen {
		if now.Sub(last) > time.Minute {
			delete(playerStatsFallbackLimiter.lastSeen, seenUID)
		}
	}

	return true
}

func (h *AnticheatHandler) buildPlayerAnticheatStats(now time.Time) (playerAnticheatStatsResponse, error) {
	start := now.Add(-24 * time.Hour)
	bansToday, err := h.acSystem.Repository.CountBansInRange(start, now)
	if err != nil {
		return playerAnticheatStatsResponse{}, err
	}

	return playerAnticheatStatsResponse{
		BansToday:        bansToday,
		SystemUptimeDays: h.acSystem.UptimeDays(now),
	}, nil
}

func buildAdminConfigResponse(config *anticheat.RiskScoringConfig) gin.H {
	thresholds := config.SanctionThresholds
	return gin.H{
		"dimensions":          config.Dimensions,
		"sanction_thresholds": thresholds,
		"enabled_strategies":  config.EnabledStrategies,
		"unban":               config.UnbanConfig,
		"unban_config":        config.UnbanConfig,
		"sanctions": gin.H{
			"observe": thresholds.ObserveMin,
			"warning": thresholds.WarningMin,
			"mute":    thresholds.MuteMin,
			"ban":     thresholds.BanMin,
		},
	}
}

func parseOptionalDate(value string, endOfDay bool) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}

	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return &t, nil
	}

	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, err
	}
	if endOfDay {
		t = t.Add(24*time.Hour - time.Nanosecond)
	}
	return &t, nil
}

func parseAuditFilter(c *gin.Context, includePagination bool) (repository.AuditLogFilter, int, error) {
	filter := repository.AuditLogFilter{}

	playerID := strings.TrimSpace(firstNonEmpty(c.Query("player_id"), c.Query("player_uid")))
	if playerID != "" {
		uid64, err := strconv.ParseUint(playerID, 10, 32)
		if err != nil {
			return filter, 0, fmt.Errorf("invalid player_id")
		}
		uid := uint(uid64)
		filter.PlayerUID = &uid
	}

	start, err := parseOptionalDate(firstNonEmpty(c.Query("start_date"), c.Query("start")), false)
	if err != nil {
		return filter, 0, fmt.Errorf("invalid start_date")
	}
	end, err := parseOptionalDate(firstNonEmpty(c.Query("end_date"), c.Query("end")), true)
	if err != nil {
		return filter, 0, fmt.Errorf("invalid end_date")
	}
	filter.StartTime = start
	filter.EndTime = end

	filter.ActionType = strings.TrimSpace(firstNonEmpty(c.Query("action_type"), c.Query("action")))
	if statuses := strings.TrimSpace(c.Query("compensation_status")); statuses != "" {
		for _, status := range strings.Split(statuses, ",") {
			status = strings.TrimSpace(status)
			if status != "" {
				filter.CompensationStatuses = append(filter.CompensationStatuses, status)
			}
		}
	}

	page := 1
	limit := 20
	if includePagination {
		if raw := c.DefaultQuery("page", "1"); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed < 1 {
				return filter, 0, fmt.Errorf("invalid page")
			}
			page = parsed
		}
		if raw := c.DefaultQuery("limit", "20"); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed < 1 {
				return filter, 0, fmt.Errorf("invalid limit")
			}
			limit = parsed
		}
		if limit > 1000 {
			limit = 1000
		}
		filter.Limit = limit
		filter.Offset = (page - 1) * limit
	}

	return filter, page, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func auditAction(log database.CheatAuditLog) string {
	if log.SanctionType != "" {
		return log.SanctionType
	}
	return log.EventType
}

func auditDetails(log database.CheatAuditLog) string {
	if log.Remark != "" {
		return log.Remark
	}
	if len(log.Details) > 0 {
		return string(log.Details)
	}
	return ""
}

func auditLogDTO(log database.CheatAuditLog) gin.H {
	return gin.H{
		"id":                   log.ID,
		"player_id":            log.PlayerUID,
		"player_uid":           log.PlayerUID,
		"action":               auditAction(log),
		"action_type":          auditAction(log),
		"event_type":           log.EventType,
		"reason":               log.Remark,
		"details":              auditDetails(log),
		"created_at":           log.CreatedAt,
		"approval_note":        log.ApprovalNote,
		"compensation_amount":  log.CompensationAmount,
		"compensation_status":  log.CompensationStatus,
		"compensation_message": log.CompensationMessage,
		"compensation_note":    log.CompensationNote,
		"compensation_date":    log.CompensationDate,
	}
}

func sanctionForRiskScore(score database.CheatRiskScore, sanctions map[uint][]database.CheatSanction) *database.CheatSanction {
	items := sanctions[score.ID]
	if len(items) == 0 {
		return nil
	}
	return &items[0]
}

func firstNonZeroTime(values ...time.Time) time.Time {
	for _, value := range values {
		if !value.IsZero() {
			return value
		}
	}
	return time.Time{}
}

func parseCheatUIDs(raw database.JSON) []uint {
	if len(raw) == 0 {
		return nil
	}

	var ints []int
	if err := json.Unmarshal(raw, &ints); err == nil {
		uids := make([]uint, 0, len(ints))
		for _, uid := range ints {
			if uid > 0 {
				uids = append(uids, uint(uid))
			}
		}
		return uids
	}

	var uints []uint
	if err := json.Unmarshal(raw, &uints); err == nil {
		return uints
	}

	return nil
}

func detectionDTO(score database.CheatRiskScore, sanction *database.CheatSanction) gin.H {
	sanctionType := "none"
	var sanctionID *uint
	var sanctionStatus string
	var sanctionReason string
	var sanctionUntil *time.Time
	if sanction != nil {
		sanctionType = sanction.SanctionType
		sanctionID = &sanction.ID
		sanctionStatus = sanction.Status
		sanctionReason = sanction.Reason
		sanctionUntil = sanction.EffectiveUntil
	}
	return gin.H{
		"id":                       score.ID,
		"room_id":                  score.RoomID,
		"player_id":                score.PlayerUID,
		"player_uid":               score.PlayerUID,
		"risk_score":               score.RiskScore,
		"response_time_score":      score.ResponseTimeDim,
		"frequency_score":          score.FrequencyDim,
		"win_rate_score":           score.WinRateDim,
		"pattern_score":            score.PatternDim,
		"account_age_score":        score.AccountAgeDim,
		"response_time_dim":        score.ResponseTimeDim,
		"frequency_dim":            score.FrequencyDim,
		"win_rate_dim":             score.WinRateDim,
		"pattern_dim":              score.PatternDim,
		"account_age_dim":          score.AccountAgeDim,
		"sanction_id":              sanctionID,
		"sanction_type":            sanctionType,
		"sanction_status":          sanctionStatus,
		"sanction_reason":          sanctionReason,
		"sanction_effective_until": sanctionUntil,
		"detection_time":           score.DetectionTime,
		"created_at":               firstNonZeroTime(score.DetectionTime, score.CreatedAt),
		"source":                   "risk_score",
	}
}

// GetDetectionList 查询检测列表
func gameHistoryDetectionDTO(history database.GameHistory, playerUID uint) gin.H {
	createdAt := firstNonZeroTime(history.FinishedAt, history.CreatedAt, history.StartedAt)
	id := fmt.Sprintf("history-%d-%d", history.ID, playerUID)
	return gin.H{
		"id":                  id,
		"history_id":          history.ID,
		"room_id":             history.RoomID,
		"player_id":           playerUID,
		"player_uid":          playerUID,
		"risk_score":          20.0,
		"sanction_type":       "observe",
		"sanction_status":     "detected",
		"sanction_reason":     "fast reaction replay marker",
		"detection_time":      createdAt,
		"created_at":          createdAt,
		"source":              "game_history",
		"cheat_detected":      history.CheatDetected,
		"response_time_score": nil,
		"frequency_score":     nil,
		"win_rate_score":      nil,
		"pattern_score":       nil,
		"account_age_score":   nil,
	}
}

func sortDetectionsByCreatedAt(items []gin.H) {
	sort.SliceStable(items, func(i, j int) bool {
		left, _ := items[i]["created_at"].(time.Time)
		right, _ := items[j]["created_at"].(time.Time)
		return left.After(right)
	})
}

func parseDetectionListPagination(c *gin.Context) (int, int, error) {
	page := 1
	limit := 50
	if raw := c.DefaultQuery("page", "1"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			return 0, 0, fmt.Errorf("invalid page")
		}
		page = parsed
	}
	if raw := c.DefaultQuery("limit", "50"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			return 0, 0, fmt.Errorf("invalid limit")
		}
		limit = parsed
	}
	if limit > 1000 {
		limit = 1000
	}
	return page, limit, nil
}

func (h *AnticheatHandler) GetDetectionList(c *gin.Context) {
	page, limit, err := parseDetectionListPagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queryLimit := page * limit
	riskScores, err := h.acSystem.Repository.GetRiskScoresByPlayer(0, queryLimit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ids := make([]uint, 0, len(riskScores))
	for _, score := range riskScores {
		ids = append(ids, score.ID)
	}
	sanctions, err := h.acSystem.Repository.GetSanctionsByRiskScoreIDs(ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	detections := make([]gin.H, 0, len(riskScores))
	seenRiskByRoomPlayer := make(map[string]struct{}, len(riskScores))
	statusFilter := strings.TrimSpace(c.Query("status"))
	for _, score := range riskScores {
		item := detectionDTO(score, sanctionForRiskScore(score, sanctions))
		if statusFilter == "" || statusFilter == "all" || item["sanction_type"] == statusFilter {
			detections = append(detections, item)
		}
		seenRiskByRoomPlayer[fmt.Sprintf("%s:%d", score.RoomID, score.PlayerUID)] = struct{}{}
	}

	histories, err := h.acSystem.Repository.GetCheatDetectedGameHistories(queryLimit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, history := range histories {
		for _, uid := range parseCheatUIDs(history.CheatUIDs) {
			if _, exists := seenRiskByRoomPlayer[fmt.Sprintf("%s:%d", history.RoomID, uid)]; exists {
				continue
			}
			item := gameHistoryDetectionDTO(history, uid)
			if statusFilter == "" || statusFilter == "all" || item["sanction_type"] == statusFilter {
				detections = append(detections, item)
			}
		}
	}

	sortDetectionsByCreatedAt(detections)
	total := len(detections)
	start := (page - 1) * limit
	end := start + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	paged := detections[start:end]

	c.JSON(http.StatusOK, gin.H{
		"count":      len(paged),
		"data":       paged,
		"detections": paged,
		"page":       page,
		"limit":      limit,
		"total":      total,
	})
}

// GetDetectionDetail 查询检测详情
func (h *AnticheatHandler) GetDetectionDetail(c *gin.Context) {
	if strings.HasPrefix(c.Param("id"), "history-") {
		parts := strings.Split(c.Param("id"), "-")
		if len(parts) != 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		historyID, historyErr := strconv.ParseUint(parts[1], 10, 32)
		playerUID, playerErr := strconv.ParseUint(parts[2], 10, 32)
		if historyErr != nil || playerErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		history, err := h.acSystem.Repository.GetGameHistoryByID(uint(historyID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Detection not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"risk_score": gameHistoryDetectionDTO(*history, uint(playerUID)),
			"sanctions":  []database.CheatSanction{},
		})
		return
	}

	riskScoreID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	riskScore, err := h.acSystem.Repository.GetRiskScoreByID(uint(riskScoreID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Detection not found"})
		return
	}

	// 获取相关的处罚记录
	sanctions, _ := h.acSystem.Decider.GetActiveSanctionsForPlayer(riskScore.PlayerUID)

	c.JSON(http.StatusOK, gin.H{
		"risk_score": riskScore,
		"sanctions":  sanctions,
	})
}

// ReviewDetection 人工审核检测结果
func (h *AnticheatHandler) ReviewDetection(c *gin.Context) {
	riskScoreID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req struct {
		Decision string `json:"decision"` // "confirm", "overturn"
		Remark   string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = h.acSystem.Repository.GetRiskScoreByID(uint(riskScoreID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Detection not found"})
		return
	}

	// TODO: 实现人工审核逻辑

	c.JSON(http.StatusOK, gin.H{
		"message": "Review completed",
		"status":  req.Decision,
	})
}

// GetAppealsList 查询申诉列表
func (h *AnticheatHandler) GetAppealsList(c *gin.Context) {
	limit := 50
	if l := c.DefaultQuery("limit", "50"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 1000 {
			limit = val
		}
	}

	appeals, err := h.acSystem.AppealManager.GetPendingAppeals(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]gin.H, 0, len(appeals))
	for _, appeal := range appeals {
		items = append(items, appealDTO(appeal))
	}

	c.JSON(http.StatusOK, gin.H{
		"appeals": items,
		"count":   len(items),
		"data":    items,
		"total":   len(items),
	})
}

func appealDTO(appeal database.CheatAppeal) gin.H {
	return gin.H{
		"id":                  appeal.ID,
		"room_id":             appeal.RoomID,
		"player_id":           appeal.PlayerUID,
		"player_uid":          appeal.PlayerUID,
		"risk_score_id":       appeal.RiskScoreID,
		"sanction_id":         appeal.SanctionID,
		"reason":              appeal.Reason,
		"evidence":            appeal.Evidence,
		"status":              appeal.Status,
		"reviewer_uid":        appeal.ReviewerUID,
		"reviewed_at":         appeal.ReviewedAt,
		"review_remark":       appeal.ReviewRemark,
		"compensation_amount": appeal.CompensationAmount,
		"compensation_status": appeal.CompensationStatus,
		"compensation_note":   appeal.CompensationNote,
		"submitted_at":        appeal.SubmittedAt,
		"created_at":          appeal.CreatedAt,
	}
}

// ApproveAppeal 批准申诉 - 支持配置补偿金额和文案
func (h *AnticheatHandler) ApproveAppeal(c *gin.Context) {
	appealID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req struct {
		Note                string `json:"note"`                 // 审核备注
		CompensationAmount  *int   `json:"compensation_amount"`  // 补偿金额（可选，使用默认值）
		CompensationMessage string `json:"compensation_message"` // 补偿文案
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取申诉详情以确认记录存在
	if _, err := h.acSystem.Repository.GetAppealByID(uint(appealID)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appeal not found"})
		return
	}

	// 获取当前用户ID（假设从上下文中获取）
	reviewerUID := uint(1)

	// 获取配置用于验证和默认值
	config := h.acSystem.Config.GetConfig()

	// 验证补偿金额范围
	compensationAmount := config.UnbanConfig.CompensationAmount
	if req.CompensationAmount != nil {
		if *req.CompensationAmount < config.UnbanConfig.MinAmount || *req.CompensationAmount > config.UnbanConfig.MaxAmount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Compensation amount out of range"})
			return
		}
		compensationAmount = *req.CompensationAmount
	}

	// 验证补偿文案长度
	compensationMessage := req.CompensationMessage
	if compensationMessage == "" {
		compensationMessage = config.UnbanConfig.DefaultMessage
	}
	if len(compensationMessage) > config.UnbanConfig.MessageMaxLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Compensation message too long"})
		return
	}

	outcome, err := h.acSystem.AppealManager.ApproveAppealWithCompensation(uint(appealID), reviewerUID, req.Note, compensationAmount, compensationMessage, h.acSystem.Decider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":             "Appeal approved successfully",
		"compensation_status": outcome.CompensationStatus,
		"compensation_note":   outcome.CompensationNote,
		"idempotent":          outcome.Idempotent,
		"compensation": gin.H{
			"amount":  compensationAmount,
			"message": compensationMessage,
		},
	})
}

// RejectAppeal 拒绝申诉
func (h *AnticheatHandler) RejectAppeal(c *gin.Context) {
	appealID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req struct {
		Remark string `json:"remark"`
		Note   string `json:"note"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Remark == "" {
		req.Remark = req.Note
	}

	appeal, err := h.acSystem.Repository.GetAppealByID(uint(appealID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appeal not found"})
		return
	}

	// TODO: 获取当前用户ID
	reviewerUID := uint(1)

	if err := h.acSystem.AppealManager.RejectAppeal(uint(appealID), reviewerUID, req.Remark); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	auditLog := &database.CheatAuditLog{
		EventType:   "review",
		RoomID:      appeal.RoomID,
		PlayerUID:   appeal.PlayerUID,
		OperatorUID: &reviewerUID,
		AppealID:    &appeal.ID,
		OldStatus:   appeal.Status,
		NewStatus:   "rejected",
		Remark:      req.Remark,
	}
	if err := h.acSystem.Repository.SaveAuditLog(auditLog); err != nil {
		log.Printf("Failed to write appeal rejection audit log: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Appeal rejected"})
}

// GetConfig 获取当前配置
func (h *AnticheatHandler) GetConfig(c *gin.Context) {
	config := h.acSystem.Config.GetConfig()
	c.JSON(http.StatusOK, buildAdminConfigResponse(config))
}

// UpdateConfig 更新配置
func (h *AnticheatHandler) UpdateConfig(c *gin.Context) {
	current := h.acSystem.Config.GetConfig()
	var req struct {
		Dimensions         map[string]anticheat.DimensionConfig `json:"dimensions"`
		SanctionThresholds anticheat.SanctionThresholds         `json:"sanction_thresholds"`
		EnabledStrategies  []string                             `json:"enabled_strategies"`
		Unban              *anticheat.UnbanConfig               `json:"unban"`
		UnbanConfig        *anticheat.UnbanConfig               `json:"unban_config"`
		Sanctions          map[string]float64                   `json:"sanctions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	next := &anticheat.RiskScoringConfig{
		Dimensions:         current.Dimensions,
		SanctionThresholds: current.SanctionThresholds,
		EnabledStrategies:  current.EnabledStrategies,
		UnbanConfig:        current.UnbanConfig,
	}
	if req.Dimensions != nil {
		next.Dimensions = req.Dimensions
	}
	if req.SanctionThresholds != (anticheat.SanctionThresholds{}) {
		next.SanctionThresholds = req.SanctionThresholds
	}
	if len(req.Sanctions) > 0 {
		if value, ok := req.Sanctions["observe"]; ok {
			next.SanctionThresholds.ObserveMin = value
		}
		if value, ok := req.Sanctions["warning"]; ok {
			next.SanctionThresholds.WarningMin = value
		}
		if value, ok := req.Sanctions["mute"]; ok {
			next.SanctionThresholds.MuteMin = value
		}
		if value, ok := req.Sanctions["ban"]; ok {
			next.SanctionThresholds.BanMin = value
		}
	}
	if req.EnabledStrategies != nil {
		next.EnabledStrategies = req.EnabledStrategies
	}
	if req.Unban != nil {
		next.UnbanConfig = *req.Unban
	} else if req.UnbanConfig != nil {
		next.UnbanConfig = *req.UnbanConfig
	}

	if err := h.acSystem.Config.ReplaceConfig(next); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.acSystem.Repository != nil {
		details, _ := json.Marshal(gin.H{
			"changed_keys": []string{"dimensions", "sanction_thresholds", "enabled_strategies", "unban"},
			"old":          buildAdminConfigResponse(current),
			"new":          buildAdminConfigResponse(next),
			"source":       "admin_anticheat_panel",
		})
		audit := &database.CheatAuditLog{
			EventType:   "config_change",
			OperatorUID: getOperatorUID(c),
			Details:     details,
			Remark:      "Anticheat configuration updated",
		}
		if err := h.acSystem.Repository.SaveAuditLog(audit); err != nil {
			log.Printf("Failed to write anticheat config audit log: %v", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration updated",
		"config":  buildAdminConfigResponse(h.acSystem.Config.GetConfig()),
	})
}

// GetAuditLog 查询审计日志
func (h *AnticheatHandler) GetAuditLog(c *gin.Context) {
	filter, page, err := parseAuditFilter(c, true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logs, total, err := h.acSystem.Repository.QueryAuditLogs(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]gin.H, 0, len(logs))
	for _, log := range logs {
		items = append(items, auditLogDTO(log))
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(items),
		"data":  items,
		"logs":  items,
		"page":  page,
		"limit": filter.Limit,
		"total": total,
	})
}

// ExportAuditLog exports the filtered audit log as CSV.
func (h *AnticheatHandler) ExportAuditLog(c *gin.Context) {
	filter, _, err := parseAuditFilter(c, false)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logs, err := h.acSystem.Repository.ExportAuditLogs(filter, 10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := fmt.Sprintf("anticheat_audit_log_%s.csv", time.Now().Format("2006-01-02"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Header("Content-Type", "text/csv; charset=utf-8")

	writer := csv.NewWriter(c.Writer)
	_ = writer.Write([]string{
		"player_id",
		"action",
		"reason",
		"created_at",
		"approval_note",
		"compensation_amount",
		"compensation_status",
		"compensation_message",
		"compensation_date",
	})
	for _, log := range logs {
		approvalNote := ""
		if log.ApprovalNote != nil {
			approvalNote = *log.ApprovalNote
		}
		compensationAmount := ""
		if log.CompensationAmount != nil {
			compensationAmount = strconv.Itoa(*log.CompensationAmount)
		}
		compensationStatus := ""
		if log.CompensationStatus != nil {
			compensationStatus = *log.CompensationStatus
		}
		compensationMessage := ""
		if log.CompensationMessage != nil {
			compensationMessage = *log.CompensationMessage
		}
		compensationDate := ""
		if log.CompensationDate != nil {
			compensationDate = log.CompensationDate.Format(time.RFC3339)
		}

		_ = writer.Write([]string{
			strconv.FormatUint(uint64(log.PlayerUID), 10),
			auditAction(log),
			auditDetails(log),
			log.CreatedAt.Format(time.RFC3339),
			approvalNote,
			compensationAmount,
			compensationStatus,
			compensationMessage,
			compensationDate,
		})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		log.Printf("⚠️  Failed to write anticheat audit CSV: %v", err)
	}
}

// SubmitAppeal 提交申诉（玩家端）
func (h *AnticheatHandler) SubmitAppeal(c *gin.Context) {
	if h.acSystem == nil || h.acSystem.AppealManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat appeal manager not initialized"})
		return
	}

	playerUID, ok := getAuthenticatedUID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "player authentication required"})
		return
	}

	var req struct {
		RiskScoreID uint   `json:"risk_score_id"`
		SanctionID  *uint  `json:"sanction_id"`
		Reason      string `json:"reason"`
		Evidence    string `json:"evidence"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reason is required"})
		return
	}

	appeal, err := h.acSystem.AppealManager.SubmitAppeal(c.Param("roomId"), playerUID, req.RiskScoreID, req.SanctionID, req.Reason, req.Evidence)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Appeal submitted",
		"appeal":  appeal,
		"data":    appeal,
	})
}

// GetPlayerAppeals 获取玩家的申诉历史
func (h *AnticheatHandler) GetPlayerAppeals(c *gin.Context) {
	if h.acSystem == nil || h.acSystem.AppealManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat appeal manager not initialized"})
		return
	}

	playerUID, ok := getAuthenticatedUID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "player authentication required"})
		return
	}

	appeals, err := h.acSystem.AppealManager.GetPlayerAppeals(playerUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.reconcileApprovedAppealEffects(playerUID, appeals); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"appeals":        appeals,
		"count":          len(appeals),
		"current_status": latestAppealStatus(appeals),
		"data":           appeals,
		"total":          len(appeals),
	})
}

func (h *AnticheatHandler) ClaimAppealCompensation(c *gin.Context) {
	if h.acSystem == nil || h.acSystem.AppealManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat appeal manager not initialized"})
		return
	}

	playerUID, ok := getAuthenticatedUID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "player authentication required"})
		return
	}

	appealID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	outcome, err := h.acSystem.AppealManager.ClaimCompensation(uint(appealID), playerUID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":             "Compensation claimed",
		"compensation_status": outcome.CompensationStatus,
		"compensation_note":   outcome.CompensationNote,
		"idempotent":          outcome.Idempotent,
		"appeal":              outcome.Appeal,
		"compensation": gin.H{
			"amount": outcome.CompensationAmount,
		},
	})
}

func (h *AnticheatHandler) reconcileApprovedAppealEffects(playerUID uint, appeals []database.CheatAppeal) error {
	if h.acSystem == nil || h.acSystem.Repository == nil {
		return nil
	}
	hasApproved := false
	for _, appeal := range appeals {
		if appeal.PlayerUID != playerUID || appeal.Status != "approved" {
			continue
		}
		hasApproved = true
		if appeal.SanctionID > 0 {
			if err := h.acSystem.Repository.UpdateSanctionStatus(appeal.SanctionID, "revoked"); err != nil {
				return err
			}
		}
	}
	if !hasApproved {
		return nil
	}
	if err := h.acSystem.Repository.RevokeActiveBanSanctionsByPlayer(playerUID); err != nil {
		return err
	}
	return h.acSystem.Repository.ClearPlayerAccountBan(playerUID)
}

func latestAppealStatus(appeals []database.CheatAppeal) string {
	if len(appeals) == 0 {
		return "none"
	}
	return appeals[0].Status
}

func getOperatorUID(c *gin.Context) *uint {
	uid, ok := getAuthenticatedUID(c)
	if !ok {
		return nil
	}
	return &uid
}

func (h *AnticheatHandler) BanFromAnticheatPanel(c *gin.Context) {
	if h.acSystem == nil || h.acSystem.Repository == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat system not initialized"})
		return
	}

	var req struct {
		PlayerUID   uint   `json:"player_uid"`
		BannedUntil string `json:"banned_until"`
		Reason      string `json:"reason"`
		RoomID      string `json:"room_id"`
		RiskScoreID *uint  `json:"risk_score_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.PlayerUID == 0 || strings.TrimSpace(req.BannedUntil) == "" || strings.TrimSpace(req.Reason) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "player_uid, banned_until and reason are required"})
		return
	}
	bannedUntil, err := time.Parse(time.RFC3339, req.BannedUntil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid banned_until"})
		return
	}
	if !bannedUntil.After(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "banned_until must be in the future"})
		return
	}

	if err := repository.NewUserRepository().UpdateBanStatusWithReason(req.PlayerUID, &bannedUntil, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	sanction := &database.CheatSanction{
		RoomID:         req.RoomID,
		PlayerUID:      req.PlayerUID,
		SanctionType:   "ban",
		RiskScore:      0,
		Reason:         req.Reason,
		EffectiveUntil: &bannedUntil,
		Status:         "active",
	}
	if req.RiskScoreID != nil {
		sanction.RiskScoreID = *req.RiskScoreID
	}
	if err := h.acSystem.Repository.SaveSanction(sanction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	details, _ := json.Marshal(gin.H{"expires_at": bannedUntil, "source": "admin_anticheat_panel"})
	audit := &database.CheatAuditLog{
		EventType:    "ban",
		RoomID:       req.RoomID,
		PlayerUID:    req.PlayerUID,
		OperatorUID:  getOperatorUID(c),
		RiskScoreID:  req.RiskScoreID,
		SanctionType: "ban",
		NewStatus:    "active",
		Details:      details,
		Remark:       req.Reason,
	}
	if err := h.acSystem.Repository.SaveAuditLog(audit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "player banned", "audit": auditLogDTO(*audit)})
}

func (h *AnticheatHandler) UnbanFromAnticheatPanel(c *gin.Context) {
	if h.acSystem == nil || h.acSystem.Repository == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat system not initialized"})
		return
	}

	var req struct {
		PlayerUID uint   `json:"player_uid"`
		Reason    string `json:"reason"`
		RoomID    string `json:"room_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.PlayerUID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "player_uid is required"})
		return
	}
	if strings.TrimSpace(req.Reason) == "" {
		req.Reason = "Manual unban from anticheat panel"
	}

	if err := repository.NewUserRepository().UpdateBanStatusWithReason(req.PlayerUID, nil, ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.acSystem.Repository.RevokeActiveBanSanctionsByPlayer(req.PlayerUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	details, _ := json.Marshal(gin.H{"source": "admin_anticheat_panel"})
	audit := &database.CheatAuditLog{
		EventType:   "unban",
		RoomID:      req.RoomID,
		PlayerUID:   req.PlayerUID,
		OperatorUID: getOperatorUID(c),
		OldStatus:   "active",
		NewStatus:   "revoked",
		Details:     details,
		Remark:      req.Reason,
	}
	if err := h.acSystem.Repository.SaveAuditLog(audit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "player unbanned", "audit": auditLogDTO(*audit)})
}

// GetPlayerSanctions 获取玩家当前处罚
func (h *AnticheatHandler) GetPlayerSanctions(c *gin.Context) {
	if h.acSystem == nil || h.acSystem.Decider == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat sanction decider not initialized"})
		return
	}

	playerUID, ok := getAuthenticatedUID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "player authentication required"})
		return
	}

	sanctions, err := h.acSystem.Decider.GetActiveSanctionsForPlayer(playerUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":     len(sanctions),
		"data":      sanctions,
		"sanctions": sanctions,
		"total":     len(sanctions),
	})
}

// GetAnticheatStats 获取管理员面板反作弊统计信息
func (h *AnticheatHandler) GetAnticheatStats(c *gin.Context) {
	// 获取今日封禁数量
	bansTodayEnd := time.Now()
	bansTodayStart := bansTodayEnd.Add(-24 * time.Hour)

	bansToday, err := h.acSystem.Repository.CountBansInRange(bansTodayStart, bansTodayEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取启用的检测规则
	enabledStrategies := h.acSystem.Config.GetEnabledStrategies()

	// 获取近期的处罚统计
	stats := map[string]interface{}{
		"bans_today":          bansToday,
		"system_uptime_days":  h.acSystem.UptimeDays(bansTodayEnd),
		"enabled_strategies":  enabledStrategies,
		"sanction_thresholds": h.acSystem.Config.GetSanctionThresholds(),
		"dimensions":          h.acSystem.Config.GetDimensions(),
		"query_range": gin.H{
			"start": bansTodayStart,
			"end":   bansTodayEnd,
		},
	}

	c.JSON(http.StatusOK, stats)
}

// GetPlayerAnticheatStats 获取玩家反作弊统计信息（公开给玩家的数据）
func (h *AnticheatHandler) GetPlayerAnticheatStats(c *gin.Context) {
	now := time.Now()
	uid, hasUID := getAuthenticatedUID(c)
	if hasUID {
		allowed, err := cache.AllowPlayerAnticheatStatsRequest(c.Request.Context(), uid, time.Second)
		if err != nil {
			log.Printf("⚠️  Player anticheat stats Redis rate limit unavailable: %v", err)
			allowed = allowPlayerStatsRequestFallback(uid, now)
		}
		if !allowed {
			c.Header("Retry-After", "1")
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
	}

	if cachedStats, err := cache.GetPlayerAnticheatStatsCache(c.Request.Context()); err == nil && cachedStats != "" {
		c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(cachedStats))
		return
	} else if err != nil {
		log.Printf("⚠️  Player anticheat stats cache unavailable: %v", err)
	}

	stats, err := h.buildPlayerAnticheatStats(now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if data, err := json.Marshal(stats); err == nil {
		if err := cache.SetPlayerAnticheatStatsCache(c.Request.Context(), string(data), 5*time.Minute); err != nil {
			log.Printf("⚠️  Failed to cache player anticheat stats: %v", err)
		}
	}

	c.JSON(http.StatusOK, stats)
}
