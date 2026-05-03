package repository

import (
	"fmt"
	"testing"
	"time"

	"chemistryuno/backend/database"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCheatRepositoryTest(t *testing.T) (*gorm.DB, *CheatRepository) {
	t.Helper()

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        fmt.Sprintf("file:cheat_repository_%s?mode=memory&cache=shared", t.Name()),
	}, &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite test db: %v", err)
	}
	if err := database.MigrateCheatTables(db); err != nil {
		t.Fatalf("migrate cheat tables: %v", err)
	}
	return db, NewCheatRepository(db)
}

func strPtr(value string) *string {
	return &value
}

func intPtr(value int) *int {
	return &value
}

func TestQueryAuditLogsFiltersByPlayerDateActionAndCompensationStatus(t *testing.T) {
	db, repo := setupCheatRepositoryTest(t)
	now := time.Now().UTC()

	fixtures := []database.CheatAuditLog{
		{
			EventType:          "review",
			RoomID:             "room_1",
			PlayerUID:          2001,
			SanctionType:       "unban",
			Remark:             "match",
			CompensationAmount: intPtr(100),
			CompensationStatus: strPtr("ok"),
			CreatedAt:          now.Add(-2 * time.Hour),
		},
		{
			EventType:          "review",
			RoomID:             "room_2",
			PlayerUID:          2001,
			SanctionType:       "unban",
			Remark:             "wrong status",
			CompensationAmount: intPtr(100),
			CompensationStatus: strPtr("failed"),
			CreatedAt:          now.Add(-90 * time.Minute),
		},
		{
			EventType:          "review",
			RoomID:             "room_3",
			PlayerUID:          2002,
			SanctionType:       "unban",
			Remark:             "wrong player",
			CompensationAmount: intPtr(100),
			CompensationStatus: strPtr("ok"),
			CreatedAt:          now.Add(-80 * time.Minute),
		},
		{
			EventType:          "review",
			RoomID:             "room_4",
			PlayerUID:          2001,
			SanctionType:       "ban",
			Remark:             "wrong action",
			CompensationAmount: intPtr(100),
			CompensationStatus: strPtr("ok"),
			CreatedAt:          now.Add(-70 * time.Minute),
		},
		{
			EventType:          "review",
			RoomID:             "room_5",
			PlayerUID:          2001,
			SanctionType:       "unban",
			Remark:             "outside date",
			CompensationAmount: intPtr(100),
			CompensationStatus: strPtr("ok"),
			CreatedAt:          now.Add(-48 * time.Hour),
		},
	}
	for i := range fixtures {
		if err := db.Create(&fixtures[i]).Error; err != nil {
			t.Fatalf("create audit fixture %d: %v", i, err)
		}
	}

	playerUID := uint(2001)
	start := now.Add(-3 * time.Hour)
	end := now.Add(-1 * time.Hour)
	logs, total, err := repo.QueryAuditLogs(AuditLogFilter{
		PlayerUID:            &playerUID,
		StartTime:            &start,
		EndTime:              &end,
		ActionType:           "unban",
		CompensationStatuses: []string{"ok"},
		Limit:                20,
	})
	if err != nil {
		t.Fatalf("query audit logs: %v", err)
	}
	if total != 1 || len(logs) != 1 {
		t.Fatalf("expected one filtered audit log, got total=%d len=%d", total, len(logs))
	}
	if logs[0].RoomID != "room_1" {
		t.Fatalf("expected room_1, got %s", logs[0].RoomID)
	}
}

func TestExportAuditLogsRespectsFiltersAndLimit(t *testing.T) {
	db, repo := setupCheatRepositoryTest(t)
	now := time.Now().UTC()
	for i := 0; i < 3; i++ {
		status := "ok"
		if i == 2 {
			status = "failed"
		}
		log := database.CheatAuditLog{
			EventType:          "review",
			RoomID:             "room_export",
			PlayerUID:          3001,
			SanctionType:       "unban",
			CompensationStatus: strPtr(status),
			CreatedAt:          now.Add(time.Duration(-i) * time.Minute),
		}
		if err := db.Create(&log).Error; err != nil {
			t.Fatalf("create audit fixture: %v", err)
		}
	}

	logs, err := repo.ExportAuditLogs(AuditLogFilter{
		CompensationStatuses: []string{"ok"},
	}, 1)
	if err != nil {
		t.Fatalf("export audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected export limit to return 1 log, got %d", len(logs))
	}
	if logs[0].CompensationStatus == nil || *logs[0].CompensationStatus != "ok" {
		t.Fatalf("expected exported log to have ok status, got %#v", logs[0].CompensationStatus)
	}
}

func TestQueryAuditLogsWithTenThousandRecordsUnderFiveHundredMilliseconds(t *testing.T) {
	db, repo := setupCheatRepositoryTest(t)
	now := time.Now().UTC()
	statusOK := "ok"
	statusFailed := "failed"
	amount := 100
	logs := make([]database.CheatAuditLog, 0, 10000)
	for i := 0; i < 10000; i++ {
		status := &statusOK
		if i%5 == 0 {
			status = &statusFailed
		}
		playerUID := uint(4000 + (i % 100))
		logs = append(logs, database.CheatAuditLog{
			EventType:          "review",
			RoomID:             "room_load",
			PlayerUID:          playerUID,
			SanctionType:       "unban",
			CompensationAmount: &amount,
			CompensationStatus: status,
			CreatedAt:          now.Add(-time.Duration(i) * time.Second),
		})
	}
	if err := db.CreateInBatches(logs, 500).Error; err != nil {
		t.Fatalf("seed 10k audit logs: %v", err)
	}

	playerUID := uint(4042)
	start := time.Now()
	results, total, err := repo.QueryAuditLogs(AuditLogFilter{
		PlayerUID:            &playerUID,
		ActionType:           "unban",
		CompensationStatuses: []string{"ok"},
		Limit:                20,
	})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("query 10k audit logs: %v", err)
	}
	if total == 0 || len(results) == 0 {
		t.Fatal("expected filtered results from 10k audit log query")
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("expected 10k audit query under 500ms, got %s", elapsed)
	}
}

func TestAuditTrailCompensationOutcomeIsAppendOnlyForRepositoryUpdates(t *testing.T) {
	db, repo := setupCheatRepositoryTest(t)
	pending := "pending"
	message := "original message"
	amount := 100
	audit := database.CheatAuditLog{
		EventType:           "review",
		RoomID:              "room_immutable",
		PlayerUID:           5001,
		SanctionType:        "unban",
		Remark:              "original approval",
		CompensationAmount:  &amount,
		CompensationStatus:  &pending,
		CompensationMessage: &message,
		CreatedAt:           time.Now().UTC(),
	}
	if err := db.Create(&audit).Error; err != nil {
		t.Fatalf("create audit log: %v", err)
	}

	if err := repo.UpdateAuditCompensation(audit.ID, "ok", "补偿成功"); err != nil {
		t.Fatalf("update audit compensation: %v", err)
	}

	var saved database.CheatAuditLog
	if err := db.First(&saved, audit.ID).Error; err != nil {
		t.Fatalf("load audit log: %v", err)
	}
	if saved.CompensationStatus == nil || *saved.CompensationStatus != "ok" {
		t.Fatalf("expected status ok, got %#v", saved.CompensationStatus)
	}
	if saved.CompensationAmount == nil || *saved.CompensationAmount != amount {
		t.Fatalf("compensation amount should remain immutable, got %#v", saved.CompensationAmount)
	}
	if saved.CompensationMessage == nil || *saved.CompensationMessage != message {
		t.Fatalf("compensation message should remain immutable, got %#v", saved.CompensationMessage)
	}
	if saved.Remark != "original approval" {
		t.Fatalf("approval remark should remain immutable, got %q", saved.Remark)
	}
}
