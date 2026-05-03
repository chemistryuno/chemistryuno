package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetVersionContractUsesEnvironmentValues(t *testing.T) {
	t.Setenv("APP_VERSION", "9.9.9")
	t.Setenv("APP_VERSION_NAME", "Contract")
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/api/version", GetVersion)

	req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body["version"] != "9.9.9" {
		t.Fatalf("unexpected version: %q", body["version"])
	}
	if body["versionName"] != "Contract" {
		t.Fatalf("unexpected versionName: %q", body["versionName"])
	}
	if body["fullVersion"] != "V9.9.9 Contract" {
		t.Fatalf("unexpected fullVersion: %q", body["fullVersion"])
	}
}

func TestCreateRoomContractRejectsInvalidMaxPlayers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/api/rooms", func(c *gin.Context) {
		c.Set("uid", 100000000)
		CreateRoom(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/rooms", nil)
	req.Body = http.NoBody
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for invalid body, got %d: %s", rec.Code, rec.Body.String())
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["error"] == "" {
		t.Fatalf("expected validation error payload, got %v", body)
	}
}
