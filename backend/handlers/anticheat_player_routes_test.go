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
		Decider:       anticheat.NewSanctionDecider(nil, cheatRepo),
		Repository:    cheatRepo,
	}
	handler := NewAnticheatHandler(system)

	router := gin.New()
	router.GET("/api/player/appeals", func(c *gin.Context) {
		c.Set("uid", 1001)
		handler.GetPlayerAppeals(c)
	})
	router.GET("/api/player/appeals/entry", func(c *gin.Context) {
		c.Set("uid", 1001)
		handler.GetAppealEntryStatus(c)
	})
	router.GET("/api/player/appeals/anonymous", handler.GetPlayerAppeals)
	router.POST("/api/player/appeals/:id/claim", func(c *gin.Context) {
		c.Set("uid", 1001)
		handler.ClaimAppealCompensation(c)
	})
	router.POST("/api/player/appeals/:id/claim/anonymous", handler.ClaimAppealCompensation)
	router.POST("/api/game/:roomId/appeal", func(c *gin.Context) {
		c.Set("uid", 1001)
		handler.SubmitAppeal(c)
	})
	router.POST("/api/game/:roomId/appeal/anonymous", handler.SubmitAppeal)

	return router, db
}

func seedBannedAppealContext(t *testing.T, db *gorm.DB, uid uint, roomID string, riskScoreID uint) {
	t.Helper()
	bannedUntil := time.Now().Add(2 * time.Hour)
	if err := db.AutoMigrate(&database.User{}, &database.GameHistory{}); err != nil {
		t.Fatalf("migrate appeal context tables: %v", err)
	}
	if err := db.Create(&database.User{UID: uid, Username: "appeal-user", BannedUntil: &bannedUntil, BanReason: "anticheat ban"}).Error; err != nil {
		t.Fatalf("create banned user: %v", err)
	}
	score := database.CheatRiskScore{
		ID:             riskScoreID,
		RoomID:         roomID,
		PlayerUID:      uid,
		ReplayID:       "77",
		GameHistoryID:  77,
		OperationIndex: 3,
		PrimaryEvidence: anticheat.MarshalReplayEvidenceAnchor(database.ReplayEvidenceAnchor{
			RoomID:           roomID,
			GameHistoryID:    77,
			ReplayID:         "77",
			EventIndex:       3,
			EventID:          "evt-3",
			EventType:        "play_card",
			PlayerUID:        uid,
			EventTimestampMs: 1710000000000,
			ActionSummary:    "played H2O",
		}),
		RelatedEvidence: anticheat.MarshalReplayEvidenceAnchors([]database.ReplayEvidenceAnchor{{
			RoomID:        roomID,
			GameHistoryID: 77,
			ReplayID:      "77",
			EventIndex:    3,
			EventID:       "evt-3",
			EventType:     "play_card",
			PlayerUID:     uid,
		}}),
		RiskScore:          88,
		SuggestedAction:    "ban",
		PunishmentDecision: "ban",
		ReviewStatus:       "processed",
		DetectionTime:      time.Now().Add(-time.Minute),
	}
	if err := db.Create(&score).Error; err != nil {
		t.Fatalf("create risk score: %v", err)
	}
	players, _ := json.Marshal([]uint{uid})
	if err := db.Create(&database.GameHistory{
		RoomID:     roomID,
		Players:    players,
		StartedAt:  time.Now().Add(-2 * time.Minute),
		FinishedAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("create game history: %v", err)
	}
}

func seedActiveBanSanction(t *testing.T, db *gorm.DB, uid uint, roomID string, riskScoreID uint, bannedUntil time.Time) database.CheatSanction {
	t.Helper()
	sanction := database.CheatSanction{
		RoomID:         roomID,
		PlayerUID:      uid,
		RiskScoreID:    riskScoreID,
		SanctionType:   "ban",
		RiskScore:      90,
		Reason:         "active anticheat ban",
		EffectiveUntil: &bannedUntil,
		Status:         "active",
	}
	if err := db.Create(&sanction).Error; err != nil {
		t.Fatalf("create active ban sanction: %v", err)
	}
	return sanction
}

func assertAccountBanActive(t *testing.T, db *gorm.DB, uid uint) {
	t.Helper()
	var user database.User
	if err := db.First(&user, uid).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user.BannedUntil == nil || !user.BannedUntil.After(time.Now()) || user.BanReason == "" {
		t.Fatalf("expected account ban to remain active, got until=%v reason=%q", user.BannedUntil, user.BanReason)
	}
}

func assertSanctionActive(t *testing.T, db *gorm.DB, sanctionID uint) {
	t.Helper()
	var sanction database.CheatSanction
	if err := db.First(&sanction, sanctionID).Error; err != nil {
		t.Fatalf("load sanction: %v", err)
	}
	if sanction.Status != "active" || sanction.SanctionType != "ban" {
		t.Fatalf("expected ban sanction to remain active, got type=%q status=%q", sanction.SanctionType, sanction.Status)
	}
}

func assertNoUnbanAudit(t *testing.T, db *gorm.DB, uid uint) {
	t.Helper()
	var unbanAudits int64
	if err := db.Model(&database.CheatAuditLog{}).Where("player_uid = ? AND event_type = ?", uid, "unban").Count(&unbanAudits).Error; err != nil {
		t.Fatalf("count unban audits: %v", err)
	}
	if unbanAudits != 0 {
		t.Fatalf("expected no unban audit events, got %d", unbanAudits)
	}
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
	seedBannedAppealContext(t, db, 1001, "room-1", 44)
	var bannedUser database.User
	if err := db.First(&bannedUser, 1001).Error; err != nil {
		t.Fatalf("load banned user: %v", err)
	}
	sanction := seedActiveBanSanction(t, db, 1001, "room-1", 44, *bannedUser.BannedUntil)
	body := map[string]any{
		"player_uid":    9999,
		"risk_score_id": 44,
		"sanction_id":   sanction.ID,
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
	if appeal.RoomID != "room-1" || appeal.Status != "pending" || appeal.SanctionID != sanction.ID {
		t.Fatalf("unexpected appeal state: room=%q status=%q", appeal.RoomID, appeal.Status)
	}
	var roomIDs []string
	if err := json.Unmarshal(appeal.RoomIDs, &roomIDs); err != nil {
		t.Fatalf("decode locked room ids: %v", err)
	}
	if len(roomIDs) != 1 || roomIDs[0] != "room-1" {
		t.Fatalf("expected locked room list from server, got %#v", roomIDs)
	}
	assertAccountBanActive(t, db, 1001)
	assertSanctionActive(t, db, sanction.ID)
	assertNoUnbanAudit(t, db, 1001)
}

func TestSubmitAppealRejectsUnbannedPlayer(t *testing.T) {
	router, _ := setupPlayerAppealHandlerTest(t)
	payload, _ := json.Marshal(map[string]any{
		"reason": "I am not banned.",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/game/room-1/appeal", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected unbanned appeal to be forbidden, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSubmitAppealRequiresOnlyReasonForBannedPlayer(t *testing.T) {
	router, db := setupPlayerAppealHandlerTest(t)
	bannedUntil := time.Now().Add(2 * time.Hour)
	if err := db.AutoMigrate(&database.User{}); err != nil {
		t.Fatalf("migrate user table: %v", err)
	}
	if err := db.Create(&database.User{
		UID:         1001,
		Username:    "reason-only-appeal",
		BannedUntil: &bannedUntil,
		BanReason:   "manual account ban",
	}).Error; err != nil {
		t.Fatalf("create banned user: %v", err)
	}
	payload, _ := json.Marshal(map[string]any{
		"reason": "Please review my account ban.",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/game/account/appeal", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected reason-only appeal success, got %d: %s", w.Code, w.Body.String())
	}
	var appeal database.CheatAppeal
	if err := db.First(&appeal).Error; err != nil {
		t.Fatalf("load saved appeal: %v", err)
	}
	if appeal.RoomID != "account" || appeal.PlayerUID != 1001 || appeal.RiskScoreID != 0 || appeal.SanctionID != 0 || appeal.Evidence != "" || appeal.Status != "pending" {
		t.Fatalf("unexpected reason-only appeal: %+v", appeal)
	}
	assertAccountBanActive(t, db, 1001)
	assertNoUnbanAudit(t, db, 1001)
}

func TestAppealEntryUsesActiveBanSanctionWhenAccountBanMissing(t *testing.T) {
	router, db := setupPlayerAppealHandlerTest(t)
	if err := db.AutoMigrate(&database.User{}, &database.GameHistory{}); err != nil {
		t.Fatalf("migrate context tables: %v", err)
	}
	if err := db.Create(&database.User{UID: 1001, Username: "sanction-only-ban"}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	score := database.CheatRiskScore{
		ID:                 55,
		RoomID:             "room-sanction-only",
		PlayerUID:          1001,
		RiskScore:          91,
		SuggestedAction:    "ban",
		PunishmentDecision: "ban",
		ReviewStatus:       "processed",
		DetectionTime:      time.Now().Add(-time.Minute),
	}
	if err := db.Create(&score).Error; err != nil {
		t.Fatalf("create risk score: %v", err)
	}
	bannedUntil := time.Now().Add(2 * time.Hour)
	if err := db.Create(&database.CheatSanction{
		RoomID:         "room-sanction-only",
		PlayerUID:      1001,
		RiskScoreID:    score.ID,
		SanctionType:   "ban",
		RiskScore:      91,
		Reason:         "active anticheat sanction",
		EffectiveUntil: &bannedUntil,
		Status:         "active",
	}).Error; err != nil {
		t.Fatalf("create ban sanction: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/player/appeals/entry", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected appeal entry success, got %d: %s", w.Code, w.Body.String())
	}

	var payload struct {
		IsBanned          bool     `json:"is_banned"`
		BanReason         string   `json:"ban_reason"`
		LatestRiskScoreID uint     `json:"latest_risk_score_id"`
		RoomIDs           []string `json:"room_ids"`
		CanSubmit         bool     `json:"can_submit"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !payload.IsBanned || !payload.CanSubmit || payload.BanReason != "active anticheat sanction" {
		t.Fatalf("expected active sanction to drive banned entry state, got %+v", payload)
	}
	if payload.LatestRiskScoreID != score.ID || len(payload.RoomIDs) != 1 || payload.RoomIDs[0] != "room-sanction-only" {
		t.Fatalf("unexpected appeal context from sanction-only ban: %+v", payload)
	}
}

func TestAppealEntryPreservesAccountBanAndActiveSanction(t *testing.T) {
	router, db := setupPlayerAppealHandlerTest(t)
	seedBannedAppealContext(t, db, 1001, "room-entry-preserve", 58)
	var bannedUser database.User
	if err := db.First(&bannedUser, 1001).Error; err != nil {
		t.Fatalf("load banned user: %v", err)
	}
	sanction := seedActiveBanSanction(t, db, 1001, "room-entry-preserve", 58, *bannedUser.BannedUntil)

	req := httptest.NewRequest(http.MethodGet, "/api/player/appeals/entry", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected appeal entry success, got %d: %s", w.Code, w.Body.String())
	}
	var payload struct {
		IsBanned  bool     `json:"is_banned"`
		CanSubmit bool     `json:"can_submit"`
		RoomIDs   []string `json:"room_ids"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !payload.IsBanned || !payload.CanSubmit || len(payload.RoomIDs) == 0 || payload.RoomIDs[0] != "room-entry-preserve" {
		t.Fatalf("unexpected appeal entry state: %+v", payload)
	}
	assertAccountBanActive(t, db, 1001)
	assertSanctionActive(t, db, sanction.ID)
	assertNoUnbanAudit(t, db, 1001)
}

func TestAppealEntryUsesActiveBanSanctionWithoutDecider(t *testing.T) {
	router, db := setupPlayerAppealHandlerTest(t)
	router = gin.New()
	repo := repository.NewCheatRepository(db)
	handler := NewAnticheatHandler(&anticheat.System{
		AppealManager: anticheat.NewAppealManager(repo, repository.NewUserRepository(db)),
		Repository:    repo,
	})
	router.GET("/api/player/appeals/entry", func(c *gin.Context) {
		c.Set("uid", 1001)
		handler.GetAppealEntryStatus(c)
	})

	if err := db.AutoMigrate(&database.User{}, &database.GameHistory{}); err != nil {
		t.Fatalf("migrate context tables: %v", err)
	}
	if err := db.Create(&database.User{UID: 1001, Username: "sanction-only-ban"}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	score := database.CheatRiskScore{
		ID:                 56,
		RoomID:             "room-no-decider",
		PlayerUID:          1001,
		RiskScore:          91,
		SuggestedAction:    "ban",
		PunishmentDecision: "ban",
		ReviewStatus:       "processed",
		DetectionTime:      time.Now().Add(-time.Minute),
	}
	if err := db.Create(&score).Error; err != nil {
		t.Fatalf("create risk score: %v", err)
	}
	bannedUntil := time.Now().Add(2 * time.Hour)
	if err := db.Create(&database.CheatSanction{
		RoomID:         "room-no-decider",
		PlayerUID:      1001,
		RiskScoreID:    score.ID,
		SanctionType:   "ban",
		RiskScore:      91,
		Reason:         "active ban without decider",
		EffectiveUntil: &bannedUntil,
		Status:         "active",
	}).Error; err != nil {
		t.Fatalf("create ban sanction: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/player/appeals/entry", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected appeal entry success, got %d: %s", w.Code, w.Body.String())
	}
	var payload struct {
		IsBanned  bool   `json:"is_banned"`
		BanReason string `json:"ban_reason"`
		CanSubmit bool   `json:"can_submit"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !payload.IsBanned || !payload.CanSubmit || payload.BanReason != "active ban without decider" {
		t.Fatalf("expected repository fallback to detect active ban, got %+v", payload)
	}
}

func TestAppealEntryTreatsBlankSanctionStatusAsActive(t *testing.T) {
	router, db := setupPlayerAppealHandlerTest(t)
	if err := db.AutoMigrate(&database.User{}, &database.GameHistory{}); err != nil {
		t.Fatalf("migrate context tables: %v", err)
	}
	if err := db.Create(&database.User{UID: 1001, Username: "legacy-sanction"}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	score := database.CheatRiskScore{
		ID:                 57,
		RoomID:             "room-legacy-status",
		PlayerUID:          1001,
		RiskScore:          91,
		SuggestedAction:    "ban",
		PunishmentDecision: "ban",
		ReviewStatus:       "processed",
		DetectionTime:      time.Now().Add(-time.Minute),
	}
	if err := db.Create(&score).Error; err != nil {
		t.Fatalf("create risk score: %v", err)
	}
	bannedUntil := time.Now().Add(2 * time.Hour)
	if err := db.Create(&database.CheatSanction{
		RoomID:         "room-legacy-status",
		PlayerUID:      1001,
		RiskScoreID:    score.ID,
		SanctionType:   "ban",
		RiskScore:      91,
		Reason:         "legacy active ban",
		EffectiveUntil: &bannedUntil,
		Status:         "",
	}).Error; err != nil {
		t.Fatalf("create legacy ban sanction: %v", err)
	}
	if err := db.Model(&database.CheatSanction{}).Where("risk_score_id = ?", score.ID).Update("status", "").Error; err != nil {
		t.Fatalf("force blank status: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/player/appeals/entry", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected appeal entry success, got %d: %s", w.Code, w.Body.String())
	}
	var payload struct {
		IsBanned  bool   `json:"is_banned"`
		BanReason string `json:"ban_reason"`
		CanSubmit bool   `json:"can_submit"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !payload.IsBanned || !payload.CanSubmit || payload.BanReason != "legacy active ban" {
		t.Fatalf("expected blank status sanction to count as active ban, got %+v", payload)
	}
}

func TestClaimAppealCompensationUsesAuthenticatedUID(t *testing.T) {
	router, db := setupPlayerAppealHandlerTest(t)
	database.DB = db
	if err := db.AutoMigrate(&database.User{}, &database.FuelCompensationRecord{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}
	if err := db.Create(&database.User{UID: 1001, Username: "claimant", Points: 1000}).Error; err != nil {
		t.Fatalf("create claimant: %v", err)
	}
	appeal := database.CheatAppeal{
		RoomID:             "room-claim",
		PlayerUID:          1001,
		Reason:             "accepted",
		Status:             "approved",
		CompensationAmount: 125,
		CompensationStatus: "pending",
	}
	if err := db.Create(&appeal).Error; err != nil {
		t.Fatalf("create appeal: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/player/appeals/"+strconv.Itoa(int(appeal.ID))+"/claim", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected compensation claim success, got %d: %s", w.Code, w.Body.String())
	}
	var payload struct {
		CompensationStatus string `json:"compensation_status"`
		Compensation       struct {
			Amount int `json:"amount"`
		} `json:"compensation"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode claim payload: %v", err)
	}
	if payload.CompensationStatus != "ok" || payload.Compensation.Amount != 125 {
		t.Fatalf("unexpected claim payload: %+v", payload)
	}

	var user database.User
	if err := db.First(&user, 1001).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user.Fuel != 125 || user.Points != 1125 {
		t.Fatalf("expected claimed balances, got fuel=%d points=%.0f", user.Fuel, user.Points)
	}
	var savedAppeal database.CheatAppeal
	if err := db.First(&savedAppeal, appeal.ID).Error; err != nil {
		t.Fatalf("load saved appeal: %v", err)
	}
	if savedAppeal.CompensationStatus != "ok" {
		t.Fatalf("expected saved appeal compensation ok, got %q", savedAppeal.CompensationStatus)
	}

	duplicateReq := httptest.NewRequest(http.MethodPost, "/api/player/appeals/"+strconv.Itoa(int(appeal.ID))+"/claim", nil)
	duplicateW := httptest.NewRecorder()
	router.ServeHTTP(duplicateW, duplicateReq)
	if duplicateW.Code != http.StatusOK {
		t.Fatalf("expected duplicate claim success, got %d: %s", duplicateW.Code, duplicateW.Body.String())
	}
	if err := db.First(&user, 1001).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if user.Fuel != 125 || user.Points != 1125 {
		t.Fatalf("duplicate claim should not change balances, got fuel=%d points=%.0f", user.Fuel, user.Points)
	}
}

func TestClaimAppealCompensationRejectsAnonymousAndOtherPlayers(t *testing.T) {
	router, db := setupPlayerAppealHandlerTest(t)
	database.DB = db
	if err := db.AutoMigrate(&database.User{}, &database.FuelCompensationRecord{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}
	if err := db.Create(&database.User{UID: 2002, Username: "other", Points: 1000}).Error; err != nil {
		t.Fatalf("create other player: %v", err)
	}
	appeal := database.CheatAppeal{
		RoomID:             "room-claim",
		PlayerUID:          2002,
		Reason:             "accepted",
		Status:             "approved",
		CompensationAmount: 125,
		CompensationStatus: "pending",
	}
	if err := db.Create(&appeal).Error; err != nil {
		t.Fatalf("create appeal: %v", err)
	}

	anonymousReq := httptest.NewRequest(http.MethodPost, "/api/player/appeals/"+strconv.Itoa(int(appeal.ID))+"/claim/anonymous", nil)
	anonymousW := httptest.NewRecorder()
	router.ServeHTTP(anonymousW, anonymousReq)
	if anonymousW.Code != http.StatusUnauthorized {
		t.Fatalf("expected anonymous claim 401, got %d", anonymousW.Code)
	}

	otherReq := httptest.NewRequest(http.MethodPost, "/api/player/appeals/"+strconv.Itoa(int(appeal.ID))+"/claim", nil)
	otherW := httptest.NewRecorder()
	router.ServeHTTP(otherW, otherReq)
	if otherW.Code != http.StatusBadRequest {
		t.Fatalf("expected other player claim rejection, got %d: %s", otherW.Code, otherW.Body.String())
	}

	var savedAppeal database.CheatAppeal
	if err := db.First(&savedAppeal, appeal.ID).Error; err != nil {
		t.Fatalf("load saved appeal: %v", err)
	}
	if savedAppeal.CompensationStatus != "pending" {
		t.Fatalf("rejected claim should leave pending status, got %q", savedAppeal.CompensationStatus)
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

func TestGetPlayerAppealsDoesNotRepairApprovedAppealBanState(t *testing.T) {
	router, db := setupPlayerAppealHandlerTest(t)
	database.DB = db
	if err := db.AutoMigrate(&database.User{}, &database.FuelCompensationRecord{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}
	bannedUntil := time.Now().Add(2 * time.Hour)
	if err := db.Create(&database.User{
		UID:         1001,
		Username:    "approved-player",
		BannedUntil: &bannedUntil,
		BanReason:   "legacy ban",
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&database.CheatSanction{
		RoomID:         "room-approved",
		PlayerUID:      1001,
		RiskScoreID:    1,
		SanctionType:   "ban",
		RiskScore:      90,
		Reason:         "legacy active ban",
		EffectiveUntil: &bannedUntil,
		Status:         "active",
	}).Error; err != nil {
		t.Fatalf("create sanction: %v", err)
	}
	if err := db.Create(&database.CheatAppeal{
		RoomID:    "room-approved",
		PlayerUID: 1001,
		Reason:    "accepted",
		Status:    "approved",
	}).Error; err != nil {
		t.Fatalf("create approved appeal: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/player/appeals", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected appeal history success, got %d: %s", w.Code, w.Body.String())
	}

	var user database.User
	if err := db.First(&user, 1001).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user.BannedUntil == nil || !user.BannedUntil.After(time.Now()) || user.BanReason != "legacy ban" {
		t.Fatalf("expected appeal history read to preserve account ban, got until=%v reason=%q", user.BannedUntil, user.BanReason)
	}
	var activeBans int64
	if err := db.Model(&database.CheatSanction{}).Where("player_uid = ? AND sanction_type = ? AND status = ?", 1001, "ban", "active").Count(&activeBans).Error; err != nil {
		t.Fatalf("count active bans: %v", err)
	}
	if activeBans != 1 {
		t.Fatalf("expected appeal history read to preserve active bans, got %d", activeBans)
	}
	assertNoUnbanAudit(t, db, 1001)
}

func TestAppealWorkflowSubmitAdminListRejectPlayerHistoryAndAudit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := database.MigrateCheatTables(db); err != nil {
		t.Fatalf("migrate cheat tables: %v", err)
	}
	seedBannedAppealContext(t, db, 1001, "room-workflow", 77)

	repo := repository.NewCheatRepository(db)
	userRepo := repository.NewUserRepository(db)
	configMgr, err := anticheat.NewConfigManager("")
	if err != nil {
		t.Fatalf("config manager: %v", err)
	}
	system := &anticheat.System{
		AppealManager: anticheat.NewAppealManager(repo, userRepo),
		Repository:    repo,
		Config:        configMgr,
		StartedAt:     time.Now(),
	}
	handler := NewAnticheatHandler(system)

	router := gin.New()
	router.POST("/api/game/:roomId/appeal", func(c *gin.Context) {
		c.Set("uid", 1001)
		handler.SubmitAppeal(c)
	})
	router.GET("/api/admin/anticheat/appeals", handler.GetAppealsList)
	router.POST("/api/admin/anticheat/appeals/:id/reject", handler.RejectAppeal)
	router.GET("/api/player/appeals", func(c *gin.Context) {
		c.Set("uid", 1001)
		handler.GetPlayerAppeals(c)
	})

	submitBody, _ := json.Marshal(map[string]any{
		"risk_score_id": 77,
		"reason":        "false positive",
		"evidence":      "stable replay",
	})
	submitReq := httptest.NewRequest(http.MethodPost, "/api/game/room-workflow/appeal", bytes.NewReader(submitBody))
	submitReq.Header.Set("Content-Type", "application/json")
	submitW := httptest.NewRecorder()
	router.ServeHTTP(submitW, submitReq)
	if submitW.Code != http.StatusOK {
		t.Fatalf("submit appeal failed: %d %s", submitW.Code, submitW.Body.String())
	}

	var submitPayload struct {
		Appeal database.CheatAppeal `json:"appeal"`
	}
	if err := json.Unmarshal(submitW.Body.Bytes(), &submitPayload); err != nil {
		t.Fatalf("decode submit: %v", err)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/admin/anticheat/appeals", nil)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("admin list failed: %d %s", listW.Code, listW.Body.String())
	}
	var listPayload struct {
		Appeals []struct {
			PlayerID  uint `json:"player_id"`
			PlayerUID uint `json:"player_uid"`
		} `json:"appeals"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(listW.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if listPayload.Total != 1 || len(listPayload.Appeals) != 1 {
		t.Fatalf("expected appeal in admin list, got total=%d len=%d", listPayload.Total, len(listPayload.Appeals))
	}
	if listPayload.Appeals[0].PlayerID != 1001 || listPayload.Appeals[0].PlayerUID != 1001 {
		t.Fatalf("expected admin list player id aliases, got %+v", listPayload.Appeals[0])
	}

	rejectBody, _ := json.Marshal(map[string]any{"note": "not enough evidence"})
	rejectReq := httptest.NewRequest(http.MethodPost, "/api/admin/anticheat/appeals/"+strconv.Itoa(int(submitPayload.Appeal.ID))+"/reject", bytes.NewReader(rejectBody))
	rejectReq.Header.Set("Content-Type", "application/json")
	rejectW := httptest.NewRecorder()
	router.ServeHTTP(rejectW, rejectReq)
	if rejectW.Code != http.StatusOK {
		t.Fatalf("reject appeal failed: %d %s", rejectW.Code, rejectW.Body.String())
	}

	historyReq := httptest.NewRequest(http.MethodGet, "/api/player/appeals", nil)
	historyW := httptest.NewRecorder()
	router.ServeHTTP(historyW, historyReq)
	if historyW.Code != http.StatusOK {
		t.Fatalf("player history failed: %d %s", historyW.Code, historyW.Body.String())
	}
	var historyPayload struct {
		Appeals []database.CheatAppeal `json:"appeals"`
	}
	if err := json.Unmarshal(historyW.Body.Bytes(), &historyPayload); err != nil {
		t.Fatalf("decode history: %v", err)
	}
	if len(historyPayload.Appeals) != 1 || historyPayload.Appeals[0].Status != "rejected" || historyPayload.Appeals[0].ReviewRemark != "not enough evidence" {
		t.Fatalf("unexpected player appeal history: %+v", historyPayload.Appeals)
	}

	var auditCount int64
	if err := db.Model(&database.CheatAuditLog{}).Where("appeal_id = ?", submitPayload.Appeal.ID).Count(&auditCount).Error; err != nil {
		t.Fatalf("count audit logs: %v", err)
	}
	if auditCount < 2 {
		t.Fatalf("expected submit and reject audit logs, got %d", auditCount)
	}
	var rejectAudit database.CheatAuditLog
	if err := db.Where("appeal_id = ? AND new_status = ?", submitPayload.Appeal.ID, "rejected").First(&rejectAudit).Error; err != nil {
		t.Fatalf("load reject audit: %v", err)
	}
	if rejectAudit.ReplayID != "77" || rejectAudit.GameHistoryID != 77 || len(rejectAudit.PrimaryEvidence) == 0 || len(rejectAudit.RelatedEvidence) == 0 {
		t.Fatalf("expected reject audit to retain replay evidence, got replay=%q history=%d primary=%s related=%s", rejectAudit.ReplayID, rejectAudit.GameHistoryID, string(rejectAudit.PrimaryEvidence), string(rejectAudit.RelatedEvidence))
	}
}
