package handlers

import (
	"chemistryuno/database"
	"chemistryuno/models"
	"chemistryuno/utils"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

// Setup2FA 生成2FA密钥和二维码URL
func Setup2FA(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}
	username := c.GetString("username")

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "ChemistryUno",
		AccountName: username,
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
	_, err = database.DB.Exec("UPDATE users SET two_factor_secret = ? WHERE UID = ?", key.Secret(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存2FA密钥到数据库失败"})
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
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证码格式错误"})
		return
	}

	var secret string
	err := database.DB.QueryRow("SELECT two_factor_secret FROM users WHERE UID = ?", uid).Scan(&secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}

	valid := totp.Validate(req.Code, secret)
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "验证码无效"})
		return
	}

	_, err = database.DB.Exec("UPDATE users SET two_factor_enabled = 1 WHERE UID = ?", uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
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

	var secret string
	err := database.DB.QueryRow("SELECT two_factor_secret FROM users WHERE UID = ?", uid).Scan(&secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}

	valid := totp.Validate(req.Code, secret)
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "验证码无效"})
		return
	}

	_, err = database.DB.Exec("UPDATE users SET two_factor_enabled = 0, two_factor_secret = '' WHERE UID = ?", uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA已禁用"})
}

// Verify2FALogin 登录时的2FA验证
func Verify2FALogin(c *gin.Context) {
	var req struct {
		UID  int    `json:"uid" binding:"required"`
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	var user models.User
	err := database.DB.QueryRow(
		"SELECT UID, username, avatar, is_admin, role, two_factor_secret FROM users WHERE UID = ?",
		req.UID,
	).Scan(&user.UID, &user.Username, &user.Avatar, &user.IsAdmin, &user.Role, &user.TwoFactorSecret)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		return
	}

	valid := totp.Validate(req.Code, user.TwoFactorSecret)
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "验证码无效"})
		return
	}

	// 生成token
	token, err := utils.GenerateToken(int(user.UID), user.Username, user.IsAdmin, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"uid":      user.UID,
			"username": user.Username,
			"avatar":   user.Avatar,
			"is_admin": user.IsAdmin,
			"role":     user.Role,
		},
	})
}
