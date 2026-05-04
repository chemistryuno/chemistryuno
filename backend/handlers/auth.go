package handlers

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/middleware"
	"chemistryuno/backend/models"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/utils"
	"chemistryuno/backend/websocket"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

// setSecureAuthCookie stores auth tokens in HttpOnly cookies.
func setSecureAuthCookie(c *gin.Context, name string, value string, maxAge int) {
	secure := c.Request.URL.Scheme == "https" || c.Request.Header.Get("X-Forwarded-Proto") == "https"
	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteNoneMode
	}

	c.SetSameSite(sameSite)
	c.SetCookie(name, value, maxAge, "/", "", secure, true)
}

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

// usernameRegex 用户名只允许英文字母、数字和下划线
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// nicknameRegex 昵称只允许中英文字母、数字和下划线
var nicknameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\x{4e00}-\x{9fa5}]+$`)

// 用户注册
func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证用户名格式
	if !usernameRegex.MatchString(req.Username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名只能包含英文字母、数字和下划线"})
		return
	}

	// 验证昵称格式和长度
	if req.Nickname != "" {
		if len([]rune(req.Nickname)) > 20 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "昵称不能超过20个字符"})
			return
		}
		if !nicknameRegex.MatchString(req.Nickname) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "昵称只能包含中英文字母、数字和下划线"})
			return
		}
	}

	userRepo := repository.NewUserRepository()

	// 检查用户名是否已存在
	exists, err := userRepo.ExistsByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
		return
	}

	// 处理可选邮箱
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email != "" {
		// 验证邮箱格式
		if !strings.Contains(req.Email, "@") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email format"})
			return
		}
		// 检查邮箱唯一性
		emailExists, err := userRepo.ExistsByEmail(req.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		if emailExists {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		// 如果启用SMTP且提供了邮箱，验证邮箱验证码
		if utils.IsSMTPConfigured() {
			if req.Code == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "email verification code is required when email is provided"})
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
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// 邮箱注册无需密保
	hashedAnswer := ""
	if req.Email == "" {
		if req.SecurityQuestion == "" || req.SecurityAnswer == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "security question and answer are required for username registration"})
			return
		}
		// 哈希密保答案
		hashedAnswer, err = utils.HashPassword(req.SecurityAnswer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process security answer"})
			return
		}
	}

	// 获取最大 UID 并计算新 UID
	maxUID, err := userRepo.GetMaxUID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error when fetching max uid"})
		return
	}
	newUID := maxUID + 1
	if newUID < 100000000 {
		newUID = 100000000
	}

	// 如果昵称为空，生成随机昵称并确保在数据库中不存在（尝试若干次）
	if strings.TrimSpace(req.Nickname) == "" {
		nickname, err := utils.GenerateUniqueRandomNickname("研究员", req.Username, userRepo.ExistsByNickname)
		if err == nil {
			req.Nickname = nickname
		} else {
			// 最后回退到用户名作为昵称
			req.Nickname = req.Username
		}
	}

	user := &database.User{
		UID:              newUID,
		Username:         req.Username,
		Email:            req.Email,
		Nickname:         req.Nickname,
		Password:         hashedPassword,
		Avatar:           "\U0001F9EA",
		Role:             "user",
		Points:           1000,
		MonthlyPoints:    1000,
		Level:            1,
		SecurityQuestion: req.SecurityQuestion,
		SecurityAnswer:   hashedAnswer,
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

	identifier := strings.TrimSpace(req.Identifier)
	log.Println("🔍 登录尝试")
	log.Printf("   账户: %s", identifier)
	log.Printf("   IP: %s", c.ClientIP())

	userRepo := repository.NewUserRepository()

	// 自动判断是邮箱还是用户名
	var dbUser *database.User
	var findErr error
	if strings.Contains(identifier, "@") {
		email := strings.ToLower(identifier)
		dbUser, findErr = userRepo.FindByEmail(email)
	} else {
		dbUser, findErr = userRepo.FindByUsername(identifier)
	}
	if findErr != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	user := models.User{
		UID:              int(dbUser.UID),
		Username:         dbUser.Username,
		Email:            dbUser.Email,
		Nickname:         dbUser.Nickname,
		PasswordHash:     dbUser.Password,
		Avatar:           dbUser.Avatar,
		Role:             models.NormalizeRole(dbUser.Role),
		TwoFactorEnabled: dbUser.TwoFactorEnabled,
		TwoFactorSecret:  dbUser.TwoFactorSecret,
		BannedUntil:      dbUser.BannedUntil,
		FrozenUntil:      dbUser.FrozenUntil,
	}

	now := time.Now()
	// 封禁用户不再被禁止登录
	// if user.BannedUntil != nil && now.Before(*user.BannedUntil) {
	// ...
	// }

	if user.FrozenUntil != nil && now.Before(*user.FrozenUntil) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":        "account is temporarily frozen",
			"frozen_until": user.FrozenUntil.Format(time.RFC3339),
		})
		return
	}

	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	if user.UID < 0 {
		log.Printf("❌ AI 账户不能登录 (UID=%d, IP=%s)", user.UID, c.ClientIP())
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

	// 生成access token（15分钟）和refresh token（7天）
	accessToken, err := utils.GenerateAccessToken(int(user.UID), user.Email, user.Role, sid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate access token"})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(int(user.UID), sid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate refresh token"})
		return
	}

	// 检测异常登录
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	anomalousLoginDetector := middleware.GetGlobalAnomalousLoginDetector()
	anomalies := anomalousLoginDetector.CheckAnomalousLogin(user.UID, clientIP, userAgent)

	// 如果检测到异常，记录反馈
	if len(anomalies) > 0 {
		loginRecorder := repository.NewAnomalousLoginRecorder(database.DB)
		go func() {
			if err := loginRecorder.RecordAnomalousLogin(uint(user.UID), clientIP, anomalies); err != nil {
				log.Printf("⚠️ 记录异常登录失败: %v", err)
			}
		}()
		log.Printf("⚠️ 检测到异常登录 UID=%d: %v", user.UID, anomalies)
	}

	isReturningPlayer := false
	daysSinceLastLogin := 0
	if dbUser.LastOfflineAt != nil {
		daysSince := time.Since(*dbUser.LastOfflineAt).Hours() / 24
		daysSinceLastLogin = int(daysSince)
		if daysSince >= 30 {
			isReturningPlayer = true
			log.Printf("🔁 用户回归: %s (UID=%d), %d 天未畛线", user.Nickname, user.UID, daysSinceLastLogin)
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

	log.Println("✅ 登录成功")
	log.Printf("   账户: %s (UID=%d)", user.Username, user.UID)
	log.Printf("   SID: %s", sid)
	log.Printf("   IP: %s", clientIP)

	// 设置安全的HttpOnly Cookie存储token
	setSecureAuthCookie(c, "access_token", accessToken, 900)         // 15分钟
	setSecureAuthCookie(c, "refresh_token", refreshToken, 7*24*3600) // 7天

	c.JSON(http.StatusOK, gin.H{
		"token_type": "Bearer",
		"expires_in": 900, // 900秒 = 15分钟
		"user": gin.H{
			"uid":                user.UID,
			"username":           user.Username,
			"email":              user.Email,
			"nickname":           user.Nickname,
			"avatar":             user.Avatar,
			"is_admin":           user.HasAdminAccess(),
			"role":               user.NormalizedRole(),
			"two_factor_enabled": user.TwoFactorEnabled,
		},
		"announcements":         announcements,
		"is_returning_player":   isReturningPlayer,
		"days_since_last_login": daysSinceLastLogin,
	})
}

// RefreshToken 使用refresh token获取新的access token
func RefreshToken(c *gin.Context) {
	// 从cookie中读取refresh_token
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing refresh_token in cookie"})
		return
	}

	// 解析refresh token
	claims, err := utils.ParseToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh_token"})
		return
	}

	// 检查token类型
	if claims.TokenType != "refresh" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token type, expected refresh token"})
		return
	}

	// 检查SID是否有效
	if claims.SID == "" || !utils.IsSessionValid(claims.SID) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session expired, please login again"})
		return
	}

	// 验证session属于该用户
	if !utils.ValidateSessionForUser(claims.SID, claims.UID) {
		log.Printf("❌ 会话验证失败: UID=%d, SID=%s", claims.UID, claims.SID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session validation failed"})
		return
	}

	// 获取用户信息以获得最新的email, role等
	userRepo := repository.NewUserRepository()
	user, err := userRepo.FindByUID(uint(claims.UID))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	// 检查账户状态
	now := time.Now()
	// 注意：封禁用户不再阻断刷新 token，这样他们可以继续留在系统中
	if user.FrozenUntil != nil && now.Before(*user.FrozenUntil) {
		c.JSON(http.StatusForbidden, gin.H{"error": "account is frozen"})
		return
	}

	// 生成新的access token
	accessToken, err := utils.GenerateAccessToken(claims.UID, user.Email, models.NormalizeRole(user.Role), claims.SID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate access token"})
		return
	}

	// 设置新的access_token cookie（安全属性）
	setSecureAuthCookie(c, "access_token", accessToken, 900)

	c.JSON(http.StatusOK, gin.H{
		"token_type": "Bearer",
		"expires_in": 900, // 900秒 = 15分钟
	})
}

func ChangePassword(c *gin.Context) {
	uid := c.GetInt("uid")

	var req struct {
		Code           string `json:"code"`         // 2FA 或 邮箱验证码
		OldPassword    string `json:"old_password"` // 旧密码
		NewPassword    string `json:"new_password" binding:"required,min=6"`
		UseEmail       bool   `json:"use_email"`       // 是否使用邮箱验证模式
		SecurityAnswer string `json:"security_answer"` // 密保答案（无2FA且无邮箱时）
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

	// 验证逻辑优先顺序: 2FA > 邮箱验证码 > 密保问题 > 旧密码
	if user.TwoFactorEnabled {
		// 开启了2FA，强制使用2FA
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
	} else if req.UseEmail && user.Email != "" {
		// 使用邮箱验证码
		if req.Code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请提供邮箱验证码"})
			return
		}
		consumed, err := consumeVerificationCode(user.Email, req.Code, "change_password")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		if !consumed {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "verification code invalid or expired"})
			return
		}
	} else if user.Email == "" && user.SecurityQuestion != "" {
		// 无邮箱时，使用密保问题验证
		if req.SecurityAnswer == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":                     "请提供密保答案以授权密码修改",
				"require_security_question": true,
				"security_question":         user.SecurityQuestion,
			})
			return
		}
		if !utils.CheckPassword(req.SecurityAnswer, user.SecurityAnswer) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "密保答案错误"})
			return
		}
	} else {
		// 传统模式：要求旧密码
		if req.OldPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请提供当前密码"})
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

	// 验证昵称格式和长度
	if req.Nickname != "" {
		if len([]rune(req.Nickname)) > 20 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "昵称不能超过20个字符"})
			return
		}
		if !nicknameRegex.MatchString(req.Nickname) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "昵称仅允许中英文字母、数字和下划线"})
			return
		}
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
			log.Printf("❌ 邮件发送失败: %v", err)
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

	// 注销账号需要验证
	var req struct {
		Code           string `json:"code"`            // 邮箱验证码
		SecurityAnswer string `json:"security_answer"` // 密保答案（无邮箱时）
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供验证信息"})
		return
	}

	userRepo := repository.NewUserRepository()
	user, err := userRepo.FindByUID(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	if utils.IsSMTPConfigured() && user.Email != "" {
		// 有邮箱时使用邮箱验证码
		if req.Code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请提供邮箱验证码"})
			return
		}
		consumed, err := consumeVerificationCode(user.Email, req.Code, "delete_account")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		if !consumed {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "verification code invalid or expired"})
			return
		}
	} else if user.Email == "" && user.SecurityQuestion != "" {
		// 无邮箱时使用密保问题
		if req.SecurityAnswer == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":                     "请提供密保答案以确认注销",
				"require_security_question": true,
				"security_question":         user.SecurityQuestion,
			})
			return
		}
		if !utils.CheckPassword(req.SecurityAnswer, user.SecurityAnswer) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "密保答案错误"})
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
		log.Printf("❌ 查询数据库失败: %v", err)
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
		"banned_until":   user.BannedUntil,
		"ban_reason":     user.BanReason,
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

// GetSecurityQuestion 获取指定用户的密保问题（用于忘记密码流程，不需要登录）
func GetSecurityQuestion(c *gin.Context) {
	username := strings.TrimSpace(c.Query("username"))
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	userRepo := repository.NewUserRepository()
	var dbUser *database.User
	var err error
	if strings.Contains(username, "@") {
		dbUser, err = userRepo.FindByEmail(strings.ToLower(username))
	} else {
		dbUser, err = userRepo.FindByUsername(username)
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"has_security_question": false})
		return
	}

	if dbUser.SecurityQuestion == "" {
		c.JSON(http.StatusOK, gin.H{"has_security_question": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"has_security_question": true,
		"security_question":     dbUser.SecurityQuestion,
	})
}

// ResetPasswordBySecurityQuestion 通过密保问题重置密码（无邮箱忘记密码时使用）
func ResetPasswordBySecurityQuestion(c *gin.Context) {
	var req models.ResetPasswordBySecurityQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid params"})
		return
	}

	userRepo := repository.NewUserRepository()
	dbUser, err := userRepo.FindByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account verification failed"})
		return
	}

	if dbUser.SecurityQuestion == "" || dbUser.SecurityAnswer == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "this account has no security question set"})
		return
	}

	if !utils.CheckPassword(req.SecurityAnswer, dbUser.SecurityAnswer) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect security answer"})
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

	sessionRepo := repository.NewSessionRepository()
	_ = sessionRepo.DeleteByUserUID(dbUser.UID)

	c.JSON(http.StatusOK, gin.H{"message": "password reset success, please login again"})
}

// GetMySecurityQuestion 获取当前已登录用户的密保问题及账户安全状态
func GetMySecurityQuestion(c *gin.Context) {
	uid := c.GetInt("uid")

	userRepo := repository.NewUserRepository()
	user, err := userRepo.FindByUID(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"security_question":     user.SecurityQuestion,
		"has_security_question": user.SecurityQuestion != "",
		"has_email":             user.Email != "",
	})
}

// UpdateSecurityQuestion 更新当前用户的密保问题和答案
func UpdateSecurityQuestion(c *gin.Context) {
	uid := c.GetInt("uid")

	var req models.UpdateSecurityQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userRepo := repository.NewUserRepository()
	user, err := userRepo.FindByUID(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if req.CurrentPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "current password is required to update security question"})
		return
	}
	if !utils.CheckPassword(req.CurrentPassword, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect current password"})
		return
	}

	hashedAnswer, err := utils.HashPassword(req.SecurityAnswer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process security answer"})
		return
	}

	if err := userRepo.UpdateSecurityQuestion(uint(uid), req.SecurityQuestion, hashedAnswer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update security question"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "security question updated successfully"})
}

// SetEmail 为无邮箱用户设置首个邮箱地址
func SetEmail(c *gin.Context) {
	uid := c.GetInt("uid")

	var req models.SetEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.NewEmail = strings.ToLower(strings.TrimSpace(req.NewEmail))

	userRepo := repository.NewUserRepository()
	user, err := userRepo.FindByUID(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if user.Email != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account already has an email, use change-email instead"})
		return
	}

	// 无2FA时，若有密保问题则需验证
	if !user.TwoFactorEnabled && user.SecurityQuestion != "" {
		if req.SecurityAnswer == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":             "security answer required",
				"security_question": user.SecurityQuestion,
			})
			return
		}
		if !utils.CheckPassword(req.SecurityAnswer, user.SecurityAnswer) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect security answer"})
			return
		}
	}

	emailExists, err := userRepo.ExistsByEmail(req.NewEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if emailExists {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered by another account"})
		return
	}

	if utils.IsSMTPConfigured() {
		consumed, err := consumeVerificationCode(req.NewEmail, req.NewCode, "change_email_new")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		if !consumed {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "verification code invalid or expired"})
			return
		}
	}

	if err := userRepo.UpdateEmail(uint(uid), req.NewEmail); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "email set successfully", "email": req.NewEmail})
}
