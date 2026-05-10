package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAdminBanTest(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&database.User{}, &database.UserSession{}); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	database.DB = db
	userRepo = repository.NewUserRepository(db)

	users := []database.User{
		{UID: 9001, Username: "admin", Role: "admin", IsAdmin: true},
		{UID: 1001, Username: "player", Role: "user"},
	}
	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			t.Fatalf("create user: %v", err)
		}
	}
	if err := db.Create(&database.UserSession{
		ID:         "target-session",
		UserUID:    1001,
		UserAgent:  "test",
		IPAddress:  "127.0.0.1",
		LastActive: time.Now(),
	}).Error; err != nil {
		t.Fatalf("create target session: %v", err)
	}

	router := gin.New()
	router.POST("/api/admin/users/ban", func(c *gin.Context) {
		c.Set("uid", 9001)
		c.Set("role", "admin")
		BanUser(c)
	})
	return router, db
}

func TestBanUserDoesNotDeleteActiveSessions(t *testing.T) {
	router, db := setupAdminBanTest(t)
	body := map[string]any{
		"target_uid":   1001,
		"banned_until": time.Now().Add(2 * time.Hour).Format(time.RFC3339),
		"reason":       "admin ban without forced logout",
	}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/ban", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected ban success, got %d: %s", w.Code, w.Body.String())
	}

	var user database.User
	if err := db.First(&user, 1001).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user.BannedUntil == nil || !user.BannedUntil.After(time.Now()) || user.BanReason != "admin ban without forced logout" {
		t.Fatalf("expected account ban to be written, got until=%v reason=%q", user.BannedUntil, user.BanReason)
	}

	var sessionCount int64
	if err := db.Model(&database.UserSession{}).Where("user_uid = ?", 1001).Count(&sessionCount).Error; err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if sessionCount != 1 {
		t.Fatalf("expected existing session to remain after ban, got %d", sessionCount)
	}
}

func TestBanUserSendsBanNotification(t *testing.T) {
	bannedUntil := time.Now().Add(2 * time.Hour).UTC()
	message := banNotificationMessage(&bannedUntil, "notify banned player")
	if message.Type != "ban_notification" {
		t.Fatalf("expected ban_notification message, got %q", message.Type)
	}
	data, ok := message.Data.(gin.H)
	if !ok {
		t.Fatalf("expected gin.H notification data, got %#v", message.Data)
	}
	if data["redirect_to"] != "/" || data["ban_reason"] != "notify banned player" || data["banned_until"] == "" {
		t.Fatalf("unexpected ban notification data: %#v", data)
	}
}
