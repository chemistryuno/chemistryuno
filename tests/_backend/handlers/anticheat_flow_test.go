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

type anticheatFlowHarness struct {
	router  *gin.Engine
	db      *gorm.DB
	system  *anticheat.System
	repo    *repository.CheatRepository
	fixture anticheatFlowFixture
}

type anticheatFlowFixture struct {
	playerUID    uint
	reporterUID  uint
	adminUID     uint
	unbannedUID  uint
	roomID       string
	followupRoom string
	baseTime     time.Time
	historyID    uint
	primary      anticheat.ReplayEvidenceAnchor
	reportAnchor anticheat.ReplayEvidenceAnchor
}

func setupAnticheatFlowHarness(t *testing.T) *anticheatFlowHarness {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&database.User{}, &database.FuelCompensationRecord{}, &database.GameHistory{}, &database.Feedback{}); err != nil {
		t.Fatalf("migrate base tables: %v", err)
	}
	if err := database.MigrateCheatTables(db); err != nil {
		t.Fatalf("migrate cheat tables: %v", err)
	}
	database.DB = db

	fixture := anticheatFlowFixture{
		playerUID:    1001,
		reporterUID:  2002,
		adminUID:     9001,
		unbannedUID:  3003,
		roomID:       "flow-room-1",
		followupRoom: "flow-room-2",
		baseTime:     time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC),
	}
	users := []database.User{
		{UID: fixture.playerUID, Username: "flow-player", Role: "user", Points: 1000},
		{UID: fixture.reporterUID, Username: "flow-reporter", Role: "user", Points: 1000},
		{UID: fixture.adminUID, Username: "flow-admin", Role: "admin", IsAdmin: true},
		{UID: fixture.unbannedUID, Username: "flow-clean", Role: "user", Points: 1000},
	}
	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			t.Fatalf("create user %d: %v", users[i].UID, err)
		}
	}

	history := seedFlowGameHistory(t, db, fixture.roomID, fixture.playerUID, fixture.baseTime)
	fixture.historyID = history.ID
	fixture.primary = anticheat.NormalizeReplayEvidenceAnchor(anticheat.ReplayEvidenceAnchor{
		RoomID:           fixture.roomID,
		GameHistoryID:    history.ID,
		ReplayID:         strconv.Itoa(int(history.ID)),
		EventIndex:       42,
		EventID:          "evt-cheat-42",
		EventType:        "play_card",
		PlayerUID:        fixture.playerUID,
		EventTimestampMs: fixture.baseTime.Add(840 * time.Millisecond).UnixMilli(),
		TurnNumber:       6,
		ActionSummary:    "played impossible reaction within 20ms",
	})
	fixture.reportAnchor = anticheat.NormalizeReplayEvidenceAnchor(anticheat.ReplayEvidenceAnchor{
		RoomID:           fixture.roomID,
		GameHistoryID:    history.ID,
		ReplayID:         strconv.Itoa(int(history.ID)),
		EventIndex:       43,
		EventID:          "evt-report-43",
		EventType:        "player_report",
		PlayerUID:        fixture.playerUID,
		EventTimestampMs: fixture.baseTime.Add(900 * time.Millisecond).UnixMilli(),
		TurnNumber:       6,
		ActionSummary:    "reporter flagged the same replay point",
	})
	seedFlowReport(t, db, fixture)

	config := deterministicAnticheatConfig()
	configMgr, err := anticheat.NewConfigManager("")
	if err != nil {
		t.Fatalf("config manager: %v", err)
	}
	if err := configMgr.ReplaceConfig(config); err != nil {
		t.Fatalf("replace config: %v", err)
	}
	engine := anticheat.NewRiskScoringEngine(config)
	if err := anticheat.NewBuiltInStrategies().RegisterAll(engine); err != nil {
		t.Fatalf("register strategies: %v", err)
	}
	repo := repository.NewCheatRepository(db)
	userRepo := repository.NewUserRepository(db)
	system := &anticheat.System{
		Engine:        engine,
		Config:        configMgr,
		Decider:       anticheat.NewSanctionDecider(config, repo),
		AppealManager: anticheat.NewAppealManager(repo, userRepo),
		AuditLogger:   anticheat.NewAuditLogger(repo),
		Repository:    repo,
		StartedAt:     fixture.baseTime,
	}
	handler := NewAnticheatHandler(system)

	router := gin.New()
	router.GET("/api/admin/anticheat/detection/:id", handler.GetDetectionDetail)
	router.POST("/api/admin/anticheat/detection/:id/review", func(c *gin.Context) {
		c.Set("uid", fixture.adminUID)
		handler.ReviewDetection(c)
	})
	router.POST("/api/admin/anticheat/detection/:id/punishment", func(c *gin.Context) {
		c.Set("uid", fixture.adminUID)
		handler.ChangeDetectionPunishment(c)
	})
	router.GET("/api/admin/anticheat/audit-log", handler.GetAuditLog)
	router.GET("/api/player/appeals/entry", func(c *gin.Context) {
		c.Set("uid", fixture.playerUID)
		handler.GetAppealEntryStatus(c)
	})
	router.GET("/api/player/appeals/entry/unbanned", func(c *gin.Context) {
		c.Set("uid", fixture.unbannedUID)
		handler.GetAppealEntryStatus(c)
	})
	router.POST("/api/game/:roomId/appeal", func(c *gin.Context) {
		c.Set("uid", fixture.playerUID)
		handler.SubmitAppeal(c)
	})
	router.POST("/api/game/:roomId/appeal/unbanned", func(c *gin.Context) {
		c.Set("uid", fixture.unbannedUID)
		handler.SubmitAppeal(c)
	})

	return &anticheatFlowHarness{
		router:  router,
		db:      db,
		system:  system,
		repo:    repo,
		fixture: fixture,
	}
}

func deterministicAnticheatConfig() *anticheat.RiskScoringConfig {
	config := anticheat.NewDefaultConfig()
	config.SanctionThresholds = anticheat.SanctionThresholds{
		ObserveMin: 20,
		ObserveMax: 39.99,
		WarningMin: 40,
		WarningMax: 59.99,
		MuteMin:    60,
		MuteMax:    79.99,
		BanMin:     80,
		BanMax:     100,
	}
	return config
}

func seedFlowGameHistory(t *testing.T, db *gorm.DB, roomID string, playerUID uint, baseTime time.Time) database.GameHistory {
	t.Helper()
	players, _ := json.Marshal([]uint{playerUID, 2002})
	cheatUIDs, _ := json.Marshal([]uint{playerUID})
	replayLog, _ := json.Marshal([]map[string]any{
		{
			"event_index":    42,
			"event_id":       "evt-cheat-42",
			"event":          "play_card",
			"uid":            playerUID,
			"unix_ms":        baseTime.Add(840 * time.Millisecond).UnixMilli(),
			"action_summary": "played impossible reaction within 20ms",
		},
		{
			"event_index":    43,
			"event_id":       "evt-report-43",
			"event":          "player_report",
			"uid":            playerUID,
			"unix_ms":        baseTime.Add(900 * time.Millisecond).UnixMilli(),
			"action_summary": "reporter flagged the same replay point",
		},
	})
	history := database.GameHistory{
		RoomID:              roomID,
		WinnerUID:           &playerUID,
		ReplayLog:           string(replayLog),
		ReplayPermanent:     true,
		CheatDetected:       true,
		CheatUIDs:           cheatUIDs,
		Players:             players,
		OriginalPlayerCount: 2,
		StartedAt:           baseTime.Add(-5 * time.Minute),
		FinishedAt:          baseTime.Add(5 * time.Minute),
	}
	if err := db.Create(&history).Error; err != nil {
		t.Fatalf("create game history: %v", err)
	}
	return history
}

func seedFlowReport(t *testing.T, db *gorm.DB, fixture anticheatFlowFixture) {
	t.Helper()
	report := database.Feedback{
		UserUID:         fixture.reporterUID,
		Type:            "cheat",
		Content:         "suspicious fast reactions",
		RoomID:          fixture.roomID,
		ReportedUID:     fixture.playerUID,
		ReplayID:        strconv.Itoa(int(fixture.historyID)),
		GameHistoryID:   fixture.historyID,
		PrimaryEvidence: anticheat.MarshalReplayEvidenceAnchor(fixture.reportAnchor),
		Status:          "confirmed",
		CreatedAt:       fixture.baseTime.Add(time.Minute),
	}
	if err := db.Create(&report).Error; err != nil {
		t.Fatalf("create report fixture: %v", err)
	}
}

func highRiskCheatingContext(fixture anticheatFlowFixture, includeReports bool) *anticheat.DetectionContext {
	operationTimes := make([]time.Time, 80)
	for i := range operationTimes {
		operationTimes[i] = fixture.baseTime.Add(time.Duration(i) * 20 * time.Millisecond)
	}
	context := &anticheat.DetectionContext{
		PlayerUID:       int(fixture.playerUID),
		RoomID:          fixture.roomID,
		ReplayID:        strconv.Itoa(int(fixture.historyID)),
		GameHistoryID:   fixture.historyID,
		OperationIndex:  fixture.primary.EventIndex,
		PrimaryEvidence: fixture.primary,
		RelatedEvidence: []anticheat.ReplayEvidenceAnchor{fixture.primary},
		OperationCount:  len(operationTimes),
		TimestampOffset: 10 * time.Second,
		// 新指标体系：全维度极端异常，确保达到封号阈值。
		TotalDecisions:          25,
		OptimalDecisions:        25, // decision_optimality 满
		ComplexDecisionCount:    12,
		SuperhumanDecisionCount: 12, // think_time 全超人
		HasRecentPerf:           true,
		RecentGames:             20,
		RecentWinRate:           1.0, // recent_performance 满
		OpponentStrength:        1.5,
		HasMultiAccount:         true,
		MultiAccountScore:       100, // multi_account 满
		WinCount:                20,
		TotalGames:              20,
		AccountAgeDays:          0,
		OperationTimes:          operationTimes,
	}
	if includeReports {
		context.ReportCount = 5
		context.ReportSummary = "5 validated player reports tied to replay evidence"
		context.ReportEvidence = []anticheat.ReplayEvidenceAnchor{fixture.reportAnchor}
	}
	return context
}

func (h *anticheatFlowHarness) processHighRiskDetection(t *testing.T, includeReports bool) (*anticheat.RiskScoringResult, *anticheat.Decision, database.CheatRiskScore, database.CheatSanction) {
	t.Helper()
	result, decision, err := h.system.ProcessGameEnd(h.fixture.roomID, h.fixture.playerUID, highRiskCheatingContext(h.fixture, includeReports))
	if err != nil {
		t.Fatalf("process game end: %v", err)
	}
	if result == nil || decision == nil {
		t.Fatalf("expected process result and decision, got result=%v decision=%v", result, decision)
	}
	var score database.CheatRiskScore
	if err := h.db.Where("player_uid = ? AND room_id = ?", h.fixture.playerUID, h.fixture.roomID).First(&score).Error; err != nil {
		t.Fatalf("load risk score: %v", err)
	}
	var sanction database.CheatSanction
	if err := h.db.Where("risk_score_id = ? AND player_uid = ?", score.ID, h.fixture.playerUID).First(&sanction).Error; err != nil {
		t.Fatalf("load sanction: %v", err)
	}
	return result, decision, score, sanction
}

func performFlowRequest(router *gin.Engine, method string, path string, body any) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		payload, _ := json.Marshal(body)
		reader = bytes.NewReader(payload)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func assertFlowAuditEvidence(t *testing.T, audit database.CheatAuditLog, expectedEvent string, expectedAdmin uint) {
	t.Helper()
	if audit.EventType != expectedEvent {
		t.Fatalf("expected audit event %q, got %q", expectedEvent, audit.EventType)
	}
	if expectedAdmin > 0 && (audit.OperatorUID == nil || *audit.OperatorUID != expectedAdmin) {
		t.Fatalf("expected audit operator %d, got %#v", expectedAdmin, audit.OperatorUID)
	}
	if audit.ReplayID == "" || audit.GameHistoryID == 0 || len(audit.PrimaryEvidence) == 0 {
		t.Fatalf("expected replay evidence on %s audit, got replay=%q history=%d primary=%s", expectedEvent, audit.ReplayID, audit.GameHistoryID, string(audit.PrimaryEvidence))
	}
}

func TestAnticheatFlowProcessGameEndCreatesEvidenceRichRiskRecord(t *testing.T) {
	h := setupAnticheatFlowHarness(t)

	result, decision, score, sanction := h.processHighRiskDetection(t, false)

	if result.RiskScore < deterministicAnticheatConfig().SanctionThresholds.BanMin || score.RiskScore < deterministicAnticheatConfig().SanctionThresholds.BanMin {
		t.Fatalf("expected risk above ban threshold, got result=%.1f saved=%.1f", result.RiskScore, score.RiskScore)
	}
	if result.SuggestedAction != "ban" || result.SanctionType != "ban" || decision.SanctionType != "ban" {
		t.Fatalf("expected ban suggestion/decision, got result=%+v decision=%+v", result, decision)
	}
	if score.SuggestedAction != "ban" || score.PunishmentDecision != "ban" || score.ReviewStatus != "pending" {
		t.Fatalf("unexpected saved risk state: suggested=%q punishment=%q review=%q", score.SuggestedAction, score.PunishmentDecision, score.ReviewStatus)
	}
	// 指标重设计：维度明细统一存 IndicatorDetails（JSON），旧的 *Dim 列已停写。
	if len(score.IndicatorDetails) == 0 {
		t.Fatalf("expected populated indicator details: %+v", score)
	}
	anchor, ok := anticheat.UnmarshalReplayEvidenceAnchor(score.PrimaryEvidence)
	if !ok {
		t.Fatalf("expected primary evidence anchor, got %s", string(score.PrimaryEvidence))
	}
	if anchor.GameHistoryID != h.fixture.historyID || anchor.EventID != "evt-cheat-42" || anchor.EventIndex != 42 || anchor.PlayerUID != h.fixture.playerUID || anchor.NavigationURL == "" {
		t.Fatalf("unexpected primary evidence anchor: %+v", anchor)
	}
	var indicators []anticheat.RiskIndicatorDetail
	if err := json.Unmarshal(score.IndicatorDetails, &indicators); err != nil {
		t.Fatalf("decode indicator details: %v", err)
	}
	if len(indicators) < 3 {
		t.Fatalf("expected multiple indicator details, got %+v", indicators)
	}
	if sanction.SanctionType != "ban" || sanction.Status != "active" || sanction.ReplayID == "" || len(sanction.PrimaryEvidence) == 0 {
		t.Fatalf("expected active ban with replay evidence, got %+v", sanction)
	}
	var detectionAudit database.CheatAuditLog
	if err := h.db.Where("risk_score_id = ? AND event_type = ?", score.ID, "detection").First(&detectionAudit).Error; err != nil {
		t.Fatalf("load detection audit: %v", err)
	}
	assertFlowAuditEvidence(t, detectionAudit, "detection", 0)
}

func TestAnticheatFlowDetectionDetailExposesEvidenceAndReportContribution(t *testing.T) {
	h := setupAnticheatFlowHarness(t)
	_, _, score, _ := h.processHighRiskDetection(t, true)

	rec := performFlowRequest(h.router, http.MethodGet, "/api/admin/anticheat/detection/"+strconv.Itoa(int(score.ID)), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected detection detail success, got %d: %s", rec.Code, rec.Body.String())
	}
	var payload struct {
		RiskScore struct {
			ID                 uint                            `json:"id"`
			PlayerUID          uint                            `json:"player_uid"`
			RoomID             string                          `json:"room_id"`
			RiskScore          float64                         `json:"risk_score"`
			SuggestedAction    string                          `json:"suggested_action"`
			PunishmentDecision string                          `json:"punishment_decision"`
			ReplayNavigation   database.ReplayEvidenceAnchor   `json:"replay_navigation"`
			IndicatorDetails   []anticheat.RiskIndicatorDetail `json:"indicator_details"`
			ReportContribution anticheat.ReportContribution    `json:"report_contribution"`
		} `json:"risk_score"`
		Sanctions []database.CheatSanction `json:"sanctions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode detection detail: %v", err)
	}
	got := payload.RiskScore
	if got.ID != score.ID || got.PlayerUID != h.fixture.playerUID || got.RoomID != h.fixture.roomID {
		t.Fatalf("unexpected detection identity: %+v", got)
	}
	if got.RiskScore < deterministicAnticheatConfig().SanctionThresholds.BanMin || got.SuggestedAction != "ban" || got.PunishmentDecision != "ban" {
		t.Fatalf("unexpected detection decision fields: %+v", got)
	}
	if got.ReplayNavigation.GameHistoryID != h.fixture.historyID || got.ReplayNavigation.EventID != "evt-cheat-42" || got.ReplayNavigation.NavigationURL == "" {
		t.Fatalf("unexpected replay navigation: %+v", got.ReplayNavigation)
	}
	if got.ReportContribution.DeduplicatedCount != 5 || got.ReportContribution.SourceSummary == "" || len(got.ReportContribution.EvidenceAnchors) != 1 {
		t.Fatalf("unexpected report contribution: %+v", got.ReportContribution)
	}
	hasReportIndicator := false
	for _, indicator := range got.IndicatorDetails {
		if indicator.Name == "player_reports" {
			hasReportIndicator = true
			if indicator.RawValue != 5 || len(indicator.EvidenceAnchors) == 0 {
				t.Fatalf("unexpected player report indicator: %+v", indicator)
			}
		}
	}
	if !hasReportIndicator {
		t.Fatalf("expected player_reports indicator in %+v", got.IndicatorDetails)
	}
	if len(payload.Sanctions) != 1 || payload.Sanctions[0].SanctionType != "ban" {
		t.Fatalf("expected active ban sanction in detail, got %+v", payload.Sanctions)
	}
}

func TestAnticheatFlowAdminReviewPunishmentChangeAndAuditTrail(t *testing.T) {
	h := setupAnticheatFlowHarness(t)
	_, _, score, sanction := h.processHighRiskDetection(t, true)

	reviewRec := performFlowRequest(h.router, http.MethodPost, "/api/admin/anticheat/detection/"+strconv.Itoa(int(score.ID))+"/review", map[string]any{
		"decision": "confirm",
		"remark":   "confirmed replay evidence",
	})
	if reviewRec.Code != http.StatusOK {
		t.Fatalf("expected review success, got %d: %s", reviewRec.Code, reviewRec.Body.String())
	}
	var reviewed database.CheatRiskScore
	if err := h.db.First(&reviewed, score.ID).Error; err != nil {
		t.Fatalf("load reviewed score: %v", err)
	}
	if reviewed.ReviewStatus != "processed" || reviewed.PunishmentDecision != "ban" {
		t.Fatalf("expected processed ban after review, got review=%q punishment=%q", reviewed.ReviewStatus, reviewed.PunishmentDecision)
	}
	var reviewAudit database.CheatAuditLog
	if err := h.db.Where("risk_score_id = ? AND event_type = ?", score.ID, "review").First(&reviewAudit).Error; err != nil {
		t.Fatalf("load review audit: %v", err)
	}
	assertFlowAuditEvidence(t, reviewAudit, "review", h.fixture.adminUID)
	if reviewAudit.NewStatus != "processed" || reviewAudit.NewDecision != "ban" || len(reviewAudit.ReportContribution) == 0 {
		t.Fatalf("unexpected review audit fields: %+v", reviewAudit)
	}

	cancelRec := performFlowRequest(h.router, http.MethodPost, "/api/admin/anticheat/detection/"+strconv.Itoa(int(score.ID))+"/punishment", map[string]any{
		"punishment_decision": "none",
		"sanction_id":         sanction.ID,
		"reason":              "attempted cancellation",
	})
	if cancelRec.Code != http.StatusBadRequest {
		t.Fatalf("expected cancellation rejection, got %d: %s", cancelRec.Code, cancelRec.Body.String())
	}
	if err := h.db.First(&reviewed, score.ID).Error; err != nil {
		t.Fatalf("reload score after cancellation attempt: %v", err)
	}
	if reviewed.PunishmentDecision != "ban" {
		t.Fatalf("cancellation should not change punishment, got %q", reviewed.PunishmentDecision)
	}
	var rejectedAudit database.CheatAuditLog
	if err := h.db.Where("risk_score_id = ? AND event_type = ?", score.ID, "punishment_change_rejected").First(&rejectedAudit).Error; err != nil {
		t.Fatalf("load rejected punishment audit: %v", err)
	}
	assertFlowAuditEvidence(t, rejectedAudit, "punishment_change_rejected", h.fixture.adminUID)
	if rejectedAudit.OldDecision != "ban" || rejectedAudit.NewDecision != "none" {
		t.Fatalf("unexpected rejected audit decisions: old=%q new=%q", rejectedAudit.OldDecision, rejectedAudit.NewDecision)
	}

	muteUntil := time.Now().Add(30 * time.Minute).UTC().Format(time.RFC3339)
	changeRec := performFlowRequest(h.router, http.MethodPost, "/api/admin/anticheat/detection/"+strconv.Itoa(int(score.ID))+"/punishment", map[string]any{
		"punishment_decision": "mute",
		"sanction_id":         sanction.ID,
		"reason":              "reduce severity after review",
		"duration":            30,
		"effective_until":     muteUntil,
	})
	if changeRec.Code != http.StatusOK {
		t.Fatalf("expected punishment change success, got %d: %s", changeRec.Code, changeRec.Body.String())
	}
	if err := h.db.First(&reviewed, score.ID).Error; err != nil {
		t.Fatalf("reload changed score: %v", err)
	}
	if reviewed.PunishmentDecision != "mute" || reviewed.ReviewStatus != "processed" {
		t.Fatalf("expected processed mute decision, got review=%q punishment=%q", reviewed.ReviewStatus, reviewed.PunishmentDecision)
	}
	var changedSanction database.CheatSanction
	if err := h.db.First(&changedSanction, sanction.ID).Error; err != nil {
		t.Fatalf("load changed sanction: %v", err)
	}
	if changedSanction.SanctionType != "mute" || changedSanction.Status != "active" || changedSanction.Duration == nil || *changedSanction.Duration != 30 {
		t.Fatalf("unexpected changed sanction: %+v", changedSanction)
	}
	var changeAudit database.CheatAuditLog
	if err := h.db.Where("risk_score_id = ? AND event_type = ?", score.ID, "punishment_change").First(&changeAudit).Error; err != nil {
		t.Fatalf("load punishment change audit: %v", err)
	}
	assertFlowAuditEvidence(t, changeAudit, "punishment_change", h.fixture.adminUID)
	if changeAudit.OldDecision != "ban" || changeAudit.NewDecision != "mute" || changeAudit.SanctionID == nil || *changeAudit.SanctionID != sanction.ID {
		t.Fatalf("unexpected punishment change audit: %+v", changeAudit)
	}

	auditRec := performFlowRequest(h.router, http.MethodGet, "/api/admin/anticheat/audit-log?player_uid=1001&limit=20", nil)
	if auditRec.Code != http.StatusOK {
		t.Fatalf("expected audit log API success, got %d: %s", auditRec.Code, auditRec.Body.String())
	}
	var auditPayload struct {
		Logs []struct {
			EventType       string `json:"event_type"`
			OldDecision     string `json:"old_decision"`
			NewDecision     string `json:"new_decision"`
			PrimaryEvidence any    `json:"primary_evidence"`
		} `json:"logs"`
	}
	if err := json.Unmarshal(auditRec.Body.Bytes(), &auditPayload); err != nil {
		t.Fatalf("decode audit log API: %v", err)
	}
	seenReview, seenRejected, seenChange := false, false, false
	for _, log := range auditPayload.Logs {
		if log.PrimaryEvidence == nil {
			t.Fatalf("expected audit API log to include replay evidence: %+v", log)
		}
		switch log.EventType {
		case "review":
			seenReview = true
		case "punishment_change_rejected":
			seenRejected = true
		case "punishment_change":
			seenChange = log.OldDecision == "ban" && log.NewDecision == "mute"
		}
	}
	if !seenReview || !seenRejected || !seenChange {
		t.Fatalf("expected review/rejected/change audit logs, got %+v", auditPayload.Logs)
	}
}

func TestAnticheatFlowPlayerBanAppealEntryAndLockedRoomSubmission(t *testing.T) {
	h := setupAnticheatFlowHarness(t)
	_, _, score, sanction := h.processHighRiskDetection(t, true)
	seedFlowGameHistory(t, h.db, h.fixture.followupRoom, h.fixture.playerUID, time.Now().Add(time.Minute))

	var banned database.User
	if err := h.db.First(&banned, h.fixture.playerUID).Error; err != nil {
		t.Fatalf("load banned user: %v", err)
	}
	if banned.BannedUntil == nil || !banned.BannedUntil.After(time.Now()) || banned.BanReason == "" {
		t.Fatalf("expected active account ban, got until=%v reason=%q", banned.BannedUntil, banned.BanReason)
	}

	entryRec := performFlowRequest(h.router, http.MethodGet, "/api/player/appeals/entry", nil)
	if entryRec.Code != http.StatusOK {
		t.Fatalf("expected appeal entry success, got %d: %s", entryRec.Code, entryRec.Body.String())
	}
	var entry struct {
		IsBanned          bool     `json:"is_banned"`
		CanSubmit         bool     `json:"can_submit"`
		LatestRiskScoreID uint     `json:"latest_risk_score_id"`
		RoomIDs           []string `json:"room_ids"`
	}
	if err := json.Unmarshal(entryRec.Body.Bytes(), &entry); err != nil {
		t.Fatalf("decode appeal entry: %v", err)
	}
	if !entry.IsBanned || !entry.CanSubmit || entry.LatestRiskScoreID != score.ID {
		t.Fatalf("unexpected appeal entry state: %+v", entry)
	}
	if len(entry.RoomIDs) != 2 || entry.RoomIDs[0] != h.fixture.roomID || entry.RoomIDs[1] != h.fixture.followupRoom {
		t.Fatalf("expected locked room list from anticheat context, got %#v", entry.RoomIDs)
	}

	submitRec := performFlowRequest(h.router, http.MethodPost, "/api/game/account/appeal", map[string]any{
		"risk_score_id": score.ID,
		"sanction_id":   sanction.ID,
		"reason":        "the replay was a false positive",
		"evidence":      "network jitter during the locked room sequence",
	})
	if submitRec.Code != http.StatusOK {
		t.Fatalf("expected appeal submission success, got %d: %s", submitRec.Code, submitRec.Body.String())
	}
	var savedAppeal database.CheatAppeal
	if err := h.db.Where("player_uid = ?", h.fixture.playerUID).First(&savedAppeal).Error; err != nil {
		t.Fatalf("load saved appeal: %v", err)
	}
	var lockedRooms []string
	if err := json.Unmarshal(savedAppeal.RoomIDs, &lockedRooms); err != nil {
		t.Fatalf("decode locked rooms: %v", err)
	}
	if savedAppeal.RoomID != h.fixture.roomID || savedAppeal.RiskScoreID != score.ID || savedAppeal.SanctionID != sanction.ID || len(lockedRooms) != 2 {
		t.Fatalf("unexpected saved appeal context: appeal=%+v rooms=%#v", savedAppeal, lockedRooms)
	}
	if savedAppeal.ReplayID == "" || savedAppeal.GameHistoryID == 0 || len(savedAppeal.PrimaryEvidence) == 0 {
		t.Fatalf("expected appeal to retain replay evidence, got %+v", savedAppeal)
	}
}

func TestAnticheatFlowUnbannedPlayerAppealRedirect(t *testing.T) {
	h := setupAnticheatFlowHarness(t)

	entryRec := performFlowRequest(h.router, http.MethodGet, "/api/player/appeals/entry/unbanned", nil)
	if entryRec.Code != http.StatusOK {
		t.Fatalf("expected unbanned appeal entry success, got %d: %s", entryRec.Code, entryRec.Body.String())
	}
	var entry struct {
		IsBanned  bool     `json:"is_banned"`
		CanSubmit bool     `json:"can_submit"`
		RoomIDs   []string `json:"room_ids"`
	}
	if err := json.Unmarshal(entryRec.Body.Bytes(), &entry); err != nil {
		t.Fatalf("decode unbanned entry: %v", err)
	}
	if entry.IsBanned || entry.CanSubmit || len(entry.RoomIDs) != 0 {
		t.Fatalf("unexpected unbanned entry: %+v", entry)
	}

	rec := performFlowRequest(h.router, http.MethodPost, "/api/game/clean-room/appeal/unbanned", map[string]any{
		"reason": "I should use feedback instead.",
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected unbanned appeal rejection, got %d: %s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Error        string `json:"error"`
		RedirectTo   string `json:"redirect_to"`
		OpenFeedback bool   `json:"open_feedback"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode unbanned rejection: %v", err)
	}
	if payload.RedirectTo != "/feedbacks" || !payload.OpenFeedback {
		t.Fatalf("expected feedback redirect metadata, got %+v", payload)
	}
	var appealCount int64
	if err := h.db.Model(&database.CheatAppeal{}).Where("player_uid = ?", h.fixture.unbannedUID).Count(&appealCount).Error; err != nil {
		t.Fatalf("count unbanned appeals: %v", err)
	}
	if appealCount != 0 {
		t.Fatalf("unbanned player should not create appeal, got %d", appealCount)
	}
}
