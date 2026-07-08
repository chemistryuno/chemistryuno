package handlers

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ListGameVersions returns all game versions.
func ListGameVersions(c *gin.Context) {
	versions, err := repository.ActivityRepo.ListVersions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, versions)
}

// CreateGameVersion creates a new game version.
func CreateGameVersion(c *gin.Context) {
	var req struct {
		Name      string    `json:"name" binding:"required"`
		StartDate time.Time `json:"start_date" binding:"required"`
		EndDate   time.Time `json:"end_date" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v := &database.GameVersion{
		Name:      req.Name,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}
	if err := repository.ActivityRepo.CreateVersion(v); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, v)
}

// UpdateGameVersion updates a game version.
func UpdateGameVersion(c *gin.Context) {
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
	for _, key := range []string{"start_date", "end_date"} {
		if v, ok := req[key]; ok {
			if s, ok := v.(string); ok {
				t, err := time.Parse(time.RFC3339, s)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid time for " + key})
					return
				}
				req[key] = t
			}
		}
	}
	if err := repository.ActivityRepo.UpdateVersion(uint(id), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	v, _ := repository.ActivityRepo.GetVersion(uint(id))
	c.JSON(http.StatusOK, v)
}
