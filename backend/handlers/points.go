package handlers

import (
	"chemistryuno/database"
	"chemistryuno/websocket"
	"fmt"
	"net/http"

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
