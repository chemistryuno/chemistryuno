package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"chemistryuno/backend/database"

	"github.com/joho/godotenv"
)

type credentialRow struct {
	ID string `gorm:"column:id"`
}

func ensureProjectRoot() {
	if _, err := os.Stat("backend"); err == nil {
		return
	}
	if _, err := os.Stat("../backend"); err == nil {
		_ = os.Chdir("..")
	}
}

func normalizeCredentialID(raw string) (string, bool) {
	if raw == "" {
		return "", false
	}

	decodeFns := []func(string) ([]byte, error){
		base64.RawURLEncoding.DecodeString,
		base64.URLEncoding.DecodeString,
		base64.StdEncoding.DecodeString,
		base64.RawStdEncoding.DecodeString,
	}

	for _, decode := range decodeFns {
		decoded, err := decode(raw)
		if err != nil {
			continue
		}
		return base64.RawURLEncoding.EncodeToString(decoded), true
	}

	return "", false
}

func main() {
	ensureProjectRoot()
	_ = godotenv.Load("backend/.env", ".env")

	if err := database.InitDB(""); err != nil {
		log.Fatalf("database init failed: %v", err)
	}

	db := database.DB
	if db == nil {
		log.Fatal("database handle is nil")
	}

	if !db.Migrator().HasTable("webauthn_credentials") {
		log.Println("table webauthn_credentials does not exist, nothing to migrate")
		return
	}

	var rows []credentialRow
	if err := db.Table("webauthn_credentials").Select("id").Find(&rows).Error; err != nil {
		log.Fatalf("load credential ids failed: %v", err)
	}

	if len(rows) == 0 {
		log.Println("no credentials found, nothing to migrate")
		return
	}

	updated := 0
	skipped := 0
	conflicts := 0
	failed := 0

	for _, row := range rows {
		normalized, ok := normalizeCredentialID(row.ID)
		if !ok {
			skipped++
			continue
		}
		if normalized == row.ID {
			skipped++
			continue
		}

		var existing int64
		if err := db.Table("webauthn_credentials").Where("id = ?", normalized).Count(&existing).Error; err != nil {
			failed++
			log.Printf("count normalized id failed for %s: %v", row.ID, err)
			continue
		}
		if existing > 0 {
			conflicts++
			log.Printf("skip id %s: normalized id already exists", row.ID)
			continue
		}

		if err := db.Table("webauthn_credentials").Where("id = ?", row.ID).Update("id", normalized).Error; err != nil {
			failed++
			log.Printf("update failed for %s: %v", row.ID, err)
			continue
		}
		updated++
	}

	fmt.Printf("credential id migration finished: updated=%d skipped=%d conflicts=%d failed=%d total=%d\n", updated, skipped, conflicts, failed, len(rows))
}
