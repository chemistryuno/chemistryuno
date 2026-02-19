package handlers

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/game"
	"chemistryuno/backend/middleware"
	"chemistryuno/backend/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetLevelInfo 获取等级信息
func GetLevelInfo(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 修复类型转换：uid 可能是 int 或 uint
	var uidUint uint
	switch v := uid.(type) {
	case int:
		uidUint = uint(v)
	case uint:
		uidUint = v
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无效的用户ID类型"})
		return
	}

	levelInfo, err := game.GetLevelInfo(uidUint)
	if err != nil {
		// 记录详细错误信息以便调试
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取等级信息失败", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, levelInfo)
}

// GetUserLevelInfo 获取指定用户的等级信息（公开信息）
func GetUserLevelInfo(c *gin.Context) {
	uidStr := c.Param("uid")
	targetUID, err := strconv.ParseUint(uidStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	levelInfo, err := game.GetLevelInfo(uint(targetUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取等级信息失败"})
		return
	}

	c.JSON(http.StatusOK, levelInfo)
}

// GetLevelLeaderboard 获取等级排行榜
func GetLevelLeaderboard(c *gin.Context) {
	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	users, err := repository.UserRepo.GetLeaderboard("total_xp", limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取排行榜失败"})
		return
	}

	// 转换为包含等级信息的格式
	type LeaderboardEntry struct {
		UID       uint   `json:"uid"`
		Nickname  string `json:"nickname"`
		Avatar    string `json:"avatar"`
		Level     int    `json:"level"`
		TotalXP   int    `json:"total_xp"`
		Tier      string `json:"tier"`
		TierName  string `json:"tier_name"`
	}

	result := make([]LeaderboardEntry, 0, len(users))
	for _, user := range users {
		// 查询等级配置
		var levelConfig database.LevelConfig
		if err := database.DB.Where("level = ?", user.Level).First(&levelConfig).Error; err != nil {
			continue
		}

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

// GetAllLevelConfigs 获取所有等级配置（用于前端展示）
func GetAllLevelConfigs(c *gin.Context) {
	var configs []database.LevelConfig
	if err := database.DB.Order("level ASC").Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取等级配置失败"})
		return
	}

	c.JSON(http.StatusOK, configs)
}

// RegisterLevelRoutes 注册等级相关路由
func RegisterLevelRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	// 公开路由
	router.GET("/api/level/configs", GetAllLevelConfigs)
	router.GET("/api/level/leaderboard", GetLevelLeaderboard)
	router.GET("/api/level/user/:uid", GetUserLevelInfo)

	// 需要认证的路由
	authGroup := router.Group("/api/level")
	authGroup.Use(middleware.AuthMiddleware())
	{
		authGroup.GET("/info", GetLevelInfo)
	}
}
