package handlers

import (
	"chemistryuno/database"
	"chemistryuno/models"
	"chemistryuno/utils"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
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

	// 1. 检查用户名是否已存在
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

	// 2. 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 3. 插入用户
	result, err := database.DB.Exec("INSERT INTO users (username, password, avatar, role) VALUES (?, ?, ?, ?)",
		req.Username, hashedPassword, "🧪", "user")
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
	fmt.Printf("登录尝试 - 用户: %s, 来源IP: %s\n", req.Username, c.ClientIP())

	// 查询用户
	var user models.User
	err := database.DB.QueryRow(
		"SELECT UID, username, password, avatar, is_admin, role, two_factor_enabled, two_factor_secret, banned_until, frozen_until FROM users WHERE username = ?",
		req.Username,
	).Scan(&user.UID, &user.Username, &user.PasswordHash, &user.Avatar, &user.IsAdmin, &user.Role, &user.TwoFactorEnabled, &user.TwoFactorSecret, &user.BannedUntil, &user.FrozenUntil)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		return
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
	sid, _ := utils.CreateSession(user.UID, c.GetHeader("User-Agent"), c.ClientIP())

	// 生成token
	token, err := utils.GenerateToken(int(user.UID), user.Username, user.IsAdmin, user.Role, sid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	// 4. 获取当前可用公告 (登陆触发器)
	var announcements []models.Announcement
	rows, _ := database.DB.Query(`
		SELECT id, title, content, type, is_ticker, close_delay FROM announcements 
		WHERE active = 1 AND (expires_at IS NULL OR expires_at > ?)`, time.Now())
	if rows != nil {
		for rows.Next() {
			var a models.Announcement
			var title sql.NullString
			if err := rows.Scan(&a.ID, &title, &a.Content, &a.Type, &a.IsTicker, &a.CloseDelay); err == nil {
				if title.Valid {
					a.Title = title.String
				}
				announcements = append(announcements, a)
			}
		}
		rows.Close()
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

	// 获取用户信息
	var user models.User
	err := database.DB.QueryRow("SELECT password, two_factor_enabled, two_factor_secret FROM users WHERE UID = ?", uid).Scan(&user.PasswordHash, &user.TwoFactorEnabled, &user.TwoFactorSecret)
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
		if !utils.CheckPassword(req.OldPassword, user.PasswordHash) {
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
	_, err = database.DB.Exec("UPDATE users SET password = ? WHERE UID = ?", hashedPassword, uid)
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

	var user models.User
	err := database.DB.QueryRow(
		"SELECT UID, two_factor_enabled, two_factor_secret FROM users WHERE username = ?",
		req.Username,
	).Scan(&user.UID, &user.TwoFactorEnabled, &user.TwoFactorSecret)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "账户不存在"})
		return
	}

	if !user.TwoFactorEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该账户未开启 2FA，无法通过此方式找回。请联系管理员。"})
		return
	}

	// 验证 2FA 码
	valid, _ := totp.ValidateCustom(req.Code, user.TwoFactorSecret, time.Now().UTC(), totp.ValidateOpts{
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
	_, err = database.DB.Exec("UPDATE users SET password = ? WHERE UID = ?", hashedPassword, user.UID)
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

	rows, err := database.DB.Query(`
		SELECT id, user_agent, ip_address, 
		       COALESCE(last_active, NOW()) as last_active, 
		       COALESCE(created_at, NOW()) as created_at 
		FROM user_sessions 
		WHERE user_uid = ? 
		ORDER BY last_active DESC`, uid)
	if err != nil {
		fmt.Printf("查询数据库失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法加载设备列表"})
		return
	}
	defer rows.Close()

	var sessions []gin.H
	for rows.Next() {
		var s struct {
			ID         string
			UA         sql.NullString
			IP         sql.NullString
			LastActive string
			CreatedAt  string
		}
		if err := rows.Scan(&s.ID, &s.UA, &s.IP, &s.LastActive, &s.CreatedAt); err != nil {
			fmt.Printf("扫描会话行失败: %v\n", err)
			continue
		}

		// 处理时间格式
		lastActive := strings.Replace(s.LastActive, " ", "T", 1)
		if !strings.HasSuffix(lastActive, "Z") {
			lastActive += "Z"
		}
		createdAt := strings.Replace(s.CreatedAt, " ", "T", 1)
		if !strings.HasSuffix(createdAt, "Z") {
			createdAt += "Z"
		}

		sessions = append(sessions, gin.H{
			"id":          s.ID,
			"user_agent":  s.UA.String,
			"ip":          s.IP.String,
			"last_active": lastActive,
			"created_at":  createdAt,
			"is_current":  s.ID != "" && s.ID == currentSID,
		})
	}

	c.JSON(http.StatusOK, sessions)
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

	_, err := database.DB.Exec("DELETE FROM user_sessions WHERE id = ? AND user_uid = ?", req.ID, uid)
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
	_, err := database.DB.Exec("UPDATE users SET frozen_until = ? WHERE UID = ?", frozenUntil.Format("2006-01-02 15:04:05"), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "冻结失败"})
		return
	}

	// 冻结后强制登出所有当前会话
	_, _ = database.DB.Exec("DELETE FROM user_sessions WHERE user_uid = ?", uid)

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("账号已冻结，直到 %s", frozenUntil.Format("2006-01-02 15:04:05"))})
}

// 获取用户信息
func GetUserInfo(c *gin.Context) {
	uid := c.GetInt("uid")

	var user models.User
	err := database.DB.QueryRow(
		"SELECT UID, username, avatar, is_admin, role, two_factor_enabled, points, total_games, win_count, created_at FROM users WHERE UID = ?",
		uid,
	).Scan(&user.UID, &user.Username, &user.Avatar, &user.IsAdmin, &user.Role, &user.TwoFactorEnabled, &user.Points, &user.TotalGames, &user.WinCount, &user.CreatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, user)
}
