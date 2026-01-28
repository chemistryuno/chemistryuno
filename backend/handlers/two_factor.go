package handlers

import (
	"chemistryuno/database"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
)

// Setup2FA 生成2FA密钥和二维码URL
func Setup2FA(c *gin.Context) {
	uid := c.GetInt("uid")
	username := c.GetString("username")

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "ChemistryUno",
		AccountName: username,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成2FA密钥失败"})
		return
	}

	// 临时保存secret，直到验证成功
	_, err = database.DB.Exec("UPDATE users SET two_factor_secret = ? WHERE UID = ?", key.Secret(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存2FA密钥失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"secret":  key.Secret(),
		"qr_code": key.URL(),
	})
}

// Verify2FA 验证并启用2FA
func Verify2FA(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证码不能为空"})
		return
	}

	var secret string
	err := database.DB.QueryRow("SELECT two_factor_secret FROM users WHERE UID = ?", uid).Scan(&secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}

	if !totp.Validate(req.Code, secret) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "验证码错误"})
		return
	}

	_, err = database.DB.Exec("UPDATE users SET two_factor_enabled = 1 WHERE UID = ?", uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新2FA状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA已成功启用"})
}

// Disable2FA 禁用2FA
func Disable2FA(c *gin.Context) {
	uid := c.GetInt("uid")

	_, err := database.DB.Exec("UPDATE users SET two_factor_enabled = 0, two_factor_secret = '' WHERE UID = ?", uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "禁用2FA失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA已禁用"})
}
