package handlers

import (
	"chemistryuno/repository"
	"chemistryuno/websocket"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetActiveAnnouncements 获取当前有效的公告
func GetActiveAnnouncements(c *gin.Context) {
	announcements, err := repository.AnnouncementRepo.FindActive()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取公告失败"})
		return
	}
	c.JSON(http.StatusOK, announcements)
}

// GetAllAnnouncements 管理员获取所有公告
func GetAllAnnouncements(c *gin.Context) {
	announcements, err := repository.AnnouncementRepo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取公告失败"})
		return
	}
	c.JSON(http.StatusOK, announcements)
}

// CreateAnnouncement 创建公告
func CreateAnnouncement(c *gin.Context) {
	var req struct {
		Title        string `json:"title"`
		Content      string `json:"content" binding:"required"`
		Type         string `json:"type"`
		IsTicker     bool   `json:"is_ticker"`
		IsPersistent bool   `json:"is_persistent"`
		OnJoin       bool   `json:"on_join"`
		CronInterval int    `json:"cron_interval"`
		CloseDelay   int    `json:"close_delay"`
		ExpiresIn    string `json:"expires_in"`
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

	announcement := &repository.Announcement{
		Title:        req.Title,
		Content:      req.Content,
		Type:         req.Type,
		Active:       true,
		IsTicker:     req.IsTicker,
		IsPersistent: req.IsPersistent,
		OnJoin:       req.OnJoin,
		CronInterval: req.CronInterval,
		CloseDelay:   req.CloseDelay,
		ExpiresAt:    expiresAt,
	}

	err := repository.AnnouncementRepo.Create(announcement)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建公告失败: " + err.Error()})
		return
	}

	// 实时广播
	if websocket.GlobalHub != nil {
		websocket.GlobalHub.BroadcastToAll(websocket.Message{
			Type: "system_announcement",
			Data: map[string]interface{}{
				"id":            announcement.ID,
				"title":         req.Title,
				"content":       req.Content,
				"type":          req.Type,
				"is_ticker":     req.IsTicker,
				"is_persistent": req.IsPersistent,
				"close_delay":   req.CloseDelay,
				"active":        true,
				"created_at":    time.Now(),
			},
		})
	}

	c.JSON(http.StatusCreated, gin.H{"message": "公告发布成功", "id": announcement.ID})
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

	idUint, _ := strconv.ParseUint(id, 10, 32)
	err := repository.AnnouncementRepo.UpdateActive(uint(idUint), req.Active)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// DeleteAnnouncement 删除公告
func DeleteAnnouncement(c *gin.Context) {
	id := c.Param("id")
	idUint, _ := strconv.ParseUint(id, 10, 32)
	err := repository.AnnouncementRepo.Delete(uint(idUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
