package handlers

import (
	"chemistryuno/repository"
	"chemistryuno/websocket"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SendFriendRequest 发送好友请求
func SendFriendRequest(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		FriendID uint   `json:"friend_id" binding:"required"`
		Message  string `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if uint(uid) == req.FriendID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能添加自己为好友"})
		return
	}

	err := repository.FriendshipRepo.CreateRequest(uint(uid), req.FriendID, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发送请求失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "好友请求已发送"})
}

// GetPendingRequests 获取待处理的好友请求
func GetPendingRequests(c *gin.Context) {
	uid := c.GetInt("uid")
	requests, err := repository.FriendshipRepo.GetPendingRequests(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取请求失败"})
		return
	}

	// 统一返回格式，方便前端展示是谁发来的
	type Response struct {
		ID           uint   `json:"id"`
		UserID       uint   `json:"user_id"`
		Username     string `json:"username"`
		Avatar       string `json:"avatar"`
		HelloMessage string `json:"hello_message"`
	}
	var res []Response
	for _, r := range requests {
		res = append(res, Response{
			ID:           r.ID,
			UserID:       r.UserID,
			Username:     r.User.Username,
			Avatar:       r.User.Avatar,
			HelloMessage: r.HelloMessage,
		})
	}

	c.JSON(http.StatusOK, res)
}

// HandleFriendRequest 处理好友请求 (接受或拒绝)
func HandleFriendRequest(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		RequestID uint   `json:"request_id" binding:"required"`
		Action    string `json:"action" binding:"required"` // accept, decline
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 验证请求是否属于该用户
	f, err := repository.FriendshipRepo.GetFriendshipByID(req.RequestID)
	if err != nil || f.FriendID != uint(uid) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作此请求"})
		return
	}

	status := "declined"
	if req.Action == "accept" {
		status = "accepted"
	}

	err = repository.FriendshipRepo.UpdateStatus(req.RequestID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "处理失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "处理成功"})
}

// GetFriendsList 获取好友列表
func GetFriendsList(c *gin.Context) {
	uid := c.GetInt("uid")
	friends, err := repository.FriendshipRepo.GetFriends(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取好友列表失败"})
		return
	}

	var res []map[string]interface{}
	for _, f := range friends {
		isOnline := false
		if websocket.GlobalHub != nil {
			isOnline = websocket.GlobalHub.IsUIDOnline(int(f.UID))
		}
		res = append(res, map[string]interface{}{
			"uid":       f.UID,
			"username":  f.Username,
			"avatar":    f.Avatar,
			"is_online": isOnline,
		})
	}

	c.JSON(http.StatusOK, res)
}

// DeleteFriend 删除好友
func DeleteFriend(c *gin.Context) {
	uid := c.GetInt("uid")
	friendIDStr := c.Param("id")
	friendID, err := strconv.ParseUint(friendIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的好友ID"})
		return
	}

	err = repository.FriendshipRepo.DeleteFriendship(uint(uid), uint(friendID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除好友失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "好友已删除"})
}
