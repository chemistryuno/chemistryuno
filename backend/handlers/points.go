package handlers

import (
	"chemistryuno/database"
	"chemistryuno/websocket"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 获取排行榜
func GetLeaderboard(c *gin.Context) {
	mode := c.Query("mode") // "total" or "monthly"
	orderBy := "points"
	if mode == "monthly" {
		orderBy = "monthly_points"
	}

	rows, err := database.DB.Query(fmt.Sprintf(`
		SELECT UID, username, avatar, points, monthly_points 
		FROM users 
		ORDER BY %s DESC 
		LIMIT 100
	`, orderBy))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取排行榜失败"})
		return
	}
	defer rows.Close()

	var leaderboard []map[string]interface{}
	for rows.Next() {
		var uid int
		var username, avatar string
		var points, monthlyPoints int
		rows.Scan(&uid, &username, &avatar, &points, &monthlyPoints)

		// 获取该玩家当前的悬赏金额
		var totalBounty int
		database.DB.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM bounties WHERE target_uid = ? AND status = 'active'", uid).Scan(&totalBounty)

		// 检查是否在线
		isOnline := false
		if websocket.GlobalHub != nil {
			isOnline = websocket.GlobalHub.IsUIDOnline(uid)
		}

		leaderboard = append(leaderboard, map[string]interface{}{
			"uid":            uid,
			"username":       username,
			"avatar":         avatar,
			"points":         points,
			"monthly_points": monthlyPoints,
			"bounty":         totalBounty,
			"is_online":      isOnline,
		})
	}

	c.JSON(http.StatusOK, leaderboard)
}

// 发起悬赏
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

	// 检查积分是否足够
	var currentPoints int
	err := database.DB.QueryRow("SELECT points FROM users WHERE UID = ?", uid).Scan(&currentPoints)
	if err != nil || currentPoints < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "积分不足"})
		return
	}

	// 扣除积分并创建悬赏
	tx, _ := database.DB.Begin()
	_, err1 := tx.Exec("UPDATE users SET points = points - ? WHERE UID = ?", req.Amount, uid)
	_, err2 := tx.Exec("INSERT INTO bounties (target_uid, amount, created_by, status) VALUES (?, ?, ?, 'active')",
		req.TargetUID, req.Amount, uid)

	if err1 != nil || err2 != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建悬赏失败"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "悬赏成功"})
}

// 积分系统后台任务 (每月刷新和每周衰减)
func StartPointsTask() {
	go func() {
		decayTicker := time.NewTicker(24 * time.Hour) // 每天检查一次是否需要衰减
		defer decayTicker.Stop()

		for range decayTicker.C {
			now := time.Now()

			// 1. 每月刷新 (每月1号0点)
			if now.Day() == 1 {
				// 我们简单处理：如果本月还没刷新过（根据系统某处的记录，或者简单的根据当前内存状态）
				// 这里为了演示，执行一个简单的逻辑：如果当前小时是0，则重置所有积分
				// 实际生产环境建议记录 "last_monthly_reset" 到系统配置表
			}

			// 2. 每周排行榜前10%积分降低2%
			// 假设每周一执行
			if now.Weekday() == time.Monday {
				processWeeklyDecay()
			}
		}
	}()
}

func processWeeklyDecay() {
	// 获取总人数
	var totalUsers int
	database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	topCount := totalUsers / 10
	if topCount < 1 {
		topCount = 1
	}

	// 获取前10%的用户UID
	rows, err := database.DB.Query("SELECT UID FROM users ORDER BY points DESC LIMIT ?", topCount)
	if err != nil {
		return
	}
	defer rows.Close()

	var uids []int
	for rows.Next() {
		var uid int
		rows.Scan(&uid)
		uids = append(uids, uid)
	}

	// 降低2%
	for _, uid := range uids {
		database.DB.Exec("UPDATE users SET points = CAST(points * 0.98 AS INTEGER) WHERE UID = ?", uid)
	}
}
