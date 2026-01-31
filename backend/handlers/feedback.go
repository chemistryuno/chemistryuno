package handlers

import (
	"chemistryuno/database"
	"chemistryuno/models"
	"chemistryuno/websocket"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateFeedback(c *gin.Context) {
	var feedback models.Feedback
	if err := c.ShouldBindJSON(&feedback); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid := c.GetInt("uid")
	_, err := database.DB.Exec("INSERT INTO feedbacks (user_id, content, type) VALUES (?, ?, ?)",
		uid, feedback.Content, feedback.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交反馈失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "反馈已提交，感谢您的建议！"})
}

func GetAllFeedbacks(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT f.id, f.user_id, u.username, f.content, f.type, f.status, f.processed_by, p.username, f.processed_at, f.last_urged_at, f.urge_count, f.resolution_note, f.created_at
		FROM feedbacks f
		JOIN users u ON f.user_id = u.UID
		LEFT JOIN users p ON f.processed_by = p.UID
		WHERE f.remove_at IS NULL
		ORDER BY f.created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取反馈失败"})
		return
	}
	defer rows.Close()

	var feedbacks []models.Feedback
	for rows.Next() {
		var f models.Feedback
		var processedBy sql.NullInt64
		var processedAt sql.NullString
		var lastUrged sql.NullString
		var resolution sql.NullString
		if err := rows.Scan(&f.ID, &f.UserID, &f.Username, &f.Content, &f.Type, &f.Status, &processedBy, &sql.NullString{}, &processedAt, &lastUrged, &f.UrgeCount, &resolution, &f.CreatedAt); err != nil {
			continue
		}
		if processedBy.Valid {
			pb := int(processedBy.Int64)
			f.ProcessedBy = &pb
		}
		if processedAt.Valid {
			f.ProcessedAt = &processedAt.String
		}
		if lastUrged.Valid {
			f.LastUrgedAt = &lastUrged.String
		}
		if resolution.Valid {
			f.ResolutionNote = &resolution.String
		}
		feedbacks = append(feedbacks, f)
	}

	c.JSON(http.StatusOK, feedbacks)
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
	now := time.Now().UTC().Format("2006-01-02 15:04:05")

	// 默认处理说明
	note := req.Note
	if note == "" {
		if req.Status == "accepted" {
			note = "您的反馈已受理"
		} else if req.Status == "dismissed" {
			note = "您的反馈不予受理"
		}
	}

	// 删除时间：72 小时后从服务器移除
	removeAt := time.Now().UTC().Add(72 * time.Hour).Format("2006-01-02 15:04:05")
	res, err := database.DB.Exec("UPDATE feedbacks SET status = ?, processed_by = ?, processed_at = ?, resolution_note = ?, remove_at = ? WHERE id = ?", req.Status, uid, now, note, removeAt, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新反馈状态失败"})
		return
	}

	// notify the feedback owner via websocket
	var owner int
	if err := database.DB.QueryRow("SELECT user_id FROM feedbacks WHERE id = ?", id).Scan(&owner); err == nil {
		websocket.GlobalHub.SendToUID(owner, gin.H{"type": "feedback_update", "feedback_id": id, "status": req.Status, "resolution_note": note})
	}

	// return rows affected
	if ra, _ := res.RowsAffected(); ra > 0 {
		c.JSON(http.StatusOK, gin.H{"message": "反馈状态已更新"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "未能更新反馈"})
}

// GetMyFeedbacks 返回当前用户的反馈列表
func GetMyFeedbacks(c *gin.Context) {
	uid := c.GetInt("uid")
	rows, err := database.DB.Query(`
		SELECT id, user_id, (SELECT username FROM users WHERE UID = user_id), content, type, status, processed_by, processed_at, last_urged_at, urge_count, resolution_note, created_at
		FROM feedbacks
		WHERE user_id = ? AND remove_at IS NULL
		ORDER BY created_at DESC
	`, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取反馈失败"})
		return
	}
	defer rows.Close()

	var feedbacks []models.Feedback
	for rows.Next() {
		var f models.Feedback
		var processedBy sql.NullInt64
		var processedAt sql.NullString
		var lastUrged sql.NullString
		var resolution sql.NullString
		if err := rows.Scan(&f.ID, &f.UserID, &f.Username, &f.Content, &f.Type, &f.Status, &processedBy, &processedAt, &lastUrged, &f.UrgeCount, &resolution, &f.CreatedAt); err != nil {
			continue
		}
		if processedBy.Valid {
			pb := int(processedBy.Int64)
			f.ProcessedBy = &pb
		}
		if processedAt.Valid {
			f.ProcessedAt = &processedAt.String
		}
		if lastUrged.Valid {
			f.LastUrgedAt = &lastUrged.String
		}
		if resolution.Valid {
			f.ResolutionNote = &resolution.String
		}
		feedbacks = append(feedbacks, f)
	}

	c.JSON(http.StatusOK, feedbacks)
}

// UrgeFeedback 允许用户每 4 小时催促一次指定反馈
func UrgeFeedback(c *gin.Context) {
	id := c.Param("id")
	uid := c.GetInt("uid")

	var owner int
	var lastUrged sql.NullString
	if err := database.DB.QueryRow("SELECT user_id, last_urged_at FROM feedbacks WHERE id = ?", id).Scan(&owner, &lastUrged); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "反馈不存在"})
		return
	}
	if owner != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权催促此反馈"})
		return
	}

	now := time.Now().UTC()
	if lastUrged.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", lastUrged.String); err == nil {
			if now.Sub(t) < 4*time.Hour {
				next := t.Add(4 * time.Hour)
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "请稍后再催促", "next_allowed_at": next.Format("2006-01-02 15:04:05")})
				return
			}
		}
	}

	nowStr := now.Format("2006-01-02 15:04:05")
	_, err := database.DB.Exec("UPDATE feedbacks SET last_urged_at = ?, urge_count = urge_count + 1 WHERE id = ?", nowStr, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "催促失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "催促已发送"})
}
