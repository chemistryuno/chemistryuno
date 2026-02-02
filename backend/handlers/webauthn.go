package handlers

import (
	"chemistryuno/database"
	"chemistryuno/models"
	"chemistryuno/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

var (
	webAuthn *webauthn.WebAuthn
	// 存储注册/登录会话的临时数据
	sessionStore = make(map[string]*webauthn.SessionData)
	sessionMutex sync.RWMutex
)

func InitWebAuthn() {
	var err error
	webAuthn, err = webauthn.New(&webauthn.Config{
		RPDisplayName: "Chemistry Uno",
		RPID:          "localhost",
		RPOrigins: []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			"http://localhost:5000",
			"http://127.0.0.1:5000",
			"http://localhost:8080",
			"http://127.0.0.1:8080",
			"http://localhost:4173",
			"http://127.0.0.1:4173",
		},
	})
	if err != nil {
		fmt.Printf("WebAuthn 初始化失败: %v\n", err)
	}
}

// BeginRegistration 开始 WebAuthn 注册
func BeginRegistration(c *gin.Context) {
	username, _ := c.Get("username")
	user, err := getUserByUsername(username.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 如果 WebAuthnIDRaw 为空，生成一个
	if user.WebAuthnIDRaw == "" {
		user.WebAuthnIDRaw = uuid.New().String()
		database.DB.Exec("UPDATE users SET webauthn_id = ? WHERE UID = ?", user.WebAuthnIDRaw, user.UID)
	}

	// 获取已有的凭证
	user.Credentials = getUserCredentials(user.UID)

	options, sessionData, err := webAuthn.BeginRegistration(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 存储会话
	sessionID := uuid.New().String()
	sessionMutex.Lock()
	sessionStore[sessionID] = sessionData
	sessionMutex.Unlock()

	c.SetCookie("webauthn_session", sessionID, 300, "/", "", false, true)
	c.JSON(http.StatusOK, options)
}

// FinishRegistration 完成 WebAuthn 注册
func FinishRegistration(c *gin.Context) {
	username, _ := c.Get("username")
	user, err := getUserByUsername(username.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	sessionID, err := c.Cookie("webauthn_session")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话已过期"})
		return
	}

	sessionMutex.RLock()
	sessionData, ok := sessionStore[sessionID]
	sessionMutex.RUnlock()

	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话数据不存在"})
		return
	}

	sessionMutex.Lock()
	delete(sessionStore, sessionID)
	sessionMutex.Unlock()

	// 安全打印调试日志：检查 webAuthn 对象是否初始化
	var rpid string
	if webAuthn != nil && webAuthn.Config != nil {
		rpid = webAuthn.Config.RPID
	} else {
		rpid = "NIL (Initialization Failed)"
	}
	fmt.Printf("WebAuthn FinishRegistration - Origin: %s, RPID: %s\n", c.Request.Header.Get("Origin"), rpid)

	if webAuthn == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebAuthn 服务未初始化，请检查配置"})
		return
	}

	credential, err := webAuthn.FinishRegistration(user, *sessionData, c.Request)
	if err != nil {
		fmt.Printf("WebAuthn FinishRegistration 失败: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 保存凭证到数据库
	err = saveCredential(user.UID, credential)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存凭证失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "硬件密钥注册成功"})
}

// BeginLogin 开始 WebAuthn 登录
func BeginLogin(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少用户名"})
		return
	}

	user, err := getUserByUsername(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	user.Credentials = getUserCredentials(user.UID)
	if len(user.Credentials) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该用户未绑定硬件密钥"})
		return
	}

	options, sessionData, err := webAuthn.BeginLogin(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sessionID := uuid.New().String()
	sessionMutex.Lock()
	sessionStore[sessionID] = sessionData
	sessionMutex.Unlock()

	c.SetCookie("webauthn_session", sessionID, 300, "/", "", false, true)
	c.JSON(http.StatusOK, options)
}

// FinishLogin 完成 WebAuthn 登录
func FinishLogin(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少用户名"})
		return
	}

	user, err := getUserByUsername(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	sessionID, err := c.Cookie("webauthn_session")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话已过期"})
		return
	}

	sessionMutex.RLock()
	sessionData, ok := sessionStore[sessionID]
	sessionMutex.RUnlock()

	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话数据不存在"})
		return
	}

	sessionMutex.Lock()
	delete(sessionStore, sessionID)
	sessionMutex.Unlock()

	user.Credentials = getUserCredentials(user.UID)
	credential, err := webAuthn.FinishLogin(user, *sessionData, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新签次数（可选，但推荐）
	updateCredentialSignCount(credential.ID, credential.Authenticator.SignCount)

	// 登录成功，生成 JWT
	token, err := utils.GenerateToken(user.UID, user.Username, user.IsAdmin, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 token 失败"})
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

// 辅助函数

func getUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := database.DB.QueryRow(`
		SELECT UID, username, avatar, is_admin, role, points, monthly_points, negative_play_count, banned_until, webauthn_id, created_at
		FROM users WHERE username = ?`, username).Scan(
		&user.UID, &user.Username, &user.Avatar, &user.IsAdmin, &user.Role, &user.Points, &user.MonthlyPoints,
		&user.NegativePlayCount, &user.BannedUntil, &user.WebAuthnIDRaw, &user.CreatedAt)
	return &user, err
}

func getUserCredentials(uid int) []webauthn.Credential {
	rows, err := database.DB.Query(`
		SELECT id, public_key, attestation_type, transport, sign_count, user_present, user_verified, backup_eligible, backup_state, clone_warning
		FROM user_credentials WHERE user_uid = ?`, uid)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var credentials []webauthn.Credential
	for rows.Next() {
		var cred webauthn.Credential
		var transportJSON string
		err := rows.Scan(
			&cred.ID, &cred.PublicKey, &cred.AttestationType, &transportJSON,
			&cred.Authenticator.SignCount, &cred.Flags.UserPresent,
			&cred.Flags.UserVerified, &cred.Flags.BackupEligible,
			&cred.Flags.BackupState, &cred.Authenticator.CloneWarning)
		if err == nil {
			if transportJSON != "" {
				json.Unmarshal([]byte(transportJSON), &cred.Transport)
			}
			credentials = append(credentials, cred)
		}
	}
	return credentials
}

func saveCredential(uid int, cred *webauthn.Credential) error {
	transportJSON, _ := json.Marshal(cred.Transport)
	_, err := database.DB.Exec(`
		INSERT INTO user_credentials (
			id, user_uid, public_key, attestation_type, transport, sign_count, 
			user_present, user_verified, backup_eligible, backup_state, clone_warning
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		cred.ID, uid, cred.PublicKey, cred.AttestationType, string(transportJSON),
		cred.Authenticator.SignCount, cred.Flags.UserPresent,
		cred.Flags.UserVerified, cred.Flags.BackupEligible,
		cred.Flags.BackupState, cred.Authenticator.CloneWarning)
	return err
}

func updateCredentialSignCount(id []byte, signCount uint32) {
	database.DB.Exec("UPDATE user_credentials SET sign_count = ? WHERE id = ?", signCount, id)
}

// ListCredentials 获取用户的硬件密钥列表
func ListCredentials(c *gin.Context) {
	username, _ := c.Get("username")
	user, err := getUserByUsername(username.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	rows, err := database.DB.Query(`
		SELECT id, attestation_type, created_at 
		FROM user_credentials WHERE user_uid = ?`, user.UID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []gin.H
	for rows.Next() {
		var id []byte
		var attestationType string
		var createdAt string
		if err := rows.Scan(&id, &attestationType, &createdAt); err == nil {
			results = append(results, gin.H{
				"id":   fmt.Sprintf("%x", id), // 转为十六进制显示
				"type": attestationType,
				"date": createdAt,
			})
		}
	}
	c.JSON(http.StatusOK, results)
}

// RemoveCredential 删除硬件密钥
func RemoveCredential(c *gin.Context) {
	username, _ := c.Get("username")
	user, err := getUserByUsername(username.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	credIDHex := c.Param("id")
	// TODO: 实现根据十六进制 ID 删除凭证
	_, err = database.DB.Exec("DELETE FROM user_credentials WHERE user_uid = ? AND hex(id) = upper(?)", user.UID, credIDHex)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密钥已删除"})
}
