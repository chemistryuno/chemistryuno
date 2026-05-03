package handlers

import (
	"chemistryuno/backend/anticheat"
	"chemistryuno/backend/cache"
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/websocket"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

// GetDetectionList 查询检测列表
func (h *AnticheatHandler) GetDetectionList(c *gin.Context) {
	limit := 50
	if l := c.DefaultQuery("limit", "50"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 1000 {
			limit = val
		}
	}

	riskScores, err := h.acSystem.Repository.GetRiskScoresByPlayer(0, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(riskScores),
		"data":  riskScores,
	})
}

// GetDetectionDetail 查询检测详情
func (h *AnticheatHandler) GetDetectionDetail(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{
		"count": len(appeals),
		"data":  appeals,
	})
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

	// 获取申诉详情以获取玩家ID
	appeal, err := h.acSystem.Repository.GetAppealByID(uint(appealID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appeal not found"})
		return
	}

	// 检查幂等性 - 防止重复发放补偿
	ctx := context.Background()
	eventID := "appeal_" + strconv.FormatUint(appealID, 10)
	isDuplicate, err := cache.CheckUnbanCompensationIdempotency(ctx, appeal.PlayerUID, eventID)
	if err != nil {
		log.Printf("⚠️  Idempotency check unavailable: %v", err)
		isDuplicate = false
	}
	if isDuplicate {
		c.JSON(http.StatusOK, gin.H{
			"message":    "Appeal already approved - compensation already issued",
			"idempotent": true,
		})
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

	if outcome.CompensationStatus == "ok" {
		if err := cache.SetUnbanCompensationIdempotency(ctx, appeal.PlayerUID, eventID, config.UnbanConfig.IdempotencyTTL); err != nil {
			log.Printf("Warning: Failed to set idempotency cache: %v", err)
		}
		if websocket.GlobalHub != nil {
			websocket.GlobalHub.SendToUID(int(appeal.PlayerUID), websocket.Message{
				Type:    "unban_compensation",
				Message: compensationMessage,
				Data: gin.H{
					"amount":    compensationAmount,
					"appeal_id": appealID,
				},
			})
		}
	} else {
		_ = cache.DeleteUnbanCompensationIdempotency(ctx, appeal.PlayerUID, eventID)
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
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 获取当前用户ID
	reviewerUID := uint(1)

	if err := h.acSystem.AppealManager.RejectAppeal(uint(appealID), reviewerUID, req.Remark); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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

	c.JSON(http.StatusOK, gin.H{
		"count": len(appeals),
		"data":  appeals,
	})
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
		"count": len(sanctions),
		"data":  sanctions,
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
