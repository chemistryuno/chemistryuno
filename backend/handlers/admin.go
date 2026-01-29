package handlers

import (
	"chemistryuno/database"
	"chemistryuno/models"
	"chemistryuno/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

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

// 获取所有化学反应 (Admin/Co-worker)
func GetReactions(c *gin.Context) {
	// 同一 group_id 的反应只显示一次
	rows, err := database.DB.Query(`
		SELECT MIN(r.id), r.display, r.status, r.group_id, r.created_by, u.username, MIN(r.created_at)
		FROM reactions r
		LEFT JOIN users u ON r.created_by = u.UID
		GROUP BY r.display, r.status, r.group_id, r.created_by, u.username
		ORDER BY MIN(r.created_at) DESC
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
			display, status, groupID, creatorName, createdAt string
		)
		if err := rows.Scan(&id, &display, &status, &groupID, &createdBy, &creatorName, &createdAt); err != nil {
			continue
		}
		reactions = append(reactions, map[string]interface{}{
			"id":           id,
			"display":      display,
			"status":       status,
			"group_id":     groupID,
			"created_by":   createdBy,
			"creator_name": creatorName,
			"created_at":   createdAt,
		})
	}

	c.JSON(http.StatusOK, reactions)
}

// 审核通过并允许编辑方程式 (仅限 Admin)
func ApproveReaction(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	groupID := c.Param("group_id")

	var req struct {
		Display string `json:"display"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的方程式"})
		return
	}

	// 如果提供了 display，说明管理员修改了方程式
	if req.Display != "" {
		// 自动识别新的 r1, r2
		rlist := parseReactants(req.Display)
		if len(rlist) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无法识别修改后的反应物"})
			return
		}

		r1 := rlist[0]
		r2 := r1
		if len(rlist) > 1 {
			r2 = rlist[1]
		}

		tx, err := database.DB.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库事务失败"})
			return
		}

		// 更新 group 中的所有记录
		// 我们需要重新处理双向映射，最简单的方法是删除旧的重新插入
		var creatorID int
		err = tx.QueryRow("SELECT created_by FROM reactions WHERE group_id = ? LIMIT 1", groupID).Scan(&creatorID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查找原作者失败"})
			return
		}

		_, err = tx.Exec("DELETE FROM reactions WHERE group_id = ?", groupID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新旧数据失败"})
			return
		}

		// 重新插入
		_, err = tx.Exec(`
			INSERT INTO reactions (r1, r2, display, status, group_id, created_by)
			VALUES (?, ?, ?, ?, ?, ?)
		`, r1, r2, req.Display, "approved", groupID, creatorID)

		if err == nil && r1 != r2 {
			_, err = tx.Exec(`
				INSERT INTO reactions (r1, r2, display, status, group_id, created_by)
				VALUES (?, ?, ?, ?, ?, ?)
			`, r2, r1, req.Display, "approved", groupID, creatorID)
		}

		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "存入修改后的反应失败"})
			return
		}

		tx.Commit()
		c.JSON(http.StatusOK, gin.H{"message": "修改并已通过审核"})
		return
	}

	// 否则直接通过
	_, err := database.DB.Exec("UPDATE reactions SET status = 'approved' WHERE group_id = ?", groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "审核操作失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "审核通过"})
}

// 解析 display 得到 r1 和 r2
func parseReactants(display string) []string {
	// 找到等号、箭头或波浪号（允许各种常见化学方程式符号）
	sep := "="
	if !strings.Contains(display, "=") {
		if strings.Contains(display, "->") {
			sep = "->"
		} else if strings.Contains(display, "→") {
			sep = "→"
		} else {
			return nil
		}
	}

	parts := strings.Split(display, sep)
	reactantPart := strings.TrimSpace(parts[0])
	reactants := strings.Split(reactantPart, "+")

	var result []string
	// 正则：忽略开头的数字系数（如 2H2O 匹配出 H2O）
	re := regexp.MustCompile(`^\d*(.*)$`)
	for _, r := range reactants {
		trimmed := strings.TrimSpace(r)
		match := re.FindStringSubmatch(trimmed)
		if len(match) > 1 {
			substance := strings.TrimSpace(match[1])
			if substance != "" {
				result = append(result, substance)
			}
		}
	}
	return result
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

	// 自动识别 r1, r2
	rlist := parseReactants(req.Display)
	if len(rlist) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法自动识别反应物，请检查方程式格式（如: A + B = C）"})
		return
	}

	r1 := rlist[0]
	r2 := r1
	if len(rlist) > 1 {
		r2 = rlist[1]
	}

	// 确定状态：admin 直接通过，co-worker 需要审核
	status := "pending"
	if role == "admin" {
		status = "approved"
	}

	// 生成 group_id
	groupID := fmt.Sprintf("%d-%d", uid, time.Now().UnixNano())

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库事务开启失败"})
		return
	}

	// 插入 r1 -> r2
	_, err = tx.Exec(`
		INSERT INTO reactions (r1, r2, display, status, group_id, created_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`, r1, r2, req.Display, status, groupID, uid)

	if err == nil && r1 != r2 {
		// 插入 r2 -> r1 (双向映射)
		_, err = tx.Exec(`
			INSERT INTO reactions (r1, r2, display, status, group_id, created_by)
			VALUES (?, ?, ?, ?, ?, ?)
		`, r2, r1, req.Display, status, groupID, uid)
	}

	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加反应失败"})
		return
	}

	tx.Commit()
	msg := "反应已提交，等待管理员审核"
	if status == "approved" {
		msg = "反应已成功加入核心数据库"
	}
	c.JSON(http.StatusOK, gin.H{"message": msg})
}

// 删除或拒绝化学反应 (Admin/Co-worker可以删除自己的)
func DeleteReaction(c *gin.Context) {
	reactionID := c.Param("id") // 接收 id，根据 id 查出 group_id 后删全组
	role := c.GetString("role")
	uid := c.GetInt("uid")

	var groupID string
	var createdBy int
	err := database.DB.QueryRow("SELECT group_id, created_by FROM reactions WHERE id = ?", reactionID).Scan(&groupID, &createdBy)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到该反应"})
		return
	}

	// 检查权限：admin 或者 提交者本人
	if role != "admin" && uid != createdBy {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	_, err = database.DB.Exec("DELETE FROM reactions WHERE group_id = ?", groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已从实验室档案中抹除"})
}
