package handlers

import (
	"chemistryuno/database"
	"chemistryuno/models"
	"chemistryuno/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
)

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/auth/google/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	microsoftOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/auth/microsoft/callback",
		ClientID:     os.Getenv("MICROSOFT_CLIENT_ID"),
		ClientSecret: os.Getenv("MICROSOFT_CLIENT_SECRET"),
		Scopes:       []string{"User.Read"},
		Endpoint:     microsoft.AzureADEndpoint("common"),
	}

	// 随机字符串用于状态验证
	oauthStateString = "random_state_string"
)

// GoogleLogin 重定向到 Google 登录页面
func GoogleLogin(c *gin.Context) {
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback 处理 Google 回调
func GoogleCallback(c *gin.Context) {
	state := c.Query("state")
	if state != oauthStateString {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid oauth state"})
		return
	}

	code := c.Query("code")
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Code exchange failed"})
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer response.Body.Close()

	var userInfo struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(response.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user info"})
		return
	}

	handleSocialLogin(c, "google", userInfo.ID, userInfo.Email, userInfo.Name, userInfo.Picture)
}

// MicrosoftLogin 重定向到 Microsoft 登录页面
func MicrosoftLogin(c *gin.Context) {
	url := microsoftOauthConfig.AuthCodeURL(oauthStateString)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// MicrosoftCallback 处理 Microsoft 回调
func MicrosoftCallback(c *gin.Context) {
	state := c.Query("state")
	if state != oauthStateString {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid oauth state"})
		return
	}

	code := c.Query("code")
	token, err := microsoftOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Code exchange failed"})
		return
	}

	client := microsoftOauthConfig.Client(context.Background(), token)
	response, err := client.Get("https://graph.microsoft.com/v1.0/me")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer response.Body.Close()

	var userInfo struct {
		ID                string `json:"id"`
		DisplayName       string `json:"displayName"`
		UserPrincipalName string `json:"userPrincipalName"`
	}
	if err := json.NewDecoder(response.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user info"})
		return
	}

	handleSocialLogin(c, "microsoft", userInfo.ID, userInfo.UserPrincipalName, userInfo.DisplayName, "")
}

// handleSocialLogin 统一处理社交登录逻辑
func handleSocialLogin(c *gin.Context, provider, providerID, email, name, avatar string) {
	var user models.User
	var query string
	switch provider {
	case "google":
		query = "SELECT UID, username, avatar, is_admin, role FROM users WHERE google_id = ?"
	case "microsoft":
		query = "SELECT UID, username, avatar, is_admin, role FROM users WHERE microsoft_id = ?"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown provider"})
		return
	}

	err := database.DB.QueryRow(query, providerID).Scan(&user.UID, &user.Username, &user.Avatar, &user.IsAdmin, &user.Role)

	if err != nil {
		// 用户不存在，创建新用户或关联现有邮箱用户
		err = database.DB.QueryRow("SELECT UID, username, avatar, is_admin, role FROM users WHERE username = ?", email).Scan(&user.UID, &user.Username, &user.Avatar, &user.IsAdmin, &user.Role)
		if err == nil {
			// 关联现有邮箱用户
			updateQuery := fmt.Sprintf("UPDATE users SET %s_id = ? WHERE UID = ?", provider)
			database.DB.Exec(updateQuery, providerID, user.UID)
		} else {
			// 创建新用户
			if avatar == "" {
				avatar = "🧪"
			}
			res, err := database.DB.Exec(
				fmt.Sprintf("INSERT INTO users (username, password, avatar, role, %s_id) VALUES (?, ?, ?, ?, ?)", provider),
				email, "OAUTH_USER", avatar, "user", providerID,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
				return
			}
			newID, _ := res.LastInsertId()
			user.UID = int(newID)
			user.Username = email
			user.Avatar = avatar
			user.IsAdmin = false
			user.Role = "user"
		}
	}

	// 生成token
	token, err := utils.GenerateToken(user.UID, user.Username, user.IsAdmin, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	// 重定向回前端并带上token (生产环境应该通过 secure cookie 或者带 token 的重定向)
	frontendURL := "http://localhost:5173/login?token=" + token
	c.Redirect(http.StatusTemporaryRedirect, frontendURL)
}
