package handlers

import (
	"chemistryuno/repository"
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
