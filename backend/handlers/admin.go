package handlers

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/game"
	"chemistryuno/backend/models"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/utils"
	"chemistryuno/backend/websocket"
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

func parseReactionListFilter(c *gin.Context) (repository.ReactionListFilter, bool) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}

	filter := repository.ReactionListFilter{
		Search:   strings.TrimSpace(c.Query("q")),
		Status:   strings.TrimSpace(c.Query("status")),
		Page:     page,
		PageSize: pageSize,
	}

	if raw := strings.TrimSpace(c.Query("has_invalid")); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid has_invalid value"})
			return filter, false
		}
		filter.HasInvalid = &parsed
	}

	return filter, true
}

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

	// 检查邮箱是否已存在
	exists, err := userRepo.ExistsByEmail(req.Email)
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

	legacyUsername := strings.Split(req.Email, "@")[0]
	user := &database.User{
		Username:      legacyUsername,
		Password:      hashedPassword,
		Avatar:        "🧪",
		Role:          "user",
		Points:        1000,
		MonthlyPoints: 1000,
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

// GetAdminStats 返回管理员面板所需的汇总统计数据
func GetAdminStats(c *gin.Context) {
	userCount, _ := userRepo.GetUserCount()

	var historyCount int64
	database.DB.Model(&database.GameHistory{}).Count(&historyCount)

	deckCardTypes := 0
	if deck, err := deckRepo.FindGlobalDeck(); err == nil && deck != nil {
		var cards map[string]int
		if err := json.Unmarshal(deck.Cards, &cards); err == nil {
			deckCardTypes = len(cards)
		}
	}

	activeRooms := len(game.GetAllRoomsAdmin())

	c.JSON(http.StatusOK, gin.H{
		"user_count":      userCount,
		"history_count":   historyCount,
		"deck_card_types": deckCardTypes,
		"active_rooms":    activeRooms,
	})
}

// 管理员踢出玩家（服务器级别，直接断开连接）
func KickPlayer(c *gin.Context) {
	var req struct {
		TargetUID int    `json:"target_uid" binding:"required"`
		Reason    string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 权限检查：获取执行操作的用户角色
	role, _ := c.Get("role")
	roleStr := role.(string)

	// 获取目标用户角色
	targetUser, err := userRepo.FindByUID(uint(req.TargetUID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "目标用户不存在"})
		return
	}

	// 规则 2.3: admin 无法被封禁/踢出
	if targetUser.Role == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "无法对管理员执行此操作"})
		return
	}

	// 规则 2.2: co-worker 只能踢出/封禁普通玩家 (user)
	// 规则 2.1: admin 可以封禁/踢出 co-worker
	if roleStr == "co-worker" && targetUser.Role == "co-worker" {
		c.JSON(http.StatusForbidden, gin.H{"error": "co-worker 无法踢出另一名 co-worker"})
		return
	}

	if req.Reason == "" {
		req.Reason = "您由于不正当游戏而被踢出"
	}

	// 如果玩家在游戏房间中，先从房间踢出
	if websocket.GlobalHub != nil {
		// 发送 force_logout 事件，前端会清除 SID 并跳转到登录页
		websocket.GlobalHub.SendToUID(req.TargetUID, websocket.Message{
			Type:    "force_logout",
			Message: req.Reason,
		})
	}

	// 删除该用户的所有会话
	sessionRepo := repository.NewSessionRepository()
	_ = sessionRepo.DeleteByUserUID(uint(req.TargetUID))

	c.JSON(http.StatusOK, gin.H{"message": "玩家已被踢出服务器"})
}

// 管理员封禁用户
func BanUser(c *gin.Context) {
	var req struct {
		TargetUID   int    `json:"target_uid" binding:"required"`
		BannedUntil string `json:"banned_until" binding:"required"` // ISO 8601 datetime
		Reason      string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 权限检查
	role, _ := c.Get("role")
	roleStr := role.(string)

	targetUser, err := userRepo.FindByUID(uint(req.TargetUID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "目标用户不存在"})
		return
	}

	// 管理员无法被封禁
	if targetUser.Role == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "无法封禁管理人员"})
		return
	}

	// co-worker 只能封禁 user
	if roleStr == "co-worker" && targetUser.Role == "co-worker" {
		c.JSON(http.StatusForbidden, gin.H{"error": "co-worker 无法封禁另一名 co-worker"})
		return
	}

	if req.Reason == "" {
		req.Reason = "您由于不正当游戏而被封禁"
	}

	// 解析前端传入的 ISO 8601 时间
	bannedUntil, err := time.Parse(time.RFC3339, req.BannedUntil)
	if err != nil {
		// 尝试不带时区的格式
		bannedUntil, err = time.Parse("2006-01-02T15:04:05", req.BannedUntil)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "时间格式无效"})
			return
		}
	}

	if bannedUntil.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "封禁截止时间必须晚于当前时间"})
		return
	}

	err = userRepo.UpdateBanStatusWithReason(uint(req.TargetUID), &bannedUntil, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "封禁失败"})
		return
	}

	// 通过 WebSocket 立即踢出玩家（无提示弹窗，直接断开）
	if websocket.GlobalHub != nil {
		websocket.GlobalHub.SendToUID(req.TargetUID, websocket.Message{
			Type:    "force_logout",
			Message: "您的账号已被封禁",
		})
	}

	// 封禁后强制登出
	sessionRepo := repository.NewSessionRepository()
	_ = sessionRepo.DeleteByUserUID(uint(req.TargetUID))

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("用户已被封禁至 %s", bannedUntil.Format("2006-01-02 15:04:05"))})
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
	normalizedCards, _ := game.NormalizeBuiltinDeckCards(cards)
	if len(normalizedCards) == 0 {
		normalizedCards = game.BuiltinDeckDefaults()
	}
	initialCards := deck.InitialCards
	if initialCards <= 0 {
		initialCards = 10
	}

	config := models.DeckConfig{
		ID:           int(deck.ID),
		Name:         deck.Name,
		IsGlobal:     deck.IsGlobal,
		Cards:        normalizedCards,
		InitialCards: initialCards,
		CreatedBy:    int(deck.CreatedByUID),
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

	normalizedCards, unknownCards := game.NormalizeBuiltinDeckCards(req.Cards)
	if len(unknownCards) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":          "卡组配置仅支持原有普通牌和特殊牌，插件牌请在插件中管理",
			"unknown_cards":  unknownCards,
			"allowed_scope":  "builtin_only",
			"plugin_managed": true,
		})
		return
	}
	if len(normalizedCards) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "卡组不能为空"})
		return
	}

	cardsJSON, _ := json.Marshal(normalizedCards)

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

// ResetGlobalDeckConfig 恢复全局牌组默认配置
// POST /api/admin/deck-config/reset
func ResetGlobalDeckConfig(c *gin.Context) {
	deck, err := deckRepo.FindGlobalDeck()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "全局配置不存在"})
		return
	}

	cards := game.BuiltinDeckDefaults()
	cardsJSON, _ := json.Marshal(cards)

	deck.Name = "默认牌组"
	deck.Cards = cardsJSON
	deck.InitialCards = 10

	if err := deckRepo.Update(deck); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "恢复默认牌组失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "已恢复默认牌组配置",
		"name":          deck.Name,
		"cards":         cards,
		"initial_cards": deck.InitialCards,
	})
}

// 删除用户（管理员）
func DeleteUser(c *gin.Context) {
	userUIDStr := c.Param("uid")
	uid, err := strconv.ParseUint(userUIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户UID"})
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
	userUIDStr := c.Param("uid")
	uid, err := strconv.ParseUint(userUIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户UID"})
		return
	}

	var req models.AdminChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户是否存在且不是管理员
	isAdmin, err := userRepo.FindIsAdminByUID(uint(uid))
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
	userUIDStr := c.Param("uid")
	uid, err := strconv.ParseUint(userUIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户UID"})
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

		var cheatUIDs []int
		if len(h.CheatUIDs) > 0 {
			if err := json.Unmarshal([]byte(h.CheatUIDs), &cheatUIDs); err != nil {
				cheatUIDs = []int{}
			}
		}

		winnerName := "AI"
		if h.IsInvalid {
			winnerName = "无效对局"
		} else if h.WinnerUID != nil {
			if h.WinnerName != "" {
				winnerName = h.WinnerName
			} else {
				winnerName = "未知用户"
			}
		}

		history = append(history, map[string]interface{}{
			"id":                    h.ID,
			"room_id":               h.RoomID,
			"winner_uid":            h.WinnerUID,
			"winner_name":           winnerName,
			"is_invalid":            h.IsInvalid,
			"invalid_reason":        h.InvalidReason,
			"has_replay":            h.HasReplay,
			"replay_permanent":      h.ReplayPermanent,
			"replay_expires_at":     h.ReplayExpiresAt,
			"replay_cleared_at":     h.ReplayClearedAt,
			"cheat_detected":        h.CheatDetected,
			"cheat_uids":            cheatUIDs,
			"players":               players,
			"original_player_count": h.OriginalPlayerCount,
			"quitted_count":         h.QuittedCount,
			"started_at":            h.StartedAt,
			"finished_at":           h.FinishedAt,
			"created_at":            h.FinishedAt, // 兼容前端字段名
		})
	}

	c.JSON(http.StatusOK, history)
}

// 获取所有化学反应 (Admin/Co-worker)
func GetReactions(c *gin.Context) {
	filter, ok := parseReactionListFilter(c)
	if !ok {
		return
	}

	role := c.GetString("role")
	if role != "admin" && role != "co-worker" {
		uid := uint(c.GetInt("uid"))
		filter.ViewerUID = &uid
		filter.IncludeApproved = true
	}

	if c.Query("paginated") == "1" {
		reactionList, total, err := reactionRepo.FindGroupedWithCreatorPage(filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
			return
		}

		totalPages := 0
		if total > 0 {
			totalPages = int((total + int64(filter.PageSize) - 1) / int64(filter.PageSize))
		}

		c.JSON(http.StatusOK, gin.H{
			"items": reactionList,
			"pagination": gin.H{
				"page":        filter.Page,
				"page_size":   filter.PageSize,
				"total":       total,
				"total_pages": totalPages,
			},
		})
		return
	}

	reactionList, err := reactionRepo.FindAllGroupedWithCreator()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}

	var reactions []map[string]interface{}
	for _, r := range reactionList {
		reactions = append(reactions, map[string]interface{}{
			"id":                   r.ID,
			"display":              r.Display,
			"status":               r.Status,
			"group_id":             r.GroupID,
			"has_invalid_elements": r.HasInvalidElements,
			"created_by":           r.CreatedByUID,
			"creator_name":         r.CreatorName,
			"created_at":           r.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, reactions)
}

// 获取所有已批准的反应 (Wiki)
func GetAllReactions(c *gin.Context) {
	filter, ok := parseReactionListFilter(c)
	if !ok {
		return
	}
	filter.Status = "approved"

	if c.Query("paginated") == "1" {
		reactionList, total, err := reactionRepo.FindGroupedWithCreatorPage(filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
			return
		}

		totalPages := 0
		if total > 0 {
			totalPages = int((total + int64(filter.PageSize) - 1) / int64(filter.PageSize))
		}

		c.JSON(http.StatusOK, gin.H{
			"items": reactionList,
			"pagination": gin.H{
				"page":        filter.Page,
				"page_size":   filter.PageSize,
				"total":       total,
				"total_pages": totalPages,
			},
		})
		return
	}

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
	filter, ok := parseReactionListFilter(c)
	if !ok {
		return
	}
	uidUint := uint(uid)
	filter.ViewerUID = &uidUint

	if c.Query("paginated") == "1" {
		reactionList, total, err := reactionRepo.FindGroupedWithCreatorPage(filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
			return
		}

		totalPages := 0
		if total > 0 {
			totalPages = int((total + int64(filter.PageSize) - 1) / int64(filter.PageSize))
		}

		c.JSON(http.StatusOK, gin.H{
			"items": reactionList,
			"pagination": gin.H{
				"page":        filter.Page,
				"page_size":   filter.PageSize,
				"total":       total,
				"total_pages": totalPages,
			},
		})
		return
	}

	reactionList, err := reactionRepo.FindMyReactions(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}

	var reactions []map[string]interface{}
	for _, r := range reactionList {
		reactions = append(reactions, map[string]interface{}{
			"id":                   r.ID,
			"display":              r.Display,
			"status":               r.Status,
			"has_invalid_elements": r.HasInvalidElements,
			"created_at":           r.CreatedAt,
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
	groupIDUint, err := strconv.ParseUint(groupIDStr, 10, 64)
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
	_, err = reactionRepo.GetStatusByGroupID(groupID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到该反应请求"})
		return
	}

	// 审批流校验
	newStatus := ""
	if req.Reject {
		newStatus = "rejected"
	} else {
		// 统一合并审批状态：co-worker 和 admin 都可以直接批准
		newStatus = "approved"
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
			if err := saveReactionToDBGorm(tx, rlist, req.Display, newStatus, &groupID, reaction.CreatedByUID); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "处理审批事务失败: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "审批已处理", "status": newStatus})
		if newStatus == "approved" {
			game.RebuildSubstanceCache()
		}
		return
	}

	// 如果没有修改内容，直接更新状态
	err = reactionRepo.UpdateStatusByGroupID(groupID, newStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "状态已更新", "status": newStatus})
	if newStatus == "approved" {
		// 审核通过时，确保反应物已入库（百科）
		var reactions []database.Reaction
		if err := database.DB.Where("group_id = ?", groupID).Find(&reactions).Error; err == nil {
			var formulas []string
			for _, r := range reactions {
				formulas = append(formulas, r.R1, r.R2)
			}
			game.EnsureSubstancesExist(database.DB, formulas, 100000000) // 使用系统管理员UID
		}
		game.RebuildSubstanceCache()
	}
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
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未获取到用户身份信息"})
		return
	}
	role := c.GetString("role")

	var req models.ReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	display := strings.TrimSpace(req.Display)
	if display == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "反应方程式不能为空"})
		return
	}

	// 自动识别 r1, r2
	rlist := parseReactants(display)
	if len(rlist) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法自动识别反应物，请检查方程式格式（如: A + B -> C）"})
		return
	}

	// 1. 校验质量守恒
	if ok, errInfo := validateBalance(display); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "方程式校验失败: " + errInfo})
		return
	}

	// 2. 校验反应物重复（检查已审核通过的）
	if isDup, oldDisplay := checkDuplicateReactants(display, nil); isDup {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("该反应物组合已有经认证的公式: %s", oldDisplay)})
		return
	}

	// 3. 严格校验反应物数量（仅支持双反应物）
	if len(rlist) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "当前实验室系统仅支持由两种不同物质组成的反应组合（如 A + B = C）"})
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

	// 生成 group_id (使用毫秒时间戳)
	groupIDVal := uint(time.Now().UnixNano() / 1000000)
	groupID := &groupIDVal

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 这里 uid 需要断言
		userID := uint(0)
		if u, ok := uid.(int); ok {
			userID = uint(u)
		} else if u, ok := uid.(uint); ok {
			userID = u
		}
		return saveReactionToDBGorm(tx, rlist, display, status, groupID, userID)
	})

	if err != nil {
		log.Printf("[AddReaction] Database Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存反应失败: " + err.Error()})
		return
	}

	msg := "反应已提交，等待协作者或管理员审核"
	if status == "approved" {
		msg = "反应已成功加入核心数据库"
		game.RebuildSubstanceCache()
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

	// Canonical ordering: R1 < R2 (字母序)
	r1, r2 := uniqueReactants[0], uniqueReactants[1]
	if r1 > r2 {
		r1, r2 = r2, r1
	}

	// 创建单条canonical反应（不再双向存储）
	reaction := database.Reaction{
		R1:           r1,
		R2:           r2,
		Display:      display,
		Status:       status,
		GroupID:      groupID,
		CreatedByUID: creatorID,
	}

	if err := tx.Create(&reaction).Error; err != nil {
		return err
	}

	// 自动录入物质到百科 (仅当反应被批准时)
	if status == "approved" {
		game.EnsureSubstancesExist(tx, []string{r1, r2}, creatorID)
	}

	return nil
}

// 内部校验逻辑
func countElements(expr string) map[string]int {
	counts := make(map[string]int)
	expr = game.NormalizeSubscripts(expr)
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

	if len(leftCounts) == 0 || len(rightCounts) == 0 {
		return false, "方程式左右两侧必须包含有效的化学元素"
	}

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

	// Canonical ordering
	if r1 > r2 {
		r1, r2 = r2, r1
	}

	// 查询是否已存在相同的反应物组合（使用R1/R2字段）
	query := database.DB.Model(&database.Reaction{}).
		Where("r1 = ? AND r2 = ? AND status = ?", r1, r2, "approved")

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
	} else {
		// 如果没有 group_id，直接按 ID 删除
		err = database.DB.Delete(&database.Reaction{}, uint(id)).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "已从实验室档案中抹除"})
	game.RebuildSubstanceCache()
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
	game.RebuildSubstanceCache()
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
	if successCount > 0 {
		game.RebuildSubstanceCache()
	}
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

// GetGameTimeConfigs 获取游戏时间配置
func GetGameTimeConfigs(c *gin.Context) {
	configRepo := repository.NewConfigRepository()
	configs, err := configRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取配置失败"})
		return
	}

	// 返回游戏时间相关的配置
	timeConfigs := map[string]string{
		"player_kick_timeout":    configs["player_kick_timeout"],
		"player_action_timeout":  configs["player_action_timeout"],
		"auto_start_timeout":     configs["auto_start_timeout"],
		"half_ready_timeout":     configs["half_ready_timeout"],
		"reconnect_grace_period": configs["reconnect_grace_period"],
		"points_scaling_enabled": configs["points_scaling_enabled"],
	}

	c.JSON(http.StatusOK, gin.H{"configs": timeConfigs})
}

// UpdateGameTimeConfig 更新游戏时间配置
func UpdateGameTimeConfig(c *gin.Context) {
	var req struct {
		PlayerKickTimeout    int    `json:"player_kick_timeout"`
		PlayerActionTimeout  int    `json:"player_action_timeout"`
		AutoStartTimeout     int    `json:"auto_start_timeout"`
		HalfReadyTimeout     int    `json:"half_ready_timeout"`
		ReconnectGracePeriod int    `json:"reconnect_grace_period"`
		PointsScalingEnabled string `json:"points_scaling_enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// 验证参数范围
	if req.PlayerKickTimeout < 10 || req.PlayerKickTimeout > 300 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "玩家踢出时间必须在10-300秒之间"})
		return
	}
	if req.PlayerActionTimeout < 10 || req.PlayerActionTimeout > 300 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "玩家操作时间必须在10-300秒之间"})
		return
	}
	if req.AutoStartTimeout < 5 || req.AutoStartTimeout > 60 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "自动开始时间必须在5-60秒之间"})
		return
	}
	if req.HalfReadyTimeout < 30 || req.HalfReadyTimeout > 120 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "半数准备时间必须在30-120秒之间"})
		return
	}

	configRepo := repository.NewConfigRepository()

	// 更新配置
	updates := map[string]string{
		"player_kick_timeout":    fmt.Sprintf("%d", req.PlayerKickTimeout),
		"player_action_timeout":  fmt.Sprintf("%d", req.PlayerActionTimeout),
		"auto_start_timeout":     fmt.Sprintf("%d", req.AutoStartTimeout),
		"half_ready_timeout":     fmt.Sprintf("%d", req.HalfReadyTimeout),
		"reconnect_grace_period": fmt.Sprintf("%d", req.ReconnectGracePeriod),
		"points_scaling_enabled": req.PointsScalingEnabled,
	}

	for key, value := range updates {
		if err := configRepo.SetValue(key, value); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "配置更新失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "游戏时间配置已更新，将在新游戏中生效"})
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
			"created_by":   s.CreatedByUID,
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
	elementsMap := game.ParseSubstanceForElements(req.Formula)
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
		Name:         req.Name,
		Formula:      req.Formula,
		Elements:     elementsStr,
		Status:       status,
		CreatedByUID: uint(uid),
	}

	err := substanceRepo.Create(substance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加物质失败，可能已存在"})
		return
	}

	msg := "物质已提交，等待审核"
	if status == "approved" {
		msg = "物质已成功添加到百科"
		game.RebuildSubstanceCache()
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
		elementsMap := game.ParseSubstanceForElements(req.Formula)
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
	if newStatus == "approved" {
		game.RebuildSubstanceCache()
	}
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

	elementsMap := game.ParseSubstanceForElements(req.Formula)
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
	game.RebuildSubstanceCache()
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
	game.RebuildSubstanceCache()
}

// BatchApproveSubstances 批量批准物质
func BatchApproveSubstances(c *gin.Context) {
	var req struct {
		GroupIDs []uint `json:"group_ids" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	// 批量操作最大限制（防止DOS攻击）
	const MAX_BATCH_SIZE = 100
	if len(req.GroupIDs) > MAX_BATCH_SIZE {
		c.JSON(http.StatusBadRequest, gin.H{"error": "批量操作最多支持100条记录"})
		return
	}

	// 批量更新状态
	affected, err := substanceRepo.BatchUpdateStatusByGroupIDs(req.GroupIDs, "approved")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量批准失败: " + err.Error()})
		return
	}

	// 重建物质缓存
	game.RebuildSubstanceCache()

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("批量批准成功，共影响 %d 条记录", affected),
		"count":   affected,
	})
}

// BatchRejectSubstances 批量拒绝物质
func BatchRejectSubstances(c *gin.Context) {
	var req struct {
		GroupIDs []uint `json:"group_ids" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	// 批量操作最大限制
	const MAX_BATCH_SIZE = 100
	if len(req.GroupIDs) > MAX_BATCH_SIZE {
		c.JSON(http.StatusBadRequest, gin.H{"error": "批量操作最多支持100条记录"})
		return
	}

	// 批量删除
	affected, err := substanceRepo.BatchDeleteByGroupIDs(req.GroupIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量拒绝失败: " + err.Error()})
		return
	}

	// 重建物质缓存
	game.RebuildSubstanceCache()

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("批量拒绝成功，共删除 %d 条记录", affected),
		"count":   affected,
	})
}

// BatchApproveReactions 批量批准反应
func BatchApproveReactions(c *gin.Context) {
	var req struct {
		GroupIDs []uint `json:"group_ids" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	// 批量操作最大限制
	const MAX_BATCH_SIZE = 100
	if len(req.GroupIDs) > MAX_BATCH_SIZE {
		c.JSON(http.StatusBadRequest, gin.H{"error": "批量操作最多支持100条记录"})
		return
	}

	// 批量更新状态
	affected, err := reactionRepo.BatchUpdateStatusByGroupIDs(req.GroupIDs, "approved")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量批准失败: " + err.Error()})
		return
	}

	// 批准后自动将反应物录入物质百科
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		var reactions []database.Reaction
		if err := tx.Where("group_id IN ?", req.GroupIDs).Find(&reactions).Error; err != nil {
			return err
		}

		// 收集所有涉及的物质化学式
		formulaSet := make(map[string]bool)
		for _, rxn := range reactions {
			formulaSet[rxn.R1] = true
			formulaSet[rxn.R2] = true
		}

		var formulas []string
		for f := range formulaSet {
			formulas = append(formulas, f)
		}

		// 确保物质已入库
		game.EnsureSubstancesExist(tx, formulas, 100000000)
		return nil
	})

	if err != nil {
		log.Printf("[批量批准] 物质同步失败: %v", err)
	}

	// 重建物质缓存
	game.RebuildSubstanceCache()

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("批量批准成功，共影响 %d 条记录", affected),
		"count":   affected,
	})
}

// BatchRejectReactions 批量拒绝反应
func BatchRejectReactions(c *gin.Context) {
	var req struct {
		GroupIDs []uint `json:"group_ids" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	// 批量操作最大限制
	const MAX_BATCH_SIZE = 100
	if len(req.GroupIDs) > MAX_BATCH_SIZE {
		c.JSON(http.StatusBadRequest, gin.H{"error": "批量操作最多支持100条记录"})
		return
	}

	// 批量删除
	affected, err := reactionRepo.BatchDeleteByGroupIDs(req.GroupIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量拒绝失败: " + err.Error()})
		return
	}

	// 重建物质缓存
	game.RebuildSubstanceCache()

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("批量拒绝成功，共删除 %d 条记录", affected),
		"count":   affected,
	})
}

// GetLogs 获取实时日志（管理员专用）
func GetLogs(c *gin.Context) {
	count := 100 // 默认获取最近100条
	if countStr := c.Query("count"); countStr != "" {
		if parsedCount, err := strconv.Atoi(countStr); err == nil && parsedCount > 0 {
			if parsedCount > 1000 {
				count = 1000 // 最多1000条
			} else {
				count = parsedCount
			}
		}
	}

	level := c.Query("level") // 可选：过滤日志级别

	var logs []utils.LogEntry
	if level != "" {
		logs = utils.GetLogsByLevel(level, count)
	} else {
		logs = utils.GetLogs(count)
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"count": len(logs),
	})
}

// GetLogsStream 通过流式响应实时推送日志（管理员专用）
func GetLogsStream(c *gin.Context) {
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "当前环境不支持日志流式输出"})
		return
	}

	level := strings.ToUpper(strings.TrimSpace(c.Query("level")))
	if level == "ALL" {
		level = ""
	}

	if level != "" && level != "INFO" && level != "WARNING" && level != "ERROR" && level != "DEBUG" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的日志级别过滤参数"})
		return
	}

	count := 200 // 初次连接默认回放最近200条
	if countStr := c.Query("count"); countStr != "" {
		if parsedCount, err := strconv.Atoi(countStr); err == nil && parsedCount > 0 {
			if parsedCount > 1000 {
				count = 1000
			} else {
				count = parsedCount
			}
		}
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	var initialLogs []utils.LogEntry
	if level != "" {
		initialLogs = utils.GetLogsByLevel(level, count)
	} else {
		initialLogs = utils.GetLogs(count)
	}

	// 先按时间正序回放最近日志，保证前端底部是最新。
	for i := len(initialLogs) - 1; i >= 0; i-- {
		payload, err := json.Marshal(initialLogs[i])
		if err != nil {
			continue
		}
		if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", payload); err != nil {
			return
		}
	}
	flusher.Flush()

	subID, stream := utils.SubscribeLogs()
	defer utils.UnsubscribeLogs(subID)

	keepAliveTicker := time.NewTicker(20 * time.Second)
	defer keepAliveTicker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case entry, ok := <-stream:
			if !ok {
				return
			}

			if level != "" && entry.Level != level {
				continue
			}

			payload, err := json.Marshal(entry)
			if err != nil {
				continue
			}

			if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", payload); err != nil {
				return
			}
			flusher.Flush()
		case <-keepAliveTicker.C:
			if _, err := fmt.Fprint(c.Writer, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

// ClearLogs 清空日志缓冲（管理员专用）
func ClearLogs(c *gin.Context) {
	utils.ClearLogs()
	c.JSON(http.StatusOK, gin.H{
		"message": "日志已清空",
	})
}
