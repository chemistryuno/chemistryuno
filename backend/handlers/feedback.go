package handlers

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/websocket"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateFeedback(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
		Type    string `json:"type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid := c.GetInt("uid")

	// 简化版本：直接创建反馈，不处理复杂的方程式逻辑
	feedback := &database.Feedback{
		UserUID: uint(uid),
		Content: req.Content,
		Type:    req.Type,
		Status:  "pending",
	}

	err := repository.FeedbackRepo.Create(feedback)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交反馈失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "感谢您的反馈！我们会尽快处理。"})
}

func GetAllFeedbacks(c *gin.Context) {
	feedbacks, err := repository.FeedbackRepo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取反馈列表失败"})
		return
	}

	// 组装返回数据，添加用户名信息
	type FeedbackWithUser struct {
		ID             uint       `json:"id"`
		UserUID        uint       `json:"user_uid"`
		Username       string     `json:"username"`
		Nickname       string     `json:"nickname"`
		Type           string     `json:"type"`
		Page           string     `json:"page"` // page 字段映射到 type
		Content        string     `json:"content"`
		Status         string     `json:"status"`
		ProcessedByUID *uint      `json:"processed_by_uid"`
		ProcessedAt    *time.Time `json:"processed_at"`
		LastUrgedAt    *time.Time `json:"last_urged_at"`
		UrgeCount      int        `json:"urge_count"`
		ResolutionNote string     `json:"resolution_note"`
		RemoveAt       *time.Time `json:"remove_at"`
		CreatedAt      time.Time  `json:"created_at"`
	}

	result := make([]FeedbackWithUser, 0, len(feedbacks))
	for _, fb := range feedbacks {
		// 查询用户信息
		user, err := repository.UserRepo.FindByUID(fb.UserUID)
		username := "未知用户"
		nickname := "未知用户"
		if err == nil && user != nil {
			username = user.Username
			nickname = user.Nickname
		}

		result = append(result, FeedbackWithUser{
			ID:             fb.ID,
			UserUID:        fb.UserUID,
			Username:       username,
			Nickname:       nickname,
			Type:           fb.Type,
			Page:           fb.Type, // page 字段使用 type 的值
			Content:        fb.Content,
			Status:         fb.Status,
			ProcessedByUID: fb.ProcessedByUID,
			ProcessedAt:    fb.ProcessedAt,
			LastUrgedAt:    fb.LastUrgedAt,
			UrgeCount:      fb.UrgeCount,
			ResolutionNote: fb.ResolutionNote,
			RemoveAt:       fb.RemoveAt,
			CreatedAt:      fb.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, result)
}

func UpdateFeedbackStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
		Note   string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid := c.GetInt("uid")

	// 默认处理说明
	note := req.Note
	if note == "" {
		if req.Status == "accepted" {
			note = "您的反馈已受理"
		} else if req.Status == "dismissed" {
			note = "您的反馈不予受理"
		}
	}

	idUint, _ := strconv.ParseUint(id, 10, 32)
	err := repository.FeedbackRepo.UpdateStatus(uint(idUint), req.Status, uint(uid), note)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新反馈状态失败"})
		return
	}

	// 通过websocket通知反馈所有者
	feedback, err := repository.FeedbackRepo.FindByID(uint(idUint))
	if err == nil && websocket.GlobalHub != nil {
		websocket.GlobalHub.SendToUID(int(feedback.UserUID), gin.H{"type": "feedback_update", "feedback_id": id, "status": req.Status, "resolution_note": note})
	}

	c.JSON(http.StatusOK, gin.H{"message": "反馈状态已更新"})
}

// GetMyFeedbacks 返回当前用户的反馈列表
func GetMyFeedbacks(c *gin.Context) {
	uid := c.GetInt("uid")
	feedbacks, err := repository.FeedbackRepo.FindByUserUID(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取反馈失败"})
		return
	}
	c.JSON(http.StatusOK, feedbacks)
}

// UrgeFeedback 允许用户每 4 小时催促一次指定反馈
func UrgeFeedback(c *gin.Context) {
	id := c.Param("id")
	uid := c.GetInt("uid")

	idUint, _ := strconv.ParseUint(id, 10, 32)
	feedback, err := repository.FeedbackRepo.FindByID(uint(idUint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "反馈不存在"})
		return
	}
	if int(feedback.UserUID) != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权催促此反馈"})
		return
	}

	now := time.Now().UTC()
	if feedback.LastUrgedAt != nil {
		next := feedback.LastUrgedAt.Add(4 * time.Hour)
		if now.Before(next) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "请稍后再催促", "next_allowed_at": next.Format("2006-01-02 15:04:05")})
			return
		}
	}

	err = repository.FeedbackRepo.UpdateUrge(uint(idUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "催促失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "催促已发送"})
}

// 撤回反馈
func WithdrawFeedback(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		ID int `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的参数"})
		return
	}

	// 检查权限
	feedback, err := repository.FeedbackRepo.FindByID(uint(req.ID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "反馈不存在"})
		return
	}
	if int(feedback.UserUID) != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除此反馈"})
		return
	}

	err = repository.FeedbackRepo.Delete(uint(req.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "反馈已撤回"})
}
