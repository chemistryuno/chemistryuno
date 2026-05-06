package middleware

import (
	"chemistryuno/backend/utils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func initRequestLogTestLogger(t *testing.T) {
	t.Helper()
	if err := utils.InitLogger(50); err != nil {
		t.Fatalf("init logger: %v", err)
	}
	utils.ClearLogs()
	t.Cleanup(func() {
		_ = utils.CloseLogger()
	})
}

func TestRequestLoggerAuthenticatedAttribution(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initRequestLogTestLogger(t)

	router := gin.New()
	_ = router.SetTrustedProxies([]string{"127.0.0.1"})
	router.Use(RequestLogger())
	router.GET("/api/user/:uid", func(c *gin.Context) {
		c.Set("uid", 100000101)
		c.Set("role", "user")
		c.Set(AuthStateKey, "authenticated")
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/100000101?token=secret&page=1", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0    Test")
	req.Header.Set("X-Forwarded-For", "203.0.113.10")
	req.Header.Set("Origin", "http://localhost:5000")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	logs := utils.GetLogsFiltered(utils.LogFilter{Category: "request", UID: utils.IntPtr(100000101)}, 10)
	if len(logs) != 1 {
		t.Fatalf("expected one authenticated request log, got %d", len(logs))
	}
	entry := logs[0]
	if entry.AuthState != "authenticated" || entry.Role != "user" {
		t.Fatalf("unexpected auth metadata: %+v", entry)
	}
	if entry.Source == nil || entry.Source.ForwardedFor != "203.0.113.10" || entry.Source.UserAgent != "Mozilla/5.0 Test" {
		t.Fatalf("unexpected source metadata: %+v", entry.Source)
	}
	if entry.Request == nil || entry.Request.Method != http.MethodGet || entry.Request.Route != "/api/user/:uid" || entry.Request.Status != http.StatusOK {
		t.Fatalf("unexpected request metadata: %+v", entry.Request)
	}
	encoded, _ := json.Marshal(entry)
	if strings.Contains(string(encoded), "secret") {
		t.Fatalf("request log leaked sensitive query value: %s", encoded)
	}
}

func TestRequestLoggerAnonymousAttribution(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initRequestLogTestLogger(t)

	router := gin.New()
	router.Use(RequestLogger())
	router.GET("/api/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/public", nil))

	logs := utils.GetLogsFiltered(utils.LogFilter{Category: "request", Keyword: "public"}, 10)
	if len(logs) != 1 {
		t.Fatalf("expected one anonymous request log, got %d", len(logs))
	}
	if logs[0].UID != nil || logs[0].AuthState != "anonymous" {
		t.Fatalf("expected anonymous log without trusted UID, got %+v", logs[0])
	}
}

func TestRequestLoggerRejectedSessionAttribution(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initRequestLogTestLogger(t)

	router := gin.New()
	router.Use(RequestLogger())
	router.GET("/api/private", func(c *gin.Context) {
		c.Set(AttemptedUIDKey, 100000202)
		c.Set(AuthStateKey, "invalid_session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/private", nil))

	logs := utils.GetLogsFiltered(utils.LogFilter{AttemptedUID: utils.IntPtr(100000202), StatusClass: "4xx"}, 10)
	if len(logs) != 1 {
		t.Fatalf("expected one rejected-session request log, got %d", len(logs))
	}
	entry := logs[0]
	if entry.UID != nil || entry.AttemptedUID == nil || *entry.AttemptedUID != 100000202 {
		t.Fatalf("expected attempted UID only, got %+v", entry)
	}
	if entry.AuthState != "invalid_session" || entry.Level != "WARNING" {
		t.Fatalf("unexpected rejected-session metadata: %+v", entry)
	}
}

func TestRequestLoggerPingIsDebug(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initRequestLogTestLogger(t)

	router := gin.New()
	router.Use(RequestLogger())
	router.GET("/api/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/ping", nil))

	logs := utils.GetLogsFiltered(utils.LogFilter{Category: "request", Keyword: "ping"}, 10)
	if len(logs) != 1 {
		t.Fatalf("expected one ping request log, got %d", len(logs))
	}
	if logs[0].Level != "DEBUG" {
		t.Fatalf("expected ping to be debug level, got %+v", logs[0])
	}
}
