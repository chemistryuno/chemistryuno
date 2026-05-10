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

func setupFeedbackReplayEvidenceTest(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&database.User{}, &database.Feedback{}, &database.GameHistory{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	database.DB = db
	repository.FeedbackRepo = repository.NewFeedbackRepository()
	repository.GameRepo = repository.NewGameRepository()

	router := gin.New()
	router.POST("/api/feedback", func(c *gin.Context) {
		c.Set("uid", 1001)
		CreateFeedback(c)
	})
	return router, db
}

func TestCreateFeedbackStoresValidatedReplayEvidence(t *testing.T) {
	router, db := setupFeedbackReplayEvidenceTest(t)
	now := time.Now().UTC()
	replayLog := map[string]any{
		"events": []map[string]any{
			{
				"event_index":    1,
				"event_id":       "evt-1",
				"event":          "play_card",
				"uid":            1002,
				"unix_ms":        now.UnixMilli(),
				"action_summary": "played H2O",
			},
		},
	}
	replayBytes, _ := json.Marshal(replayLog)
	history := database.GameHistory{
		RoomID:     "room-feedback",
		ReplayLog:  string(replayBytes),
		Players:    database.JSON(`[1001,1002]`),
		StartedAt:  now.Add(-time.Hour),
		FinishedAt: now,
	}
	if err := db.Create(&history).Error; err != nil {
		t.Fatalf("save history: %v", err)
	}

	body := map[string]any{
		"type":         "report",
		"content":      "reported suspicious replay point",
		"room_id":      "room-feedback",
		"reported_uid": 1002,
		"replay_anchor": map[string]any{
			"game_history_id":    history.ID,
			"event_index":        1,
			"event_id":           "evt-1",
			"event_timestamp_ms": now.UnixMilli(),
			"player_uid":         1002,
		},
	}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/feedback", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected feedback created, got %d: %s", w.Code, w.Body.String())
	}
	var feedback database.Feedback
	if err := db.First(&feedback).Error; err != nil {
		t.Fatalf("load feedback: %v", err)
	}
	if feedback.GameHistoryID != history.ID || feedback.ReplayID != "1" || feedback.RoomID != "room-feedback" || feedback.ReportedUID != 1002 {
		t.Fatalf("unexpected persisted feedback: %+v", feedback)
	}
	var anchor database.ReplayEvidenceAnchor
	if err := json.Unmarshal(feedback.PrimaryEvidence, &anchor); err != nil {
		t.Fatalf("decode evidence anchor: %v", err)
	}
	if anchor.EventID != "evt-1" || anchor.EventIndex != 1 || anchor.EvidencePrecision != "operation" {
		t.Fatalf("unexpected evidence anchor: %+v", anchor)
	}
}

func TestCreateFeedbackRejectsInvalidReplayEvidence(t *testing.T) {
	router, db := setupFeedbackReplayEvidenceTest(t)
	history := database.GameHistory{
		RoomID:     "room-feedback",
		ReplayLog:  `{"events":[{"event_index":1,"event_id":"evt-1"}]}`,
		Players:    database.JSON(`[1001,1002]`),
		StartedAt:  time.Now().Add(-time.Hour),
		FinishedAt: time.Now(),
	}
	if err := db.Create(&history).Error; err != nil {
		t.Fatalf("save history: %v", err)
	}

	body := map[string]any{
		"type":         "report",
		"content":      "invalid anchor",
		"room_id":      "room-feedback",
		"reported_uid": 1002,
		"replay_anchor": map[string]any{
			"game_history_id": history.ID,
			"event_id":        "missing-event",
		},
	}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/feedback", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d: %s", w.Code, w.Body.String())
	}
	var count int64
	if err := db.Model(&database.Feedback{}).Count(&count).Error; err != nil {
		t.Fatalf("count feedbacks: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected invalid anchor not to create feedback, got %d", count)
	}
}
