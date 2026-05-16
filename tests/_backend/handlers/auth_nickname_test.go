package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuthNicknameTest(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&database.User{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}
	database.DB = db
	repository.UserRepo = repository.NewUserRepository(db)

	router := gin.New()
	router.POST("/api/auth/register", Register)
	return router, db
}

func TestRegisterWithBlankNicknameDoesNotCopyUsername(t *testing.T) {
	router, db := setupAuthNicknameTest(t)
	body := map[string]any{
		"username":          "nickname_source_user",
		"nickname":          "",
		"password":          "secret123",
		"security_question": "q",
		"security_answer":   "a",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected register success, got %d body=%s", rec.Code, rec.Body.String())
	}

	var user database.User
	if err := db.Where("username = ?", "nickname_source_user").First(&user).Error; err != nil {
		t.Fatalf("load created user: %v", err)
	}
	if user.Nickname == "" {
		t.Fatalf("expected generated nickname, got empty")
	}
	if user.Nickname == user.Username {
		t.Fatalf("nickname must not copy username, got %q", user.Nickname)
	}
}
