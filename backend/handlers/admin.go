package handlers

import (
	"chemistryuno/database"
	"chemistryuno/models"
	"chemistryuno/utils"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 管理员创建用户
func CreateUser(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户名是否已存在
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", req.Username).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
		return
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 插入用户
	result, err := database.DB.Exec("INSERT INTO users (username, password, avatar, role) VALUES (?, ?, ?, ?)",
		req.Username, hashedPassword, "🧪", "user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	userUID, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{
		"message": "用户创建成功",
		"uid":     userUID,
	})
}

// 获取所有用户
func GetAllUsers(c *gin.Context) {
	rows, err := database.DB.Query(
		"SELECT UID, username, avatar, is_admin, role, created_at FROM users ORDER BY created_at DESC",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.UID, &user.Username, &user.Avatar, &user.IsAdmin, &user.Role, &user.CreatedAt); err != nil {
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

	_, err := database.DB.Exec("DELETE FROM users WHERE UID = ? AND is_admin = 0", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户已删除"})
}

// 管理员修改用户密码
func AdminChangePassword(c *gin.Context) {
	userID := c.Param("id")

	var req models.AdminChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户是否存在且不是管理员
	var isAdmin bool
	err := database.DB.QueryRow("SELECT is_admin FROM users WHERE UID = ?", userID).Scan(&isAdmin)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	if isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "无法修改管理员密码"})
		return
	}

	// 更新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	_, err = database.DB.Exec("UPDATE users SET password = ? WHERE UID = ?", hashedPassword, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码修改失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// 提升用户权限
func PromoteUser(c *gin.Context) {
	userID := c.Param("id")

	var req models.PromoteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新用户角色
	isAdmin := req.Role == "admin"
	_, err := database.DB.Exec("UPDATE users SET role = ?, is_admin = ? WHERE UID = ?", req.Role, isAdmin, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "权限修改失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户权限已更新"})
}

// 获取游戏历史
func GetGameHistory(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT gh.id, gh.room_id, gh.winner_uid, u.username, gh.players, gh.started_at, gh.finished_at
		FROM game_history gh
		LEFT JOIN users u ON gh.winner_uid = u.UID
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
			id, winnerUID                   int
			roomID, winnerName, playersJSON string
			startedAt, finishedAt           string
		)
		if err := rows.Scan(&id, &roomID, &winnerUID, &winnerName, &playersJSON, &startedAt, &finishedAt); err != nil {
			continue
		}
		history = append(history, map[string]interface{}{
			"id":          id,
			"room_id":     roomID,
			"winner_uid":  winnerUID,
			"winner_name": winnerName,
			"players":     playersJSON,
			"started_at":  startedAt,
			"finished_at": finishedAt,
		})
	}

	c.JSON(http.StatusOK, history)
}

// 获取所有化学反应
func GetReactions(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT r.id, r.reactant, r.product, r.type, r.created_by, u.username, r.created_at
		FROM reactions r
		LEFT JOIN users u ON r.created_by = u.UID
		ORDER BY r.created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	defer rows.Close()

	var reactions []map[string]interface{}
	for rows.Next() {
		var (
			id, createdBy                                    int
			reactant, product, rType, creatorName, createdAt string
		)
		if err := rows.Scan(&id, &reactant, &product, &rType, &createdBy, &creatorName, &createdAt); err != nil {
			continue
		}
		reactions = append(reactions, map[string]interface{}{
			"id":           id,
			"reactant":     reactant,
			"product":      product,
			"type":         rType,
			"created_by":   createdBy,
			"creator_name": creatorName,
			"created_at":   createdAt,
		})
	}

	c.JSON(http.StatusOK, reactions)
}

// 添加化学反应（co-worker或admin权限）
func AddReaction(c *gin.Context) {
	uid := c.GetInt("uid")
	role := c.GetString("role")

	// 检查权限
	if role != "admin" && role != "co-worker" {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	var req models.ReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := database.DB.Exec(`
		INSERT INTO reactions (reactant, product, type, created_by)
		VALUES (?, ?, ?, ?)
	`, req.Reactant, req.Product, req.Type, uid)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加反应失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "反应添加成功"})
}

// 删除化学反应（co-worker或admin权限）
func DeleteReaction(c *gin.Context) {
	reactionID := c.Param("id")
	role := c.GetString("role")

	// 检查权限
	if role != "admin" && role != "co-worker" {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	_, err := database.DB.Exec("DELETE FROM reactions WHERE id = ?", reactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "反应已删除"})
}
