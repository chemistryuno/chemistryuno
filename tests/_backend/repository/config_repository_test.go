package repository

import (
	"testing"

	"chemistryuno/backend/database"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestInitDefaultConfigsKeepsReconnectGracePeriodKey(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&database.SystemConfig{}); err != nil {
		t.Fatalf("migrate system configs: %v", err)
	}

	previousDB := database.DB
	database.DB = db
	t.Cleanup(func() {
		database.DB = previousDB
	})

	// Simulate legacy deployments that only had reconnect_grace_period.
	if err := db.Create(&database.SystemConfig{
		Key:   "reconnect_grace_period",
		Value: "42",
	}).Error; err != nil {
		t.Fatalf("seed reconnect_grace_period: %v", err)
	}

	repo := NewConfigRepository()
	if err := repo.InitDefaultConfigs(); err != nil {
		t.Fatalf("init default configs: %v", err)
	}

	var reconnect database.SystemConfig
	if err := db.Where("`key` = ?", "reconnect_grace_period").Take(&reconnect).Error; err != nil {
		t.Fatalf("expected reconnect_grace_period to exist after init: %v", err)
	}
	if reconnect.Value != "42" {
		t.Fatalf("expected reconnect_grace_period value to be kept as 42, got %q", reconnect.Value)
	}
}
