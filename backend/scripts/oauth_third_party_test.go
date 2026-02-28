//go:build scripts
// +build scripts

package main

import (
	"chemistryuno/backend/handlers"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func buildOAuthRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/auth/config", handlers.GetAuthConfig)
	r.GET("/auth/github/login", handlers.GitHubLogin)
	r.GET("/auth/ms/login", handlers.MicrosoftLogin)
	r.GET("/auth/google/login", handlers.GoogleLogin)
	r.GET("/auth/apple/login", handlers.AppleLogin)
	r.GET("/auth/github/callback", handlers.GitHubCallback)
	r.GET("/auth/ms/callback", handlers.MicrosoftCallback)
	r.GET("/auth/google/callback", handlers.GoogleCallback)
	r.GET("/auth/apple/callback", handlers.AppleCallback)
	r.POST("/auth/apple/callback", handlers.AppleCallback)
	return r
}

func httpGET(r http.Handler, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestThirdPartyOAuthScript(t *testing.T) {
	// 默认先清空，确保测试不受外部环境影响
	t.Setenv("GITHUB_CLIENT_ID", "")
	t.Setenv("GITHUB_CLIENT_SECRET", "")
	t.Setenv("GITHUB_REDIRECT_URI", "")
	t.Setenv("MS_CLIENT_ID", "")
	t.Setenv("MS_CLIENT_SECRET", "")
	t.Setenv("MS_REDIRECT_URI", "")
	t.Setenv("MS_TENANT_ID", "")
	t.Setenv("GOOGLE_CLIENT_ID", "")
	t.Setenv("GOOGLE_CLIENT_SECRET", "")
	t.Setenv("GOOGLE_REDIRECT_URI", "")
	t.Setenv("APPLE_CLIENT_ID", "")
	t.Setenv("APPLE_CLIENT_SECRET", "")
	t.Setenv("APPLE_REDIRECT_URI", "")
	t.Setenv("JWT_SECRET", "oauth-script-test-secret-0123456789abcdef")

	handlers.InitOauth()
	router := buildOAuthRouter()

	t.Run("auth_config_flags_follow_complete_config", func(t *testing.T) {
		// 只有 ID 时，前端按钮不应显示（避免点了必失败）
		t.Setenv("GITHUB_CLIENT_ID", "gh-id")
		t.Setenv("GITHUB_CLIENT_SECRET", "")
		t.Setenv("MS_CLIENT_ID", "ms-id")
		t.Setenv("MS_CLIENT_SECRET", "")
		t.Setenv("GOOGLE_CLIENT_ID", "google-id")
		t.Setenv("GOOGLE_CLIENT_SECRET", "")
		t.Setenv("APPLE_CLIENT_ID", "apple-id")
		t.Setenv("APPLE_CLIENT_SECRET", "")
		t.Setenv("APPLE_REDIRECT_URI", "")

		handlers.InitOauth()
		resp := httpGET(router, "/auth/config")
		if resp.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.Code)
		}

		var body map[string]bool
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("unmarshal config failed: %v", err)
		}

		if body["github_enabled"] || body["ms_enabled"] || body["google_enabled"] || body["apple_enabled"] {
			t.Fatalf("expected all oauth providers disabled with incomplete config, got %#v", body)
		}
	})

	t.Run("login_endpoints_redirect_when_ready", func(t *testing.T) {
		t.Setenv("GITHUB_CLIENT_ID", "gh-id")
		t.Setenv("GITHUB_CLIENT_SECRET", "gh-secret")
		t.Setenv("GITHUB_REDIRECT_URI", "http://localhost:5000/api/auth/github/callback")
		t.Setenv("MS_CLIENT_ID", "ms-id")
		t.Setenv("MS_CLIENT_SECRET", "ms-secret")
		t.Setenv("MS_REDIRECT_URI", "http://localhost:5000/api/auth/ms/callback")
		t.Setenv("MS_TENANT_ID", "common")
		t.Setenv("GOOGLE_CLIENT_ID", "google-id")
		t.Setenv("GOOGLE_CLIENT_SECRET", "google-secret")
		t.Setenv("GOOGLE_REDIRECT_URI", "http://localhost:5000/api/auth/google/callback")
		t.Setenv("APPLE_CLIENT_ID", "apple-id")
		t.Setenv("APPLE_CLIENT_SECRET", "apple-secret")
		t.Setenv("APPLE_REDIRECT_URI", "http://localhost:5000/api/auth/apple/callback")

		handlers.InitOauth()

		paths := []string{
			"/auth/github/login",
			"/auth/ms/login",
			"/auth/google/login",
			"/auth/apple/login",
		}

		for _, path := range paths {
			resp := httpGET(router, path)
			if resp.Code != http.StatusTemporaryRedirect {
				t.Fatalf("%s expected 307 redirect, got %d body=%s", path, resp.Code, resp.Body.String())
			}

			loc := resp.Header().Get("Location")
			if loc == "" || !strings.Contains(loc, "state=") {
				t.Fatalf("%s redirect location missing state: %q", path, loc)
			}
		}
	})

	t.Run("invalid_state_callbacks_return_oauth_error_page", func(t *testing.T) {
		handlers.InitOauth()
		paths := []string{
			"/auth/github/callback?state=invalid&code=fake",
			"/auth/ms/callback?state=invalid&code=fake",
			"/auth/google/callback?state=invalid&code=fake",
			"/auth/apple/callback?state=invalid&code=fake",
		}

		for _, path := range paths {
			resp := httpGET(router, path)
			if resp.Code != http.StatusBadRequest {
				t.Fatalf("%s expected 400, got %d", path, resp.Code)
			}
			body := resp.Body.String()
			if !strings.Contains(body, "oauth-error") || !strings.Contains(body, "AUTH_ERROR") {
				t.Fatalf("%s expected oauth error HTML, got: %s", path, body)
			}
		}
	})
}
