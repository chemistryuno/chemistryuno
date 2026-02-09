package handlers

import (
	"chemistryuno/backend/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetGlobalChatHistory 获取全服聊天历史
func GetGlobalChatHistory(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit > 200 {
		limit = 200
	}

	messages, err := repository.ChatRepo.GetRecentMessages(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取聊天记录失败"})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// GetPrivateChatHistory 获取与指定用户的私聊历史
func GetPrivateChatHistory(c *gin.Context) {
	uid := c.GetInt("uid")
	friendUIDStr := c.Param("friend_uid")
	friendUID, err := strconv.ParseUint(friendUIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID格式错误"})
		return
	}

	// 检查是否是好友关系
	isFriend, err := repository.FriendshipRepo.IsFriend(uint(uid), uint(friendUID))
	if err != nil || !isFriend {
		c.JSON(http.StatusForbidden, gin.H{"error": "只能查看好友的聊天记录"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit > 200 {
		limit = 200
	}

	privateChatRepo := repository.NewPrivateChatRepository()
	messages, err := privateChatRepo.GetMessagesBetweenUsers(uint(uid), uint(friendUID), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取聊天记录失败"})
		return
	}

	c.JSON(http.StatusOK, messages)
}
