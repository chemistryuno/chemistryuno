package handlers

import (
	"chemistryuno/backend/utils"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupAdminLogsTest(t *testing.T) *gin.Engine {
	t.Helper()
	if err := utils.InitLogger(50); err != nil {
		t.Fatalf("init logger: %v", err)
	}
	utils.ClearLogs()
	t.Cleanup(func() {
		_ = utils.CloseLogger()
	})

	uid := 100000101
	otherUID := 100000102
	attemptedUID := 100000103
	utils.LogStructured(utils.LogEntry{
		Level:    "INFO",
		Category: "request",
		Message:  "GET /api/rooms",
		UID:      &uid,
		Source:   &utils.LogSource{ClientIP: "10.0.0.7", ForwardedFor: "203.0.113.10", UserAgent: "room-agent"},
		Request:  &utils.LogRequest{Method: "GET", Path: "/api/rooms", Status: 404, StatusClass: "4xx"},
	})
	utils.LogStructured(utils.LogEntry{
		Level:        "WARNING",
		Category:     "request",
		Message:      "GET /api/private",
		AttemptedUID: &attemptedUID,
		AuthState:    "invalid_session",
		Source:       &utils.LogSource{ClientIP: "10.0.0.8"},
		Request:      &utils.LogRequest{Method: "GET", Path: "/api/private", Status: 401, StatusClass: "4xx"},
	})
	utils.LogStructured(utils.LogEntry{
		Level:    "INFO",
		Category: "websocket",
		Message:  "websocket message uid=100000102",
		UID:      &otherUID,
		Source:   &utils.LogSource{ClientIP: "10.0.0.9"},
		WebSocket: &utils.LogWebSocket{
			Event: "message",
			Type:  "chat",
		},
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/admin/logs", GetLogs)
	router.GET("/api/admin/logs/stream", GetLogsStream)
	router.POST("/api/admin/logs/clear", ClearLogs)
	return router
}

func decodeAdminLogsResponse(t *testing.T, rec *httptest.ResponseRecorder) []map[string]interface{} {
	t.Helper()
	var body struct {
		Logs []map[string]interface{} `json:"logs"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode logs response: %v\nbody=%s", err, rec.Body.String())
	}
	return body.Logs
}

func decodeAdminLogStreamData(t *testing.T, body string) []map[string]interface{} {
	t.Helper()
	var logs []map[string]interface{}
	for _, line := range strings.Split(body, "\n") {
		payload, ok := strings.CutPrefix(line, "data: ")
		if !ok {
			continue
		}
		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(payload), &entry); err != nil {
			t.Fatalf("decode stream entry: %v\npayload=%s", err, payload)
		}
		logs = append(logs, entry)
	}
	return logs
}

func TestAdminLogsFilterByUIDSourceStatusAndKeyword(t *testing.T) {
	router := setupAdminLogsTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/logs?uid=100000101&source_ip=203.0.113&category=request&status_class=4xx&q=rooms", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	logs := decodeAdminLogsResponse(t, rec)
	if len(logs) != 1 {
		t.Fatalf("expected one filtered log, got %d: %#v", len(logs), logs)
	}
	if logs[0]["uid"] != float64(100000101) || logs[0]["category"] != "request" {
		t.Fatalf("unexpected filtered log: %#v", logs[0])
	}
}

func TestAdminLogsFilterByAttemptedUID(t *testing.T) {
	router := setupAdminLogsTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/logs?attempted_uid=100000103", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	logs := decodeAdminLogsResponse(t, rec)
	if len(logs) != 1 {
		t.Fatalf("expected one attempted UID log, got %d", len(logs))
	}
	if logs[0]["attempted_uid"] != float64(100000103) || logs[0]["auth_state"] != "invalid_session" {
		t.Fatalf("unexpected attempted UID log: %#v", logs[0])
	}
}

func TestAdminLogsStreamInitialSnapshotUsesSameFilters(t *testing.T) {
	router := setupAdminLogsTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/logs/stream?uid=100000101&source_ip=203.0.113&category=request&status_class=4xx&q=rooms&count=100", nil)
	ctx, cancel := context.WithCancel(req.Context())
	cancel()
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if contentType := rec.Header().Get("Content-Type"); !strings.Contains(contentType, "text/event-stream") {
		t.Fatalf("expected event-stream content type, got %q", contentType)
	}
	logs := decodeAdminLogStreamData(t, rec.Body.String())
	if len(logs) != 1 {
		t.Fatalf("expected one filtered stream snapshot log, got %d body=%s", len(logs), rec.Body.String())
	}
	if logs[0]["uid"] != float64(100000101) || logs[0]["category"] != "request" {
		t.Fatalf("unexpected streamed filtered log: %#v", logs[0])
	}
}

func TestAdminLogsClearPreservesBehavior(t *testing.T) {
	router := setupAdminLogsTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/logs/clear", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	if logs := utils.GetLogs(10); len(logs) != 0 {
		t.Fatalf("expected logs to be cleared, got %d", len(logs))
	}
}
