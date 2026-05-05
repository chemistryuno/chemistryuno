package handlers

import (
	"chemistryuno/backend/anticheat"
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/websocket"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateFeedback(c *gin.Context) {
	var req struct {
		Content      string                         `json:"content" binding:"required"`
		Type         string                         `json:"type" binding:"required"`
		RoomID       string                         `json:"room_id"`
		ReportedUID  uint                           `json:"reported_uid"`
		ReplayAnchor *database.ReplayEvidenceAnchor `json:"replay_anchor"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid := c.GetInt("uid")
	var primaryEvidence database.JSON
	replayID := ""
	gameHistoryID := uint(0)
	roomID := strings.TrimSpace(req.RoomID)
	if req.ReplayAnchor != nil {
		anchor, err := validateFeedbackReplayAnchor(uint(uid), req.Type, *req.ReplayAnchor, roomID, req.ReportedUID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		primaryEvidence = anticheat.MarshalReplayEvidenceAnchor(anchor)
		replayID = anchor.ReplayID
		gameHistoryID = anchor.GameHistoryID
		if roomID == "" {
			roomID = anchor.RoomID
		}
	}

	// 简化版本：直接创建反馈，不处理复杂的方程式逻辑
	feedback := &database.Feedback{
		UserUID:         uint(uid),
		Content:         req.Content,
		Type:            req.Type,
		RoomID:          roomID,
		ReportedUID:     req.ReportedUID,
		ReplayID:        replayID,
		GameHistoryID:   gameHistoryID,
		PrimaryEvidence: primaryEvidence,
		Status:          "pending",
	}

	err := repository.FeedbackRepo.Create(feedback)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交反馈失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "感谢您的反馈！我们会尽快处理。"})
}

func validateFeedbackReplayAnchor(reporterUID uint, feedbackType string, anchor database.ReplayEvidenceAnchor, roomID string, reportedUID uint) (database.ReplayEvidenceAnchor, error) {
	anchor.RoomID = firstNonEmpty(anchor.RoomID, roomID)
	if strings.TrimSpace(feedbackType) != "report" {
		return anticheat.NormalizeReplayEvidenceAnchor(anchor), nil
	}

	historyID := anchor.GameHistoryID
	if historyID == 0 && strings.TrimSpace(anchor.ReplayID) != "" {
		if parsed, err := strconv.ParseUint(anchor.ReplayID, 10, 32); err == nil {
			historyID = uint(parsed)
		}
	}
	if historyID == 0 {
		return anticheat.NormalizeReplayEvidenceAnchor(anchor), nil
	}

	history, err := repository.GameRepo.FindByID(historyID)
	if err != nil {
		return database.ReplayEvidenceAnchor{}, fmt.Errorf("invalid replay evidence anchor")
	}
	if history.ReplayLog == "" {
		return database.ReplayEvidenceAnchor{}, fmt.Errorf("referenced replay is unavailable")
	}
	if anchor.RoomID != "" && anchor.RoomID != history.RoomID {
		return database.ReplayEvidenceAnchor{}, fmt.Errorf("replay evidence does not belong to the reported room")
	}
	players := parseIntSliceJSON(history.Players)
	if !containsUID(players, int(reporterUID)) {
		return database.ReplayEvidenceAnchor{}, fmt.Errorf("reporter cannot access referenced replay")
	}
	if reportedUID > 0 && !containsUID(players, int(reportedUID)) {
		return database.ReplayEvidenceAnchor{}, fmt.Errorf("reported player is not in referenced replay")
	}

	anchor.RoomID = history.RoomID
	anchor.GameHistoryID = history.ID
	if anchor.ReplayID == "" {
		anchor.ReplayID = strconv.FormatUint(uint64(history.ID), 10)
	}
	if (anchor.EventID != "" || anchor.EventIndex > 0 || anchor.EventTimestampMs > 0) && !replayAnchorExistsInLog(history.ReplayLog, anchor) {
		return database.ReplayEvidenceAnchor{}, fmt.Errorf("replay evidence anchor was not found in replay")
	}
	return anticheat.NormalizeReplayEvidenceAnchor(anchor), nil
}

func replayAnchorExistsInLog(replayLog string, anchor database.ReplayEvidenceAnchor) bool {
	var replay map[string]interface{}
	if err := json.Unmarshal([]byte(replayLog), &replay); err != nil {
		return false
	}
	rawEvents, ok := replay["events"].([]interface{})
	if !ok {
		return false
	}
	if anchor.EventIndex > 0 && anchor.EventID == "" && anchor.EventTimestampMs == 0 {
		return anchor.EventIndex <= len(rawEvents)
	}
	for index, raw := range rawEvents {
		event, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if anchor.EventID != "" && stringFromEvent(event, "event_id") != anchor.EventID {
			continue
		}
		if anchor.EventIndex > 0 {
			eventIndex := intFromEvent(event, "event_index")
			if eventIndex == 0 {
				eventIndex = index + 1
			}
			if eventIndex != anchor.EventIndex {
				continue
			}
		}
		if anchor.EventTimestampMs > 0 {
			eventTimestamp := int64FromEvent(event, "unix_ms")
			if eventTimestamp == 0 {
				eventTimestamp = timestampMsFromEvent(event, "timestamp")
			}
			if eventTimestamp != anchor.EventTimestampMs {
				continue
			}
		}
		return true
	}
	return false
}

func stringFromEvent(event map[string]interface{}, key string) string {
	if value, ok := event[key].(string); ok {
		return value
	}
	return ""
}

func intFromEvent(event map[string]interface{}, key string) int {
	switch value := event[key].(type) {
	case int:
		return value
	case float64:
		return int(value)
	case json.Number:
		parsed, _ := value.Int64()
		return int(parsed)
	default:
		return 0
	}
}

func int64FromEvent(event map[string]interface{}, key string) int64 {
	switch value := event[key].(type) {
	case int64:
		return value
	case int:
		return int64(value)
	case float64:
		return int64(value)
	case json.Number:
		parsed, _ := value.Int64()
		return parsed
	default:
		return 0
	}
}

func timestampMsFromEvent(event map[string]interface{}, key string) int64 {
	raw, ok := event[key].(string)
	if !ok || raw == "" {
		return 0
	}
	if parsed, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return parsed.UnixMilli()
	}
	return 0
}

func GetAllFeedbacks(c *gin.Context) {
	feedbacks, err := repository.FeedbackRepo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取反馈列表失败"})
		return
	}

	// 组装返回数据，添加用户名信息
	type FeedbackWithUser struct {
		ID              uint        `json:"id"`
		UserUID         uint        `json:"user_uid"`
		Username        string      `json:"username"`
		Nickname        string      `json:"nickname"`
		Type            string      `json:"type"`
		Page            string      `json:"page"` // page 字段映射到 type
		Content         string      `json:"content"`
		RoomID          string      `json:"room_id,omitempty"`
		ReportedUID     uint        `json:"reported_uid,omitempty"`
		ReplayID        string      `json:"replay_id,omitempty"`
		GameHistoryID   uint        `json:"game_history_id,omitempty"`
		PrimaryEvidence interface{} `json:"primary_evidence,omitempty"`
		Status          string      `json:"status"`
		ProcessedByUID  *uint       `json:"processed_by_uid"`
		ProcessedAt     *time.Time  `json:"processed_at"`
		LastUrgedAt     *time.Time  `json:"last_urged_at"`
		UrgeCount       int         `json:"urge_count"`
		ResolutionNote  string      `json:"resolution_note"`
		RemoveAt        *time.Time  `json:"remove_at"`
		CreatedAt       time.Time   `json:"created_at"`
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
			ID:              fb.ID,
			UserUID:         fb.UserUID,
			Username:        username,
			Nickname:        nickname,
			Type:            fb.Type,
			Page:            fb.Type, // page 字段使用 type 的值
			Content:         fb.Content,
			Status:          fb.Status,
			RoomID:          fb.RoomID,
			ReportedUID:     fb.ReportedUID,
			ReplayID:        fb.ReplayID,
			GameHistoryID:   fb.GameHistoryID,
			PrimaryEvidence: jsonRawOrNil(fb.PrimaryEvidence),
			ProcessedByUID:  fb.ProcessedByUID,
			ProcessedAt:     fb.ProcessedAt,
			LastUrgedAt:     fb.LastUrgedAt,
			UrgeCount:       fb.UrgeCount,
			ResolutionNote:  fb.ResolutionNote,
			RemoveAt:        fb.RemoveAt,
			CreatedAt:       fb.CreatedAt,
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

// 消除系统警告（特别是异常登陆警告）
func DismissFeedback(c *gin.Context) {
	uid := c.GetInt("uid")
	feedbackID := c.Param("id")

	id, err := strconv.ParseUint(feedbackID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的反馈ID"})
		return
	}

	// 检查反馈是否存在且属于当前用户
	feedback, err := repository.FeedbackRepo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "反馈不存在"})
		return
	}

	if int(feedback.UserUID) != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权消除此警告"})
		return
	}

	// 系统警告可以直接删除
	if feedback.Type == "system_alert" || feedback.Type == "security_alert" {
		err = repository.FeedbackRepo.Delete(uint(id))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "警告已消除"})
		return
	}

	// 其他类型的消息标记为已消除
	err = repository.FeedbackRepo.UpdateStatus(uint(id), "dismissed", 0, "用户已消除")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "反馈已消除"})
}
