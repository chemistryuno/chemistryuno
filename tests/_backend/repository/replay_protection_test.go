package repository

import (
	"errors"
	"testing"
	"time"

	"chemistryuno/backend/database"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupReplayProtectionTest(t *testing.T) (*gorm.DB, *GameRepository, *CheatRepository) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&database.GameHistory{}); err != nil {
		t.Fatalf("migrate game history: %v", err)
	}
	if err := database.MigrateCheatTables(db); err != nil {
		t.Fatalf("migrate cheat tables: %v", err)
	}
	return db, &GameRepository{db: db}, NewCheatRepository(db)
}

func TestProtectedReplayCleanupAndManualClearAreRejected(t *testing.T) {
	db, gameRepo, cheatRepo := setupReplayProtectionTest(t)
	now := time.Now().UTC()
	history := database.GameHistory{
		RoomID:          "room-protected",
		ReplayLog:       `{"events":[{"event_index":1,"event_id":"evt-1"}]}`,
		ReplayExpiresAt: ptrTime(now.Add(-time.Hour)),
		Players:         database.JSON(`[1001,1002]`),
		StartedAt:       now.Add(-2 * time.Hour),
		FinishedAt:      now.Add(-time.Hour),
	}
	if err := db.Create(&history).Error; err != nil {
		t.Fatalf("save history: %v", err)
	}
	if err := cheatRepo.SaveRiskScore(&database.CheatRiskScore{
		RoomID:        history.RoomID,
		PlayerUID:     1001,
		ReplayID:      "custom-replay",
		GameHistoryID: history.ID,
		RiskScore:     90,
	}); err != nil {
		t.Fatalf("save risk score: %v", err)
	}

	count, err := gameRepo.CleanupExpiredReplays(now)
	if err != nil {
		t.Fatalf("cleanup expired replays: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected protected replay to be skipped, cleared %d", count)
	}

	var refreshed database.GameHistory
	if err := db.First(&refreshed, history.ID).Error; err != nil {
		t.Fatalf("reload history: %v", err)
	}
	if refreshed.ReplayLog == "" {
		t.Fatal("expected protected replay log to remain")
	}

	if err := gameRepo.ClearReplayByID(history.ID); !errors.Is(err, ErrReplayProtectedByAnticheat) {
		t.Fatalf("expected protected replay error, got %v", err)
	}

	var skipped, rejected int64
	if err := db.Model(&database.CheatAuditLog{}).Where("event_type = ?", "replay_cleanup_skipped").Count(&skipped).Error; err != nil {
		t.Fatalf("count skipped audit: %v", err)
	}
	if err := db.Model(&database.CheatAuditLog{}).Where("event_type = ?", "replay_clear_rejected").Count(&rejected).Error; err != nil {
		t.Fatalf("count rejected audit: %v", err)
	}
	if skipped != 1 || rejected != 1 {
		t.Fatalf("expected cleanup skip and clear rejection audits, got skipped=%d rejected=%d", skipped, rejected)
	}
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

