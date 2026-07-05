package handlers

import (
	"chemistryuno/backend/cache"
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/websocket"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var errInsufficientPoints = errors.New("insufficient points")

func GetLeaderboard(c *gin.Context) {
	uid := c.GetInt("uid")
	mode := c.Query("mode")
	orderBy := "points"
	if mode == "monthly" {
		orderBy = "monthly_points"
	}

	// 使用缓存获取排行榜
	leaderboardData, err := repository.GetLeaderboardWithCache(c.Request.Context(), orderBy, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取排行榜失败"})
		return
	}

	// 预先获取等级配置以避免循环查询
	levelConfigs, err := repository.GetLevelConfigsWithCache(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取排行榜失败"})
		return
	}
	configMap := make(map[int]database.LevelConfig)
	for _, conf := range levelConfigs {
		configMap[conf.Level] = conf
	}

	var leaderboard []map[string]interface{}
	foundSelf := false
	for _, entry := range leaderboardData {
		if int(entry.UID) == uid {
			foundSelf = true
		}

		conf := configMap[entry.Level]
		totalBounty, _ := repository.GetBountyTotalWithCache(c.Request.Context(), entry.UID)
		isOnline := false
		if websocket.GlobalHub != nil {
			isOnline = websocket.GlobalHub.IsUIDOnline(int(entry.UID))
		}

		// 获取用户完整信息以检查封禁状态
		user, _ := repository.UserRepo.FindByUID(entry.UID)
		isBanned := false
		if user != nil {
			isBanned = user.BannedUntil != nil && time.Now().Before(*user.BannedUntil)
		}

		leaderboard = append(leaderboard, map[string]interface{}{
			"uid":             entry.UID,
			"username":        entry.Username,
			"nickname":        entry.Nickname,
			"avatar":          entry.Avatar,
			"points":          entry.Points,
			"monthly_points":  entry.MonthlyPoints,
			"level":           entry.Level,
			"tier":            conf.Tier,
			"tier_name":       conf.TierName,
			"win_count":       entry.WinCount,
			"total_games":     entry.TotalGames,
			"bounty":          totalBounty,
			"is_online":       isOnline,
			"is_banned":       isBanned,
			"last_offline_at": entry.LastOfflineAt,
		})
	}

	// 如果自己不在前100名中，则单独获取并追加
	var selfInfo map[string]interface{}
	if !foundSelf && uid > 0 {
		var user database.User
		err := database.DB.Where("uid = ?", uid).First(&user).Error
		if err == nil {
			// 计算排名：优先 ZREVRANK，fallback 到 SQL COUNT
			var rank int64
			zrank, zErr := cache.ZREVRANKLeaderboard(c.Request.Context(), orderBy, user.UID)
			if zErr == nil && zrank > 0 {
				rank = zrank - 1
			} else {
				score := user.Points
				if mode == "monthly" {
					score = user.MonthlyPoints
				}
				if err := database.DB.Model(&database.User{}).Where(orderBy+" > ?", score).Count(&rank).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "获取排行榜失败"})
					return
				}
			}

			conf := configMap[user.Level]
			totalBounty, _ := repository.GetBountyTotalWithCache(c.Request.Context(), user.UID)
			isOnline := false
			if websocket.GlobalHub != nil {
				isOnline = websocket.GlobalHub.IsUIDOnline(int(user.UID))
			}
			isBanned := user.BannedUntil != nil && time.Now().Before(*user.BannedUntil)

			selfInfo = map[string]interface{}{
				"uid":             user.UID,
				"username":        user.Username,
				"nickname":        user.Nickname,
				"avatar":          user.Avatar,
				"points":          user.Points,
				"monthly_points":  user.MonthlyPoints,
				"level":           user.Level,
				"tier":            conf.Tier,
				"tier_name":       conf.TierName,
				"win_count":       user.WinCount,
				"total_games":     user.TotalGames,
				"bounty":          totalBounty,
				"is_online":       isOnline,
				"is_banned":       isBanned,
				"last_offline_at": user.LastOfflineAt,
				"rank":            rank + 1,
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": leaderboard,
		"self":        selfInfo,
	})
}

func CreateBounty(c *gin.Context) {
	var req struct {
		TargetUID int `json:"target_uid" binding:"required"`
		Amount    int `json:"amount" binding:"required,min=100"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	uid := c.GetInt("uid")
	if uid == req.TargetUID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能给自己设置悬赏"})
		return
	}

	_, err := repository.UserRepo.FindByUID(uint(uid))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		updateResult := tx.Model(&database.User{}).
			Where("uid = ? AND points >= ?", uid, req.Amount).
			Updates(map[string]interface{}{
				"points":         gorm.Expr("points - ?", req.Amount),
				"monthly_points": gorm.Expr("monthly_points - ?", req.Amount),
			})
		if updateResult.Error != nil {
			return updateResult.Error
		}
		if updateResult.RowsAffected == 0 {
			return errInsufficientPoints
		}

		bounty := &database.Bounty{
			TargetUID: uint(req.TargetUID),
			Amount:    req.Amount,
			IssuerUID: uint(uid),
			Status:    "active",
		}
		return tx.Create(bounty).Error
	})
	if err != nil {
		if errors.Is(err, errInsufficientPoints) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "积分不足"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "设置悬赏失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "悬赏已设置"})
}
