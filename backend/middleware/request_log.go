package middleware

import (
	"chemistryuno/backend/utils"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	AuthStateKey    = "auth_state"
	AttemptedUIDKey = "attempted_uid"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)
		path := utils.SanitizeURLPath(requestPath(c.Request))
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		level := requestLogLevel(status)
		if isLowNoiseEndpoint(c.Request.URL.Path) {
			level = "DEBUG"
		}

		entry := utils.LogEntry{
			Level:     level,
			Category:  "request",
			Message:   fmt.Sprintf("%s %s -> %d (%dms)", c.Request.Method, path, status, latency.Milliseconds()),
			UID:       trustedUID(c),
			Role:      c.GetString("role"),
			AuthState: authState(c),
			Source:    utils.SourceFromRequest(c.Request, c.ClientIP()),
			Request: &utils.LogRequest{
				Method:      c.Request.Method,
				Path:        path,
				Route:       route,
				Status:      status,
				StatusClass: utils.StatusClass(status),
				LatencyMs:   latency.Milliseconds(),
				BytesIn:     requestBytesIn(c.Request),
				BytesOut:    responseBytesOut(c.Writer.Size()),
			},
		}

		if entry.UID == nil {
			entry.AttemptedUID = attemptedUID(c)
		}

		utils.LogStructured(entry)
	}
}

func requestLogLevel(status int) string {
	if status >= http.StatusInternalServerError {
		return "ERROR"
	}
	if status >= http.StatusBadRequest {
		return "WARNING"
	}
	return "INFO"
}

func trustedUID(c *gin.Context) *int {
	value, exists := c.Get("uid")
	if !exists {
		return nil
	}
	uid, ok := value.(int)
	if !ok {
		return nil
	}
	return utils.IntPtr(uid)
}

func attemptedUID(c *gin.Context) *int {
	value, exists := c.Get(AttemptedUIDKey)
	if !exists {
		return nil
	}
	uid, ok := value.(int)
	if !ok {
		return nil
	}
	return utils.IntPtr(uid)
}

func authState(c *gin.Context) string {
	if state := strings.TrimSpace(c.GetString(AuthStateKey)); state != "" {
		return state
	}
	if _, exists := c.Get("uid"); exists {
		return "authenticated"
	}
	if c.Query("token") == "" && c.GetHeader("Authorization") == "" {
		if _, err := c.Cookie("access_token"); err != nil {
			return "anonymous"
		}
	}
	return "unauthenticated"
}

func requestPath(r *http.Request) string {
	if r == nil || r.URL == nil {
		return ""
	}
	if r.URL.RawQuery == "" {
		return r.URL.Path
	}
	return r.URL.Path + "?" + r.URL.RawQuery
}

func requestBytesIn(r *http.Request) int64 {
	if r == nil || r.ContentLength < 0 {
		return 0
	}
	return r.ContentLength
}

func responseBytesOut(size int) int {
	if size < 0 {
		return 0
	}
	return size
}

func isLowNoiseEndpoint(path string) bool {
	return path == "/api/ping" || path == "/api/health"
}
