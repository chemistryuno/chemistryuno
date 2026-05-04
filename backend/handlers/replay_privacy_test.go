package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupReplayPrivacyTest(t *testing.T) *gin.Engine {
	t.Helper()

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        fmt.Sprintf("file:replay_privacy_%s?mode=memory&cache=shared", strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())),
	}, &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite test db: %v", err)
	}
	if err := db.AutoMigrate(&database.User{}, &database.GameHistory{}); err != nil {
		t.Fatalf("migrate replay privacy tables: %v", err)
	}

	database.DB = db
	repository.UserRepo = repository.NewUserRepository(db)
	repository.GameRepo = repository.NewGameRepository()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("uid", 1001)
		c.Next()
	})
	router.GET("/user/game-history", GetMyGameHistory)
	router.GET("/user/game-history/:id/replay", GetMyGameReplay)
	router.GET("/admin/game-history/:id/replay", GetAdminGameReplay)
	return router
}

func seedReplayPrivacyHistory(t *testing.T) uint {
	t.Helper()

	user := database.User{
		UID:      1001,
		Username: "player1001",
		Nickname: "Player 1001",
	}
	if err := repository.UserRepo.Create(&user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour).UTC()
	players, _ := json.Marshal([]int{1001, 1002})
	cheatUIDs, _ := json.Marshal([]int{1001})
	replay := `{"version":1,"cheat_detected":true,"cheat_uids":[1001],"events":[],"participants":[{"uid":1001,"nickname":"Player 1001","is_ai":false}]}`
	history := database.GameHistory{
		RoomID:          "privacy_room",
		Players:         players,
		ReplayLog:       replay,
		ReplayPermanent: true,
		ReplayExpiresAt: &expiresAt,
		CheatDetected:   true,
		CheatUIDs:       cheatUIDs,
		StartedAt:       time.Now().Add(-time.Hour),
		FinishedAt:      time.Now(),
	}
	if err := repository.GameRepo.Create(&history); err != nil {
		t.Fatalf("create game history: %v", err)
	}
	return history.ID
}

func decodeReplayPrivacyResponse(t *testing.T, rec *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v\nbody=%s", err, rec.Body.String())
	}
	return body
}

func TestPlayerGameHistoryOmitsAnticheatFields(t *testing.T) {
	router := setupReplayPrivacyTest(t)
	seedReplayPrivacyHistory(t)

	req := httptest.NewRequest(http.MethodGet, "/user/game-history", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode history response: %v", err)
	}
	if len(body) != 1 {
		t.Fatalf("expected one history item, got %d", len(body))
	}
	forbidden := []string{"cheat_detected", "cheat_uids", "replay_permanent"}
	for _, key := range forbidden {
		if _, exists := body[0][key]; exists {
			t.Fatalf("player history leaked %s: %#v", key, body[0])
		}
	}
	if _, exists := body[0]["replay_expires_at"]; !exists {
		t.Fatalf("player history should include normal replay expiration: %#v", body[0])
	}
}

func TestPlayerReplayResponseSanitizesAnticheatFields(t *testing.T) {
	router := setupReplayPrivacyTest(t)
	historyID := seedReplayPrivacyHistory(t)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/user/game-history/%d/replay", historyID), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	body := decodeReplayPrivacyResponse(t, rec)
	forbidden := []string{"cheat_detected", "cheat_uids", "replay_permanent"}
	for _, key := range forbidden {
		if _, exists := body[key]; exists {
			t.Fatalf("player replay leaked top-level %s: %#v", key, body)
		}
	}
	replay, ok := body["replay"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected replay object, got %#v", body["replay"])
	}
	forbiddenReplayKeys := []string{"cheat_detected", "cheat_uids"}
	for _, key := range forbiddenReplayKeys {
		if _, exists := replay[key]; exists {
			t.Fatalf("player replay leaked embedded %s: %#v", key, replay)
		}
	}
	if _, exists := body["replay_expires_at"]; !exists {
		t.Fatalf("player replay should include normal replay expiration: %#v", body)
	}
}

func TestAdminReplayResponseRetainsAnticheatFields(t *testing.T) {
	router := setupReplayPrivacyTest(t)
	historyID := seedReplayPrivacyHistory(t)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/game-history/%d/replay", historyID), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	body := decodeReplayPrivacyResponse(t, rec)
	for _, key := range []string{"cheat_detected", "cheat_uids", "replay_permanent"} {
		if _, exists := body[key]; !exists {
			t.Fatalf("admin replay missing %s: %#v", key, body)
		}
	}
	if body["cheat_detected"] != true {
		t.Fatalf("admin replay should retain cheat_detected=true, got %#v", body["cheat_detected"])
	}
}
