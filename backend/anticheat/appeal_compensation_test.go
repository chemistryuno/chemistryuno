package anticheat

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAppealCompensationTest(t *testing.T) (*gorm.DB, *AppealManager, *repository.CheatRepository) {
	t.Helper()

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        fmt.Sprintf("file:appeal_compensation_%s?mode=memory&cache=shared", t.Name()),
	}, &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite test db: %v", err)
	}

	if err := db.AutoMigrate(&database.User{}, &database.FuelCompensationRecord{}); err != nil {
		t.Fatalf("migrate user/fuel tables: %v", err)
	}
	if err := database.MigrateCheatTables(db); err != nil {
		t.Fatalf("migrate cheat tables: %v", err)
	}

	database.DB = db
	cheatRepo := repository.NewCheatRepository(db)
	userRepo := repository.NewUserRepository()
	return db, NewAppealManager(cheatRepo, userRepo), cheatRepo
}

func createAppealCompensationFixture(t *testing.T, db *gorm.DB, cheatRepo *repository.CheatRepository, uid uint) database.CheatAppeal {
	t.Helper()

	user := database.User{
		UID:      uid,
		Username: fmt.Sprintf("player_%d", uid),
		Nickname: fmt.Sprintf("Player %d", uid),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	appeal := database.CheatAppeal{
		RoomID:    "room_appeal",
		PlayerUID: uid,
		Reason:    "false positive",
		Evidence:  "clean replay",
		Status:    "pending",
	}
	if err := cheatRepo.SaveAppeal(&appeal); err != nil {
		t.Fatalf("create appeal: %v", err)
	}
	return appeal
}

func TestApproveAppealWithCompensationStoresPendingClaim(t *testing.T) {
	db, manager, cheatRepo := setupAppealCompensationTest(t)
	appeal := createAppealCompensationFixture(t, db, cheatRepo, 1001)
	bannedUntil := time.Now().Add(2 * time.Hour)
	if err := db.Model(&database.User{}).Where("uid = ?", appeal.PlayerUID).Updates(map[string]interface{}{
		"banned_until": &bannedUntil,
		"ban_reason":   "active ban",
	}).Error; err != nil {
		t.Fatalf("seed account ban: %v", err)
	}

	outcome, err := manager.ApproveAppealWithCompensation(appeal.ID, 9001, "approved", 150, "restored", nil)
	if err != nil {
		t.Fatalf("approve appeal: %v", err)
	}
	if outcome.CompensationStatus != "pending" {
		t.Fatalf("expected pending compensation, got %q", outcome.CompensationStatus)
	}

	var user database.User
	if err := db.First(&user, appeal.PlayerUID).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user.Fuel != 0 {
		t.Fatalf("approval should not issue fuel before claim, got %d", user.Fuel)
	}
	if user.Points != 1000 {
		t.Fatalf("approval should not change visible points before claim, got %.0f", user.Points)
	}
	if user.BannedUntil != nil || user.BanReason != "" {
		t.Fatalf("expected account ban cleared, got until=%v reason=%q", user.BannedUntil, user.BanReason)
	}

	var savedAppeal database.CheatAppeal
	if err := db.First(&savedAppeal, appeal.ID).Error; err != nil {
		t.Fatalf("load saved appeal: %v", err)
	}
	if savedAppeal.Status != "approved" || savedAppeal.CompensationStatus != "pending" || savedAppeal.CompensationAmount != 150 {
		t.Fatalf("unexpected saved appeal compensation state: status=%q compensation=%q amount=%d", savedAppeal.Status, savedAppeal.CompensationStatus, savedAppeal.CompensationAmount)
	}

	var records int64
	if err := db.Model(&database.FuelCompensationRecord{}).Where("user_uid = ?", appeal.PlayerUID).Count(&records).Error; err != nil {
		t.Fatalf("count compensation records: %v", err)
	}
	if records != 0 {
		t.Fatalf("approval should not create compensation records, got %d", records)
	}
}

func TestClaimCompensationIssuesFuelAndIsIdempotent(t *testing.T) {
	db, manager, cheatRepo := setupAppealCompensationTest(t)
	appeal := createAppealCompensationFixture(t, db, cheatRepo, 1005)

	if _, err := manager.ApproveAppealWithCompensation(appeal.ID, 9001, "approved", 150, "restored", nil); err != nil {
		t.Fatalf("approve appeal: %v", err)
	}

	claim, err := manager.ClaimCompensation(appeal.ID, appeal.PlayerUID)
	if err != nil {
		t.Fatalf("claim compensation: %v", err)
	}
	if claim.CompensationStatus != "ok" || claim.Idempotent {
		t.Fatalf("expected first claim to be non-idempotent ok, got status=%q idempotent=%v", claim.CompensationStatus, claim.Idempotent)
	}

	var user database.User
	if err := db.First(&user, appeal.PlayerUID).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user.Fuel != 150 {
		t.Fatalf("expected fuel 150 after claim, got %d", user.Fuel)
	}
	if user.Points != 1150 {
		t.Fatalf("expected visible points 1150 after claim, got %.0f", user.Points)
	}

	duplicate, err := manager.ClaimCompensation(appeal.ID, appeal.PlayerUID)
	if err != nil {
		t.Fatalf("duplicate claim compensation: %v", err)
	}
	if duplicate.CompensationStatus != "ok" || !duplicate.Idempotent {
		t.Fatalf("expected duplicate claim to be idempotent ok, got status=%q idempotent=%v", duplicate.CompensationStatus, duplicate.Idempotent)
	}
	if err := db.First(&user, appeal.PlayerUID).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if user.Fuel != 150 || user.Points != 1150 {
		t.Fatalf("duplicate claim should not add balances again, got fuel=%d points=%.0f", user.Fuel, user.Points)
	}
	var records int64
	if err := db.Model(&database.FuelCompensationRecord{}).Where("user_uid = ?", appeal.PlayerUID).Count(&records).Error; err != nil {
		t.Fatalf("count compensation records: %v", err)
	}
	if records != 1 {
		t.Fatalf("expected one compensation record, got %d", records)
	}

	reapprove, err := manager.ApproveAppealWithCompensation(appeal.ID, 9001, "approved again", 200, "restored again", nil)
	if err != nil {
		t.Fatalf("reapprove claimed appeal: %v", err)
	}
	if reapprove.CompensationStatus != "ok" || !reapprove.Idempotent {
		t.Fatalf("expected reapproval after claim to be idempotent ok, got status=%q idempotent=%v", reapprove.CompensationStatus, reapprove.Idempotent)
	}
	if err := db.First(&user, appeal.PlayerUID).Error; err != nil {
		t.Fatalf("reload user after reapproval: %v", err)
	}
	if user.Fuel != 150 || user.Points != 1150 {
		t.Fatalf("reapproval after claim should not change balances, got fuel=%d points=%.0f", user.Fuel, user.Points)
	}
	var savedAppeal database.CheatAppeal
	if err := db.First(&savedAppeal, appeal.ID).Error; err != nil {
		t.Fatalf("load appeal after reapproval: %v", err)
	}
	if savedAppeal.CompensationStatus != "ok" || savedAppeal.CompensationAmount != 150 {
		t.Fatalf("reapproval should preserve claimed compensation, got status=%q amount=%d", savedAppeal.CompensationStatus, savedAppeal.CompensationAmount)
	}
}

func TestClaimCompensationFailureLeavesApprovedAppealRetryable(t *testing.T) {
	db, manager, cheatRepo := setupAppealCompensationTest(t)
	appeal := createAppealCompensationFixture(t, db, cheatRepo, 1002)

	outcome, err := manager.ApproveAppealWithCompensation(appeal.ID, 9001, "approved", 0, "restored", nil)
	if err != nil {
		t.Fatalf("approval should not fail when compensation amount is zero: %v", err)
	}
	if outcome.CompensationStatus != "pending" {
		t.Fatalf("expected pending compensation, got %q", outcome.CompensationStatus)
	}

	var savedAppeal database.CheatAppeal
	if err := db.First(&savedAppeal, appeal.ID).Error; err != nil {
		t.Fatalf("load appeal: %v", err)
	}
	if savedAppeal.Status != "approved" {
		t.Fatalf("appeal should be approved, got %q", savedAppeal.Status)
	}
	if savedAppeal.CompensationStatus != "pending" {
		t.Fatalf("expected persisted pending compensation status, got %q", savedAppeal.CompensationStatus)
	}

	var audit database.CheatAuditLog
	if err := db.Where("appeal_id = ?", appeal.ID).First(&audit).Error; err != nil {
		t.Fatalf("load audit: %v", err)
	}
	if audit.CompensationStatus == nil || *audit.CompensationStatus != "pending" {
		t.Fatalf("expected audit compensation pending status, got %#v", audit.CompensationStatus)
	}

	if _, err := manager.ClaimCompensation(appeal.ID, appeal.PlayerUID); err == nil {
		t.Fatal("expected zero amount claim to fail")
	}

	retry, err := manager.ApproveAppealWithCompensation(appeal.ID, 9001, "retry compensation", 80, "retry restored", nil)
	if err != nil {
		t.Fatalf("retry approval should update pending compensation: %v", err)
	}
	if retry.CompensationStatus != "pending" {
		t.Fatalf("expected retry compensation to stay pending, got %q", retry.CompensationStatus)
	}
	claim, err := manager.ClaimCompensation(appeal.ID, appeal.PlayerUID)
	if err != nil {
		t.Fatalf("retry claim should succeed: %v", err)
	}
	if claim.CompensationStatus != "ok" {
		t.Fatalf("expected retry claim compensation to be ok, got %q", claim.CompensationStatus)
	}

	var user database.User
	if err := db.First(&user, appeal.PlayerUID).Error; err != nil {
		t.Fatalf("load user after retry: %v", err)
	}
	if user.Fuel != 80 {
		t.Fatalf("expected retry to issue 80 fuel, got %d", user.Fuel)
	}
	if user.Points != 1080 {
		t.Fatalf("expected retry to issue visible points, got %.0f", user.Points)
	}
}

func TestApproveAppealUsesUpdatedCompensationConfig(t *testing.T) {
	db, manager, cheatRepo := setupAppealCompensationTest(t)
	appeal := createAppealCompensationFixture(t, db, cheatRepo, 1003)
	config := NewDefaultConfig()
	config.UnbanConfig.CompensationAmount = 275
	config.UnbanConfig.DefaultMessage = "custom policy"

	if err := manager.ApproveAppeal(appeal.ID, 9001, "approved with policy", nil, config); err != nil {
		t.Fatalf("approve appeal with config: %v", err)
	}

	var savedAppeal database.CheatAppeal
	if err := db.First(&savedAppeal, appeal.ID).Error; err != nil {
		t.Fatalf("load appeal: %v", err)
	}
	if savedAppeal.CompensationAmount != 275 {
		t.Fatalf("expected compensation amount from config, got %d", savedAppeal.CompensationAmount)
	}

	var audit database.CheatAuditLog
	if err := db.Where("appeal_id = ?", appeal.ID).First(&audit).Error; err != nil {
		t.Fatalf("load audit: %v", err)
	}
	if audit.CompensationMessage == nil || *audit.CompensationMessage != "custom policy" {
		t.Fatalf("expected audit compensation message from config, got %#v", audit.CompensationMessage)
	}
}

func TestUserRepositoryAddFuelDuplicateCompensationID(t *testing.T) {
	db, _, _ := setupAppealCompensationTest(t)
	repo := repository.NewUserRepository()
	user := database.User{UID: 1004, Username: "fuel-player", Nickname: "Fuel Player"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	if _, err := repo.AddFuel(user.UID, 50, "appeal_42"); err != nil {
		t.Fatalf("first AddFuel should succeed: %v", err)
	}
	if _, err := repo.AddFuel(user.UID, 50, "appeal_42"); !errors.Is(err, repository.ErrFuelCompensationAlreadyIssued) {
		t.Fatalf("expected duplicate AddFuel to return ErrFuelCompensationAlreadyIssued, got %v", err)
	}

	var saved database.User
	if err := db.First(&saved, user.UID).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if saved.Fuel != 50 {
		t.Fatalf("duplicate AddFuel should not change fuel, got %d", saved.Fuel)
	}
	if saved.Points != 1050 {
		t.Fatalf("duplicate AddFuel should only add visible points once, got %.0f", saved.Points)
	}
}

func TestApproveAppealCompensationEdgeCases(t *testing.T) {
	t.Run("zero amount remains pending but cannot be claimed", func(t *testing.T) {
		db, manager, cheatRepo := setupAppealCompensationTest(t)
		appeal := createAppealCompensationFixture(t, db, cheatRepo, 1101)

		outcome, err := manager.ApproveAppealWithCompensation(appeal.ID, 9001, "zero amount", 0, "zero amount message", nil)
		if err != nil {
			t.Fatalf("approval should remain successful for zero compensation: %v", err)
		}
		if outcome.CompensationStatus != "pending" {
			t.Fatalf("expected pending zero-amount compensation, got %q", outcome.CompensationStatus)
		}
		if _, err := manager.ClaimCompensation(appeal.ID, appeal.PlayerUID); err == nil {
			t.Fatal("expected zero amount claim to fail")
		}
	})

	t.Run("max amount and special message persist", func(t *testing.T) {
		db, manager, cheatRepo := setupAppealCompensationTest(t)
		appeal := createAppealCompensationFixture(t, db, cheatRepo, 1102)
		message := "特殊字符 <> & \" ' / newline\n补偿确认"

		outcome, err := manager.ApproveAppealWithCompensation(appeal.ID, 9001, "max amount", 10000, message, nil)
		if err != nil {
			t.Fatalf("approve max amount: %v", err)
		}
		if outcome.CompensationStatus != "pending" {
			t.Fatalf("expected pending max amount compensation, got %q", outcome.CompensationStatus)
		}

		var audit database.CheatAuditLog
		if err := db.Where("appeal_id = ?", appeal.ID).First(&audit).Error; err != nil {
			t.Fatalf("load audit: %v", err)
		}
		if audit.CompensationAmount == nil || *audit.CompensationAmount != 10000 {
			t.Fatalf("expected audit amount 10000, got %#v", audit.CompensationAmount)
		}
		if audit.CompensationMessage == nil || *audit.CompensationMessage != message {
			t.Fatalf("expected special message to persist, got %#v", audit.CompensationMessage)
		}
	})
}

func TestRapidCompensationClaimsDoNotDoubleIssueFuel(t *testing.T) {
	db, manager, cheatRepo := setupAppealCompensationTest(t)
	appeal := createAppealCompensationFixture(t, db, cheatRepo, 1201)

	if _, err := manager.ApproveAppealWithCompensation(appeal.ID, 9001, "rapid retry", 75, "rapid compensation", nil); err != nil {
		t.Fatalf("approve appeal: %v", err)
	}
	for i := 0; i < 5; i++ {
		outcome, err := manager.ClaimCompensation(appeal.ID, appeal.PlayerUID)
		if err != nil {
			t.Fatalf("rapid claim %d failed: %v", i, err)
		}
		if outcome.CompensationStatus != "ok" {
			t.Fatalf("rapid claim %d expected ok, got %q", i, outcome.CompensationStatus)
		}
	}

	var user database.User
	if err := db.First(&user, appeal.PlayerUID).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user.Fuel != 75 {
		t.Fatalf("rapid approvals should issue fuel once, got %d", user.Fuel)
	}
	if user.Points != 1075 {
		t.Fatalf("rapid approvals should issue visible points once, got %.0f", user.Points)
	}
}
