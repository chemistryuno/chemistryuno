package handlers

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/models"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/utils"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
)

var (
	githubOauthConfig *oauth2.Config
	msOauthConfig     *oauth2.Config
	googleOauthConfig *oauth2.Config
	appleOauthConfig  *oauth2.Config
)

func hasConfigValue(v string) bool {
	return strings.TrimSpace(v) != ""
}

func isGitHubOAuthReady() bool {
	return hasConfigValue(os.Getenv("GITHUB_CLIENT_ID")) &&
		hasConfigValue(os.Getenv("GITHUB_CLIENT_SECRET"))
}

func isMicrosoftOAuthReady() bool {
	return hasConfigValue(os.Getenv("MS_CLIENT_ID")) &&
		hasConfigValue(os.Getenv("MS_CLIENT_SECRET"))
}

func isGoogleOAuthReady() bool {
	return hasConfigValue(os.Getenv("GOOGLE_CLIENT_ID")) &&
		hasConfigValue(os.Getenv("GOOGLE_CLIENT_SECRET"))
}

func isAppleOAuthReady() bool {
	// 当前实现仅支持静态 APPLE_CLIENT_SECRET
	return hasConfigValue(os.Getenv("APPLE_CLIENT_ID")) &&
		hasConfigValue(os.Getenv("APPLE_CLIENT_SECRET")) &&
		hasConfigValue(os.Getenv("APPLE_REDIRECT_URI"))
}

// InitOauth 初始化 OAuth 配置
func InitOauth() {
	log.Println("🔐 初始化 OAuth 配置...")

	githubOauthConfig = &oauth2.Config{
		ClientID:     strings.TrimSpace(os.Getenv("GITHUB_CLIENT_ID")),
		ClientSecret: strings.TrimSpace(os.Getenv("GITHUB_CLIENT_SECRET")),
		RedirectURL:  strings.TrimSpace(os.Getenv("GITHUB_REDIRECT_URI")), // 这里确保读取 .env
		Scopes:       []string{"user:email", "read:user"},
		Endpoint:     github.Endpoint,
	}

	msOauthConfig = &oauth2.Config{
		ClientID:     strings.TrimSpace(os.Getenv("MS_CLIENT_ID")),
		ClientSecret: strings.TrimSpace(os.Getenv("MS_CLIENT_SECRET")),
		RedirectURL:  strings.TrimSpace(os.Getenv("MS_REDIRECT_URI")), // 这里确保读取 .env
		Scopes:       []string{"User.Read"},
		Endpoint:     microsoft.AzureADEndpoint(strings.TrimSpace(os.Getenv("MS_TENANT_ID"))),
	}

	googleOauthConfig = &oauth2.Config{
		ClientID:     strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID")),
		ClientSecret: strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET")),
		RedirectURL:  strings.TrimSpace(os.Getenv("GOOGLE_REDIRECT_URI")), // 这里确保读取 .env
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	appleOauthConfig = &oauth2.Config{
		ClientID:    strings.TrimSpace(os.Getenv("APPLE_CLIENT_ID")),
		RedirectURL: strings.TrimSpace(os.Getenv("APPLE_REDIRECT_URI")), // 这里确保读取 .env
		Scopes:      []string{"name", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://appleid.apple.com/auth/authorize",
			TokenURL: "https://appleid.apple.com/auth/token",
		},
	}

	// 统计已配置的 OAuth 提供商
	enabledProviders := []string{}
	disabledProviders := []string{}

	if isGitHubOAuthReady() {
		log.Println("   ✅ GitHub OAuth 已启用")
		enabledProviders = append(enabledProviders, "GitHub")
	} else {
		disabledProviders = append(disabledProviders, "GitHub")
	}

	if isMicrosoftOAuthReady() {
		log.Println("   ✅ Microsoft OAuth 已启用")
		enabledProviders = append(enabledProviders, "Microsoft")
	} else {
		disabledProviders = append(disabledProviders, "Microsoft")
	}

	if isGoogleOAuthReady() {
		log.Println("   ✅ Google OAuth 已启用")
		enabledProviders = append(enabledProviders, "Google")
	} else {
		disabledProviders = append(disabledProviders, "Google")
	}

	if isAppleOAuthReady() {
		log.Println("   ✅ Apple OAuth 已启用")
		enabledProviders = append(enabledProviders, "Apple")
	} else {
		disabledProviders = append(disabledProviders, "Apple")
	}

	// 汇总提示
	if len(enabledProviders) > 0 {
		log.Printf("✅ OAuth 配置完成，已启用 %d 个提供商: %s", len(enabledProviders), strings.Join(enabledProviders, ", "))
	}

	if len(disabledProviders) > 0 {
		log.Printf("💡 未配置的 OAuth 提供商: %s (登录按钮将自动隐藏)", strings.Join(disabledProviders, ", "))
		log.Println("   如需启用，请在 .env 文件中配置相应的 CLIENT_ID 和 CLIENT_SECRET")
	}
}

// generateStateToken 生成加密的 state 令牌，支持绑定模式
func generateStateToken(c *gin.Context, intent string) (string, error) {
	uid := 0
	if intent == "bind" {
		// 从 context 中获取 uid (由 AuthMiddleware 设置)
		if val, exists := c.Get("uid"); exists {
			uid = val.(int)
		} else {
			return "", fmt.Errorf("未授权的绑定请求")
		}
	}
	return utils.GenerateOAuthState(intent, uid)
}

// getBaseURL 获取当前请求的基础 URL (例如 http://localhost:5000)
func getBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	host := c.Request.Host
	if xForwardedHost := c.GetHeader("X-Forwarded-Host"); xForwardedHost != "" {
		host = xForwardedHost
	}

	// 优先使用环境变量配置的域名（如果存在）
	if configuredHost := os.Getenv("APP_HOST"); configuredHost != "" {
		return configuredHost
	}

	return fmt.Sprintf("%s://%s", scheme, host)
}

// GitHubLogin 重定向到 GitHub 登录
func GitHubLogin(c *gin.Context) {
	if !isGitHubOAuthReady() {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "GitHub OAuth 未配置"})
		return
	}

	intent := c.Query("intent")
	if intent == "" {
		if strings.Contains(c.Request.URL.Path, "/bind") {
			intent = "bind"
		} else {
			intent = "login"
		}
	}

	state, err := generateStateToken(c, intent)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 动态修正配置中的 RedirectURL，如果环境变量为空或为了保持环境一致性
	config := *githubOauthConfig
	if config.RedirectURL == "" {
		config.RedirectURL = getBaseURL(c) + "/api/auth/github/callback"
	}

	url := config.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// MicrosoftLogin 重定向到 Microsoft 登录
func MicrosoftLogin(c *gin.Context) {
	if !isMicrosoftOAuthReady() {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Microsoft OAuth 未配置"})
		return
	}

	intent := c.Query("intent")
	if intent == "" {
		if strings.Contains(c.Request.URL.Path, "/bind") {
			intent = "bind"
		} else {
			intent = "login"
		}
	}

	state, err := generateStateToken(c, intent)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	config := *msOauthConfig
	if config.RedirectURL == "" {
		config.RedirectURL = getBaseURL(c) + "/api/auth/ms/callback"
	}

	url := config.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleLogin 重定向到 Google 登录
func GoogleLogin(c *gin.Context) {
	if !isGoogleOAuthReady() {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Google OAuth 未配置"})
		return
	}

	intent := c.Query("intent")
	if intent == "" {
		if strings.Contains(c.Request.URL.Path, "/bind") {
			intent = "bind"
		} else {
			intent = "login"
		}
	}

	state, err := generateStateToken(c, intent)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	config := *googleOauthConfig
	if config.RedirectURL == "" {
		config.RedirectURL = getBaseURL(c) + "/api/auth/google/callback"
	}

	url := config.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// AppleLogin 重定向到 Apple 登录
func AppleLogin(c *gin.Context) {
	if !isAppleOAuthReady() {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Apple OAuth 未配置"})
		return
	}

	intent := c.Query("intent")
	if intent == "" {
		if strings.Contains(c.Request.URL.Path, "/bind") {
			intent = "bind"
		} else {
			intent = "login"
		}
	}

	state, err := generateStateToken(c, intent)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	config := *appleOauthConfig
	if config.RedirectURL == "" {
		config.RedirectURL = getBaseURL(c) + "/api/auth/apple/callback"
	}

	url := config.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GitHubCallback 处理 GitHub 回调
func GitHubCallback(c *gin.Context) {
	stateToken := c.Query("state")
	stateClaims, err := utils.VerifyOAuthState(stateToken)
	if err != nil {
		sendOAuthError(c, http.StatusBadRequest, "无效的状态参数或会话已超时")
		return
	}

	config := *githubOauthConfig
	if config.RedirectURL == "" {
		config.RedirectURL = getBaseURL(c) + "/api/auth/github/callback"
	}

	code := c.Query("code")
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		sendOAuthError(c, http.StatusInternalServerError, "无法换取 GitHub 访问令牌")
		return
	}

	client := config.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		sendOAuthError(c, http.StatusInternalServerError, "获取 GitHub 用户信息失败")
		return
	}
	defer resp.Body.Close()

	var ghUser struct {
		ID       int    `json:"id"`
		Login    string `json:"login"`
		Email    string `json:"email"`
		Avatar   string `json:"avatar_url"`
		Nickname string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ghUser); err != nil {
		sendOAuthError(c, http.StatusInternalServerError, "反序列化 GitHub 用户数据失败")
		return
	}

	// 如果 email 为空，尝试单独获取
	if ghUser.Email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err == nil {
			defer emailResp.Body.Close()
			var emails []struct {
				Email   string `json:"email"`
				Primary bool   `json:"primary"`
			}
			if err := json.NewDecoder(emailResp.Body).Decode(&emails); err == nil {
				for _, e := range emails {
					if e.Primary {
						ghUser.Email = e.Email
						break
					}
				}
			}
		}
	}

	handleOAuthUser(c, "github", fmt.Sprintf("%d", ghUser.ID), ghUser.Login, ghUser.Email, ghUser.Nickname, stateClaims)
}

// MicrosoftCallback 处理 Microsoft 回调
func MicrosoftCallback(c *gin.Context) {
	stateToken := c.Query("state")
	stateClaims, err := utils.VerifyOAuthState(stateToken)
	if err != nil {
		sendOAuthError(c, http.StatusBadRequest, "无效的状态参数或会话已超时")
		return
	}

	config := *msOauthConfig
	if config.RedirectURL == "" {
		config.RedirectURL = getBaseURL(c) + "/api/auth/ms/callback"
	}

	code := c.Query("code")
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		sendOAuthError(c, http.StatusInternalServerError, "无法换取 Microsoft 访问令牌")
		return
	}

	client := config.Client(context.Background(), token)
	resp, err := client.Get("https://graph.microsoft.com/v1.0/me")
	if err != nil {
		sendOAuthError(c, http.StatusInternalServerError, "获取 Microsoft 用户信息失败")
		return
	}
	defer resp.Body.Close()

	var msUser struct {
		ID                string `json:"id"`
		DisplayName       string `json:"displayName"`
		Mail              string `json:"mail"`
		UserPrincipalName string `json:"userPrincipalName"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&msUser); err != nil {
		sendOAuthError(c, http.StatusInternalServerError, "反序列化 Microsoft 用户数据失败")
		return
	}

	email := msUser.Mail
	if email == "" {
		email = msUser.UserPrincipalName
	}

	handleOAuthUser(c, "microsoft", msUser.ID, msUser.UserPrincipalName, email, msUser.DisplayName, stateClaims)
}

// GoogleCallback 处理 Google 回调
func GoogleCallback(c *gin.Context) {
	stateToken := c.Query("state")
	stateClaims, err := utils.VerifyOAuthState(stateToken)
	if err != nil {
		sendOAuthError(c, http.StatusBadRequest, "无效的状态参数或会话已超时")
		return
	}

	config := *googleOauthConfig
	if config.RedirectURL == "" {
		config.RedirectURL = getBaseURL(c) + "/api/auth/google/callback"
	}

	code := c.Query("code")
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		sendOAuthError(c, http.StatusInternalServerError, "无法换取 Google 访问令牌")
		return
	}

	client := config.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		sendOAuthError(c, http.StatusInternalServerError, "获取 Google 用户信息失败")
		return
	}
	defer resp.Body.Close()

	var googleUser struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		sendOAuthError(c, http.StatusInternalServerError, "反序列化 Google 用户数据失败")
		return
	}

	handleOAuthUser(c, "google", googleUser.ID, googleUser.Email, googleUser.Email, googleUser.Name, stateClaims)
}

// AppleCallback 处理 Apple 回调
func AppleCallback(c *gin.Context) {
	// Apple 默认可能使用 POST，但如果配置正确可以使用 GET
	stateToken := c.DefaultPostForm("state", c.Query("state"))
	code := c.DefaultPostForm("code", c.Query("code"))

	stateClaims, err := utils.VerifyOAuthState(stateToken)
	if err != nil {
		sendOAuthError(c, http.StatusBadRequest, "无效的状态参数或会话已超时")
		return
	}

	// 注意：Apple 换取 Token 需要 ClientSecret (JWT 签名)
	// 这里简化处理，假设在 InitOauth 中生成了或有辅助函数
	appleClientSecret, err := getAppleClientSecret()
	if err != nil {
		sendOAuthError(c, http.StatusInternalServerError, "无法生成 Apple 认证密钥: "+err.Error())
		return
	}

	conf := *appleOauthConfig
	conf.ClientSecret = appleClientSecret

	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		sendOAuthError(c, http.StatusInternalServerError, "无法换取 Apple 访问令牌: "+err.Error())
		return
	}

	// Apple 的用户信息主要在 id_token 中
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		sendOAuthError(c, http.StatusInternalServerError, "Apple 响应中缺失 id_token")
		return
	}

	// 解析 ID Token 获取 email 和 sub
	claims := strings.Split(idToken, ".")
	if len(claims) < 2 {
		sendOAuthError(c, http.StatusInternalServerError, "无效的 Apple ID Token")
		return
	}

	payload, _ := base64.RawURLEncoding.DecodeString(claims[1])
	var appleClaims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
	}
	json.Unmarshal(payload, &appleClaims)

	// Apple 第一次授权时可能会提供 user 字段 (包含姓名)
	var nickname string
	userJSON := c.DefaultPostForm("user", "")
	if userJSON != "" {
		var u struct {
			Name struct {
				FirstName string `json:"firstName"`
				LastName  string `json:"lastName"`
			} `json:"name"`
		}
		json.Unmarshal([]byte(userJSON), &u)
		nickname = u.Name.FirstName + " " + u.Name.LastName
	}

	handleOAuthUser(c, "apple", appleClaims.Sub, appleClaims.Email, appleClaims.Email, nickname, stateClaims)
}

// getAppleClientSecret 生成 Apple Client Secret (根据环境变量中的私钥)
// 这是一个简化的实现框架
func getAppleClientSecret() (string, error) {
	// 如果已经在 .env 中配置了静态的 CLIENT_SECRET (虽然不推荐但可以)，直接返回
	staticSecret := os.Getenv("APPLE_CLIENT_SECRET")
	if staticSecret != "" {
		return staticSecret, nil
	}

	// 实际生产中应使用 jwt.NewWithClaims 和私钥文件生成
	// 这里返回空并提示配置，直到用户提供私钥
	return "", fmt.Errorf("未配置 Apple 私钥，无法动态生成 ClientSecret")
}

func handleOAuthUser(c *gin.Context, provider, providerID, username, email, nickname string, state *utils.StateClaims) {
	userRepo := repository.NewUserRepository()

	// 查找是否已有绑定此提供商的用户
	var user *database.User
	var err error

	if provider == "github" {
		user, err = userRepo.FindByGithubID(providerID)
	} else if provider == "microsoft" {
		user, err = userRepo.FindByMicrosoftID(providerID)
	} else if provider == "google" {
		user, err = userRepo.FindByGoogleID(providerID)
	} else if provider == "apple" {
		user, err = userRepo.FindByAppleID(providerID)
	}

	if err != nil {
		sendOAuthError(c, http.StatusInternalServerError, "研究员数据库索引故障: "+err.Error())
		return
	}

	// 处理绑定逻辑
	if state != nil && state.Intent == "bind" {
		if state.UID <= 0 {
			sendOAuthError(c, http.StatusBadRequest, "身份令牌解析异常，请尝试重新登录后绑定")
			return
		}

		if user != nil && int(user.UID) != state.UID {
			// 该 OAuth 账号已绑定到其他用户
			sendOAuthError(c, http.StatusBadRequest, "此第三方账号已被其他研究员档案占用，无法重复关联")
			return
		}

		// 执行绑定
		targetUser, err := userRepo.FindByUID(uint(state.UID))
		if err != nil {
			sendOAuthError(c, http.StatusInternalServerError, "获取当前研究员档案失败，同步中断")
			return
		}

		if provider == "github" {
			targetUser.GithubID = providerID
		} else if provider == "microsoft" {
			targetUser.MicrosoftID = providerID
		} else if provider == "google" {
			targetUser.GoogleID = providerID
		} else if provider == "apple" {
			targetUser.AppleID = providerID
		}
		if err := database.DB.Save(targetUser).Error; err != nil {
			sendOAuthError(c, http.StatusInternalServerError, "绑定失败，请稍后重试")
			return
		}

		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, `
			<script>
				if (window.opener) window.opener.postMessage({ type: 'oauth-bind-success' }, '*');
				window.close();
			</script>
			<div style="font-family:sans-serif;text-align:center;padding-top:100px;color:#059669;">
				<h3>同步成功</h3>
				<p>档案已成功关联，正在返回实验室...</p>
			</div>
		`)
		return
	}

	// 登录逻辑 (intent == "login" 或默认)
	email = strings.ToLower(strings.TrimSpace(email))

	// 如果无法提供邮箱，由于系统强制要求邮箱，告知前端需要补充信息
	if email == "" {
		sendOAuthError(c, http.StatusBadRequest, "NEED_EMAIL")
		return
	}

	if user == nil {
		// 尝试通过邮箱关联
		user, _ = userRepo.FindByEmail(email)

		if user != nil {
			// 关联现有用户
			if provider == "github" {
				user.GithubID = providerID
			} else if provider == "microsoft" {
				user.MicrosoftID = providerID
			} else if provider == "google" {
				user.GoogleID = providerID
			} else if provider == "apple" {
				user.AppleID = providerID
			}
			if err := database.DB.Save(user).Error; err != nil {
				sendOAuthError(c, http.StatusInternalServerError, "关联已有账号失败")
				return
			}
		} else {
			// 创建新用户
				// 如果 OAuth 提供方没有传回可用昵称，则生成随机昵称并确保唯一性
				if strings.TrimSpace(nickname) == "" {
					base := "研究员"
					candidate := utils.GenerateRandomNickname(base)
					for i := 0; i < 8; i++ {
						exists, _ := userRepo.ExistsByNickname(candidate)
						if !exists {
							nickname = candidate
							break
						}
						candidate = utils.GenerateRandomNickname(base)
					}
					if nickname == "" {
						// 回退到邮箱本地部分或 provider 标识
						parts := strings.SplitN(email, "@", 2)
						if parts[0] != "" {
							nickname = parts[0]
						} else {
							nickname = provider + "_user"
						}
					}
				}

			// 确保用户名唯一
			baseUsername := strings.Split(email, "@")[0]
			if baseUsername == "" {
				baseUsername = "oauth_user"
			}
			finalUsername := baseUsername
			for i := 1; ; i++ {
				exists, _ := userRepo.ExistsByUsername(finalUsername)
				if !exists {
					break
				}
				finalUsername = fmt.Sprintf("%s_%d", baseUsername, i)
			}

			// 获取最大 UID 并计算新 UID
			maxUID, _ := userRepo.GetMaxUID() // OAuth 模式下为了流程顺畅，如果获取失败则回退到 0 让数据库自增
			newUID := maxUID + 1
			if newUID < 100000000 {
				newUID = 100000000
			}

			user = &database.User{
				UID:           newUID,
				Username:      finalUsername,
				Email:         email,
				Nickname:      nickname,
				Avatar:        "🔬", // OAuth 用户默认头像
				Role:          "user",
				Points:        1000, // 初始积分
				MonthlyPoints: 1000, // 初始月积分
				Level:         1,    // 确保有初始等级
				XP:            0,
				OAuthProvider: provider,
			}

			if provider == "github" {
				user.GithubID = providerID
			} else if provider == "microsoft" {
				user.MicrosoftID = providerID
			} else if provider == "google" {
				user.GoogleID = providerID
			} else if provider == "apple" {
				user.AppleID = providerID
			}

			if err := userRepo.Create(user); err != nil {
				sendOAuthError(c, http.StatusInternalServerError, "创建新研究员档案失败")
				return
			}
		}
	}

	// 登录成功，生成会话 (Fix: generating SID for OAuth login)
	sid, err := utils.CreateSession(int(user.UID), c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil || sid == "" {
		sendOAuthError(c, http.StatusInternalServerError, "实验室会话创建失败，请稍后重试")
		return
	}

	// 生成 Token，现在包含有效的 sid
	accessToken, err := utils.GenerateAccessToken(int(user.UID), user.Email, models.NormalizeRole(user.Role), sid)
	if err != nil {
		sendOAuthError(c, http.StatusInternalServerError, "实验室访问令牌签署失败")
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(int(user.UID), sid)
	if err != nil {
		sendOAuthError(c, http.StatusInternalServerError, "实验室刷新令牌签署失败")
		return
	}

	// 设置安全的HttpOnly Cookie存储token
	setSecureAuthCookie(c, "access_token", accessToken, 900)         // 15分钟
	setSecureAuthCookie(c, "refresh_token", refreshToken, 7*24*3600) // 7天

	// 将 user 序列化为 JSON，以便在 JS 中安全嵌入（防止特殊字符破坏脚本）
	userJSON := ToJSON(user)
	payloadJSON, _ := json.Marshal(map[string]interface{}{
		"type": "oauth-success",
		"user": json.RawMessage(userJSON),
	})

	// 计算前端回调页基础 URL（用于 window.opener 不可用时的降级重定向）
	frontendBaseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("FRONTEND_BASE_URL")), "/")
	if frontendBaseURL == "" {
		// 未配置时，从请求的 Referer 或 Origin 推断前端地址
		origin := c.GetHeader("Origin")
		if origin == "" {
			referer := c.GetHeader("Referer")
			if referer != "" {
				// 只取协议+主机+端口部分
				if idx := strings.Index(referer, "/"); idx >= 0 {
					// scheme://host 部分
					parts := strings.SplitN(referer, "/", 4)
					if len(parts) >= 3 {
						origin = parts[0] + "//" + parts[2]
					}
				}
			}
		}
		if origin != "" {
			frontendBaseURL = origin
		}
	}
	fallbackURL := frontendBaseURL + "/oauth-callback"

	// 通信与跳转
	// postMessage 优先（弹窗场景），window.opener 不可用时降级重定向（跨域 / 独立标签场景）
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><title>授权成功</title></head>
<body>
<div style="font-family:sans-serif;text-align:center;padding-top:100px;color:#2563eb;">
  <h3>访问批准</h3>
  <p>欢迎回来，正在同步进入实验室...</p>
</div>
<script>
  (function() {
    var payload = %s;
    if (window.opener && !window.opener.closed) {
      try {
        window.opener.postMessage(payload, '*');
        setTimeout(function() { window.close(); }, 100);
        return;
      } catch(e) {
        // opener 跨域访问受限，降级到重定向
      }
    }
    // 降级方案：重定向到oauth-callback页面，token已通过cookie设置
    window.location.replace(%s);
  })();
</script>
</body>
</html>`, string(payloadJSON), func() string {
		b, _ := json.Marshal(fallbackURL)
		return string(b)
	}()))
}

// UnbindOAuth 处理解绑逻辑
func UnbindOAuth(c *gin.Context) {
	provider := c.Query("provider")
	uid := c.GetInt("uid")

	userRepo := repository.NewUserRepository()
	user, err := userRepo.FindByUID(uint(uid))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	if provider == "github" {
		user.GithubID = ""
	} else if provider == "microsoft" || provider == "ms" {
		user.MicrosoftID = ""
	} else if provider == "google" {
		user.GoogleID = ""
	} else if provider == "apple" {
		user.AppleID = ""
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的服务商"})
		return
	}

	if err := database.DB.Save(user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解绑失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "解绑成功"})
}

// ToJSON 辅助函数
func ToJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// sendOAuthError 向父窗口发送错误消息并显示友好的错误页面
func sendOAuthError(c *gin.Context, status int, msg string) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(status, fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<title>认证错误</title>
			<style>
				body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; background: #f8fafc; color: #1e293b; }
				.card { background: white; padding: 2rem; border-radius: 1rem; box-shadow: 0 10px 15px -3px rgba(0,0,0,0.1); text-align: center; max-width: 400px; width: 90%%; }
				h2 { color: #ef4444; margin-top: 0; }
				p { line-height: 1.5; color: #64748b; }
				.timer { font-size: 0.875rem; color: #94a3b8; margin-top: 1.5rem; }
			</style>
		</head>
		<body>
			<div class="card">
				<h2>认证失败 / AUTH_ERROR</h2>
				<p>%s</p>
				<div class="timer">此窗口将在 3 秒后尝试自动关闭...</div>
			</div>
			<script>
				if (window.opener) {
					window.opener.postMessage({ type: 'oauth-error', error: '%s' }, '*');
				}
				setTimeout(() => window.close(), 3000);
			</script>
		</body>
		</html>
	`, msg, msg))
}
