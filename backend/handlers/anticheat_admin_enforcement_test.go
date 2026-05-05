package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"chemistryuno/backend/anticheat"
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAdminEnforcementTest(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&database.User{}, &database.FuelCompensationRecord{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}
	if err := database.MigrateCheatTables(db); err != nil {
		t.Fatalf("migrate cheat tables: %v", err)
	}
	database.DB = db

	if err := db.Create(&database.User{UID: 1001, Username: "player", Role: "user"}).Error; err != nil {
		t.Fatalf("create player: %v", err)
	}

	repo := repository.NewCheatRepository(db)
	handler := NewAnticheatHandler(&anticheat.System{Repository: repo})
	router := gin.New()
	router.POST("/api/admin/anticheat/ban", func(c *gin.Context) {
		c.Set("uid", 9001)
		handler.BanFromAnticheatPanel(c)
	})
	router.POST("/api/admin/anticheat/unban", func(c *gin.Context) {
		c.Set("uid", 9001)
		handler.UnbanFromAnticheatPanel(c)
	})
	return router, db
}

func TestAnticheatPanelBanWritesAuditLog(t *testing.T) {
	router, db := setupAdminEnforcementTest(t)
	body := map[string]any{
		"player_uid":   1001,
		"banned_until": time.Now().Add(2 * time.Hour).Format(time.RFC3339),
		"reason":       "manual anticheat enforcement",
		"room_id":      "room-a",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/anticheat/ban", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected ban success, got %d: %s", w.Code, w.Body.String())
	}

	var audit database.CheatAuditLog
	if err := db.Where("player_uid = ? AND event_type = ?", 1001, "ban").First(&audit).Error; err != nil {
		t.Fatalf("load ban audit: %v", err)
	}
	if audit.OperatorUID == nil || *audit.OperatorUID != 9001 {
		t.Fatalf("expected operator uid 9001, got %#v", audit.OperatorUID)
	}
	if audit.Remark != "manual anticheat enforcement" || audit.RoomID != "room-a" {
		t.Fatalf("unexpected audit context: remark=%q room=%q", audit.Remark, audit.RoomID)
	}
	var sanction database.CheatSanction
	if err := db.Where("player_uid = ? AND sanction_type = ? AND status = ?", 1001, "ban", "active").First(&sanction).Error; err != nil {
		t.Fatalf("expected manual ban to create active sanction: %v", err)
	}
	if sanction.EffectiveUntil == nil {
		t.Fatal("expected manual ban sanction to have effective_until")
	}
}

func TestAnticheatPanelUnbanWritesAuditLog(t *testing.T) {
	router, db := setupAdminEnforcementTest(t)
	bannedUntil := time.Now().Add(2 * time.Hour)
	if err := db.Model(&database.User{}).Where("uid = ?", 1001).Updates(map[string]interface{}{
		"banned_until": &bannedUntil,
		"ban_reason":   "active ban",
	}).Error; err != nil {
		t.Fatalf("seed account ban: %v", err)
	}
	if err := repository.NewCheatRepository(db).SaveSanction(&database.CheatSanction{
		RoomID:         "room-b",
		PlayerUID:      1001,
		RiskScoreID:    77,
		SanctionType:   "ban",
		RiskScore:      90,
		Reason:         "active sanction",
		EffectiveUntil: &bannedUntil,
		Status:         "active",
	}); err != nil {
		t.Fatalf("seed active sanction: %v", err)
	}

	body := map[string]any{
		"player_uid": 1001,
		"reason":     "appeal accepted manually",
		"room_id":    "room-b",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/anticheat/unban", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected unban success, got %d: %s", w.Code, w.Body.String())
	}

	var audit database.CheatAuditLog
	if err := db.Where("player_uid = ? AND event_type = ?", 1001, "unban").First(&audit).Error; err != nil {
		t.Fatalf("load unban audit: %v", err)
	}
	if audit.NewStatus != "revoked" || audit.Remark != "appeal accepted manually" {
		t.Fatalf("unexpected unban audit: status=%q remark=%q", audit.NewStatus, audit.Remark)
	}

	var user database.User
	if err := db.First(&user, 1001).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user.BannedUntil != nil || user.BanReason != "" {
		t.Fatalf("expected account ban cleared, got until=%v reason=%q", user.BannedUntil, user.BanReason)
	}
	var activeBans int64
	if err := db.Model(&database.CheatSanction{}).Where("player_uid = ? AND sanction_type = ? AND status = ?", 1001, "ban", "active").Count(&activeBans).Error; err != nil {
		t.Fatalf("count active bans: %v", err)
	}
	if activeBans != 0 {
		t.Fatalf("expected active ban sanctions revoked, got %d", activeBans)
	}
}

func TestDetectionListIncludesRealRiskAndSanctionData(t *testing.T) {
	_, db := setupAdminEnforcementTest(t)
	repo := repository.NewCheatRepository(db)
	handler := NewAnticheatHandler(&anticheat.System{Repository: repo})

	score := database.CheatRiskScore{
		RoomID:          "room-risk",
		PlayerUID:       1001,
		RiskScore:       91.5,
		ResponseTimeDim: 40,
		FrequencyDim:    20,
		WinRateDim:      15,
		PatternDim:      10,
		AccountAgeDim:   6.5,
	}
	if err := repo.SaveRiskScore(&score); err != nil {
		t.Fatalf("save risk score: %v", err)
	}
	if err := repo.SaveSanction(&database.CheatSanction{
		RoomID:       "room-risk",
		PlayerUID:    1001,
		RiskScoreID:  score.ID,
		SanctionType: "ban",
		RiskScore:    91.5,
		Reason:       "high risk score",
		Status:       "active",
	}); err != nil {
		t.Fatalf("save sanction: %v", err)
	}

	router := gin.New()
	router.GET("/api/admin/anticheat/detection-list", handler.GetDetectionList)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/anticheat/detection-list", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected detection list success, got %d: %s", w.Code, w.Body.String())
	}

	var payload struct {
		Detections []struct {
			PlayerUID    uint    `json:"player_uid"`
			RiskScore    float64 `json:"risk_score"`
			SanctionType string  `json:"sanction_type"`
		} `json:"detections"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode detection list: %v", err)
	}
	if payload.Total != 1 || len(payload.Detections) != 1 {
		t.Fatalf("expected one detection, got total=%d len=%d", payload.Total, len(payload.Detections))
	}
	got := payload.Detections[0]
	if got.PlayerUID != 1001 || got.RiskScore != 91.5 || got.SanctionType != "ban" {
		t.Fatalf("unexpected detection payload: %+v", got)
	}
}

func TestDetectionListIncludesGameHistoryCheatMarkers(t *testing.T) {
	_, db := setupAdminEnforcementTest(t)
	if err := db.AutoMigrate(&database.GameHistory{}); err != nil {
		t.Fatalf("migrate game history: %v", err)
	}
	repo := repository.NewCheatRepository(db)
	handler := NewAnticheatHandler(&anticheat.System{Repository: repo})

	cheatUIDs, _ := json.Marshal([]int{1001})
	players, _ := json.Marshal([]int{1001, 1002})
	history := database.GameHistory{
		RoomID:              "room-history",
		CheatDetected:       true,
		CheatUIDs:           cheatUIDs,
		Players:             players,
		OriginalPlayerCount: 2,
		StartedAt:           time.Now().Add(-10 * time.Minute),
		FinishedAt:          time.Now(),
	}
	if err := db.Create(&history).Error; err != nil {
		t.Fatalf("save game history: %v", err)
	}

	router := gin.New()
	router.GET("/api/admin/anticheat/detection-list", handler.GetDetectionList)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/anticheat/detection-list", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected detection list success, got %d: %s", w.Code, w.Body.String())
	}

	var payload struct {
		Detections []struct {
			ID           string `json:"id"`
			PlayerUID    uint   `json:"player_uid"`
			RoomID       string `json:"room_id"`
			Source       string `json:"source"`
			SanctionType string `json:"sanction_type"`
		} `json:"detections"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode detection list: %v", err)
	}
	if payload.Total != 1 || len(payload.Detections) != 1 {
		t.Fatalf("expected one detection, got total=%d len=%d", payload.Total, len(payload.Detections))
	}
	got := payload.Detections[0]
	if got.PlayerUID != 1001 || got.RoomID != "room-history" || got.Source != "game_history" || got.SanctionType != "observe" {
		t.Fatalf("unexpected history detection payload: %+v", got)
	}
	if got.ID == "" {
		t.Fatalf("expected synthetic history detection id")
	}
}

func TestBanStatsIncludeManualPanelBans(t *testing.T) {
	router, _ := setupAdminEnforcementTest(t)
	body := map[string]any{
		"player_uid":   1001,
		"banned_until": time.Now().Add(2 * time.Hour).Format(time.RFC3339),
		"reason":       "manual anticheat enforcement",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/anticheat/ban", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected ban success, got %d: %s", w.Code, w.Body.String())
	}

	repo := repository.NewCheatRepository(database.DB)
	count, err := repo.CountBansInRange(time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("count bans: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one manual panel ban in stats, got %d", count)
	}
}
