package handlers

import (
	"chemistryuno/database"
	"chemistryuno/models"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 获取所有用户
func GetAllUsers(c *gin.Context) {
	rows, err := database.DB.Query(
		"SELECT id, username, avatar, is_admin, created_at FROM users ORDER BY created_at DESC",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Avatar, &user.IsAdmin, &user.CreatedAt); err != nil {
			continue
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}

// 获取全局牌组配置
func GetGlobalDeckConfig(c *gin.Context) {
	var config models.DeckConfig
	var cardsJSON string

	err := database.DB.QueryRow(
		"SELECT id, name, is_global, cards, created_by, created_at FROM deck_configs WHERE is_global = 1 LIMIT 1",
	).Scan(&config.ID, &config.Name, &config.IsGlobal, &cardsJSON, &config.CreatedBy, &config.CreatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "全局配置不存在"})
		return
	}

	json.Unmarshal([]byte(cardsJSON), &config.Cards)
	c.JSON(http.StatusOK, config)
}

// 更新全局牌组配置
func UpdateGlobalDeckConfig(c *gin.Context) {
	var req struct {
		Name  string         `json:"name" binding:"required"`
		Cards map[string]int `json:"cards" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cardsJSON, _ := json.Marshal(req.Cards)

	_, err := database.DB.Exec(
		"UPDATE deck_configs SET name = ?, cards = ? WHERE is_global = 1",
		req.Name, string(cardsJSON),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "全局牌组配置更新成功"})
}

// 删除用户（管理员）
func DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	_, err := database.DB.Exec("DELETE FROM users WHERE id = ? AND is_admin = 0", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户已删除"})
}

// 获取游戏历史
func GetGameHistory(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT gh.id, gh.room_id, gh.winner_id, u.username, gh.players, gh.started_at, gh.finished_at
		FROM game_history gh
		LEFT JOIN users u ON gh.winner_id = u.id
		ORDER BY gh.finished_at DESC
		LIMIT 100
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	defer rows.Close()

	var history []map[string]interface{}
	for rows.Next() {
		var (
			id, winnerID                     int
			roomID, winnerName, playersJSON  string
			startedAt, finishedAt            string
		)
		if err := rows.Scan(&id, &roomID, &winnerID, &winnerName, &playersJSON, &startedAt, &finishedAt); err != nil {
			continue
		}
		history = append(history, map[string]interface{}{
			"id":          id,
			"room_id":     roomID,
			"winner_id":   winnerID,
			"winner_name": winnerName,
			"players":     playersJSON,
			"started_at":  startedAt,
			"finished_at": finishedAt,
		})
	}

	c.JSON(http.StatusOK, history)
}
