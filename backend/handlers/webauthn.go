package handlers

import (
	"bytes"
	"chemistryuno/database"
	"chemistryuno/models"
	"chemistryuno/repository"
	"chemistryuno/utils"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

var (
	webAuthn     *webauthn.WebAuthn
	sessionStore = make(map[string]*webauthn.SessionData)
	sessionMutex sync.RWMutex
)

func InitWebAuthn() {
	var err error

	rpid := os.Getenv("WEBAUTHN_RPID")
	if rpid == "" {
		rpid = "localhost"
	}

	origins := []string{
		"http://localhost:5000",
		"http://127.0.0.1:5000",
	}

	if rpid != "" && rpid != "localhost" && rpid != "127.0.0.1" {
		origins = append(origins, "http://"+rpid, "https://"+rpid)
	}

	extraOriginsStr := os.Getenv("WEBAUTHN_ORIGIN")
	if extraOriginsStr != "" {
		parts := strings.Split(extraOriginsStr, ",")
		for _, p := range parts {
			trimmed := strings.TrimSpace(strings.TrimSuffix(p, "/"))
			if trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
	}

	// 去重
	uniqueOrigins := make([]string, 0)
	originMap := make(map[string]bool)
	for _, o := range origins {
		if !originMap[o] {
			originMap[o] = true
			uniqueOrigins = append(uniqueOrigins, o)
		}
	}

	webAuthn, err = webauthn.New(&webauthn.Config{
		RPDisplayName: "Chemistry Uno",
		RPID:          rpid,
		RPOrigins:     uniqueOrigins,
	})
	if err != nil {
		log.Printf("WebAuthn 初始化失败: %v", err)
	} else {
		log.Printf("WebAuthn 已初始化 - RPID: %s", rpid)
	}
}

// BeginRegistration 开始注册 (需要认证)
func BeginRegistration(c *gin.Context) {
	uid := c.GetInt("uid")
	user, err := repository.UserRepo.FindByUID(uint(uid))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	if user.WebAuthnID == "" {
		newID := uuid.New().String()
		user.WebAuthnID = newID
		repository.UserRepo.UpdateWebAuthnID(uint(uid), newID)
	}

	credentials, _ := repository.WebAuthnRepo.FindByUserUID(uint(uid))
	waUser := &WebAuthnUser{User: user, Credentials: credentials}

	if webAuthn == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebAuthn 服务未初始化"})
		return
	}

	options, sessionData, err := webAuthn.BeginRegistration(waUser)
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

// FinishRegistration 完成注册 (需要认证)
func FinishRegistration(c *gin.Context) {
	uid := c.GetInt("uid")
	user, err := repository.UserRepo.FindByUID(uint(uid))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	sessionID, err := c.Cookie("webauthn_session")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话已过期"})
		return
	}

	sessionMutex.Lock()
	sessionData, ok := sessionStore[sessionID]
	if ok {
		delete(sessionStore, sessionID)
	}
	sessionMutex.Unlock()

	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话数据不存在"})
		return
	}

	if webAuthn == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebAuthn 服务未初始化"})
		return
	}

	credentials, _ := repository.WebAuthnRepo.FindByUserUID(uint(uid))
	waUser := &WebAuthnUser{User: user, Credentials: credentials}

	credential, err := webAuthn.FinishRegistration(waUser, *sessionData, c.Request)
	if err != nil {
		log.Printf("WebAuthn FinishRegistration 失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 保存凭证（使用 base64 URL-safe 编码存储二进制 ID）
	cred := &database.WebAuthnCredential{
		ID:              base64.RawURLEncoding.EncodeToString(credential.ID),
		UserUID:         uint(uid),
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		AAGUID:          credential.Authenticator.AAGUID,
		SignCount:       credential.Authenticator.SignCount,
		CloneWarning:    credential.Authenticator.CloneWarning,
	}
	err = repository.WebAuthnRepo.Create(cred)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存凭证失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "硬件密钥注册成功"})
}

// BeginLogin 开始登录 (根据是否提供用户名选择可发现或指定用户登录)
func BeginLogin(c *gin.Context) {
	if webAuthn == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebAuthn 服务未初始化"})
		return
	}

	username := c.Query("username")
	var options *protocol.CredentialAssertion
	var sessionData *webauthn.SessionData
	var err error

	if username != "" {
		user, err2 := repository.UserRepo.FindByUsername(username)
		if err2 == nil {
			credentials, _ := repository.WebAuthnRepo.FindByUserUID(uint(user.UID))
			waUser := &WebAuthnUser{User: user, Credentials: credentials}
			options, sessionData, err = webAuthn.BeginLogin(waUser)
		} else {
			// 如果用户未找到，回退到可发现登录
			options, sessionData, err = webAuthn.BeginDiscoverableLogin()
		}
	} else {
		// 不指定用户，使用 discoverable credentials
		options, sessionData, err = webAuthn.BeginDiscoverableLogin()
	}

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

// FinishLogin 完成登录 (无需认证)
func FinishLogin(c *gin.Context) {
	sessionID, err := c.Cookie("webauthn_session")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话已过期"})
		return
	}

	sessionMutex.Lock()
	sessionData, ok := sessionStore[sessionID]
	if ok {
		delete(sessionStore, sessionID)
	}
	sessionMutex.Unlock()

	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话数据不存在"})
		return
	}

	if webAuthn == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebAuthn 服务未初始化"})
		return
	}

	// 🔐 修复：确保请求体可以被多次读取（一次用于提取 UserHandle，一次用于 webauthn.FinishLogin）
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var user *database.User
	username := c.Query("username")

	if username != "" {
		// 如果提供了用户名，直接查找用户
		user, err = repository.UserRepo.FindByUsername(username)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
	} else {
		// 否则使用 userHandle 从 credential 中识别用户
		parsedResponse, err := protocol.ParseCredentialRequestResponse(c.Request)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "解析响应失败: " + err.Error()})
			return
		}

		// 从 userHandle 中获取用户ID
		userHandle := parsedResponse.Response.UserHandle
		if len(userHandle) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少用户标识"})
			return
		}

		uid := binary.LittleEndian.Uint64(userHandle)
		user, err = repository.UserRepo.FindByUID(uint(uid))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
	}

	// 恢复请求体，供下一次 FinishLogin 使用
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	credentials, _ := repository.WebAuthnRepo.FindByUserUID(uint(user.UID))
	if len(credentials) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该用户未绑定硬件密钥"})
		return
	}

	waUser := &WebAuthnUser{User: user, Credentials: credentials}

	credential, err := webAuthn.FinishLogin(waUser, *sessionData, c.Request)
	if err != nil {
		log.Printf("WebAuthn FinishLogin 失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新签名计数（使用 base64 URL-safe 编码）
	repository.WebAuthnRepo.UpdateSignCount(base64.RawURLEncoding.EncodeToString(credential.ID), credential.Authenticator.SignCount)

	// 检查封禁状态
	now := time.Now()
	if user.BannedUntil != nil && now.Before(*user.BannedUntil) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "您的账号已被封禁，直到 " + user.BannedUntil.Format("2006-01-02 15:04:05"),
		})
		return
	}

	// 检查冻结状态
	if user.FrozenUntil != nil && now.Before(*user.FrozenUntil) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "您的账号当前处于冷冻状态，直到 " + user.FrozenUntil.Format("2006-01-02 15:04:05"),
		})
		return
	}

	// 生成会话和token
	sid, err := utils.CreateSession(int(user.UID), c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil || sid == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建会话失败"})
		return
	}
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

	fmt.Printf("WebAuthn 登录成功 - 用户: %s (UID=%d), SID: %s, IP: %s\n", user.Username, user.UID, sid, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"uid":                int(user.UID),
			"username":           user.Username,
			"avatar":             user.Avatar,
			"is_admin":           user.IsAdmin,
			"role":               user.Role,
			"two_factor_enabled": user.TwoFactorEnabled,
		},
		"announcements": announcements,
	})
}

// ListCredentials 列出凭证
func ListCredentials(c *gin.Context) {
	uid := c.GetInt("uid")
	credentials, err := repository.WebAuthnRepo.FindByUserUID(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	result := []gin.H{}
	for _, cred := range credentials {
		result = append(result, gin.H{
			"id":         cred.ID,
			"date":       cred.CreatedAt.Format("2006-01-02"),
			"created_at": cred.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, result)
}

// RemoveCredential 移除凭证
func RemoveCredential(c *gin.Context) {
	// credential ID 已经是 base64 URL-safe 编码的字符串，直接使用
	credID := c.Param("id")

	// 验证是否为有效的 base64 URL-safe 字符串
	if _, err := base64.RawURLEncoding.DecodeString(credID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的凭证ID"})
		return
	}

	err := repository.WebAuthnRepo.Delete(credID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "凭证已移除"})
}

// WebAuthnUser 实现 webauthn.User 接口
type WebAuthnUser struct {
	User        *database.User
	Credentials []database.WebAuthnCredential
}

func (u *WebAuthnUser) WebAuthnID() []byte {
	// 使用8字节的uint编码
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(u.User.UID))
	return buf
}

func (u *WebAuthnUser) WebAuthnName() string {
	return u.User.Username
}

func (u *WebAuthnUser) WebAuthnDisplayName() string {
	return u.User.Username
}

func (u *WebAuthnUser) WebAuthnIcon() string {
	return u.User.Avatar
}

func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	creds := make([]webauthn.Credential, len(u.Credentials))
	for i, c := range u.Credentials {
		// 从 base64 URL-safe 编码的字符串解码为二进制
		credID, err := base64.RawURLEncoding.DecodeString(c.ID)
		if err != nil {
			log.Printf("解码 credential ID 失败: %v", err)
			credID = []byte(c.ID) // 降级处理
		}
		creds[i] = webauthn.Credential{
			ID:              credID,
			PublicKey:       c.PublicKey,
			AttestationType: c.AttestationType,
			Authenticator: webauthn.Authenticator{
				AAGUID:       c.AAGUID,
				SignCount:    c.SignCount,
				CloneWarning: c.CloneWarning,
			},
		}
	}
	return creds
}

func timeNow() time.Time {
	return time.Now()
}
