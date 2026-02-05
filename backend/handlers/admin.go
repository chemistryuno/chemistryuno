package handlers

import (
	"chemistryuno/database"
	"chemistryuno/game"
	"chemistryuno/models"
	"chemistryuno/repository"
	"chemistryuno/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	userRepo      *repository.UserRepository
	reactionRepo  *repository.ReactionRepository
	substanceRepo *repository.SubstanceRepository
	gameRepo      *repository.GameRepository
	deckRepo      *repository.DeckRepository
)

// InitAdminHandlers 初始化admin handlers的依赖
func InitAdminHandlers() {
	// 初始化Repository
	userRepo = repository.NewUserRepository()
	reactionRepo = repository.NewReactionRepository()
	substanceRepo = repository.NewSubstanceRepository()
	gameRepo = repository.NewGameRepository()
	deckRepo = repository.NewDeckRepository()
}

// 管理员创建用户
func CreateUser(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户名是否已存在
	exists, err := userRepo.ExistsByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
		return
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 创建用户
	user := &database.User{
		Username: req.Username,
		Password: hashedPassword,
		Avatar:   "🧪",
		Role:     "user",
	}

	err = userRepo.Create(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "用户创建成功",
		"uid":     user.UID,
	})
}

// 获取所有用户
func GetAllUsers(c *gin.Context) {
	users, err := userRepo.GetAllUsersOrderByCreatedAt()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// 管理员踢出玩家
func KickPlayer(c *gin.Context) {
	var req struct {
		RoomID    string `json:"room_id" binding:"required"`
		TargetUID int    `json:"target_uid" binding:"required"`
		Reason    string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if req.Reason == "" {
		req.Reason = "你由于违规游戏而被踢出"
	}

	err := game.AdminKickPlayer(req.RoomID, req.TargetUID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "玩家已踢出"})
}

// 管理员封禁用户
func BanUser(c *gin.Context) {
	var req struct {
		TargetUID int    `json:"target_uid" binding:"required"`
		Hours     int    `json:"hours"` // 0 为永久
		Reason    string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if req.Reason == "" {
		req.Reason = "你由于违规游戏而被封禁"
	}

	var bannedUntil *time.Time
	if req.Hours > 0 {
		t := time.Now().Add(time.Duration(req.Hours) * time.Hour)
		bannedUntil = &t
	} else if req.Hours == -1 {
		// 实际上前端可以传-1表示永久
		t := time.Now().AddDate(100, 0, 0) // 100年后，相当于永久
		bannedUntil = &t
	}

	err := userRepo.UpdateBanStatusWithReason(uint(req.TargetUID), bannedUntil, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "封禁失败"})
		return
	}

	// 封禁后强制登出
	sessionRepo := repository.NewSessionRepository()
	_ = sessionRepo.DeleteByUserID(uint(req.TargetUID))

	msg := "用户已被永久封禁"
	if req.Hours > 0 {
		msg = fmt.Sprintf("用户已被封禁 %d 小时", req.Hours)
	}

	c.JSON(http.StatusOK, gin.H{"message": msg})
}

// 获取全局牌组配置
func GetGlobalDeckConfig(c *gin.Context) {
	deck, err := deckRepo.FindGlobalDeck()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "全局配置不存在"})
		return
	}

	// 解析Cards JSON
	var cards map[string]int
	json.Unmarshal([]byte(deck.Cards), &cards)

	config := models.DeckConfig{
		ID:           int(deck.ID),
		Name:         deck.Name,
		IsGlobal:     deck.IsGlobal,
		Cards:        cards,
		InitialCards: deck.InitialCards,
		CreatedBy:    int(deck.CreatedBy),
		CreatedAt:    deck.CreatedAt,
	}

	c.JSON(http.StatusOK, config)
}

// 更新全局牌组配置
func UpdateGlobalDeckConfig(c *gin.Context) {
	var req struct {
		Name         string         `json:"name" binding:"required"`
		Cards        map[string]int `json:"cards" binding:"required"`
		InitialCards int            `json:"initial_cards"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.InitialCards <= 0 {
		req.InitialCards = 10
	}

	cardsJSON, _ := json.Marshal(req.Cards)

	// 更新数据库
	deck, err := deckRepo.FindGlobalDeck()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "全局配置不存在"})
		return
	}

	deck.Name = req.Name
	deck.Cards = cardsJSON
	deck.InitialCards = req.InitialCards

	err = deckRepo.Update(deck)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "全局牌组配置更新成功"})
}

// 删除用户（管理员）
func DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	uid, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	err = userRepo.DeleteNonAdmin(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户已删除"})
}

// 管理员修改用户密码
func AdminChangePassword(c *gin.Context) {
	userID := c.Param("id")
	uid, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req models.AdminChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户是否存在且不是管理员
	isAdmin, err := userRepo.FindIsAdminByID(uint(uid))
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

	err = userRepo.UpdatePassword(uint(uid), hashedPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码修改失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// 提升用户权限
func PromoteUser(c *gin.Context) {
	userID := c.Param("id")
	uid, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req models.PromoteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新用户角色
	isAdmin := req.Role == "admin"
	err = userRepo.UpdateRole(uint(uid), req.Role, isAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "权限修改失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户权限已更新"})
}

// 获取游戏历史
func GetGameHistory(c *gin.Context) {
	historyList, err := gameRepo.FindAllWithWinner(100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误: " + err.Error()})
		return
	}

	var history []map[string]interface{}
	for _, h := range historyList {
		// 解析玩家列表 JSON
		var players []int
		if err := json.Unmarshal([]byte(h.Players), &players); err != nil {
			log.Printf("解析玩家列表失败: %v", err)
			players = []int{}
		}

		history = append(history, map[string]interface{}{
			"id":          h.ID,
			"room_id":     h.RoomID,
			"winner_uid":  h.WinnerUID,
			"winner_name": h.WinnerName,
			"players":     players,
			"started_at":  h.StartedAt,
			"finished_at": h.FinishedAt,
			"created_at":  h.FinishedAt, // 兼容前端字段名
		})
	}

	c.JSON(http.StatusOK, history)
}

// 获取所有化学反应 (Admin/Co-worker)
func GetReactions(c *gin.Context) {
	reactionList, err := reactionRepo.FindAllGroupedWithCreator()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}

	var reactions []map[string]interface{}
	for _, r := range reactionList {
		reactions = append(reactions, map[string]interface{}{
			"id":           r.ID,
			"display":      r.Display,
			"status":       r.Status,
			"group_id":     r.GroupID,
			"created_by":   r.CreatedBy,
			"creator_name": r.CreatorName,
			"created_at":   r.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, reactions)
}

// 获取所有已批准的反应 (Wiki)
func GetAllReactions(c *gin.Context) {
	reactionList, err := reactionRepo.FindApprovedGrouped()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}

	c.JSON(http.StatusOK, reactionList)
}

// 获取我提交的反应
func GetMyReactions(c *gin.Context) {
	uid := c.GetInt("uid")
	reactionList, err := reactionRepo.FindMyReactions(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}

	var reactions []map[string]interface{}
	for _, r := range reactionList {
		reactions = append(reactions, map[string]interface{}{
			"id":         r.ID,
			"display":    r.Display,
			"status":     r.Status,
			"created_at": r.CreatedAt,
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

	groupIDStr := c.Param("group_id")
	groupIDUint, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的group_id"})
		return
	}
	groupID := uint(groupIDUint)

	var req struct {
		Display string `json:"display"`
		Reject  bool   `json:"reject"` // 拒绝功能
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}

	// 检查当前状态是否符合审批流
	currentStatus, err := reactionRepo.GetStatusByGroupID(groupID)
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
		rlist := parseReactants(req.Display)
		if len(rlist) != 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "修改后的方程式必须包含且仅包含两种不同的反应物"})
			return
		}

		// 增加冲突校验
		if isDup, oldDisplay := checkDuplicateReactants(req.Display, &groupID); isDup {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("修改后的组合已存在: %s", oldDisplay)})
			return
		}

		err = database.DB.Transaction(func(tx *gorm.DB) error {
			// 获取创建者ID
			var reaction database.Reaction
			if err := tx.Where("group_id = ?", groupID).First(&reaction).Error; err != nil {
				return fmt.Errorf("查找原作者失败")
			}

			// 删除旧记录
			if err := tx.Where("group_id = ?", groupID).Delete(&database.Reaction{}).Error; err != nil {
				return err
			}

			// 创建新反应记录
			if err := saveReactionToDBGorm(tx, rlist, req.Display, newStatus, &groupID, reaction.CreatedBy); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "处理审批事务失败: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "审批已处理", "status": newStatus})
		return
	}

	// 如果没有修改内容，直接更新状态
	err = reactionRepo.UpdateStatusByGroupID(groupID, newStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "状态已更新", "status": newStatus})
}

// 解析 display 得到 r1 和 r2
func parseReactants(display string) []string {
	// 找到等号、箭头或波浪号（支持多种中英文化学方程式符号）
	sep := ""
	if strings.Contains(display, "=") {
		sep = "="
	} else if strings.Contains(display, "＝") {
		sep = "＝"
	} else if strings.Contains(display, "->") {
		sep = "->"
	} else if strings.Contains(display, "→") {
		sep = "→"
	}

	if sep == "" {
		return nil
	}

	parts := strings.Split(display, sep)
	reactantPart := strings.TrimSpace(parts[0])
	reactants := strings.Split(reactantPart, "+")

	var result []string
	unique := make(map[string]bool)
	// 正则：忽略开头的数字系数（如 2H2O 匹配出 H2O）
	re := regexp.MustCompile(`^\d*(.*)$`)
	for _, r := range reactants {
		trimmed := strings.TrimSpace(r)
		match := re.FindStringSubmatch(trimmed)
		if len(match) > 1 {
			substance := strings.TrimSpace(match[1])
			if substance != "" && !unique[substance] {
				unique[substance] = true
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
	if isDup, oldDisplay := checkDuplicateReactants(req.Display, nil); isDup {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("该反应物组合已存在: %s", oldDisplay)})
		return
	}

	// 3. 严格校验反应物数量（仅支持双反应物）
	if len(rlist) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "当前仅支持由两种不同物质组成的反应组合（如 A + B = C）"})
		return
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
	groupIDVal := uint(time.Now().UnixNano() / 1000000)
	groupID := &groupIDVal

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		return saveReactionToDBGorm(tx, rlist, req.Display, status, groupID, uint(uid))
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存反应失败: " + err.Error()})
		return
	}

	msg := "反应已提交，等待管理员审核"
	if status == "approved" {
		msg = "反应已成功加入核心数据库"
	}
	c.JSON(http.StatusOK, gin.H{"message": msg})
}

// 统一保存反应到数据库的工具函数 (GORM版本)
func saveReactionToDBGorm(tx *gorm.DB, reactants []string, display, status string, groupID *uint, creatorID uint) error {
	// 1. 去重并清理反应物
	uniqueReactantsMap := make(map[string]bool)
	var uniqueReactants []string
	for _, r := range reactants {
		r = strings.TrimSpace(r)
		if r != "" && !uniqueReactantsMap[r] {
			uniqueReactantsMap[r] = true
			uniqueReactants = append(uniqueReactants, r)
		}
	}

	// 严格要求必须有两个不同的反应物参与组合
	if len(uniqueReactants) != 2 {
		return fmt.Errorf("当前系统仅支持\"双反应物\"组合（如 A + B = C），请确保反应物恰好为两种不同的物质")
	}

	// 存储双向排列组合 (r1, r2) 和 (r2, r1)
	// 这样可以保证 A 上能接 B，B 上也能接 A
	var reactions []database.Reaction
	for i := 0; i < len(uniqueReactants); i++ {
		for j := 0; j < len(uniqueReactants); j++ {
			if i == j {
				continue
			}
			reactions = append(reactions, database.Reaction{
				Reactants: uniqueReactants[i],
				Products:  uniqueReactants[j],
				Display:   display,
				Status:    status,
				GroupID:   groupID,
				CreatedBy: creatorID,
			})
		}
	}

	return tx.Create(&reactions).Error
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

		// 循环处理括号，从内向外
		reBracket := regexp.MustCompile(`\(([^()]+)\)(\d*)`)
		for {
			loc := reBracket.FindStringSubmatchIndex(substance)
			if loc == nil {
				break
			}
			inner := substance[loc[2]:loc[3]]
			mult := 1
			if loc[4] != -1 && substance[loc[4]:loc[5]] != "" {
				mult, _ = strconv.Atoi(substance[loc[4]:loc[5]])
			}

			// 解析括号内部元素并暂存入 counts
			reElem := regexp.MustCompile(`([A-Z][a-z]*)(\d*)`)
			eMatches := reElem.FindAllStringSubmatch(inner, -1)
			for _, m := range eMatches {
				eCount := 1
				if m[2] != "" {
					eCount, _ = strconv.Atoi(m[2])
				}
				counts[m[1]] += eCount * mult * coeff
			}
			// 移除已处理的括号部分
			substance = substance[:loc[0]] + substance[loc[1]:]
		}

		// 处理剩余的非括号部分
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
	cleanDisplay = strings.ReplaceAll(cleanDisplay, "→", "->")
	parts := strings.Split(cleanDisplay, "->")
	if len(parts) != 2 {
		return false, "方程式格式错误，应包含 '->'、'=' 或 '→'"
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

func checkDuplicateReactants(display string, excludeGroupID *uint) (bool, string) {
	rList := parseReactants(display)
	if len(rList) < 2 {
		return false, ""
	}
	sort.Strings(rList)
	r1, r2 := rList[0], rList[1]

	// 查询是否已存在相同的反应物组合
	query := database.DB.Model(&database.Reaction{}).
		Where("((reactants = ? AND products = ?) OR (reactants = ? AND products = ?)) AND status = ?",
			r1, r2, r2, r1, "approved")

	if excludeGroupID != nil {
		query = query.Where("group_id != ?", *excludeGroupID)
	}

	var reaction database.Reaction
	err := query.First(&reaction).Error
	if err == nil {
		return true, reaction.Display
	}

	return false, ""
}

// 删除或拒绝化学反应 (Admin/Co-worker可以删除自己的)
func DeleteReaction(c *gin.Context) {
	reactionID := c.Param("id") // 接收 id，根据 id 查出 group_id 后删全组
	role := c.GetString("role")
	uid := c.GetInt("uid")

	id, err := strconv.Atoi(reactionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的反应ID"})
		return
	}

	groupID, createdBy, err := reactionRepo.GetGroupIDAndCreatorByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到该反应"})
		return
	}

	// 检查权限：admin 或者 提交者本人
	if role != "admin" && uint(uid) != createdBy {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	if groupID != nil {
		err = reactionRepo.DeleteByGroupID(*groupID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "已从实验室档案中抹除"})
}

// 直接修改化学反应 (Admin/Co-worker)
func UpdateReaction(c *gin.Context) {
	reactionID := c.Param("id")
	role := c.GetString("role")

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
	if len(rlist) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "系统当前仅支持双反应物组合"})
		return
	}

	id, err := strconv.Atoi(reactionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的反应ID"})
		return
	}

	groupID, creatorID, err := reactionRepo.GetGroupIDAndCreatorByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到原反应记录"})
		return
	}

	// 增补：校验修改后的反应是否与其他反应冲突
	if isDup, oldDisplay := checkDuplicateReactants(req.Display, groupID); isDup {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("修改后的组合已存在: %s", oldDisplay)})
		return
	}

	status := "pending_coworker"
	if role == "admin" {
		status = "approved"
	} else if role == "co-worker" {
		status = "pending_admin"
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// 删除旧记录
		if groupID != nil {
			if err := tx.Where("group_id = ?", *groupID).Delete(&database.Reaction{}).Error; err != nil {
				return err
			}
		}

		// 统一设为对应角色的预期状态
		if err := saveReactionToDBGorm(tx, rlist, req.Display, status, groupID, creatorID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存修改失败: " + err.Error()})
		return
	}

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

	status := "pending_coworker"
	if role == "admin" {
		status = "approved"
	} else if role == "co-worker" {
		status = "pending_admin"
	}

	successCount := 0
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		for i, req := range reqs {
			rlist := parseReactants(req.Display)
			if len(rlist) != 2 {
				continue
			}

			// 批量导入也进行基础校验，不满足的跳过
			if ok, _ := validateBalance(req.Display); !ok {
				continue
			}
			if isDup, _ := checkDuplicateReactants(req.Display, nil); isDup {
				continue
			}

			groupIDVal := uint(time.Now().UnixNano()/1000000 + int64(i))
			groupID := &groupIDVal

			if err := saveReactionToDBGorm(tx, rlist, req.Display, status, groupID, uint(uid)); err == nil {
				successCount++
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量提交失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("成功导入 %d 条反应", successCount),
		"count":   successCount,
	})
}

// GetSystemConfigs 获取所有系统基础配置
func GetSystemConfigs(c *gin.Context) {
	var configs []database.SystemConfig
	err := database.DB.Find(&configs).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取配置失败"})
		return
	}

	configMap := make(map[string]string)
	for _, cfg := range configs {
		configMap[cfg.Key] = cfg.Value
	}
	c.JSON(http.StatusOK, gin.H{"configs": configMap})
}

// UpdateSystemConfig 更新指定的系统配置
func UpdateSystemConfig(c *gin.Context) {
	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	config := database.SystemConfig{
		Key:   req.Key,
		Value: req.Value,
	}

	err := database.DB.Save(&config).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "配置更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "配置更新成功"})
}

// SubstanceRequest 定义物质请求结构
type SubstanceRequest struct {
	Formula string `json:"formula" binding:"required"`
	Name    string `json:"name" binding:"required"`
}

// GetSubstances 获取所有物质
func GetSubstances(c *gin.Context) {
	substanceList, err := substanceRepo.FindAllWithCreator()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误: " + err.Error()})
		return
	}

	results := make([]map[string]interface{}, 0)
	for _, s := range substanceList {
		results = append(results, map[string]interface{}{
			"id":           s.ID,
			"formula":      s.Formula,
			"name":         s.Name,
			"elements":     s.Elements,
			"status":       s.Status,
			"created_by":   s.CreatedBy,
			"creator_name": s.CreatorName,
			"created_at":   s.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, results)
}

// AddSubstance 添加新物质
func AddSubstance(c *gin.Context) {
	uid := c.GetInt("uid")
	role := c.GetString("role")

	var req SubstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 自动分析元素
	elementsMap := parseSubstanceForElements(req.Formula)
	var elementsArr []string
	for e := range elementsMap {
		elementsArr = append(elementsArr, e)
	}
	elementsStr := strings.Join(elementsArr, ",")

	// 确定初始状态
	status := "pending_coworker"
	if role == "admin" {
		status = "approved"
	} else if role == "co-worker" {
		status = "pending_admin"
	}

	substance := &database.Substance{
		Name:        req.Name,
		Formula:     req.Formula,
		Elements:    elementsStr,
		Description: req.Formula,
		Status:      status,
		CreatedBy:   uint(uid),
	}

	err := substanceRepo.Create(substance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加物质失败，可能已存在"})
		return
	}

	msg := "物质已提交，等待审核"
	if status == "approved" {
		msg = "物质已成功添加到百科"
	}
	c.JSON(http.StatusCreated, gin.H{"message": msg, "status": status})
}

// ApproveSubstance 审批物质 (Admin/Co-worker)
func ApproveSubstance(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" && role != "co-worker" {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要协作者或管理员权限"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的物质ID"})
		return
	}

	var req struct {
		Formula string `json:"formula"`
		Name    string `json:"name"`
		Reject  bool   `json:"reject"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}

	currentStatus, err := substanceRepo.GetStatus(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到该物质请求"})
		return
	}

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
			if currentStatus != "pending_admin" && currentStatus != "pending_coworker" {
				c.JSON(http.StatusForbidden, gin.H{"error": "该物质不需要管理员再审批"})
				return
			}
			newStatus = "approved"
		}
	}

	// 如果提供了修改内容
	if !req.Reject && (req.Formula != "" || req.Name != "") {
		elementsMap := parseSubstanceForElements(req.Formula)
		var elementsArr []string
		for e := range elementsMap {
			elementsArr = append(elementsArr, e)
		}
		elementsStr := strings.Join(elementsArr, ",")

		err = substanceRepo.UpdateWithElements(uint(id), req.Formula, req.Name, elementsStr, newStatus)
	} else {
		err = substanceRepo.UpdateStatus(uint(id), newStatus)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "审批已处理", "status": newStatus})
}

// UpdateSubstance 更新物质
func UpdateSubstance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的物质ID"})
		return
	}

	role := c.GetString("role")
	if role != "admin" && role != "co-worker" {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权编辑"})
		return
	}

	var req SubstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	elementsMap := parseSubstanceForElements(req.Formula)
	var elementsArr []string
	for e := range elementsMap {
		elementsArr = append(elementsArr, e)
	}
	elementsStr := strings.Join(elementsArr, ",")

	err = substanceRepo.UpdateFormula(uint(id), req.Formula, req.Name, elementsStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新物质失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// DeleteSubstance 删除物质
func DeleteSubstance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的物质ID"})
		return
	}

	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有管理员可以删除物质"})
		return
	}

	err = substanceRepo.Delete(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// 内部工具：解析化学式获取涉及的元素
func parseSubstanceForElements(substance string) map[string]bool {
	result := make(map[string]bool)
	i := 0
	for i < len(substance) {
		c := substance[i]
		if c >= 'A' && c <= 'Z' {
			start := i
			i++
			for i < len(substance) && substance[i] >= 'a' && substance[i] <= 'z' {
				i++
			}
			element := substance[start:i]
			result[element] = true
		} else {
			i++
		}
	}
	return result
}
