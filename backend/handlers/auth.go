package handlers

import (
	"chemistryuno/database"
	"chemistryuno/models"
	"chemistryuno/repository"
	"chemistryuno/utils"
	"chemistryuno/websocket"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
)

// 用户注册
func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("注册尝试 - 用户: %s, 来源IP: %s\n", req.Username, c.ClientIP())

	userRepo := repository.NewUserRepository()

	// 1. 检查用户名是否已存在
	exists, err := userRepo.ExistsByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
		return
	}

	// 2. 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 3. 创建用户
	user := &database.User{
		Username:      req.Username,
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
		"message": "注册成功",
		"uid":     user.UID,
	})
}

// 用户登录
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("登录尝试 - 用户: %s, 来源IP: %s\n", req.Username, c.ClientIP())

	userRepo := repository.NewUserRepository()
	dbUser, err := userRepo.FindByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		return
	}

	// 转换为models.User
	user := models.User{
		UID:              int(dbUser.UID),
		Username:         dbUser.Username,
		PasswordHash:     dbUser.Password,
		Avatar:           dbUser.Avatar,
		IsAdmin:          dbUser.IsAdmin,
		Role:             dbUser.Role,
		TwoFactorEnabled: dbUser.TwoFactorEnabled,
		TwoFactorSecret:  dbUser.TwoFactorSecret,
		BannedUntil:      dbUser.BannedUntil,
		FrozenUntil:      dbUser.FrozenUntil,
	}

	// 检查封禁状态
	now := time.Now()
	if user.BannedUntil != nil && now.Before(*user.BannedUntil) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": fmt.Sprintf("您的账号已被封禁，直到 %s", user.BannedUntil.Format("2006-01-02 15:04:05")),
		})
		return
	}

	// 检查冻结状态
	if user.FrozenUntil != nil && now.Before(*user.FrozenUntil) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": fmt.Sprintf("您的账号当前处于冷冻状态，直到 %s", user.FrozenUntil.Format("2006-01-02 15:04:05")),
		})
		return
	}

	// 密码登录
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 如果开启了2FA
	if user.TwoFactorEnabled {
		c.JSON(http.StatusOK, gin.H{
			"two_factor_required": true,
			"uid":                 user.UID,
		})
		return
	}

	// 生成会话
	sid, err := utils.CreateSession(user.UID, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil || sid == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建会话失败"})
		return
	}

	// 生成token
	token, err := utils.GenerateToken(int(user.UID), user.Username, user.IsAdmin, user.Role, sid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	// 4. 获取当前可用公告 (登陆触发器)
	announcementRepo := repository.NewAnnouncementRepository()
	dbAnnouncements, _ := announcementRepo.FindActive()
	var announcements []models.Announcement
	for _, a := range dbAnnouncements {
		announcements = append(announcements, models.Announcement{
			ID:         int(a.ID),
			Title:      a.Title,
			Content:    a.Content,
			Type:       a.Type,
			IsTicker:   a.IsTicker,
			CloseDelay: a.CloseDelay,
		})
	}

	fmt.Printf("登录成功 - 用户: %s (UID=%d), SID: %s, IP: %s\n", user.Username, user.UID, sid, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"uid":                user.UID,
			"username":           user.Username,
			"avatar":             user.Avatar,
			"is_admin":           user.IsAdmin,
			"role":               user.Role,
			"two_factor_enabled": user.TwoFactorEnabled,
		},
		"announcements": announcements,
	})
}

// 修改密码
func ChangePassword(c *gin.Context) {
	uid := c.GetInt("uid")

	var req struct {
		Code        string `json:"code"` // Changed from OldPassword to Code (optional if 2FA not enabled)
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userRepo := repository.NewUserRepository()

	// 获取用户信息
	user, err := userRepo.FindByID(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}

	// 验证逻辑
	if user.TwoFactorEnabled {
		// 如果开启了2FA，强制使用2FA验证码重置密码，不校验旧密码
		if req.Code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请提供 2FA 验证码以授权密码修改"})
			return
		}
		valid, _ := totp.ValidateCustom(req.Code, user.TwoFactorSecret, time.Now().UTC(), totp.ValidateOpts{
			Period: 30,
			Skew:   2,
			Digits: 6,
		})
		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "2FA 验证码无效"})
			return
		}
	} else {
		// 如果未开启2FA，仍然使用旧密码验证（作为兜底）
		if !utils.CheckPassword(req.OldPassword, user.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "当前密码错误"})
			return
		}
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 更新密码
	err = userRepo.UpdatePassword(uint(uid), hashedPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新密码失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// 通过2FA重置密码
func ResetPasswordBy2FA(c *gin.Context) {
	var req struct {
		Username    string `json:"username" binding:"required"`
		Code        string `json:"code" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}

	userRepo := repository.NewUserRepository()
	dbUser, err := userRepo.FindByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "账户不存在"})
		return
	}

	if !dbUser.TwoFactorEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该账户未开启 2FA，无法通过此方式找回。请联系管理员。"})
		return
	}

	// 验证 2FA 码
	valid, _ := totp.ValidateCustom(req.Code, dbUser.TwoFactorSecret, time.Now().UTC(), totp.ValidateOpts{
		Period: 30,
		Skew:   2,
		Digits: 6,
	})

	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "验证码无效"})
		return
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码处理失败"})
		return
	}

	// 更新密码
	err = userRepo.UpdatePassword(dbUser.UID, hashedPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新密码失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码重置成功，请使用新密码登录"})
}

// 更新头像
func UpdateAvatar(c *gin.Context) {
	uid := c.GetInt("uid")

	var req models.UpdateAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userRepo := repository.NewUserRepository()
	err := userRepo.UpdateAvatar(uint(uid), req.Avatar)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新头像失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "头像更新成功", "avatar": req.Avatar})
}

// 注销账号
func DeleteAccount(c *gin.Context) {
	uid := c.GetInt("uid")

	userRepo := repository.NewUserRepository()
	err := userRepo.Delete(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注销账号失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "账号已注销"})
}

// 获取会话列表
func GetSessions(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权的访问"})
		return
	}

	sidVal, _ := c.Get("sid")
	currentSID := ""
	if sidVal != nil {
		currentSID = sidVal.(string)
	}

	sessionRepo := repository.NewSessionRepository()
	sessions, err := sessionRepo.FindByUserID(uint(uid.(int)))
	if err != nil {
		fmt.Printf("查询数据库失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法加载设备列表"})
		return
	}

	var result []gin.H
	for _, s := range sessions {
		result = append(result, gin.H{
			"id":          s.ID,
			"user_agent":  s.UserAgent,
			"ip":          s.IPAddress,
			"last_active": s.LastActive.Format(time.RFC3339),
			"created_at":  s.CreatedAt.Format(time.RFC3339),
			"is_current":  s.ID == currentSID,
		})
	}

	c.JSON(http.StatusOK, result)
}

// 撤销会话（登出设备）
func RevokeSession(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		ID string `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	sessionRepo := repository.NewSessionRepository()
	err := sessionRepo.DeleteByIDAndUserID(req.ID, uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "登出失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已成功在该设备上登出"})
}

// 冻结账号
func FreezeAccount(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		Hours int `json:"hours" binding:"required,min=1,max=24"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "冻结时长必须在1-24小时之间"})
		return
	}

	frozenUntil := time.Now().Add(time.Duration(req.Hours) * time.Hour)

	userRepo := repository.NewUserRepository()
	err := userRepo.UpdateFreezeStatus(uint(uid), &frozenUntil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "冻结失败"})
		return
	}

	// 冻结后强制登出所有当前会话
	sessionRepo := repository.NewSessionRepository()
	_ = sessionRepo.DeleteByUserID(uint(uid))

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("账号已冻结，直到 %s", frozenUntil.Format("2006-01-02 15:04:05"))})
}

// 获取用户信息
func GetUserInfo(c *gin.Context) {
	uid := c.GetInt("uid")

	userRepo := repository.NewUserRepository()
	user, err := userRepo.FindByID(uint(uid))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// SearchUsers 搜索用户
func SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索内容不能为空"})
		return
	}

	userRepo := repository.NewUserRepository()
	users, err := userRepo.SearchUsers(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索失败"})
		return
	}

	var result []map[string]interface{}
	for _, user := range users {
		totalBounty, _ := repository.BountyRepo.GetTotalBounty(user.UID)
		isOnline := false
		if websocket.GlobalHub != nil {
			isOnline = websocket.GlobalHub.IsUIDOnline(int(user.UID))
		}

		result = append(result, map[string]interface{}{
			"uid":            user.UID,
			"username":       user.Username,
			"avatar":         user.Avatar,
			"points":         user.Points,
			"monthly_points": user.MonthlyPoints,
			"win_count":      user.WinCount,
			"total_games":    user.TotalGames,
			"bounty":         totalBounty,
			"is_online":      isOnline,
		})
	}

	c.JSON(http.StatusOK, result)
}
