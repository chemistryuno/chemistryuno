package handlers

import (
	"chemistryuno/backend/repository"
	"chemistryuno/backend/websocket"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SendFriendRequest 发送好友请求
func SendFriendRequest(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		FriendUID uint   `json:"friend_uid" binding:"required"`
		Message   string `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if uint(uid) == req.FriendUID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能添加自己为好友"})
		return
	}

	err := repository.FriendshipRepo.CreateRequest(uint(uid), req.FriendUID, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发送请求失败"})
		return
	}

	// 尝试通知目标用户有新的好友请求
	if websocket.GlobalHub != nil {
		websocket.GlobalHub.SendToUser(int(req.FriendUID), websocket.Message{
			Type: "friend_request",
			Data: map[string]interface{}{
				"from_uid": uid,
			},
		})
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
		UserUID      uint   `json:"user_uid"`
		Username     string `json:"username"`
		Nickname     string `json:"nickname"`
		Avatar       string `json:"avatar"`
		HelloMessage string `json:"hello_message"`
	}
	var res []Response
	for _, r := range requests {
		username := "已注销的用户"
		nickname := ""
		avatar := "🧪"
		if r.User.Username != "" {
			username = r.User.Username
			nickname = r.User.Nickname
			avatar = r.User.Avatar
		}
		res = append(res, Response{
			ID:           r.ID,
			UserUID:      r.UserUID,
			Username:     username,
			Nickname:     nickname,
			Avatar:       avatar,
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
	if err != nil || f.FriendUID != uint(uid) {
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

	// 尝试通过 WebSocket 通知对方
	if websocket.GlobalHub != nil {
		websocket.GlobalHub.SendToUser(int(f.UserUID), websocket.Message{
			Type: "friend_request_handled",
			Data: map[string]interface{}{
				"id":     req.RequestID,
				"action": req.Action,
				"by_uid": uid,
			},
		})
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
			"nickname":  f.Nickname,
			"avatar":    f.Avatar,
			"is_online": isOnline,
		})
	}

	c.JSON(http.StatusOK, res)
}

// DeleteFriend 删除好友
func DeleteFriend(c *gin.Context) {
	uid := c.GetInt("uid")
	friendUIDStr := c.Param("id")
	friendUID, err := strconv.ParseUint(friendUIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的好友ID"})
		return
	}

	err = repository.FriendshipRepo.DeleteFriendship(uint(uid), uint(friendUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除好友失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "好友已删除"})
}

// SetFriendRemark 设置好友备注
func SetFriendRemark(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		FriendUID uint   `json:"friend_uid" binding:"required"`
		Remark    string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 限制备注长度
	if len([]rune(req.Remark)) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "备注不能超过50个字符"})
		return
	}

	err := repository.FriendshipRepo.SetRemark(uint(uid), req.FriendUID, req.Remark)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "设置备注失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "备注已设置"})
}

// GetFriendsListWithRemarks 获取好友列表（包含备注）
func GetFriendsListWithRemarks(c *gin.Context) {
	uid := c.GetInt("uid")
	friends, err := repository.FriendshipRepo.GetFriendsWithRemarks(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取好友列表失败"})
		return
	}

	// 添加在线状态
	for i := range friends {
		isOnline := false
		if websocket.GlobalHub != nil {
			isOnline = websocket.GlobalHub.IsUIDOnline(int(friends[i]["uid"].(uint)))
		}
		friends[i]["is_online"] = isOnline
	}

	c.JSON(http.StatusOK, friends)
}
