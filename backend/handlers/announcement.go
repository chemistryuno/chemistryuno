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

// GetActiveAnnouncements 获取当前有效的公告
func GetActiveAnnouncements(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT id, title, content, type, active, is_ticker, created_at, expires_at 
		FROM announcements 
		WHERE active = 1 AND (expires_at IS NULL OR expires_at > ?)
		ORDER BY created_at DESC`, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取公告失败"})
		return
	}
	defer rows.Close()

	var announcements []models.Announcement
	for rows.Next() {
		var a models.Announcement
		var expiresAt sql.NullTime
		var title sql.NullString
		if err := rows.Scan(&a.ID, &title, &a.Content, &a.Type, &a.Active, &a.IsTicker, &a.CreatedAt, &expiresAt); err != nil {
			continue
		}
		if title.Valid {
			a.Title = title.String
		}
		if expiresAt.Valid {
			a.ExpiresAt = &expiresAt.Time
		}
		announcements = append(announcements, a)
	}

	c.JSON(http.StatusOK, announcements)
}

// GetAllAnnouncements 管理员获取所有公告
func GetAllAnnouncements(c *gin.Context) {
	rows, err := database.DB.Query("SELECT id, title, content, type, active, is_ticker, created_at, expires_at FROM announcements ORDER BY created_at DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取公告失败"})
		return
	}
	defer rows.Close()

	var announcements []models.Announcement
	for rows.Next() {
		var a models.Announcement
		var expiresAt sql.NullTime
		var title sql.NullString
		if err := rows.Scan(&a.ID, &title, &a.Content, &a.Type, &a.Active, &a.IsTicker, &a.CreatedAt, &expiresAt); err != nil {
			continue
		}
		if title.Valid {
			a.Title = title.String
		}
		if expiresAt.Valid {
			a.ExpiresAt = &expiresAt.Time
		}
		announcements = append(announcements, a)
	}

	c.JSON(http.StatusOK, announcements)
}

// CreateAnnouncement 创建新公告
func CreateAnnouncement(c *gin.Context) {
	var req struct {
		Title     string `json:"title"`
		Content   string `json:"content" binding:"required"`
		Type      string `json:"type"`
		IsTicker  bool   `json:"is_ticker"`
		ExpiresIn string `json:"expires_in"` // 持续时间，例如 "24h"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	var expiresAt *time.Time
	if req.ExpiresIn != "" {
		duration, err := time.ParseDuration(req.ExpiresIn)
		if err == nil {
			t := time.Now().Add(duration)
			expiresAt = &t
		}
	}

	res, err := database.DB.Exec("INSERT INTO announcements (title, content, type, is_ticker, expires_at) VALUES (?, ?, ?, ?, ?)",
		req.Title, req.Content, req.Type, req.IsTicker, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建公告失败"})
		return
	}

	id, _ := res.LastInsertId()

	// 实时广播
	if websocket.GlobalHub != nil {
		websocket.GlobalHub.BroadcastToAll(websocket.Message{
			Type: "system_announcement",
			Data: map[string]interface{}{
				"id":         id,
				"title":      req.Title,
				"content":    req.Content,
				"type":       req.Type,
				"is_ticker":  req.IsTicker,
				"active":     true,
				"created_at": time.Now(),
			},
		})
	}

	c.JSON(http.StatusCreated, gin.H{"message": "公告发布成功", "id": id})
}

// UpdateAnnouncementStatus 更新公告状态
func UpdateAnnouncementStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Active bool `json:"active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	_, err := database.DB.Exec("UPDATE announcements SET active = ? WHERE id = ?", req.Active, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// DeleteAnnouncement 删除公告
func DeleteAnnouncement(c *gin.Context) {
	id := c.Param("id")
	_, err := database.DB.Exec("DELETE FROM announcements WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
