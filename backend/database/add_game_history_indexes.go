package database

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// GameHistoryPlayer 游戏历史玩家关联表（用于优化查询）
type GameHistoryPlayer struct {
	ID            uint      `gorm:"primaryKey"`
	GameHistoryID uint      `gorm:"not null;index:idx_ghp_player_game,priority:1;index:idx_ghp_game"`
	PlayerUID     uint      `gorm:"not null;index:idx_ghp_player_game,priority:2"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
}

// MigrateGameHistoryIndexes adds performance indexes for game history queries
func MigrateGameHistoryIndexes(db *gorm.DB) error {
	log.Println("Running game history performance migration...")

	// 1. Create junction table for player lookups
	if !db.Migrator().HasTable(&GameHistoryPlayer{}) {
		if err := db.Migrator().CreateTable(&GameHistoryPlayer{}); err != nil {
			log.Printf("Failed to create game_history_players table: %v", err)
			return err
		}
		log.Println("✅ Created game_history_players junction table")
	}

	// 2. Create composite index on (player_uid, game_history_id)
	// This is already defined in the struct tags, but ensure it exists
	if !db.Migrator().HasIndex(&GameHistoryPlayer{}, "idx_ghp_player_game") {
		if err := db.Migrator().CreateIndex(&GameHistoryPlayer{}, "idx_ghp_player_game"); err != nil {
			log.Printf("⚠️  Failed to create composite index: %v", err)
		} else {
			log.Println("✅ Created composite index on (player_uid, game_history_id)")
		}
	}

	// 3. Create index on game_history_id for reverse lookups
	if !db.Migrator().HasIndex(&GameHistoryPlayer{}, "idx_ghp_game") {
		if err := db.Migrator().CreateIndex(&GameHistoryPlayer{}, "idx_ghp_game"); err != nil {
			log.Printf("⚠️  Failed to create game_history_id index: %v", err)
		} else {
			log.Println("✅ Created index on game_history_id")
		}
	}

	// 4. Add MySQL-specific functional index on players JSON column
	// This is optional and only works on MySQL 8.0+
	if db.Dialector.Name() == "mysql" {
		// Check MySQL version
		var version string
		if err := db.Raw("SELECT VERSION()").Scan(&version).Error; err == nil {
			log.Printf("MySQL version: %s", version)

			// Try to create functional index (will fail gracefully on older MySQL)
			// Note: GORM doesn't support functional indexes directly, so we use raw SQL
			indexSQL := `CREATE INDEX idx_game_history_players_json
				ON game_history ((CAST(players AS JSON)))`

			// Check if index already exists
			var indexExists int64
			db.Raw(`SELECT COUNT(*) FROM information_schema.statistics
				WHERE table_schema = DATABASE()
				AND table_name = 'game_history'
				AND index_name = 'idx_game_history_players_json'`).Scan(&indexExists)

			if indexExists == 0 {
				if err := db.Exec(indexSQL).Error; err != nil {
					log.Printf("ℹ️  Functional index not created (requires MySQL 8.0+): %v", err)
				} else {
					log.Println("✅ Created MySQL functional index on players JSON column")
				}
			} else {
				log.Println("✅ MySQL functional index already exists")
			}
		}
	} else {
		log.Println("ℹ️  Skipping MySQL functional index (not using MySQL)")
	}

	log.Println("Game history performance migration completed")
	return nil
}

// PopulateGameHistoryPlayers backfills the junction table from existing game history
// This should be run AFTER the table is created, optionally in a background job
func PopulateGameHistoryPlayers(db *gorm.DB) error {
	log.Println("Starting game_history_players backfill...")

	// Check if table exists
	if !db.Migrator().HasTable(&GameHistoryPlayer{}) {
		return fmt.Errorf("game_history_players table does not exist")
	}

	// Get all game history records that haven't been indexed yet
	var histories []GameHistory
	if err := db.Find(&histories).Error; err != nil {
		return fmt.Errorf("failed to load game histories: %v", err)
	}

	log.Printf("Found %d game history records to process", len(histories))

	indexed := 0
	skipped := 0

	for _, history := range histories {
		// Check if already indexed
		var count int64
		db.Model(&GameHistoryPlayer{}).Where("game_history_id = ?", history.ID).Count(&count)
		if count > 0 {
			skipped++
			continue
		}

		// Parse players JSON
		var players []int
		if err := json.Unmarshal([]byte(history.Players), &players); err != nil {
			log.Printf("⚠️  Failed to parse players for game %d: %v", history.ID, err)
			continue
		}

		// Insert junction records
		for _, playerUID := range players {
			junction := GameHistoryPlayer{
				GameHistoryID: history.ID,
				PlayerUID:     uint(playerUID),
			}
			if err := db.Create(&junction).Error; err != nil {
				log.Printf("⚠️  Failed to create junction for game %d, player %d: %v", history.ID, playerUID, err)
			}
		}

		indexed++
		if indexed%100 == 0 {
			log.Printf("Progress: indexed %d/%d records", indexed, len(histories))
		}
	}

	log.Printf("✅ Backfill complete: indexed %d, skipped %d (already indexed)", indexed, skipped)
	return nil
}
