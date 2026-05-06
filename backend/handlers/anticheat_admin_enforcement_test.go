package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
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
	router.POST("/api/admin/anticheat/detection/:id/review", func(c *gin.Context) {
		c.Set("uid", 9001)
		handler.ReviewDetection(c)
	})
	router.POST("/api/admin/anticheat/detection/:id/punishment", func(c *gin.Context) {
		c.Set("uid", 9001)
		handler.ChangeDetectionPunishment(c)
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

func TestAnticheatPanelBanRetainsRiskEvidence(t *testing.T) {
	router, db := setupAdminEnforcementTest(t)
	repo := repository.NewCheatRepository(db)
	primary := anticheat.MarshalReplayEvidenceAnchor(database.ReplayEvidenceAnchor{
		RoomID:           "room-evidence",
		GameHistoryID:    77,
		ReplayID:         "77",
		EventIndex:       3,
		EventID:          "evt-3",
		EventType:        "play_card",
		PlayerUID:        1001,
		EventTimestampMs: 1710000000000,
		ActionSummary:    "played H2O",
	})
	related := anticheat.MarshalReplayEvidenceAnchors([]database.ReplayEvidenceAnchor{{
		RoomID:        "room-evidence",
		GameHistoryID: 77,
		ReplayID:      "77",
		EventIndex:    3,
		EventID:       "evt-3",
		EventType:     "play_card",
		PlayerUID:     1001,
	}})
	score := database.CheatRiskScore{
		RoomID:          "room-evidence",
		PlayerUID:       1001,
		ReplayID:        "77",
		GameHistoryID:   77,
		OperationIndex:  3,
		PrimaryEvidence: primary,
		RelatedEvidence: related,
		RiskScore:       91,
	}
	if err := repo.SaveRiskScore(&score); err != nil {
		t.Fatalf("save source risk score: %v", err)
	}

	body := map[string]any{
		"player_uid":    1001,
		"banned_until":  time.Now().Add(2 * time.Hour).Format(time.RFC3339),
		"reason":        "manual evidence enforcement",
		"risk_score_id": score.ID,
	}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/anticheat/ban", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected ban success, got %d: %s", w.Code, w.Body.String())
	}

	var sanction database.CheatSanction
	if err := db.Where("risk_score_id = ? AND sanction_type = ?", score.ID, "ban").First(&sanction).Error; err != nil {
		t.Fatalf("load evidence sanction: %v", err)
	}
	if sanction.ReplayID != "77" || sanction.GameHistoryID != 77 || len(sanction.PrimaryEvidence) == 0 {
		t.Fatalf("expected sanction to retain replay evidence, got replay=%q history=%d evidence=%s", sanction.ReplayID, sanction.GameHistoryID, string(sanction.PrimaryEvidence))
	}

	var audit database.CheatAuditLog
	if err := db.Where("risk_score_id = ? AND event_type = ?", score.ID, "ban").First(&audit).Error; err != nil {
		t.Fatalf("load evidence audit: %v", err)
	}
	if audit.ReplayID != "77" || audit.GameHistoryID != 77 || len(audit.PrimaryEvidence) == 0 || len(audit.RelatedEvidence) == 0 {
		t.Fatalf("expected audit to retain replay evidence, got replay=%q history=%d primary=%s related=%s", audit.ReplayID, audit.GameHistoryID, string(audit.PrimaryEvidence), string(audit.RelatedEvidence))
	}
}

func TestAnticheatDetectionReviewBanEnforcesAccountBan(t *testing.T) {
	router, db := setupAdminEnforcementTest(t)
	repo := repository.NewCheatRepository(db)
	primary := anticheat.MarshalReplayEvidenceAnchor(database.ReplayEvidenceAnchor{
		RoomID:        "room-review",
		GameHistoryID: 88,
		ReplayID:      "88",
		EventIndex:    4,
		EventID:       "evt-review-ban",
		EventType:     "play_card",
		PlayerUID:     1001,
	})
	score := database.CheatRiskScore{
		RoomID:             "room-review",
		PlayerUID:          1001,
		ReplayID:           "88",
		GameHistoryID:      88,
		OperationIndex:     4,
		PrimaryEvidence:    primary,
		RiskScore:          93,
		SuggestedAction:    "ban",
		PunishmentDecision: "none",
		ReviewStatus:       "pending",
		SuggestionReason:   "high confidence replay evidence",
	}
	if err := repo.SaveRiskScore(&score); err != nil {
		t.Fatalf("save review risk score: %v", err)
	}

	body := map[string]any{
		"decision": "confirm",
		"remark":   "confirmed ban after replay review",
	}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/anticheat/detection/"+strconv.Itoa(int(score.ID))+"/review", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected review success, got %d: %s", w.Code, w.Body.String())
	}

	var user database.User
	if err := db.First(&user, 1001).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user.BannedUntil == nil || !user.BannedUntil.After(time.Now()) || user.BanReason != "confirmed ban after replay review" {
		t.Fatalf("expected account ban from review, got until=%v reason=%q", user.BannedUntil, user.BanReason)
	}
	var reviewed database.CheatRiskScore
	if err := db.First(&reviewed, score.ID).Error; err != nil {
		t.Fatalf("load reviewed score: %v", err)
	}
	if reviewed.ReviewStatus != "processed" || reviewed.PunishmentDecision != "ban" {
		t.Fatalf("expected review to keep ban decision, got review=%q punishment=%q", reviewed.ReviewStatus, reviewed.PunishmentDecision)
	}
	var sanction database.CheatSanction
	if err := db.Where("risk_score_id = ? AND sanction_type = ? AND status = ?", score.ID, "ban", "active").First(&sanction).Error; err != nil {
		t.Fatalf("expected review to create active ban sanction: %v", err)
	}
	if sanction.EffectiveUntil == nil || sanction.ReplayID != "88" || len(sanction.PrimaryEvidence) == 0 {
		t.Fatalf("expected review sanction to retain ban expiry and evidence, got %+v", sanction)
	}
	var banAudit database.CheatAuditLog
	if err := db.Where("risk_score_id = ? AND event_type = ?", score.ID, "ban").First(&banAudit).Error; err != nil {
		t.Fatalf("expected review ban audit: %v", err)
	}
	if banAudit.OperatorUID == nil || *banAudit.OperatorUID != 9001 || banAudit.SanctionID == nil {
		t.Fatalf("unexpected review ban audit: %+v", banAudit)
	}
}

func TestAnticheatPunishmentChangeBanEnforcesAccountBan(t *testing.T) {
	router, db := setupAdminEnforcementTest(t)
	repo := repository.NewCheatRepository(db)
	score := database.CheatRiskScore{
		RoomID:             "room-change",
		PlayerUID:          1001,
		ReplayID:           "99",
		GameHistoryID:      99,
		OperationIndex:     7,
		PrimaryEvidence:    anticheat.MarshalReplayEvidenceAnchor(database.ReplayEvidenceAnchor{RoomID: "room-change", GameHistoryID: 99, ReplayID: "99", EventIndex: 7, PlayerUID: 1001}),
		RiskScore:          82,
		SuggestedAction:    "mute",
		PunishmentDecision: "mute",
		ReviewStatus:       "processed",
	}
	if err := repo.SaveRiskScore(&score); err != nil {
		t.Fatalf("save change risk score: %v", err)
	}
	duration := 45
	effectiveUntil := time.Now().Add(time.Duration(duration) * time.Minute).UTC().Format(time.RFC3339)
	body := map[string]any{
		"punishment_decision": "ban",
		"reason":              "escalated after admin review",
		"duration":            duration,
		"effective_until":     effectiveUntil,
	}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/anticheat/detection/"+strconv.Itoa(int(score.ID))+"/punishment", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected punishment change success, got %d: %s", w.Code, w.Body.String())
	}

	var user database.User
	if err := db.First(&user, 1001).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user.BannedUntil == nil || !user.BannedUntil.After(time.Now()) || user.BanReason != "escalated after admin review" {
		t.Fatalf("expected account ban from punishment change, got until=%v reason=%q", user.BannedUntil, user.BanReason)
	}
	var sanction database.CheatSanction
	if err := db.Where("risk_score_id = ? AND sanction_type = ? AND status = ?", score.ID, "ban", "active").First(&sanction).Error; err != nil {
		t.Fatalf("expected punishment change to create active ban sanction: %v", err)
	}
	if sanction.Duration == nil || *sanction.Duration != duration || sanction.EffectiveUntil == nil {
		t.Fatalf("expected punishment change sanction duration/expiry, got %+v", sanction)
	}
	var changed database.CheatRiskScore
	if err := db.First(&changed, score.ID).Error; err != nil {
		t.Fatalf("load changed score: %v", err)
	}
	if changed.PunishmentDecision != "ban" {
		t.Fatalf("expected changed punishment ban, got %q", changed.PunishmentDecision)
	}
	var banAudit database.CheatAuditLog
	if err := db.Where("risk_score_id = ? AND event_type = ?", score.ID, "ban").First(&banAudit).Error; err != nil {
		t.Fatalf("expected punishment change ban audit: %v", err)
	}
	if banAudit.SanctionID == nil || banAudit.Remark != "escalated after admin review" {
		t.Fatalf("unexpected punishment change ban audit: %+v", banAudit)
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
