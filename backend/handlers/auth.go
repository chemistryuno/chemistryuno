package handlers

import (
	"chemistryuno/database"
	"chemistryuno/models"
	"chemistryuno/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 发送验证码
func SendVerificationCode(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Type  string `json:"type" binding:"required"` // register, login, reset
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱格式错误"})
		return
	}

	code := utils.GenerateCode()
	err := utils.SaveVerificationCode(req.Email, code, req.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存验证码失败"})
		return
	}

	err = utils.SendEmailCode(req.Email, code, req.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发送验证码失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "验证码已发送"})
}

// 找回密码（使用验证码重置）
func ResetPassword(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Code     string `json:"code" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 校验验证码
	if !utils.VerifyEmailCode(req.Email, req.Code, "reset") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "验证码无效或已过期"})
		return
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 更新密码
	_, err = database.DB.Exec("UPDATE users SET password = ? WHERE email = ?", hashedPassword, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重置密码失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码已重置，请重新登录"})
}

// 用户注册
func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. 校验验证码
	if !utils.VerifyEmailCode(req.Email, req.Code, "register") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "验证码无效或已过期"})
		return
	}

	// 2. 检查用户名或邮箱是否已存在
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? OR email = ?", req.Username, req.Email).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "用户名或邮箱已存在"})
		return
	}

	// 3. 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 4. 插入用户
	result, err := database.DB.Exec("INSERT INTO users (username, email, password, avatar, role) VALUES (?, ?, ?, ?, ?)",
		req.Username, req.Email, hashedPassword, "🧪", "user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	userUID, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{
		"message": "注册成功",
		"uid":     userUID,
	})
}

// 用户登录
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查询用户（支持用户名或邮箱登录）
	var user models.User
	err := database.DB.QueryRow(
		"SELECT UID, username, COALESCE(email, ''), password, avatar, is_admin, role, two_factor_enabled, two_factor_secret FROM users WHERE username = ? OR email = ?",
		req.Username, req.Username,
	).Scan(&user.UID, &user.Username, &user.Email, &user.PasswordHash, &user.Avatar, &user.IsAdmin, &user.Role, &user.TwoFactorEnabled, &user.TwoFactorSecret)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		return
	}

	// 验证逻辑
	if req.Method == "code" {
		// 验证码登录
		if !utils.VerifyEmailCode(user.Email, req.Code, "login") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "验证码无效或已过期"})
			return
		}
	} else {
		// 密码登录
		if !utils.CheckPassword(req.Password, user.PasswordHash) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
			return
		}
	}

	// 如果开启了2FA，且请求中没有验证码，且不是验证码直接登录（或者我们规定验证码登录也需要2FA，这里简单处理：验证码登录优先级高于2FA）
	if user.TwoFactorEnabled && req.Method != "code" {
		c.JSON(http.StatusOK, gin.H{
			"two_factor_required": true,
			"uid":                 user.UID,
		})
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
			"email":    user.Email,
			"avatar":   user.Avatar,
			"is_admin": user.IsAdmin,
			"role":     user.Role,
		},
	})
}

// 修改密码
func ChangePassword(c *gin.Context) {
	uid := c.GetInt("uid")

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前密码
	var currentPassword string
	err := database.DB.QueryRow("SELECT password FROM users WHERE UID = ?", uid).Scan(&currentPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}

	// 验证旧密码
	if !utils.CheckPassword(req.OldPassword, currentPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "旧密码错误"})
		return
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 更新密码
	_, err = database.DB.Exec("UPDATE users SET password = ? WHERE UID = ?", hashedPassword, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新密码失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// 更新头像
func UpdateAvatar(c *gin.Context) {
	uid := c.GetInt("uid")

	var req models.UpdateAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := database.DB.Exec("UPDATE users SET avatar = ? WHERE UID = ?", req.Avatar, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新头像失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "头像更新成功", "avatar": req.Avatar})
}

// 注销账号
func DeleteAccount(c *gin.Context) {
	uid := c.GetInt("uid")

	// 删除用户
	_, err := database.DB.Exec("DELETE FROM users WHERE UID = ?", uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注销账号失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "账号已注销"})
}

// 获取用户信息
func GetUserInfo(c *gin.Context) {
	uid := c.GetInt("uid")

	var user models.User
	err := database.DB.QueryRow(
		"SELECT UID, username, COALESCE(email, ''), avatar, is_admin, role, two_factor_enabled, created_at FROM users WHERE UID = ?",
		uid,
	).Scan(&user.UID, &user.Username, &user.Email, &user.Avatar, &user.IsAdmin, &user.Role, &user.TwoFactorEnabled, &user.CreatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, user)
}
