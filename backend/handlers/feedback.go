package handlers

import (
	"chemistryuno/database"
	"chemistryuno/models"
	"net/http"

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
		SELECT f.id, f.user_id, u.username, f.content, f.type, f.status, f.created_at 
		FROM feedbacks f
		JOIN users u ON f.user_id = u.UID
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
		if err := rows.Scan(&f.ID, &f.UserID, &f.Username, &f.Content, &f.Type, &f.Status, &f.CreatedAt); err != nil {
			continue
		}
		feedbacks = append(feedbacks, f)
	}

	c.JSON(http.StatusOK, feedbacks)
}

func UpdateFeedbackStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := database.DB.Exec("UPDATE feedbacks SET status = ? WHERE id = ?", req.Status, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新反馈状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "反馈状态已更新"})
}
