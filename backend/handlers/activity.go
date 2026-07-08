package handlers

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ListActivities returns activities; admin sees all, normal users see only active+current.
func ListActivities(c *gin.Context) {
	isAdmin := c.GetBool("is_admin") || c.GetString("role") == "co-worker" || c.GetString("role") == "admin"
	acts, err := repository.ActivityRepo.ListActivities(isAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, acts)
}

// CreateActivity creates a new activity (admin only).
func CreateActivity(c *gin.Context) {
	var req struct {
		Name      string          `json:"name" binding:"required"`
		Type      string          `json:"type" binding:"required"`
		StartTime time.Time       `json:"start_time" binding:"required"`
		EndTime   time.Time       `json:"end_time" binding:"required"`
		VersionID *uint           `json:"version_id"`
		Params    json.RawMessage `json:"params"`
		IsActive  *bool           `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	act := &database.Activity{
		Name:      req.Name,
		Type:      req.Type,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		VersionID: req.VersionID,
		Params:    database.JSON(req.Params),
		IsActive:  true,
	}
	if req.IsActive != nil {
		act.IsActive = *req.IsActive
	}
	if err := repository.ActivityRepo.CreateActivity(act); err != nil {
		if isConflictError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, act)
}

// UpdateActivity updates an existing activity.
func UpdateActivity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Parse time fields if present
	for _, key := range []string{"start_time", "end_time"} {
		if v, ok := req[key]; ok {
			if s, ok := v.(string); ok {
				t, err := time.Parse(time.RFC3339, s)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid time format for " + key})
					return
				}
				req[key] = t
			}
		}
	}
	if err := repository.ActivityRepo.UpdateActivity(uint(id), req); err != nil {
		if isConflictError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	act, _ := repository.ActivityRepo.GetActivity(uint(id))
	c.JSON(http.StatusOK, act)
}

// ToggleActivity enables/disables an activity.
func ToggleActivity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := repository.ActivityRepo.ToggleActivity(uint(id), req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func isConflictError(err error) bool {
	if err == nil {
		return false
	}
	return len(err.Error()) > 8 && err.Error()[:8] == "conflict"
}
