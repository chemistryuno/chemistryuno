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

const defaultDetectionBanDurationMinutes = 10080

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

func ChangeDetectionPunishment(c *gin.Context) {
	if globalAnticheatHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat handler not initialized"})
		return
	}
	globalAnticheatHandler.ChangeDetectionPunishment(c)
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

func GetAppealEntryStatus(c *gin.Context) {
	if globalAnticheatHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "anticheat handler not initialized"})
		return
	}
	globalAnticheatHandler.GetAppealEntryStatus(c)
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
		"risk_score":           log.RiskScore,
		"replay_id":            log.ReplayID,
		"game_history_id":      log.GameHistoryID,
		"has_replay":           replayExists(replayHistoryID(log.GameHistoryID, log.ReplayID)),
		"operation_index":      log.OperationIndex,
		"operation_timestamp":  log.OperationTimestamp,
		"primary_evidence":     replayEvidenceOrFallback(log.RoomID, log.ReplayID, log.GameHistoryID, log.OperationIndex, log.PlayerUID, log.PrimaryEvidence),
		"related_evidence":     jsonRawWithReplayAvailability(log.RelatedEvidence),
		"suggested_action":     log.SuggestedAction,
		"indicator_details":    jsonRawWithReplayAvailability(log.IndicatorDetails),
		"report_contribution":  jsonRawWithReplayAvailability(log.ReportContribution),
		"old_decision":         log.OldDecision,
		"new_decision":         log.NewDecision,
	}
}

func jsonRawOrNil(raw database.JSON) interface{} {
	if len(raw) == 0 {
		return nil
	}
	var decoded interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return string(raw)
	}
	return decoded
}

func jsonRawWithReplayAvailability(raw database.JSON) interface{} {
	return annotateReplayAvailability(jsonRawOrNil(raw))
}

func annotateReplayAvailability(value interface{}) interface{} {
	switch typed := value.(type) {
	case []interface{}:
		for i := range typed {
			typed[i] = annotateReplayAvailability(typed[i])
		}
	case map[string]interface{}:
		for key, child := range typed {
			typed[key] = annotateReplayAvailability(child)
		}
		if isReplayEvidenceMap(typed) {
			hasReplay := replayExists(replayHistoryIDFromMap(typed))
			typed["has_replay"] = hasReplay
			typed["replay_available"] = hasReplay
			if !hasReplay {
				typed["navigation_url"] = ""
			}
		}
	}
	return value
}

func isReplayEvidenceMap(value map[string]interface{}) bool {
	if _, ok := value["game_history_id"]; ok {
		return true
	}
	if _, ok := value["replay_id"]; ok {
		return true
	}
	if _, ok := value["navigation_url"]; ok {
		return true
	}
	if _, ok := value["evidence_precision"]; ok {
		return true
	}
	if _, ok := value["event_index"]; ok {
		_, hasRoom := value["room_id"]
		return hasRoom
	}
	if _, ok := value["event_id"]; ok {
		_, hasRoom := value["room_id"]
		return hasRoom
	}
	return false
}

func replayHistoryIDFromMap(value map[string]interface{}) uint {
	if id := uintFromAny(value["game_history_id"]); id > 0 {
		return id
	}
	return replayHistoryIDFromReplayID(stringFromAny(value["replay_id"]))
}

func replayHistoryID(gameHistoryID uint, replayID string) uint {
	if gameHistoryID > 0 {
		return gameHistoryID
	}
	return replayHistoryIDFromReplayID(replayID)
}

func replayHistoryIDFromReplayID(replayID string) uint {
	parsed, err := strconv.ParseUint(strings.TrimSpace(replayID), 10, 64)
	if err != nil || parsed == 0 {
		return 0
	}
	return uint(parsed)
}

func replayExists(gameHistoryID uint) bool {
	if gameHistoryID == 0 || database.DB == nil {
		return false
	}
	if !database.DB.Migrator().HasTable(&database.GameHistory{}) {
		return false
	}
	var history database.GameHistory
	if err := database.DB.Select("id", "replay_log").First(&history, gameHistoryID).Error; err != nil {
		return false
	}
	return strings.TrimSpace(history.ReplayLog) != ""
}

func uintFromAny(value interface{}) uint {
	switch typed := value.(type) {
	case uint:
		return typed
	case uint64:
		return uint(typed)
	case int:
		if typed > 0 {
			return uint(typed)
		}
	case int64:
		if typed > 0 {
			return uint(typed)
		}
	case float64:
		if typed > 0 {
			return uint(typed)
		}
	case json.Number:
		parsed, err := typed.Int64()
		if err == nil && parsed > 0 {
			return uint(parsed)
		}
	case string:
		return replayHistoryIDFromReplayID(typed)
	}
	return 0
}

func stringFromAny(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	case nil:
		return ""
	default:
		return fmt.Sprint(typed)
	}
}

func replayEvidenceOrFallback(roomID string, replayID string, gameHistoryID uint, operationIndex int, playerUID uint, primary database.JSON) gin.H {
	if anchor, ok := anticheat.UnmarshalReplayEvidenceAnchor(primary); ok {
		hasReplay := replayExists(replayHistoryID(anchor.GameHistoryID, anchor.ReplayID))
		navigationURL := anchor.NavigationURL
		if !hasReplay {
			navigationURL = ""
		}
		return gin.H{
			"room_id":             anchor.RoomID,
			"game_history_id":     anchor.GameHistoryID,
			"replay_id":           anchor.ReplayID,
			"has_replay":          hasReplay,
			"replay_available":    hasReplay,
			"event_index":         anchor.EventIndex,
			"event_id":            anchor.EventID,
			"event_type":          anchor.EventType,
			"player_uid":          anchor.PlayerUID,
			"event_timestamp_ms":  anchor.EventTimestampMs,
			"turn_number":         anchor.TurnNumber,
			"action_summary":      anchor.ActionSummary,
			"evidence_precision":  anchor.EvidencePrecision,
			"compatibility_level": anchor.CompatibilityLevel,
			"navigation_url":      navigationURL,
		}
	}
	anchor := anticheat.NormalizeReplayEvidenceAnchor(database.ReplayEvidenceAnchor{
		RoomID:             roomID,
		GameHistoryID:      gameHistoryID,
		ReplayID:           replayID,
		EventIndex:         operationIndex,
		PlayerUID:          playerUID,
		ActionSummary:      "compatibility anticheat evidence",
		EvidencePrecision:  "room",
		CompatibilityLevel: "compatibility_index",
	})
	hasReplay := replayExists(replayHistoryID(anchor.GameHistoryID, anchor.ReplayID))
	navigationURL := anchor.NavigationURL
	if !hasReplay {
		navigationURL = ""
	}
	return gin.H{
		"room_id":             anchor.RoomID,
		"game_history_id":     anchor.GameHistoryID,
		"replay_id":           anchor.ReplayID,
		"has_replay":          hasReplay,
		"replay_available":    hasReplay,
		"event_index":         anchor.EventIndex,
		"event_id":            anchor.EventID,
		"event_type":          anchor.EventType,
		"player_uid":          anchor.PlayerUID,
		"event_timestamp_ms":  anchor.EventTimestampMs,
		"turn_number":         anchor.TurnNumber,
		"action_summary":      anchor.ActionSummary,
		"evidence_precision":  anchor.EvidencePrecision,
		"compatibility_level": anchor.CompatibilityLevel,
		"navigation_url":      navigationURL,
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
	hasReplay := replayExists(replayHistoryID(score.GameHistoryID, score.ReplayID))
	return gin.H{
		"id":                       score.ID,
		"room_id":                  score.RoomID,
		"player_id":                score.PlayerUID,
		"player_uid":               score.PlayerUID,
		"risk_score":               score.RiskScore,
		"replay_id":                score.ReplayID,
		"game_history_id":          score.GameHistoryID,
		"has_replay":               hasReplay,
		"operation_index":          score.OperationIndex,
		"operation_timestamp":      score.OperationTimestamp,
		"primary_evidence":         replayEvidenceOrFallback(score.RoomID, score.ReplayID, score.GameHistoryID, score.OperationIndex, score.PlayerUID, score.PrimaryEvidence),
		"related_evidence":         jsonRawWithReplayAvailability(score.RelatedEvidence),
		"replay_navigation":        replayEvidenceOrFallback(score.RoomID, score.ReplayID, score.GameHistoryID, score.OperationIndex, score.PlayerUID, score.PrimaryEvidence),
		"indicator_details":        jsonRawWithReplayAvailability(score.IndicatorDetails),
		"report_contribution":      jsonRawWithReplayAvailability(score.ReportContribution),
		"suggested_action":         firstNonEmpty(score.SuggestedAction, sanctionType),
		"suggestion_reason":        score.SuggestionReason,
		"review_status":            firstNonEmpty(score.ReviewStatus, "pending"),
		"punishment_decision":      punishmentDecisionForScore(score, sanctionType),
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
	evidence := replayEvidenceOrFallback(history.RoomID, fmt.Sprintf("%d", history.ID), history.ID, 0, playerUID, nil)
	hasReplay := strings.TrimSpace(history.ReplayLog) != ""
	evidence["has_replay"] = hasReplay
	evidence["replay_available"] = hasReplay
	if !hasReplay {
		evidence["navigation_url"] = ""
	}
	return gin.H{
		"id":                  id,
		"history_id":          history.ID,
		"room_id":             history.RoomID,
		"player_id":           playerUID,
		"player_uid":          playerUID,
		"risk_score":          20.0,
		"replay_id":           fmt.Sprintf("%d", history.ID),
		"game_history_id":     history.ID,
		"has_replay":          hasReplay,
		"operation_index":     0,
		"operation_timestamp": nil,
		"primary_evidence":    evidence,
		"related_evidence":    []gin.H{evidence},
		"replay_navigation":   evidence,
		"indicator_details": []gin.H{
			{
				"name":             "legacy_replay_marker",
				"raw_value":        1,
				"normalized_score": 20,
				"weight":           1,
				"contribution":     20,
				"explanation":      "legacy replay marker retained for evidence lookup only",
				"evidence_anchors": []gin.H{evidence},
			},
		},
		"report_contribution": nil,
		"suggested_action":    "observe",
		"suggestion_reason":   "legacy cheat marker is not authoritative for new punishment",
		"review_status":       "pending",
		"punishment_decision": "observe",
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
		"risk_score": detectionDTO(*riskScore, sanctionForRiskScore(*riskScore, map[uint][]database.CheatSanction{riskScore.ID: sanctions})),
		"sanctions":  sanctions,
	})
}

// ReviewDetection 人工审核检测结果
func (h *AnticheatHandler) ReviewDetection(c *gin.Context) {
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
		var req struct {
			Decision string `json:"decision"`
			Remark   string `json:"remark"`
			Note     string `json:"note"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		remark := firstNonEmpty(req.Remark, req.Note, "Legacy replay marker reviewed")
		details, _ := json.Marshal(gin.H{
			"source":     "legacy_game_history",
			"history_id": history.ID,
			"decision":   firstNonEmpty(req.Decision, "confirm"),
		})
		audit := &database.CheatAuditLog{
			EventType:       "review",
			RoomID:          history.RoomID,
			PlayerUID:       uint(playerUID),
			OperatorUID:     getOperatorUID(c),
			ReplayID:        strconv.FormatUint(historyID, 10),
			SuggestedAction: "observe",
			OldStatus:       "detected",
			NewStatus:       "processed",
			NewDecision:     "observe",
			Details:         details,
			Remark:          remark,
		}
		if err := h.acSystem.Repository.SaveAuditLog(audit); err != nil {
			log.Printf("Failed to write legacy detection review audit log: %v", err)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Review completed",
			"status":  "processed",
			"audit":   auditLogDTO(*audit),
		})
		return
	}

	riskScoreID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req struct {
		Decision string `json:"decision"` // "confirm", "overturn"
		Remark   string `json:"remark"`
		Note     string `json:"note"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	score, err := h.acSystem.Repository.GetRiskScoreByID(uint(riskScoreID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Detection not found"})
		return
	}

	decision := strings.TrimSpace(req.Decision)
	if decision == "" {
		decision = "confirm"
	}
	if decision == "override" || decision == "overturn" || decision == "none" || decision == "cancel" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "processed punishment cannot be cancelled here"})
		return
	}
	punishmentDecision := punishmentDecisionForScore(*score, "observe")
	if err := h.acSystem.Repository.UpdateRiskScoreReview(uint(riskScoreID), "processed", punishmentDecision); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	remark := firstNonEmpty(req.Remark, req.Note, "Detection reviewed")
	details, _ := json.Marshal(gin.H{
		"source":              "admin_anticheat_panel",
		"decision":            decision,
		"punishment_decision": punishmentDecision,
	})
	audit := &database.CheatAuditLog{
		EventType:          "review",
		RoomID:             score.RoomID,
		PlayerUID:          score.PlayerUID,
		OperatorUID:        getOperatorUID(c),
		RiskScoreID:        &score.ID,
		RiskScore:          &score.RiskScore,
		ReplayID:           score.ReplayID,
		GameHistoryID:      score.GameHistoryID,
		OperationIndex:     score.OperationIndex,
		OperationTimestamp: score.OperationTimestamp,
		PrimaryEvidence:    score.PrimaryEvidence,
		RelatedEvidence:    score.RelatedEvidence,
		SuggestedAction:    score.SuggestedAction,
		IndicatorDetails:   score.IndicatorDetails,
		ReportContribution: score.ReportContribution,
		OldStatus:          score.ReviewStatus,
		NewStatus:          "processed",
		NewDecision:        punishmentDecision,
		Details:            details,
		Remark:             remark,
	}
	if err := h.acSystem.Repository.SaveAuditLog(audit); err != nil {
		log.Printf("Failed to write detection review audit log: %v", err)
	}
	if isBanDecision(punishmentDecision) {
		reason := firstNonEmpty(req.Remark, req.Note, score.SuggestionReason, "Anticheat review confirmed ban")
		if _, err := h.enforceDetectionBan(c, *score, reason, nil, nil, "admin_anticheat_review", 0); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Review completed",
		"status":  "processed",
	})
}

func isCancellationDecision(decision string) bool {
	switch strings.ToLower(strings.TrimSpace(decision)) {
	case "", "none", "cancel", "cancelled", "canceled", "revoke", "revoked", "unban", "no_punishment", "no-punishment":
		return true
	default:
		return false
	}
}

func isBanDecision(decision string) bool {
	return strings.EqualFold(strings.TrimSpace(decision), "ban")
}

func isActiveSanctionStatus(status string) bool {
	normalized := strings.ToLower(strings.TrimSpace(status))
	return normalized == "" || normalized == "active"
}

func punishmentDecisionForScore(score database.CheatRiskScore, fallback string) string {
	for _, value := range []string{score.PunishmentDecision, score.SuggestedAction, fallback, "observe"} {
		decision := strings.TrimSpace(value)
		if decision != "" && !isCancellationDecision(decision) {
			return decision
		}
	}
	return "observe"
}

func (h *AnticheatHandler) enforceDetectionBan(c *gin.Context, score database.CheatRiskScore, reason string, requestedUntil *time.Time, requestedDuration *int, source string, requestedSanctionID uint) (*database.CheatSanction, error) {
	if h.acSystem == nil || h.acSystem.Repository == nil {
		return nil, fmt.Errorf("anticheat system not initialized")
	}

	now := time.Now()
	if requestedUntil != nil && !requestedUntil.After(now) {
		return nil, fmt.Errorf("effective_until must be in the future")
	}
	if requestedDuration != nil && *requestedDuration <= 0 {
		return nil, fmt.Errorf("duration must be positive")
	}

	reason = firstNonEmpty(reason, score.SuggestionReason, "Anticheat review confirmed ban")
	var sanction *database.CheatSanction
	if requestedSanctionID > 0 {
		existing, err := h.acSystem.Repository.GetSanctionByID(requestedSanctionID)
		if err == nil && existing != nil {
			if existing.PlayerUID != score.PlayerUID {
				return nil, fmt.Errorf("sanction does not belong to detection player")
			}
			if existing.RiskScoreID != 0 && existing.RiskScoreID != score.ID {
				return nil, fmt.Errorf("sanction does not belong to detection")
			}
			sanction = existing
		}
	}

	if sanction == nil {
		sanctions, err := h.acSystem.Repository.GetSanctionsByRiskScoreIDs([]uint{score.ID})
		if err != nil {
			return nil, err
		}
		for i := range sanctions[score.ID] {
			item := sanctions[score.ID][i]
			if item.PlayerUID == score.PlayerUID && item.SanctionType == "ban" && isActiveSanctionStatus(item.Status) {
				sanction = &item
				break
			}
		}
	}

	duration := requestedDuration
	effectiveUntil := requestedUntil
	if effectiveUntil == nil && duration != nil {
		computed := now.Add(time.Duration(*duration) * time.Minute)
		effectiveUntil = &computed
	}
	if effectiveUntil == nil && sanction != nil && sanction.EffectiveUntil != nil && sanction.EffectiveUntil.After(now) {
		effectiveUntil = sanction.EffectiveUntil
		if duration == nil {
			duration = sanction.Duration
		}
	}
	if effectiveUntil == nil {
		defaultDuration := defaultDetectionBanDurationMinutes
		duration = &defaultDuration
		computed := now.Add(time.Duration(defaultDuration) * time.Minute)
		effectiveUntil = &computed
	}
	if duration == nil {
		minutes := int(time.Until(*effectiveUntil).Minutes())
		if minutes < 1 {
			minutes = 1
		}
		duration = &minutes
	}

	if err := repository.NewUserRepository().UpdateBanStatusWithReason(score.PlayerUID, effectiveUntil, reason); err != nil {
		return nil, err
	}
	sendBanNotification(score.PlayerUID, effectiveUntil, reason)

	sanctionCreated := false
	if sanction != nil {
		if err := h.acSystem.Repository.UpdateSanctionDecision(sanction.ID, "ban", reason, duration, effectiveUntil); err != nil {
			return nil, err
		}
		sanction.SanctionType = "ban"
		sanction.Reason = reason
		sanction.Duration = duration
		sanction.EffectiveUntil = effectiveUntil
		sanction.Status = "active"
	} else {
		sanction = &database.CheatSanction{
			RoomID:          score.RoomID,
			PlayerUID:       score.PlayerUID,
			RiskScoreID:     score.ID,
			ReplayID:        score.ReplayID,
			GameHistoryID:   score.GameHistoryID,
			PrimaryEvidence: score.PrimaryEvidence,
			SanctionType:    "ban",
			RiskScore:       score.RiskScore,
			Reason:          reason,
			Duration:        duration,
			EffectiveUntil:  effectiveUntil,
			Status:          "active",
		}
		if err := h.acSystem.Repository.SaveSanction(sanction); err != nil {
			return nil, err
		}
		sanctionCreated = true
	}

	details, _ := json.Marshal(gin.H{
		"expires_at":       effectiveUntil,
		"source":           source,
		"sanction_created": sanctionCreated,
	})
	audit := &database.CheatAuditLog{
		EventType:          "ban",
		RoomID:             score.RoomID,
		PlayerUID:          score.PlayerUID,
		OperatorUID:        getOperatorUID(c),
		RiskScoreID:        &score.ID,
		SanctionID:         &sanction.ID,
		RiskScore:          &score.RiskScore,
		ReplayID:           score.ReplayID,
		GameHistoryID:      score.GameHistoryID,
		OperationIndex:     score.OperationIndex,
		OperationTimestamp: score.OperationTimestamp,
		PrimaryEvidence:    score.PrimaryEvidence,
		RelatedEvidence:    score.RelatedEvidence,
		SuggestedAction:    score.SuggestedAction,
		IndicatorDetails:   score.IndicatorDetails,
		ReportContribution: score.ReportContribution,
		SanctionType:       "ban",
		NewStatus:          "active",
		NewDecision:        "ban",
		Details:            details,
		Remark:             reason,
	}
	if err := h.acSystem.Repository.SaveAuditLog(audit); err != nil {
		return nil, err
	}
	return sanction, nil
}

func (h *AnticheatHandler) ChangeDetectionPunishment(c *gin.Context) {
	riskScoreID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req struct {
		PunishmentDecision string `json:"punishment_decision"`
		SanctionID         uint   `json:"sanction_id"`
		Reason             string `json:"reason"`
		Duration           *int   `json:"duration"`
		EffectiveUntil     string `json:"effective_until"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if isCancellationDecision(req.PunishmentDecision) {
		h.auditRejectedPunishmentChange(c, uint(riskScoreID), req.PunishmentDecision, req.Reason)
		c.JSON(http.StatusBadRequest, gin.H{"error": "processed punishment cannot be cancelled"})
		return
	}

	score, err := h.acSystem.Repository.GetRiskScoreByID(uint(riskScoreID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Detection not found"})
		return
	}
	if score.ReviewStatus != "processed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "risk entry must be processed before changing punishment"})
		return
	}

	var effectiveUntil *time.Time
	if strings.TrimSpace(req.EffectiveUntil) != "" {
		parsed, err := time.Parse(time.RFC3339, req.EffectiveUntil)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid effective_until"})
			return
		}
		effectiveUntil = &parsed
	}
	if req.Duration != nil && *req.Duration <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "duration must be positive"})
		return
	}
	if effectiveUntil != nil && !effectiveUntil.After(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "effective_until must be in the future"})
		return
	}
	oldDecision := score.PunishmentDecision
	if err := h.acSystem.Repository.UpdateRiskScoreReview(score.ID, "processed", req.PunishmentDecision); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var enforcedSanction *database.CheatSanction
	if isBanDecision(req.PunishmentDecision) {
		enforcedSanction, err = h.enforceDetectionBan(c, *score, req.Reason, effectiveUntil, req.Duration, "admin_anticheat_punishment_change", req.SanctionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else if req.SanctionID > 0 {
		if err := h.acSystem.Repository.UpdateSanctionDecision(req.SanctionID, req.PunishmentDecision, req.Reason, req.Duration, effectiveUntil); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	audit := &database.CheatAuditLog{
		EventType:          "punishment_change",
		RoomID:             score.RoomID,
		PlayerUID:          score.PlayerUID,
		OperatorUID:        getOperatorUID(c),
		RiskScoreID:        &score.ID,
		RiskScore:          &score.RiskScore,
		ReplayID:           score.ReplayID,
		GameHistoryID:      score.GameHistoryID,
		OperationIndex:     score.OperationIndex,
		OperationTimestamp: score.OperationTimestamp,
		PrimaryEvidence:    score.PrimaryEvidence,
		RelatedEvidence:    score.RelatedEvidence,
		SuggestedAction:    score.SuggestedAction,
		IndicatorDetails:   score.IndicatorDetails,
		ReportContribution: score.ReportContribution,
		OldDecision:        oldDecision,
		NewDecision:        req.PunishmentDecision,
		Remark:             req.Reason,
	}
	if req.SanctionID > 0 {
		audit.SanctionID = &req.SanctionID
	} else if enforcedSanction != nil {
		audit.SanctionID = &enforcedSanction.ID
	}
	if err := h.acSystem.Repository.SaveAuditLog(audit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "punishment decision updated", "audit": auditLogDTO(*audit)})
}

func (h *AnticheatHandler) auditRejectedPunishmentChange(c *gin.Context, riskScoreID uint, requestedDecision string, reason string) {
	if h.acSystem == nil || h.acSystem.Repository == nil {
		return
	}
	score, err := h.acSystem.Repository.GetRiskScoreByID(riskScoreID)
	if err != nil {
		return
	}
	audit := &database.CheatAuditLog{
		EventType:          "punishment_change_rejected",
		RoomID:             score.RoomID,
		PlayerUID:          score.PlayerUID,
		OperatorUID:        getOperatorUID(c),
		RiskScoreID:        &score.ID,
		RiskScore:          &score.RiskScore,
		ReplayID:           score.ReplayID,
		GameHistoryID:      score.GameHistoryID,
		OperationIndex:     score.OperationIndex,
		OperationTimestamp: score.OperationTimestamp,
		PrimaryEvidence:    score.PrimaryEvidence,
		RelatedEvidence:    score.RelatedEvidence,
		OldDecision:        score.PunishmentDecision,
		NewDecision:        requestedDecision,
		Remark:             firstNonEmpty(reason, "processed punishment cannot be cancelled"),
	}
	if err := h.acSystem.Repository.SaveAuditLog(audit); err != nil {
		log.Printf("Failed to write rejected punishment change audit log: %v", err)
	}
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
		"room_ids":            jsonRawOrNil(appeal.RoomIDs),
		"player_id":           appeal.PlayerUID,
		"player_uid":          appeal.PlayerUID,
		"risk_score_id":       appeal.RiskScoreID,
		"sanction_id":         appeal.SanctionID,
		"replay_id":           appeal.ReplayID,
		"game_history_id":     appeal.GameHistoryID,
		"has_replay":          replayExists(replayHistoryID(appeal.GameHistoryID, appeal.ReplayID)),
		"primary_evidence":    replayEvidenceOrFallback(appeal.RoomID, appeal.ReplayID, appeal.GameHistoryID, 0, appeal.PlayerUID, appeal.PrimaryEvidence),
		"related_evidence":    jsonRawWithReplayAvailability(appeal.RelatedEvidence),
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

func stringSliceFromJSON(raw database.JSON) []string {
	if len(raw) == 0 {
		return nil
	}
	var stringsOut []string
	if err := json.Unmarshal(raw, &stringsOut); err == nil {
		return stringsOut
	}
	return nil
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
		EventType:       "review",
		RoomID:          appeal.RoomID,
		PlayerUID:       appeal.PlayerUID,
		OperatorUID:     &reviewerUID,
		RiskScoreID:     &appeal.RiskScoreID,
		SanctionID:      &appeal.SanctionID,
		AppealID:        &appeal.ID,
		ReplayID:        appeal.ReplayID,
		GameHistoryID:   appeal.GameHistoryID,
		PrimaryEvidence: appeal.PrimaryEvidence,
		RelatedEvidence: appeal.RelatedEvidence,
		OldStatus:       appeal.Status,
		NewStatus:       "rejected",
		Remark:          req.Remark,
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
		"room_id",
		"replay_id",
		"game_history_id",
		"event_index",
		"event_id",
		"evidence_precision",
		"action_summary",
		"navigation_url",
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
		evidence := replayEvidenceOrFallback(log.RoomID, log.ReplayID, log.GameHistoryID, log.OperationIndex, log.PlayerUID, log.PrimaryEvidence)
		field := func(key string) string {
			if value, ok := evidence[key]; ok && value != nil {
				return fmt.Sprint(value)
			}
			return ""
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
			log.RoomID,
			log.ReplayID,
			strconv.FormatUint(uint64(log.GameHistoryID), 10),
			field("event_index"),
			field("event_id"),
			field("evidence_precision"),
			field("action_summary"),
			field("navigation_url"),
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
	req.Reason = strings.TrimSpace(req.Reason)
	req.Evidence = strings.TrimSpace(req.Evidence)
	if req.Reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reason is required"})
		return
	}

	entry, err := h.buildAppealEntry(playerUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !entry.IsBanned {
		c.JSON(http.StatusForbidden, gin.H{"error": "appeals are only available for banned accounts", "redirect_to": "/feedbacks", "open_feedback": true})
		return
	}
	fallbackRoomID := strings.TrimSpace(c.Param("roomId"))
	if len(entry.RoomIDs) == 0 && fallbackRoomID != "" && fallbackRoomID != "account" {
		entry.RoomIDs = []string{fallbackRoomID}
	}
	if req.RiskScoreID == 0 && entry.LatestRiskScoreID != nil {
		req.RiskScoreID = *entry.LatestRiskScoreID
	}
	roomID := fallbackRoomID
	if len(entry.RoomIDs) > 0 {
		roomID = entry.RoomIDs[0]
	}
	if roomID == "" {
		roomID = "account"
	}

	appeal, err := h.acSystem.AppealManager.SubmitAppealWithRooms(roomID, playerUID, req.RiskScoreID, req.SanctionID, req.Reason, req.Evidence, entry.RoomIDs)
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

type appealEntryState struct {
	IsBanned          bool       `json:"is_banned"`
	BannedUntil       *time.Time `json:"banned_until,omitempty"`
	BanReason         string     `json:"ban_reason,omitempty"`
	LatestRiskScoreID *uint      `json:"latest_risk_score_id,omitempty"`
	FirstRoomID       string     `json:"first_room_id,omitempty"`
	RoomIDs           []string   `json:"room_ids"`
	CanSubmit         bool       `json:"can_submit"`
}

func (h *AnticheatHandler) buildAppealEntry(playerUID uint) (*appealEntryState, error) {
	bannedUntil, _, banReason, err := h.acSystem.Repository.GetUserBanStatus(playerUID)
	if err != nil {
		return nil, err
	}
	isBanned := bannedUntil != nil && bannedUntil.After(time.Now())
	entry := &appealEntryState{
		IsBanned:    isBanned,
		BannedUntil: bannedUntil,
		BanReason:   banReason,
		RoomIDs:     []string{},
		CanSubmit:   false,
	}

	if !entry.IsBanned {
		sanctions, sanctionErr := h.activeSanctionsForPlayer(playerUID)
		if sanctionErr != nil {
			return nil, sanctionErr
		}
		applyAppealEntryBanSanction(entry, sanctions)
	}

	score, err := h.acSystem.Repository.GetLatestRiskScoreByPlayer(playerUID)
	if err == nil && score != nil {
		entry.LatestRiskScoreID = &score.ID
		entry.FirstRoomID = score.RoomID
		histories, historyErr := h.acSystem.Repository.GetGameHistoriesForPlayerSince(playerUID, firstNonZeroTime(score.DetectionTime, score.CreatedAt))
		if historyErr != nil {
			return nil, historyErr
		}
		seen := map[string]struct{}{}
		if score.RoomID != "" {
			entry.RoomIDs = append(entry.RoomIDs, score.RoomID)
			seen[score.RoomID] = struct{}{}
		}
		for _, history := range histories {
			if history.RoomID == "" {
				continue
			}
			if _, exists := seen[history.RoomID]; exists {
				continue
			}
			entry.RoomIDs = append(entry.RoomIDs, history.RoomID)
			seen[history.RoomID] = struct{}{}
		}
	}
	entry.CanSubmit = entry.IsBanned
	return entry, nil
}

func (h *AnticheatHandler) activeSanctionsForPlayer(playerUID uint) ([]database.CheatSanction, error) {
	if h.acSystem.Decider != nil {
		return h.acSystem.Decider.GetActiveSanctionsForPlayer(playerUID)
	}
	sanctions, err := h.acSystem.Repository.GetActiveSanctionsByPlayer(playerUID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	activeSanctions := make([]database.CheatSanction, 0, len(sanctions))
	for i := range sanctions {
		if sanctions[i].EffectiveUntil != nil && sanctions[i].EffectiveUntil.Before(now) {
			if updateErr := h.acSystem.Repository.UpdateSanctionStatus(sanctions[i].ID, "expired"); updateErr != nil {
				log.Printf("[appeal-entry] failed to expire sanction %d: %v", sanctions[i].ID, updateErr)
			}
			continue
		}
		activeSanctions = append(activeSanctions, sanctions[i])
	}
	return activeSanctions, nil
}

func applyAppealEntryBanSanction(entry *appealEntryState, sanctions []database.CheatSanction) {
	for i := range sanctions {
		sanction := sanctions[i]
		if sanction.SanctionType != "ban" || (sanction.Status != "" && sanction.Status != "active") {
			continue
		}
		entry.IsBanned = true
		entry.BannedUntil = sanction.EffectiveUntil
		if entry.BanReason == "" {
			entry.BanReason = sanction.Reason
		}
		break
	}
}

func (h *AnticheatHandler) GetAppealEntryStatus(c *gin.Context) {
	playerUID, ok := getAuthenticatedUID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "player authentication required"})
		return
	}
	entry, err := h.buildAppealEntry(playerUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, entry)
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

	var sourceScore *database.CheatRiskScore
	if req.RiskScoreID != nil && *req.RiskScoreID > 0 {
		if score, err := h.acSystem.Repository.GetRiskScoreByID(*req.RiskScoreID); err == nil && score != nil {
			sourceScore = score
		}
	}

	if err := repository.NewUserRepository().UpdateBanStatusWithReason(req.PlayerUID, &bannedUntil, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	sendBanNotification(req.PlayerUID, &bannedUntil, req.Reason)
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
	if sourceScore != nil {
		sanction.ReplayID = sourceScore.ReplayID
		sanction.GameHistoryID = sourceScore.GameHistoryID
		sanction.PrimaryEvidence = sourceScore.PrimaryEvidence
		if sanction.RoomID == "" {
			sanction.RoomID = sourceScore.RoomID
		}
		if sanction.RiskScore == 0 {
			sanction.RiskScore = sourceScore.RiskScore
		}
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
	if sourceScore != nil {
		audit.ReplayID = sourceScore.ReplayID
		audit.GameHistoryID = sourceScore.GameHistoryID
		audit.OperationIndex = sourceScore.OperationIndex
		audit.OperationTimestamp = sourceScore.OperationTimestamp
		audit.PrimaryEvidence = sourceScore.PrimaryEvidence
		audit.RelatedEvidence = sourceScore.RelatedEvidence
		audit.IndicatorDetails = sourceScore.IndicatorDetails
		audit.ReportContribution = sourceScore.ReportContribution
		if audit.RoomID == "" {
			audit.RoomID = sourceScore.RoomID
		}
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

	sanctions, err := h.activeSanctionsForPlayer(playerUID)
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
