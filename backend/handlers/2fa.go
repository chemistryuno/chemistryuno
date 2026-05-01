package handlers

import (
	"chemistryuno/backend/models"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/utils"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

// Setup2FA 生成2FA密钥和二维码URL
func Setup2FA(c *gin.Context) {
	uid := c.GetInt("uid")
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}

	// 获取用户邮箱或昵称用于 TOTP 标识
	user, err := repository.UserRepo.FindByUID(uint(uid))
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	accountName := user.Email
	if accountName == "" {
		accountName = user.Nickname
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "ChemistryUno",
		AccountName: accountName,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成2FA密钥失败"})
		return
	}

	// 生成二维码图片 (Base64)
	var png []byte
	png, err = qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成二维码失败"})
		return
	}
	qrCodeBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)

	// 临时保存密钥到数据库，但不启用
	userRepo := repository.NewUserRepository()
	err = userRepo.Update2FASecret(uint(uid), key.Secret())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存2FA密钥失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"secret":  key.Secret(),
		"url":     key.URL(),
		"qr_code": qrCodeBase64,
	})
}

// Enable2FA 验证并正式启用2FA
func Enable2FA(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		Code     string `json:"code" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数缺失或格式错误"})
		return
	}

	userRepo := repository.NewUserRepository()
	user, err := userRepo.FindByUID(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}
	userPassword := user.Password
	secret := user.TwoFactorSecret

	// 开启2FA时必须验证当前密码
	if !utils.CheckPassword(req.Password, userPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误，身份核验失败"})
		return
	}

	if secret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先获取2FA设置二维码"})
		return
	}

	// 允许 +/- 2 个步长的误差 (30秒 * 5 = 150秒窗口)
	valid, _ := totp.ValidateCustom(req.Code, secret, time.Now().UTC(), totp.ValidateOpts{
		Period: 30,
		Skew:   2,
		Digits: 6,
	})

	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "验证码无效，请确保手机时间同步"})
		return
	}

	err = userRepo.Enable2FA(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新2FA状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA启用成功"})
}

// Disable2FA 禁用2FA
func Disable2FA(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证码参数错误"})
		return
	}

	userRepo := repository.NewUserRepository()
	user, err := userRepo.FindByUID(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取2FA密钥失败"})
		return
	}

	valid, _ := totp.ValidateCustom(req.Code, user.TwoFactorSecret, time.Now().UTC(), totp.ValidateOpts{
		Period: 30,
		Skew:   2,
		Digits: 6,
	})

	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "验证码无效"})
		return
	}

	err = userRepo.Disable2FA(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA已禁用"})
}

// Verify2FALogin 登录时的2FA验证
func Verify2FALogin(c *gin.Context) {
	var req struct {
		UID      int    `json:"uid" binding:"required"`
		Code     string `json:"code" binding:"required"`
		Password string `json:"password" binding:"required"` // 添加密码验证
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	userRepo := repository.NewUserRepository()
	dbUser, err := userRepo.FindByUID(uint(req.UID))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		return
	}

	user := models.User{
		UID:              int(dbUser.UID),
		Username:         dbUser.Username,
		Email:            dbUser.Email,
		PasswordHash:     dbUser.Password,
		Avatar:           dbUser.Avatar,
		Role:             models.NormalizeRole(dbUser.Role),
		TwoFactorEnabled: dbUser.TwoFactorEnabled,
		TwoFactorSecret:  dbUser.TwoFactorSecret,
		BannedUntil:      dbUser.BannedUntil,
		FrozenUntil:      dbUser.FrozenUntil,
	}

	// 安全检查1: 验证用户确实启用了2FA
	if !user.TwoFactorEnabled {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "该账户未启用2FA，请使用常规登录"})
		return
	}

	// 安全检查2: 验证密码
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
		return
	}

	// 封禁状态已不再阻止登录
	now := time.Now()
	// 检查冻结状态
	if user.FrozenUntil != nil && now.Before(*user.FrozenUntil) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": fmt.Sprintf("您的账号当前处于冷冻状态，直到 %s", user.FrozenUntil.Format("2006-01-02 15:04:05")),
		})
		return
	}

	// 验证2FA验证码
	valid, _ := totp.ValidateCustom(req.Code, user.TwoFactorSecret, time.Now().UTC(), totp.ValidateOpts{
		Period: 30,
		Skew:   2,
		Digits: 6,
	})

	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "验证码无效"})
		return
	}

	// 生成会话
	sid, err := utils.CreateSession(user.UID, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil || sid == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建会话失败"})
		return
	}

	// 生成access token（15分钟）和refresh token（7天）
	accessToken, err := utils.GenerateAccessToken(int(user.UID), user.Email, user.Role, sid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成access token失败"})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(int(user.UID), sid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成refresh token失败"})
		return
	}

	// 获取当前可用公告
	announcementRepo := repository.NewAnnouncementRepository()
	dbAnnouncements, _ := announcementRepo.FindActive()
	var announcements []models.Announcement
	for _, a := range dbAnnouncements {
		announcements = append(announcements, models.Announcement{
			ID:       int(a.ID),
			Title:    a.Title,
			Content:  a.Content,
			Type:     a.Type,
			IsTicker: a.IsTicker,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    900,
		"user": gin.H{
			"uid":      user.UID,
			"username": user.Username,
			"email":    user.Email,
			"avatar":   user.Avatar,
			"is_admin": user.HasAdminAccess(),
			"role":     user.NormalizedRole(),
		},
		"announcements": announcements,
	})
}
