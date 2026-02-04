package handlers

import (
	"chemistryuno/repository"
	"chemistryuno/websocket"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetLeaderboard(c *gin.Context) {
	mode := c.Query("mode")
	orderBy := "points"
	if mode == "monthly" {
		orderBy = "monthly_points"
	}

	users, err := repository.UserRepo.GetLeaderboard(orderBy, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取排行榜失败"})
		return
	}

	var leaderboard []map[string]interface{}
	for _, user := range users {
		totalBounty, _ := repository.BountyRepo.GetTotalBounty(user.UID)
		isOnline := false
		if websocket.GlobalHub != nil {
			isOnline = websocket.GlobalHub.IsUIDOnline(int(user.UID))
		}

		leaderboard = append(leaderboard, map[string]interface{}{
			"uid":            user.UID,
			"username":       user.Username,
			"avatar":         user.Avatar,
			"points":         user.Points,
			"monthly_points": user.MonthlyPoints,
			"bounty":         totalBounty,
			"is_online":      isOnline,
		})
	}

	c.JSON(http.StatusOK, leaderboard)
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

	user, err := repository.UserRepo.FindByID(uint(uid))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	if user.Points < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "积分不足"})
		return
	}

	// 扣除积分
	err = repository.UserRepo.DeductPoints(uint(uid), req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "扣除积分失败"})
		return
	}

	// 创建悬赏
	bounty := &repository.Bounty{
		TargetUID: uint(req.TargetUID),
		Amount:    req.Amount,
		CreatedBy: uint(uid),
		Status:    "active",
	}
	err = repository.BountyRepo.Create(bounty)
	if err != nil {
		// 如果创建失败，回退积分
		_ = repository.UserRepo.AddPoints(uint(uid), req.Amount)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "设置悬赏失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "悬赏已设置"})
}
