package handlers

import (
	"chemistryuno/database"
	"chemistryuno/models"
	"chemistryuno/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
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
		ORDER BY 
			CASE 
				WHEN r.status = 'pending_admin' THEN 1 
				WHEN r.status = 'pending_coworker' THEN 2 
				ELSE 3 
			END, 
			MIN(r.created_at) DESC
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

// 获取所有已批准的反应 (Wiki)
func GetAllReactions(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT id, display, r1, r2, created_at
		FROM reactions
		WHERE status = 'approved'
		ORDER BY created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	defer rows.Close()

	var reactions []map[string]interface{}
	for rows.Next() {
		var (
			id                         int
			display, r1, r2, createdAt string
		)
		if err := rows.Scan(&id, &display, &r1, &r2, &createdAt); err != nil {
			continue
		}
		reactions = append(reactions, map[string]interface{}{
			"id":         id,
			"display":    display,
			"r1":         r1,
			"r2":         r2,
			"created_at": createdAt,
		})
	}
	c.JSON(http.StatusOK, reactions)
}

// 获取我提交的反应
func GetMyReactions(c *gin.Context) {
	uid := c.GetInt("uid")
	rows, err := database.DB.Query(`
		SELECT id, display, status, created_at
		FROM reactions
		WHERE created_by = ?
		ORDER BY created_at DESC
	`, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	defer rows.Close()

	var reactions []map[string]interface{}
	for rows.Next() {
		var (
			id                         int
			display, status, createdAt string
		)
		if err := rows.Scan(&id, &display, &status, &createdAt); err != nil {
			continue
		}
		reactions = append(reactions, map[string]interface{}{
			"id":         id,
			"display":    display,
			"status":     status,
			"created_at": createdAt,
		})
	}
	c.JSON(http.StatusOK, reactions)
}

// 审核通过并允许编辑方程式 (仅限 Admin)
func ApproveReaction(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" && role != "co-worker" {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要协作者或管理员权限"})
		return
	}

	groupID := c.Param("group_id")

	var req struct {
		Display string `json:"display"`
		Reject  bool   `json:"reject"` // 拒绝功能
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}

	// 检查当前状态是否符合审批流
	var currentStatus string
	err := database.DB.QueryRow("SELECT status FROM reactions WHERE group_id = ? LIMIT 1", groupID).Scan(&currentStatus)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到该反应请求"})
		return
	}

	// 审批流校验
	newStatus := ""
	if req.Reject {
		newStatus = "rejected"
	} else {
		if role == "co-worker" {
			if currentStatus != "pending_coworker" {
				c.JSON(http.StatusForbidden, gin.H{"error": "当前状态不符合协作者审批权限"})
				return
			}
			newStatus = "pending_admin"
		} else if role == "admin" {
			// 管理员可以审批 pending_admin 或者是跳过协作者审批 pending_coworker
			if currentStatus != "pending_admin" && currentStatus != "pending_coworker" {
				c.JSON(http.StatusForbidden, gin.H{"error": "该反应不需要管理员再审批"})
				return
			}
			newStatus = "approved"
		}
	}

	// 如果提供了 display，且管理员/协作者修改了方程式
	if !req.Reject && req.Display != "" {
		// ... 保持原有的修改逻辑，但使用 newStatus ...
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

		_, err = tx.Exec(`
			INSERT INTO reactions (r1, r2, display, status, group_id, created_by)
			VALUES (?, ?, ?, ?, ?, ?)
		`, r1, r2, req.Display, newStatus, groupID, creatorID)

		if err == nil && r1 != r2 {
			_, err = tx.Exec(`
				INSERT INTO reactions (r1, r2, display, status, group_id, created_by)
				VALUES (?, ?, ?, ?, ?, ?)
			`, r2, r1, req.Display, newStatus, groupID, creatorID)
		}

		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "处理审批事物失败"})
			return
		}
		tx.Commit()
		c.JSON(http.StatusOK, gin.H{"message": "审批已处理", "status": newStatus})
		return
	}

	// 如果没有修改内容，直接更新状态
	_, err = database.DB.Exec("UPDATE reactions SET status = ? WHERE group_id = ?", newStatus, groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "状态已更新", "status": newStatus})
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

	var req models.ReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 自动识别 r1, r2
	rlist := parseReactants(req.Display)
	if len(rlist) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法自动识别反应物，请检查方程式格式（如: A + B -> C）"})
		return
	}

	// 1. 校验质量守恒
	if ok, errInfo := validateBalance(req.Display); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "方程式校验失败: " + errInfo})
		return
	}

	// 2. 校验反应物重复
	if isDup, oldDisplay := checkDuplicateReactants(req.Display); isDup {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("该反应物组合已存在: %s", oldDisplay)})
		return
	}

	r1 := rlist[0]
	r2 := r1
	if len(rlist) > 1 {
		r2 = rlist[1]
	}

	// 确定状态：
	// admin: 直接 approved
	// co-worker: pending_admin
	// user: pending_coworker
	status := "pending_coworker"
	if role == "admin" {
		status = "approved"
	} else if role == "co-worker" {
		status = "pending_admin"
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

// 内部校验逻辑
func normalizeSubscripts(s string) string {
	subs := map[rune]rune{
		'₀': '0', '₁': '1', '₂': '2', '₃': '3', '₄': '4',
		'₅': '5', '₆': '6', '₇': '7', '₈': '8', '₉': '9',
	}
	var res strings.Builder
	for _, r := range s {
		if v, ok := subs[r]; ok {
			res.WriteRune(v)
		} else {
			res.WriteRune(r)
		}
	}
	return res.String()
}

func countElements(expr string) map[string]int {
	counts := make(map[string]int)
	expr = normalizeSubscripts(expr)
	parts := strings.Split(expr, "+")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		reCoeff := regexp.MustCompile(`^(\d+)?(.*)$`)
		matches := reCoeff.FindStringSubmatch(part)
		coeff := 1
		if matches[1] != "" {
			coeff, _ = strconv.Atoi(matches[1])
		}
		substance := matches[2]
		reElem := regexp.MustCompile(`([A-Z][a-z]*)(\d*)`)
		elemMatches := reElem.FindAllStringSubmatch(substance, -1)
		for _, m := range elemMatches {
			element := m[1]
			count := 1
			if m[2] != "" {
				count, _ = strconv.Atoi(m[2])
			}
			counts[element] += count * coeff
		}
	}
	return counts
}

func validateBalance(display string) (bool, string) {
	cleanDisplay := strings.ReplaceAll(display, "＝", "->")
	cleanDisplay = strings.ReplaceAll(cleanDisplay, "=", "->")
	parts := strings.Split(cleanDisplay, "->")
	if len(parts) != 2 {
		return false, "方程式格式错误，应包含 '->' 或 '='"
	}
	leftCounts := countElements(parts[0])
	rightCounts := countElements(parts[1])
	allElements := make(map[string]bool)
	for k := range leftCounts {
		allElements[k] = true
	}
	for k := range rightCounts {
		allElements[k] = true
	}
	for el := range allElements {
		if leftCounts[el] != rightCounts[el] {
			return false, fmt.Sprintf("元素 %s 不守恒 (左:%d, 右:%d)", el, leftCounts[el], rightCounts[el])
		}
	}
	return true, ""
}

func checkDuplicateReactants(display string) (bool, string) {
	rList := parseReactants(display)
	if len(rList) < 2 {
		return false, ""
	}
	sort.Strings(rList)
	r1, r2 := rList[0], rList[1]

	var existingDisplay string
	err := database.DB.QueryRow(`
		SELECT display FROM reactions 
		WHERE status = 'approved' AND ((r1 = ? AND r2 = ?) OR (r1 = ? AND r2 = ?))
		LIMIT 1
	`, r1, r2, r2, r1).Scan(&existingDisplay)

	if err == nil {
		return true, existingDisplay
	}
	return false, ""
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

// 直接修改化学反应 (Admin/Co-worker)
func UpdateReaction(c *gin.Context) {
	reactionID := c.Param("id")
	role := c.GetString("role")
	uid := c.GetInt("uid")

	if role != "admin" && role != "co-worker" {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	var req models.ReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 校验
	if ok, errInfo := validateBalance(req.Display); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "方程式守恒校验失败: " + errInfo})
		return
	}

	rlist := parseReactants(req.Display)
	if len(rlist) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法识别反应物"})
		return
	}
	r1 := rlist[0]
	r2 := r1
	if len(rlist) > 1 {
		r2 = rlist[1]
	}

	var groupID string
	err := database.DB.QueryRow("SELECT group_id FROM reactions WHERE id = ?", reactionID).Scan(&groupID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到原反应记录"})
		return
	}

	tx, _ := database.DB.Begin()
	_, err = tx.Exec("DELETE FROM reactions WHERE group_id = ?", groupID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新过程事务异常"})
		return
	}

	// 统一设为 approved，因为是管理员或协作者直接修改
	_, err = tx.Exec(`
		INSERT INTO reactions (r1, r2, display, status, group_id, created_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`, r1, r2, req.Display, "approved", groupID, uid)
	if err == nil && r1 != r2 {
		_, err = tx.Exec(`
			INSERT INTO reactions (r1, r2, display, status, group_id, created_by)
			VALUES (?, ?, ?, ?, ?, ?)
		`, r2, r1, req.Display, "approved", groupID, uid)
	}

	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "插入新记录失败"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "方程式已直接修改并生效"})
}

// 批量导入化学反应 (JSON数组)
func BatchAddReactions(c *gin.Context) {
	uid := c.GetInt("uid")
	role := c.GetString("role")

	if role != "admin" && role != "co-worker" {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	var reqs []models.ReactionRequest
	if err := c.ShouldBindJSON(&reqs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的JSON数组或格式错误"})
		return
	}

	status := "pending"
	if role == "admin" {
		status = "approved"
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "事务开启失败"})
		return
	}

	successCount := 0
	for _, req := range reqs {
		rlist := parseReactants(req.Display)
		if len(rlist) == 0 {
			continue
		}

		// 批量导入也进行基础校验，不满足的跳过
		if ok, _ := validateBalance(req.Display); !ok {
			continue
		}
		if isDup, _ := checkDuplicateReactants(req.Display); isDup {
			continue
		}

		r1 := rlist[0]
		r2 := r1
		if len(rlist) > 1 {
			r2 = rlist[1]
		}

		groupID := fmt.Sprintf("%d-%d-%d", uid, time.Now().UnixNano(), successCount)

		_, err = tx.Exec(`
			INSERT INTO reactions (r1, r2, display, status, group_id, created_by)
			VALUES (?, ?, ?, ?, ?, ?)
		`, r1, r2, req.Display, status, groupID, uid)

		if err == nil && r1 != r2 {
			_, err = tx.Exec(`
				INSERT INTO reactions (r1, r2, display, status, group_id, created_by)
				VALUES (?, ?, ?, ?, ?, ?)
			`, r2, r1, req.Display, status, groupID, uid)
		}

		if err == nil {
			successCount++
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量提交失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("成功导入 %d 条反应", successCount),
		"count":   successCount,
	})
}
