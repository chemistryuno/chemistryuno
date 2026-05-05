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
		Repository:    cheatRepo,
	}
	handler := NewAnticheatHandler(system)

	router := gin.New()
	router.GET("/api/player/appeals", func(c *gin.Context) {
		c.Set("uid", 1001)
		handler.GetPlayerAppeals(c)
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

func TestGetPlayerAppealsRepairsApprovedAppealBanState(t *testing.T) {
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
	if user.BannedUntil != nil || user.BanReason != "" {
		t.Fatalf("expected approved appeal to repair account ban, got until=%v reason=%q", user.BannedUntil, user.BanReason)
	}
	var activeBans int64
	if err := db.Model(&database.CheatSanction{}).Where("player_uid = ? AND sanction_type = ? AND status = ?", 1001, "ban", "active").Count(&activeBans).Error; err != nil {
		t.Fatalf("count active bans: %v", err)
	}
	if activeBans != 0 {
		t.Fatalf("expected approved appeal to revoke active bans, got %d", activeBans)
	}
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
}
