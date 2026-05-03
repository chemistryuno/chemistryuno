package repository_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUserRepositoryTest(t *testing.T) (*gorm.DB, *repository.UserRepository) {
	t.Helper()

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        fmt.Sprintf("file:user_repository_%s?mode=memory&cache=shared", t.Name()),
	}, &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite test db: %v", err)
	}
	if err := db.AutoMigrate(&database.User{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}
	return db, repository.NewUserRepository(db)
}

func TestFindUsersWithBlankNickname(t *testing.T) {
	db, repo := setupUserRepositoryTest(t)
	users := []database.User{
		{UID: 100000001, Username: "blank", Nickname: ""},
		{UID: 100000002, Username: "spaces", Nickname: "   "},
		{UID: 100000003, Username: "named", Nickname: "研究员A"},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("create users: %v", err)
	}

	blankUsers, err := repo.FindUsersWithBlankNickname()
	if err != nil {
		t.Fatalf("find blank nicknames: %v", err)
	}
	if len(blankUsers) != 2 {
		t.Fatalf("expected 2 blank nickname users, got %d", len(blankUsers))
	}
	if blankUsers[0].UID != 100000001 || blankUsers[1].UID != 100000002 {
		t.Fatalf("unexpected blank users: %#v", blankUsers)
	}
}

func TestRepairMissingNicknamesIsIdempotent(t *testing.T) {
	db, repo := setupUserRepositoryTest(t)
	users := []database.User{
		{UID: 100000011, Username: "blank", Nickname: ""},
		{UID: 100000012, Username: "named", Nickname: "已有昵称"},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("create users: %v", err)
	}

	repaired, err := utils.RepairMissingNicknames(repo)
	if err != nil {
		t.Fatalf("repair missing nicknames: %v", err)
	}
	if repaired != 1 {
		t.Fatalf("expected 1 repaired user, got %d", repaired)
	}

	var blankUser database.User
	if err := db.First(&blankUser, 100000011).Error; err != nil {
		t.Fatalf("load repaired user: %v", err)
	}
	if !regexp.MustCompile(`^研究员\d{6}$`).MatchString(blankUser.Nickname) {
		t.Fatalf("unexpected repaired nickname: %q", blankUser.Nickname)
	}

	repairedAgain, err := utils.RepairMissingNicknames(repo)
	if err != nil {
		t.Fatalf("repair missing nicknames again: %v", err)
	}
	if repairedAgain != 0 {
		t.Fatalf("expected idempotent repair count 0, got %d", repairedAgain)
	}

	var namedUser database.User
	if err := db.First(&namedUser, 100000012).Error; err != nil {
		t.Fatalf("load named user: %v", err)
	}
	if namedUser.Nickname != "已有昵称" {
		t.Fatalf("expected existing nickname preserved, got %q", namedUser.Nickname)
	}
}

func TestGenerateUniqueRandomNicknameFallsBackAfterCollisions(t *testing.T) {
	exists := func(candidate string) (bool, error) {
		return strings.HasPrefix(candidate, "研究员") && regexp.MustCompile(`^研究员\d{6}$`).MatchString(candidate), nil
	}

	nickname, err := utils.GenerateUniqueRandomNickname("研究员", "100000123", exists)
	if err != nil {
		t.Fatalf("generate unique nickname: %v", err)
	}
	if nickname != "研究员100000123" {
		t.Fatalf("expected deterministic fallback nickname, got %q", nickname)
	}
}
