package handlers

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/game"
	"chemistryuno/backend/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetLevelInfo returns current user's level info.
func GetLevelInfo(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var uidUint uint
	switch v := uid.(type) {
	case int:
		if v < 0 {
			c.JSON(http.StatusOK, gin.H{
				"level":            1,
				"xp":               0,
				"total_xp":         0,
				"tier":             "ai",
				"tier_name":        "AI",
				"next_level_xp":    0,
				"progress_percent": 0,
			})
			return
		}
		uidUint = uint(v)
	case uint:
		uidUint = v
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid uid type"})
		return
	}

	levelInfo, err := game.GetLevelInfo(uidUint)
	if err != nil {
		if err.Error() == "用户不存在" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "session invalid"})
			return
		}
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get level info"})
		return
	}

	c.JSON(http.StatusOK, levelInfo)
}

// GetUserLevelInfo returns public level info of target user.
func GetUserLevelInfo(c *gin.Context) {
	uidStr := c.Param("uid")
	targetUID, err := strconv.ParseUint(uidStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uid"})
		return
	}

	levelInfo, err := game.GetLevelInfo(uint(targetUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get level info"})
		return
	}

	c.JSON(http.StatusOK, levelInfo)
}

// GetLevelLeaderboard returns leaderboard sorted by total_xp.
func GetLevelLeaderboard(c *gin.Context) {
	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	users, err := repository.UserRepo.GetLeaderboard("total_xp", limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get leaderboard"})
		return
	}

	var levelConfigs []database.LevelConfig
	if err := database.DB.Find(&levelConfigs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	levelConfigMap := make(map[int]database.LevelConfig, len(levelConfigs))
	for _, cfg := range levelConfigs {
		levelConfigMap[cfg.Level] = cfg
	}

	type LeaderboardEntry struct {
		UID      uint   `json:"uid"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
		Level    int    `json:"level"`
		TotalXP  int    `json:"total_xp"`
		Tier     string `json:"tier"`
		TierName string `json:"tier_name"`
	}

	result := make([]LeaderboardEntry, 0, len(users))
	for _, user := range users {
		levelConfig := levelConfigMap[user.Level]
		result = append(result, LeaderboardEntry{
			UID:      user.UID,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Level:    user.Level,
			TotalXP:  user.TotalXP,
			Tier:     levelConfig.Tier,
			TierName: levelConfig.TierName,
		})
	}

	c.JSON(http.StatusOK, result)
}

// GetAllLevelConfigs returns all level configs.
func GetAllLevelConfigs(c *gin.Context) {
	var configs []database.LevelConfig
	if err := database.DB.Order("level ASC").Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get level configs"})
		return
	}

	c.JSON(http.StatusOK, configs)
}
