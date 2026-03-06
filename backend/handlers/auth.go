package handlers

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/models"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/utils"
	"chemistryuno/backend/websocket"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

// consumeVerificationCode atomically deletes one matching, unexpired code.
func consumeVerificationCode(email, code, codeType string) (bool, error) {
	consumed := false
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var vCode database.VerificationCode
		err := tx.Where("email = ? AND code = ? AND type = ? AND expires_at > ?", email, code, codeType, time.Now()).
			Order("created_at DESC").
			First(&vCode).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}

		result := tx.Where("id = ?", vCode.ID).Delete(&database.VerificationCode{})
		if result.Error != nil {
			return result.Error
		}
		consumed = result.RowsAffected > 0
		return nil
	})
	if err != nil {
		return false, err
	}
	return consumed, nil
}

// 用户注册
func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	userRepo := repository.NewUserRepository()
	exists, err := userRepo.ExistsByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	if utils.IsSMTPConfigured() {
		if req.Code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "verification code is required"})
			return
		}
		consumed, err := consumeVerificationCode(req.Email, req.Code, "register")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		if !consumed {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "verification code invalid or expired"})
			return
		}
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// 获取最大 UID 并计算新 UID
	maxUID, err := userRepo.GetMaxUID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error when fetching max uid"})
		return
	}
	newUID := maxUID + 1
	if newUID < 100000000 {
		newUID = 100000000 // 建议起始 UID
	}

	baseUsername := strings.Split(req.Email, "@")[0]
	if baseUsername == "" {
		baseUsername = "researcher"
	}
	finalUsername := baseUsername
	// 确保用户名唯一，避免同名前缀冲突
	for i := 1; ; i++ {
		exists, _ := userRepo.ExistsByUsername(finalUsername)
		if !exists {
			break
		}
		finalUsername = fmt.Sprintf("%s_%d", baseUsername, i)
	}

	user := &database.User{
		UID:           newUID,
		Username:      finalUsername,
		Email:         req.Email,
		Nickname:      req.Nickname,
		Password:      hashedPassword,
		Avatar:        "\U0001F9EA",
		Role:          "user",
		Points:        1000,
		MonthlyPoints: 1000,
		Level:         1,
	}

	if err := userRepo.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "register success",
		"uid":     user.UID,
	})
}

func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	fmt.Printf("login attempt - email: %s, ip: %s\n", email, c.ClientIP())

	userRepo := repository.NewUserRepository()
	dbUser, err := userRepo.FindByEmail(email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	user := models.User{
		UID:              int(dbUser.UID),
		Username:         dbUser.Username,
		Email:            dbUser.Email,
		Nickname:         dbUser.Nickname,
		PasswordHash:     dbUser.Password,
		Avatar:           dbUser.Avatar,
		IsAdmin:          dbUser.IsAdmin,
		Role:             dbUser.Role,
		TwoFactorEnabled: dbUser.TwoFactorEnabled,
		TwoFactorSecret:  dbUser.TwoFactorSecret,
		BannedUntil:      dbUser.BannedUntil,
		FrozenUntil:      dbUser.FrozenUntil,
	}

	now := time.Now()
	if user.BannedUntil != nil && now.Before(*user.BannedUntil) {
		banReason := dbUser.BanReason
		if banReason == "" {
			banReason = "account is banned due to policy violations"
		}
		c.JSON(http.StatusForbidden, gin.H{
			"error":        fmt.Sprintf("account banned: %s", banReason),
			"banned_until": user.BannedUntil.Format(time.RFC3339),
		})
		return
	}

	if user.FrozenUntil != nil && now.Before(*user.FrozenUntil) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":        "account is temporarily frozen",
			"frozen_until": user.FrozenUntil.Format(time.RFC3339),
		})
		return
	}

	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	if user.UID < 0 {
		log.Printf("[AI login blocked] UID=%d, IP=%s", user.UID, c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "AI account cannot login"})
		return
	}

	if user.TwoFactorEnabled {
		c.JSON(http.StatusOK, gin.H{
			"two_factor_required": true,
			"uid":                 user.UID,
		})
		return
	}

	sid, err := utils.CreateSession(user.UID, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil || sid == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	token, err := utils.GenerateToken(int(user.UID), user.Email, user.IsAdmin, user.Role, sid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	isReturningPlayer := false
	daysSinceLastLogin := 0
	if dbUser.LastOfflineAt != nil {
		daysSince := time.Since(*dbUser.LastOfflineAt).Hours() / 24
		daysSinceLastLogin = int(daysSince)
		if daysSince >= 30 {
			isReturningPlayer = true
			log.Printf("[returning user] %s (UID=%d), inactive days=%d", user.Nickname, user.UID, daysSinceLastLogin)
		}
	}

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

	fmt.Printf("login success - email: %s (UID=%d), SID: %s, IP: %s\n", user.Email, user.UID, sid, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"uid":                user.UID,
			"username":           user.Username,
			"email":              user.Email,
			"nickname":           user.Nickname,
			"avatar":             user.Avatar,
			"is_admin":           user.IsAdmin,
			"role":               user.Role,
			"two_factor_enabled": user.TwoFactorEnabled,
		},
		"announcements":         announcements,
		"is_returning_player":   isReturningPlayer,
		"days_since_last_login": daysSinceLastLogin,
	})
}

func ChangePassword(c *gin.Context) {
	uid := c.GetInt("uid")

	var req struct {
		Code        string `json:"code"`         // 2FA 或 邮箱验证码
		OldPassword string `json:"old_password"` // 旧密码 (当没配置2FA且没使用邮箱验证时使用，作为备选)
		NewPassword string `json:"new_password" binding:"required,min=6"`
		UseEmail    bool   `json:"use_email"` // 是否使用邮箱验证模式
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userRepo := repository.NewUserRepository()

	// 获取用户信息
	user, err := userRepo.FindByUID(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}

	// 验证逻辑优先顺序: 邮箱验证码 > 2FA > 旧密码
	if req.UseEmail {
		if req.Code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请提供邮箱验证码"})
			return
		}
		// 验证邮箱码
		consumed, err := consumeVerificationCode(user.Email, req.Code, "change_password")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		if !consumed {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "verification code invalid or expired"})
			return
		}
	} else if user.TwoFactorEnabled {
		// 如果开启了2FA且没选择邮箱验证，强制使用2FA
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
		// 传统模式：要求旧密码
		if req.OldPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请提供当前密码或选择邮箱验证"})
			return
		}
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
		Email       string `json:"email" binding:"required,email"`
		Code        string `json:"code" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid params"})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	userRepo := repository.NewUserRepository()
	dbUser, err := userRepo.FindByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account verification failed"})
		return
	}

	if !dbUser.TwoFactorEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "2FA is not enabled for this account"})
		return
	}

	valid, _ := totp.ValidateCustom(req.Code, dbUser.TwoFactorSecret, time.Now().UTC(), totp.ValidateOpts{
		Period: 30,
		Skew:   2,
		Digits: 6,
	})
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid code"})
		return
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process password"})
		return
	}

	if err := userRepo.UpdatePassword(dbUser.UID, hashedPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password reset success, please login again"})
}

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

// UpdateProfile 更新个人资料
func UpdateProfile(c *gin.Context) {
	uid := c.GetInt("uid")

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userRepo := repository.NewUserRepository()
	err := userRepo.UpdateProfile(uint(uid), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新资料失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":              "资料更新成功",
		"nickname":             req.Nickname,
		"bio":                  req.Bio,
		"wechat":               req.Wechat,
		"qq":                   req.QQ,
		"show_email":           req.ShowEmail,
		"birthday":             req.Birthday,
		"sound_volume":         req.SoundVolume,
		"vibration_enabled":    req.VibrationEnabled,
		"enable_element_input": req.EnableElementInput,
		"custom_contact":       req.CustomContact,
	})
}

// GetAuthConfig 获取鉴权配置模式
func GetAuthConfig(c *gin.Context) {
	githubEnabled := strings.TrimSpace(os.Getenv("GITHUB_CLIENT_ID")) != "" &&
		strings.TrimSpace(os.Getenv("GITHUB_CLIENT_SECRET")) != ""
	msEnabled := strings.TrimSpace(os.Getenv("MS_CLIENT_ID")) != "" &&
		strings.TrimSpace(os.Getenv("MS_CLIENT_SECRET")) != ""
	googleEnabled := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID")) != "" &&
		strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET")) != ""
	// 当前后端实现要求 APPLE_CLIENT_SECRET（尚未启用私钥动态生成）
	appleEnabled := strings.TrimSpace(os.Getenv("APPLE_CLIENT_ID")) != "" &&
		strings.TrimSpace(os.Getenv("APPLE_CLIENT_SECRET")) != "" &&
		strings.TrimSpace(os.Getenv("APPLE_REDIRECT_URI")) != ""

	c.JSON(http.StatusOK, gin.H{
		"smtp_enabled":   utils.IsSMTPConfigured(),
		"github_enabled": githubEnabled,
		"ms_enabled":     msEnabled,
		"google_enabled": googleEnabled,
		"apple_enabled":  appleEnabled,
	})
}

// SendVerificationCode 发送邮箱验证码
func SendVerificationCode(c *gin.Context) {
	var req models.SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的参数"})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if !utils.IsSMTPConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "系统未启用邮件服务"})
		return
	}

	codeType := req.Type
	if codeType == "" {
		codeType = "register"
	}

	// 限制发送频率
	var latestCode database.VerificationCode
	err := database.DB.Where("email = ? AND created_at > ?", req.Email, time.Now().Add(-1*time.Minute)).Order("created_at desc").First(&latestCode).Error
	if err == nil {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "请求过快，请稍后再试"})
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// 如果是重置密码或修改密码，检查用户是否存在
	if codeType == "reset" || codeType == "change_password" {
		userRepo := repository.NewUserRepository()
		exists, err := userRepo.ExistsByEmail(req.Email)
		if err != nil || !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "该邮箱未注册"})
			return
		}
	}

	// 如果是更换邮箱验证原邮箱，确保是当前用户的邮箱
	if codeType == "change_email_old" {
		// 这里虽然是公开路由，但我们可以尝试从 token 中获取 uid (如果提供了)
		// 或者在发送代码时不做严格检查，但在执行更换时严格检查
		// 为了防止骚扰，这里可以不做额外检查，只要邮箱存在即可
		userRepo := repository.NewUserRepository()
		exists, err := userRepo.ExistsByEmail(req.Email)
		if err != nil || !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "此邮箱未在系统中登记"})
			return
		}
	}

	// 如果是更换邮箱验证新邮箱，确保邮箱未被占用
	if codeType == "change_email_new" {
		userRepo := repository.NewUserRepository()
		exists, err := userRepo.ExistsByEmail(req.Email)
		if err == nil && exists {
			c.JSON(http.StatusConflict, gin.H{"error": "该邮箱已被其他研究员占用"})
			return
		}
	}

	// 如果是注销账号，检查用户是否存在
	if codeType == "delete_account" {
		userRepo := repository.NewUserRepository()
		exists, err := userRepo.ExistsByEmail(req.Email)
		if err != nil || !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "该邮箱未注册"})
			return
		}
	}

	code, err := utils.GenerateVerificationCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成验证码失败"})
		return
	}
	expiresAt := time.Now().Add(10 * time.Minute)

	// 保存到数据库
	vCode := database.VerificationCode{
		Email:     req.Email,
		Code:      code,
		Type:      codeType,
		ExpiresAt: expiresAt,
	}

	if err := database.DB.Create(&vCode).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发送失败，请重试"})
		return
	}

	// 发送邮件
	subject := "ChemistryUNO研究所联络验证码"
	var body string
	if codeType == "reset" {
		subject = "研究所凭证找回"
		body = fmt.Sprintf(`
			<div style="padding: 20px; font-family: sans-serif;">
				<h2>重置您的研究员档案</h2>
				<p>您正在申请重置化学研究所账号的访问凭证，验证码为：</p>
				<h1 style="color: #e11d48; letter-spacing: 5px;">%s</h1>
				<p>该验证码将在 10 分钟后过期。如非本人操作，请立刻检查账号安全。</p>
			</div>
		`, code)
	} else if codeType == "change_password" {
		subject = "研究所安全验证码"
		body = fmt.Sprintf(`
			<div style="padding: 20px; font-family: sans-serif;">
				<h2>安全变更授权</h2>
				<p>您正在尝试修改化学研究所账号的通行密钥，验证码为：</p>
				<h1 style="color: #2563eb; letter-spacing: 5px;">%s</h1>
				<p>为了保障您的数据安全，请在 10 分钟内完成操作。如非本人操作，请忽略此邮件。</p>
			</div>
		`, code)
	} else if codeType == "change_email_old" {
		subject = "变动档案邮箱确认"
		body = fmt.Sprintf(`
			<div style="padding: 20px; font-family: sans-serif;">
				<h2>确认原有的研究通讯地址</h2>
				<p>您正在申请更换您的研究员通讯邮箱。这是对原邮箱的二次验证，验证码为：</p>
				<h1 style="color: #4b5563; letter-spacing: 5px;">%s</h1>
				<p>该验证码将在 10 分钟后过期。如果这不是您的操作，请忽略此邮件。</p>
			</div>
		`, code)
	} else if codeType == "change_email_new" {
		subject = "新档案邮箱验证"
		body = fmt.Sprintf(`
			<div style="padding: 20px; font-family: sans-serif;">
				<h2>验证新的研究通讯地址</h2>
				<p>您正尝试将此邮箱绑定为您的研究员通讯地址。验证码为：</p>
				<h1 style="color: #2563eb; letter-spacing: 5px;">%s</h1>
				<p>该验证码将在 10 分钟后过期。如果这不是您的操作，请忽略此邮件。</p>
			</div>
		`, code)
	} else if codeType == "delete_account" {
		subject = "【危险操作】研究所档案注销请求"
		body = fmt.Sprintf(`
			<div style="padding: 20px; font-family: sans-serif; border: 2px solid #ef4444; border-radius: 8px;">
				<h2 style="color: #ef4444;">档案彻底清除授权</h2>
				<p>我们收到了通过此邮箱注销研究员档案的申请。<strong>注意：此操作不可逆，您的所有实验数据、积分和成就将被永久删除。</strong></p>
				<p>如果不慎误操作，请立即关闭相关页面。若确定要注销，请在页面输入以下验证码：</p>
				<h1 style="color: #ef4444; letter-spacing: 5px; text-align: center;">%s</h1>
				<p>该验证码将在 10 分钟后失效。</p>
				<p style="font-size: 12px; color: #6b7280; margin-top: 20px;">这是一封由系统自动发出的邮件，请勿直接回复。</p>
			</div>
		`, code)
	} else {
		body = fmt.Sprintf(`
			<div style="padding: 20px; font-family: sans-serif;">
				<h2>欢迎加入化学研究所</h2>
				<p>您正在申请建立新的研究员档案，验证码为：</p>
				<h1 style="color: #2563eb; letter-spacing: 5px;">%s</h1>
				<p>该验证码将在 10 分钟后过期。如果这不是您的操作，请忽略此邮件。</p>
			</div>
		`, code)
	}

	go func() {
		if err := utils.SendEmail(req.Email, subject, body); err != nil {
			fmt.Printf("邮件发送失败: %v\n", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "验证码已发送至您的电子邮箱"})
}

// ResetPasswordByEmail 通过邮箱验证码重置密码
func ResetPasswordByEmail(c *gin.Context) {
	var req models.ResetPasswordByEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的参数"})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// 验证验证码
	consumed, err := consumeVerificationCode(req.Email, req.Code, "reset")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if !consumed {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "verification code invalid or expired"})
		return
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	userRepo := repository.NewUserRepository()
	dbUser, err := userRepo.FindByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "账户核验失败，请核对邮箱地址"})
		return
	}

	// 更新密码
	err = userRepo.UpdatePassword(dbUser.UID, hashedPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码更新失败"})
		return
	}

	// 成功后删除验证码

	// 强制登出所有会话
	sessionRepo := repository.NewSessionRepository()
	_ = sessionRepo.DeleteByUserUID(dbUser.UID)

	c.JSON(http.StatusOK, gin.H{"message": "密码重置成功，请重新登录"})
}

// ChangeEmail 更换邮箱地址
func ChangeEmail(c *gin.Context) {
	uid := c.GetInt("uid")
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权的访问"})
		return
	}

	var req models.ChangeEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的参数"})
		return
	}

	req.NewEmail = strings.ToLower(strings.TrimSpace(req.NewEmail))

	userRepo := repository.NewUserRepository()
	dbUser, err := userRepo.FindByUID(uint(uid))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "账户不存在"})
		return
	}

	if dbUser.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "当前账号未绑定邮箱，无法执行更换流程"})
		return
	}

	// 1. 验证原邮箱验证码
	oldConsumed, err := consumeVerificationCode(dbUser.Email, req.OldCode, "change_email_old")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if !oldConsumed {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "verification code invalid or expired"})
		return
	}

	// 2. 验证新邮箱验证码
	newConsumed, err := consumeVerificationCode(req.NewEmail, req.NewCode, "change_email_new")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if !newConsumed {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "verification code invalid or expired"})
		return
	}

	// 3. 检查新邮箱是否被占用
	exists, err := userRepo.ExistsByEmail(req.NewEmail)
	if err == nil && exists {
		c.JSON(http.StatusConflict, gin.H{"error": "新邮箱已被其他研究员占用"})
		return
	}

	// 4. 更新邮箱
	dbUser.Email = req.NewEmail
	err = userRepo.Update(dbUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新邮箱失败"})
		return
	}

	// 成功后删除验证码

	c.JSON(http.StatusOK, gin.H{"message": "邮箱地址已更新"})
}

// 注销账号
func DeleteAccount(c *gin.Context) {
	uid := c.GetInt("uid")

	// 注销账号需要邮箱验证
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供验证码进行验证"})
		return
	}

	userRepo := repository.NewUserRepository()
	user, err := userRepo.FindByUID(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	if utils.IsSMTPConfigured() {
		// 验证码校验
		consumed, err := consumeVerificationCode(user.Email, req.Code, "delete_account")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		if !consumed {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "verification code invalid or expired"})
			return
		}
	}

	err = userRepo.Delete(uint(uid))
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
	sessions, err := sessionRepo.FindByUserUID(uint(uid.(int)))
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
	err := sessionRepo.DeleteByIDAndUserUID(req.ID, uint(uid))
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
	_ = sessionRepo.DeleteByUserUID(uint(uid))

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("账号已冻结，直到 %s", frozenUntil.Format("2006-01-02 15:04:05"))})
}

// 获取用户信息
func GetUserInfo(c *gin.Context) {
	uid := c.GetInt("uid")

	if uid < 0 {
		c.JSON(http.StatusOK, gin.H{
			"uid":      uid,
			"nickname": "AI研究员",
			"role":     "ai",
			"points":   1000,
		})
		return
	}

	userRepo := repository.NewUserRepository()
	user, err := userRepo.FindByUID(uint(uid))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在，请重新登录"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetUserProfile 获取公开个人空间资料
func GetUserProfile(c *gin.Context) {
	uidStr := c.Param("uid")
	uid, err := strconv.ParseUint(uidStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	userRepo := repository.NewUserRepository()
	user, err := userRepo.FindByUID(uint(uid))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 屏蔽敏感信息，仅返回公开字段
	email := ""
	if user.ShowEmail {
		email = user.Email
	}

	c.JSON(http.StatusOK, gin.H{
		"uid":            user.UID,
		"nickname":       user.Nickname,
		"avatar":         user.Avatar,
		"role":           user.Role,
		"bio":            user.Bio,
		"wechat":         user.Wechat,
		"qq":             user.QQ,
		"email":          email,
		"show_email":     user.ShowEmail,
		"birthday":       user.Birthday,
		"custom_contact": user.CustomContact,
		"points":         user.Points,
		"level":          user.Level,
		"win_count":      user.WinCount,
		"total_games":    user.TotalGames,
		"created_at":     user.CreatedAt,
	})
}

// SearchUsers 搜索用户
func SearchUsers(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		query = strings.TrimSpace(c.Query("query"))
	}

	if query == "" {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	// 预先获取等级配置以避免循环查询
	var levelConfigs []database.LevelConfig
	if err := database.DB.Find(&levelConfigs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	configMap := make(map[int]database.LevelConfig)
	for _, conf := range levelConfigs {
		configMap[conf.Level] = conf
	}

	userRepo := repository.NewUserRepository()
	users, err := userRepo.SearchUsers(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索失败"})
		return
	}

	result := []map[string]interface{}{}
	for _, user := range users {
		totalBounty, _ := repository.BountyRepo.GetTotalBounty(user.UID)
		isOnline := false
		if websocket.GlobalHub != nil {
			isOnline = websocket.GlobalHub.IsUIDOnline(int(user.UID))
		}

		// 计算排名
		var rank int64
		if err := database.DB.Model(&database.User{}).Where("points > ?", user.Points).Count(&rank).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		var monthlyRank int64
		if err := database.DB.Model(&database.User{}).Where("monthly_points > ?", user.MonthlyPoints).Count(&monthlyRank).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		conf := configMap[user.Level]
		result = append(result, map[string]interface{}{
			"uid":             user.UID,
			"username":        user.Username,
			"nickname":        user.Nickname,
			"avatar":          user.Avatar,
			"points":          user.Points,
			"monthly_points":  user.MonthlyPoints,
			"win_count":       user.WinCount,
			"total_games":     user.TotalGames,
			"level":           user.Level,
			"tier":            conf.Tier,
			"tier_name":       conf.TierName,
			"bounty":          totalBounty,
			"is_online":       isOnline,
			"rank":            rank + 1,
			"monthly_rank":    monthlyRank + 1,
			"last_offline_at": user.LastOfflineAt,
		})
	}

	c.JSON(http.StatusOK, result)
}
