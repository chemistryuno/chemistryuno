package handlers

import (
	"chemistryuno/backend/repository"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetDoublePointsStatus returns whether double-points activity is active and remaining daily uses.
func GetDoublePointsStatus(c *gin.Context) {
	uid := uint(c.GetInt("uid"))
	acts, err := repository.ActivityRepo.GetActiveActivitiesByType("double_points")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(acts) == 0 {
		c.JSON(http.StatusOK, gin.H{"active": false})
		return
	}
	act := acts[0]
	dailyLimit := repository.GetDailyLimit(&act, 3)
	dateStr := time.Now().UTC().Format("2006-01-02")

	token, err := repository.ActivityRepo.GetTokenUsage(uid, act.ID, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	usedCount := 0
	if token != nil {
		usedCount = token.UsedCount
	}
	remaining := dailyLimit - usedCount
	if remaining < 0 {
		remaining = 0
	}
	c.JSON(http.StatusOK, gin.H{
		"active":      true,
		"activity_id": act.ID,
		"daily_limit": dailyLimit,
		"used_count":  usedCount,
		"remaining":   remaining,
	})
}

// TriggerDoublePoints applies double-points for a completed game.
// The caller must have completed the game (non-quitting). Points are doubled in the game engine.
// This endpoint records the token use and returns the new points delta.
func TriggerDoublePoints(c *gin.Context) {
	uid := uint(c.GetInt("uid"))
	var req struct {
		ActivityID uint   `json:"activity_id" binding:"required"`
		RoomID     string `json:"room_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	act, err := repository.ActivityRepo.GetActivity(req.ActivityID)
	if err != nil || act == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "activity not found"})
		return
	}
	if act.Type != "double_points" || !act.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "activity not available"})
		return
	}
	now := time.Now().UTC()
	if now.Before(act.StartTime) || now.After(act.EndTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "activity not in progress"})
		return
	}

	dailyLimit := repository.GetDailyLimit(act, 3)
	dateStr := now.Format("2006-01-02")

	token, err := repository.ActivityRepo.GetTokenUsage(uid, act.ID, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	usedCount := 0
	if token != nil {
		usedCount = token.UsedCount
	}
	if usedCount >= dailyLimit {
		c.JSON(http.StatusForbidden, gin.H{"error": "今日次数已用完"})
		return
	}

	if err := repository.ActivityRepo.IncrementToken(uid, act.ID, dateStr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":        true,
		"remaining": dailyLimit - usedCount - 1,
	})
}
