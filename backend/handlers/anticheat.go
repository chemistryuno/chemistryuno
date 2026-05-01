package handlers

import (
	"chemistryuno/backend/anticheat"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AnticheatHandler 反作弊处理程序
type AnticheatHandler struct {
	acSystem *anticheat.System
}

// NewAnticheatHandler 创建反作弊处理程序
func NewAnticheatHandler(acSystem *anticheat.System) *AnticheatHandler {
	return &AnticheatHandler{
		acSystem: acSystem,
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

// ApproveAppeal 批准申诉
func (h *AnticheatHandler) ApproveAppeal(c *gin.Context) {
	appealID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req struct {
		Remark             string `json:"remark"`
		CompensationAmount *int   `json:"compensation_amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 获取当前用户ID
	reviewerUID := uint(1)

	config := h.acSystem.Config.GetConfig()
	if req.CompensationAmount != nil {
		config.UnbanConfig.CompensationAmount = *req.CompensationAmount
	}

	if err := h.acSystem.AppealManager.ApproveAppeal(uint(appealID), reviewerUID, req.Remark, h.acSystem.Decider, config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Appeal approved"})
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
	c.JSON(http.StatusOK, config)
}

// UpdateConfig 更新配置
func (h *AnticheatHandler) UpdateConfig(c *gin.Context) {
	var config anticheat.RiskScoringConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.acSystem.Config.UpdateDimensionWeight("response_time", config.Dimensions["response_time"].Weight)

	c.JSON(http.StatusOK, gin.H{"message": "Configuration updated"})
}

// GetAuditLog 查询审计日志
func (h *AnticheatHandler) GetAuditLog(c *gin.Context) {
	startStr := c.DefaultQuery("start", "")
	endStr := c.DefaultQuery("end", "")

	var startTime, endTime time.Time
	startTime = time.Now().Add(-24 * time.Hour)
	endTime = time.Now()

	if startStr != "" {
		if t, err := time.Parse("2006-01-02", startStr); err == nil {
			startTime = t
		}
	}

	if endStr != "" {
		if t, err := time.Parse("2006-01-02", endStr); err == nil {
			endTime = t
		}
	}

	logs, err := h.acSystem.AuditLogger.GetAuditLogsByTimeRange(startTime, endTime, 1000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(logs),
		"data":  logs,
	})
}

// SubmitAppeal 提交申诉（玩家端）
func (h *AnticheatHandler) SubmitAppeal(c *gin.Context) {
	_ = c.Param("roomId") // roomID

	// TODO: 获取当前玩家UID

	var req struct {
		Reason   string `json:"reason"`
		Evidence string `json:"evidence"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 实现申诉提交逻辑

	c.JSON(http.StatusOK, gin.H{"message": "Appeal submitted"})
}

// GetPlayerAppeals 获取玩家的申诉历史
func (h *AnticheatHandler) GetPlayerAppeals(c *gin.Context) {
	// TODO: 获取当前玩家UID
	playerUID := uint(1)

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

// GetPlayerSanctions 获取玩家的处罚历史
func (h *AnticheatHandler) GetPlayerSanctions(c *gin.Context) {
	// TODO: 获取当前玩家UID
	playerUID := uint(1)

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
