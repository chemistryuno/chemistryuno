package database

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMigrateCheatTablesCompensationColumnsAndLegacyQueries(t *testing.T) {
	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        "file::memory:?cache=shared",
	}, &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite test db: %v", err)
	}

	if err := MigrateCheatTables(db); err != nil {
		t.Fatalf("migrate cheat tables: %v", err)
	}

	requiredColumns := []string{
		"compensation_amount",
		"compensation_status",
		"compensation_message",
		"compensation_note",
		"compensation_date",
		"approval_note",
	}
	for _, column := range requiredColumns {
		if !db.Migrator().HasColumn(&CheatAuditLog{}, column) {
			t.Fatalf("expected cheat_audit_logs.%s to exist", column)
		}
	}

	legacyLog := CheatAuditLog{
		EventType: "sanction",
		RoomID:    "room_legacy",
		PlayerUID: 1001,
		Remark:    "legacy ban without compensation fields",
		CreatedAt: time.Now(),
	}
	if err := db.Create(&legacyLog).Error; err != nil {
		t.Fatalf("insert legacy audit log: %v", err)
	}

	var logs []CheatAuditLog
	if err := db.Where("created_at BETWEEN ? AND ?", time.Now().Add(-time.Hour), time.Now().Add(time.Hour)).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		t.Fatalf("legacy time range query should still work: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 legacy audit log, got %d", len(logs))
	}
	if logs[0].CompensationAmount != nil || logs[0].CompensationStatus != nil {
		t.Fatalf("legacy audit row should keep compensation fields null")
	}
}
