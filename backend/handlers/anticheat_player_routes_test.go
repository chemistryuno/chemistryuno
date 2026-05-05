package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"chemistryuno/backend/anticheat"
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPlayerAppealHandlerTest(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := database.MigrateCheatTables(db); err != nil {
		t.Fatalf("migrate cheat tables: %v", err)
	}

	cheatRepo := repository.NewCheatRepository(db)
	userRepo := repository.NewUserRepository(db)
	system := &anticheat.System{
		AppealManager: anticheat.NewAppealManager(cheatRepo, userRepo),
		Repository:    cheatRepo,
	}
	handler := NewAnticheatHandler(system)

	router := gin.New()
	router.GET("/api/player/appeals", func(c *gin.Context) {
		c.Set("uid", 1001)
		handler.GetPlayerAppeals(c)
	})
	router.GET("/api/player/appeals/anonymous", handler.GetPlayerAppeals)
	router.POST("/api/game/:roomId/appeal", func(c *gin.Context) {
		c.Set("uid", 1001)
		handler.SubmitAppeal(c)
	})
	router.POST("/api/game/:roomId/appeal/anonymous", handler.SubmitAppeal)

	return router, db
}

func TestPlayerAppealsRequireAuthenticatedUID(t *testing.T) {
	router, _ := setupPlayerAppealHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/player/appeals/anonymous", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for anonymous appeal history, got %d", w.Code)
	}
}

func TestSubmitAppealUsesAuthenticatedUID(t *testing.T) {
	router, db := setupPlayerAppealHandlerTest(t)
	body := map[string]any{
		"player_uid":    9999,
		"risk_score_id": 44,
		"reason":        "This detection was a false positive.",
		"evidence":      "Replay timing looked normal from my side.",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/game/room-1/appeal", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected appeal submission success, got %d: %s", w.Code, w.Body.String())
	}

	var appeal database.CheatAppeal
	if err := db.First(&appeal).Error; err != nil {
		t.Fatalf("load saved appeal: %v", err)
	}
	if appeal.PlayerUID != 1001 {
		t.Fatalf("expected authenticated uid 1001, got %d", appeal.PlayerUID)
	}
	if appeal.RoomID != "room-1" || appeal.Status != "pending" {
		t.Fatalf("unexpected appeal state: room=%q status=%q", appeal.RoomID, appeal.Status)
	}
}

func TestGetPlayerAppealsReturnsOnlyAuthenticatedPlayerAndNormalizedPayload(t *testing.T) {
	router, db := setupPlayerAppealHandlerTest(t)
	fixtures := []database.CheatAppeal{
		{RoomID: "own-new", PlayerUID: 1001, RiskScoreID: 1, Reason: "latest", Status: "pending"},
		{RoomID: "other", PlayerUID: 2002, RiskScoreID: 2, Reason: "other player", Status: "pending"},
		{RoomID: "own-old", PlayerUID: 1001, RiskScoreID: 3, Reason: "older", Status: "approved"},
	}
	for i := range fixtures {
		if err := db.Create(&fixtures[i]).Error; err != nil {
			t.Fatalf("create fixture: %v", err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/player/appeals", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected appeal history success, got %d: %s", w.Code, w.Body.String())
	}

	var payload struct {
		Appeals       []database.CheatAppeal `json:"appeals"`
		Count         int                    `json:"count"`
		Total         int                    `json:"total"`
		CurrentStatus string                 `json:"current_status"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Count != 2 || payload.Total != 2 || len(payload.Appeals) != 2 {
		t.Fatalf("expected only two authenticated player appeals, got count=%d total=%d len=%d", payload.Count, payload.Total, len(payload.Appeals))
	}
	for _, appeal := range payload.Appeals {
		if appeal.PlayerUID != 1001 {
			t.Fatalf("response leaked appeal for uid %d", appeal.PlayerUID)
		}
	}
	if payload.CurrentStatus == "" || payload.CurrentStatus == "none" {
		t.Fatalf("expected current_status to reflect latest appeal, got %q", payload.CurrentStatus)
	}
}
