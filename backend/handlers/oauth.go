package handlers

import (
	"chemistryuno/database"
	"chemistryuno/repository"
	"chemistryuno/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/microsoft"
)

var (
	githubOauthConfig *oauth2.Config
	msOauthConfig     *oauth2.Config
)

// InitOauth 初始化 OAuth 配置
func InitOauth() {
	githubOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GITHUB_REDIRECT_URI"),
		Scopes:       []string{"user:email", "read:user"},
		Endpoint:     github.Endpoint,
	}

	msOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("MS_CLIENT_ID"),
		ClientSecret: os.Getenv("MS_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("MS_REDIRECT_URI"),
		Scopes:       []string{"User.Read"},
		Endpoint:     microsoft.AzureADEndpoint(os.Getenv("MS_TENANT_ID")),
	}
}

// generateStateToken 生成加密的 state 令牌，支持绑定模式
func generateStateToken(c *gin.Context, intent string) string {
	uid := 0
	if intent == "bind" {
		// 尝试从 context 中获取 uid (如果已经通过 AuthMiddleware)
		if val, exists := c.Get("uid"); exists {
			uid = val.(int)
		}
	}
	state, _ := utils.GenerateOAuthState(intent, uid)
	return state
}

// GitHubLogin 重定向到 GitHub 登录
func GitHubLogin(c *gin.Context) {
	if githubOauthConfig.ClientID == "" {
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

	state := generateStateToken(c, intent)
	url := githubOauthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// MicrosoftLogin 重定向到 Microsoft 登录
func MicrosoftLogin(c *gin.Context) {
	if msOauthConfig.ClientID == "" {
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

	state := generateStateToken(c, intent)
	url := msOauthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GitHubCallback 处理 GitHub 回调
func GitHubCallback(c *gin.Context) {
	stateToken := c.Query("state")
	stateClaims, err := utils.VerifyOAuthState(stateToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的状态参数"})
		return
	}

	code := c.Query("code")
	token, err := githubOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法换取 Token"})
		return
	}

	client := githubOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取用户信息"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "反序列化用户信息失败"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的状态参数"})
		return
	}

	code := c.Query("code")
	token, err := msOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法换取 Token"})
		return
	}

	client := msOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://graph.microsoft.com/v1.0/me")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取用户信息"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "反序列化用户信息失败"})
		return
	}

	email := msUser.Mail
	if email == "" {
		email = msUser.UserPrincipalName
	}

	handleOAuthUser(c, "microsoft", msUser.ID, msUser.UserPrincipalName, email, msUser.DisplayName, stateClaims)
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
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询失败"})
		return
	}

	// 处理绑定逻辑
	if state != nil && state.Intent == "bind" && state.UID > 0 {
		if user != nil && int(user.UID) != state.UID {
			// 该 OAuth 账号已绑定到其他用户
			c.Header("Content-Type", "text/html")
			c.String(http.StatusBadRequest, `
				<script>
					window.opener.postMessage({ type: 'oauth-error', error: '此第三方账号已被其他研究员占用' }, '*');
					window.close();
				</script>
			`)
			return
		}

		// 执行绑定
		targetUser, err := userRepo.FindByUID(uint(state.UID))
		if err != nil {
			c.String(http.StatusInternalServerError, "获取用户失败")
			return
		}

		if provider == "github" {
			targetUser.GithubID = providerID
		} else {
			targetUser.MicrosoftID = providerID
		}
		database.DB.Save(targetUser)

		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
			<script>
				window.opener.postMessage({ type: 'oauth-bind-success' }, '*');
				window.close();
			</script>
			绑定成功，正在关闭窗口...
		`)
		return
	}

	// 登录逻辑 (intent == "login")
	if user == nil {
		// 尝试通过邮箱关联
		if email != "" {
			user, _ = userRepo.FindByEmail(email)
		}

		if user != nil {
			// 关联现有用户
			if provider == "github" {
				user.GithubID = providerID
			} else {
				user.MicrosoftID = providerID
			}
			database.DB.Save(user)
		} else {
			// 创建新用户
			if nickname == "" {
				nickname = username
			}

			// 确保用户名唯一
			finalUsername := username
			for i := 1; ; i++ {
				exists, _ := userRepo.ExistsByUsername(finalUsername)
				if !exists {
					break
				}
				finalUsername = fmt.Sprintf("%s_%d", username, i)
			}

			user = &database.User{
				Username:      finalUsername,
				Email:         email,
				Nickname:      nickname,
				Avatar:        "🔬", // OAuth 用户默认头像
				Role:          "user",
				Points:        100, // 初始积分
				OAuthProvider: provider,
			}

			if provider == "github" {
				user.GithubID = providerID
			} else {
				user.MicrosoftID = providerID
			}

			if err := userRepo.Create(user); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
				return
			}
		}
	}

	// 登录成功，生成会话 (Fix: generating SID for OAuth login)
	sid, err := utils.CreateSession(int(user.UID), c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil || sid == "" {
		// 如果创建会话失败，至少尝试生成一个无会话的 Token，或者报错
		// 这里选择报错以保证一致性
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建会话失败"})
		return
	}

	// 生成 Token，现在包含有效的 sid
	token, err := utils.GenerateToken(int(user.UID), user.Username, user.IsAdmin, user.Role, sid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}

	// 这里通常会重定向回前端，并带上 Token
	// 或者通过 HTML 模板发送 postMessage 给父窗口
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, fmt.Sprintf(`
		<script>
			window.opener.postMessage({
				type: 'oauth-success',
				token: '%s',
				user: %s
			}, '*');
			window.close();
		</script>
		正在跳转...
	`, token, ToJSON(user)))
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
	} else if provider == "microsoft" {
		user.MicrosoftID = ""
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
