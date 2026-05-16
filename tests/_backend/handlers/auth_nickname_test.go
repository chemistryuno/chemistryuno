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
	previousDB := database.DB
	previousUserRepo := repository.UserRepo
	database.DB = db
	repository.UserRepo = repository.NewUserRepository(db)
	t.Cleanup(func() {
		database.DB = previousDB
		repository.UserRepo = previousUserRepo
	})

	router := gin.New()
	router.POST("/api/auth/register", Register)
	router.PUT("/api/user/profile", func(c *gin.Context) {
		c.Set("uid", 100000321)
		UpdateProfile(c)
	})
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

func TestUpdateProfileWithoutNicknamePreservesExistingNickname(t *testing.T) {
	router, db := setupAuthNicknameTest(t)
	user := database.User{
		UID:      100000321,
		Username: "profile_user",
		Nickname: "已有昵称",
		Password: "hash",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	body := map[string]any{
		"enable_element_input": false,
	}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/api/user/profile", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected profile update success, got %d body=%s", rec.Code, rec.Body.String())
	}

	var updated database.User
	if err := db.First(&updated, user.UID).Error; err != nil {
		t.Fatalf("load updated user: %v", err)
	}
	if updated.Nickname != "已有昵称" {
		t.Fatalf("expected existing nickname preserved, got %q", updated.Nickname)
	}
	if updated.EnableElementInput {
		t.Fatalf("expected enable_element_input to be updated")
	}
}

func TestUpdateProfileWithBlankNicknameIsRejected(t *testing.T) {
	router, db := setupAuthNicknameTest(t)
	user := database.User{
		UID:      100000321,
		Username: "profile_user",
		Nickname: "已有昵称",
		Password: "hash",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	body := map[string]any{
		"nickname": "   ",
	}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/api/user/profile", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected blank nickname rejection, got %d body=%s", rec.Code, rec.Body.String())
	}

	var updated database.User
	if err := db.First(&updated, user.UID).Error; err != nil {
		t.Fatalf("load updated user: %v", err)
	}
	if updated.Nickname != "已有昵称" {
		t.Fatalf("rejected update should preserve nickname, got %q", updated.Nickname)
	}
}
