package database

import (
	"log"

	"gorm.io/gorm"
)

// MigrateCheatTables 执行反作弊相关的数据库迁移
func MigrateCheatTables(db *gorm.DB) error {
	log.Println("Running anticheat migrations...")

	// 创建新表
	tables := []interface{}{
		&CheatRiskScore{},
		&CheatSanction{},
		&CheatAppeal{},
		&CheatAuditLog{},
	}

	for _, table := range tables {
		if !db.Migrator().HasTable(table) {
			if err := db.Migrator().CreateTable(table); err != nil {
				log.Printf("Failed to create table: %v", err)
				return err
			}
			log.Printf("Created table: %v", table)
		}
	}

	// 自动迁移会创建关键字段上的索引
	if err := db.Migrator().CreateIndex(&CheatRiskScore{}, "room_id"); err != nil {
		log.Printf("Warning: Failed to create index on CheatRiskScore.room_id: %v", err)
	}
	if err := db.Migrator().CreateIndex(&CheatRiskScore{}, "player_uid"); err != nil {
		log.Printf("Warning: Failed to create index on CheatRiskScore.player_uid: %v", err)
	}

	if err := db.Migrator().CreateIndex(&CheatSanction{}, "player_uid"); err != nil {
		log.Printf("Warning: Failed to create index on CheatSanction.player_uid: %v", err)
	}
	if err := db.Migrator().CreateIndex(&CheatSanction{}, "sanction_type"); err != nil {
		log.Printf("Warning: Failed to create index on CheatSanction.sanction_type: %v", err)
	}

	if err := db.Migrator().CreateIndex(&CheatAppeal{}, "player_uid"); err != nil {
		log.Printf("Warning: Failed to create index on CheatAppeal.player_uid: %v", err)
	}
	if err := db.Migrator().CreateIndex(&CheatAppeal{}, "status"); err != nil {
		log.Printf("Warning: Failed to create index on CheatAppeal.status: %v", err)
	}

	if err := db.Migrator().CreateIndex(&CheatAuditLog{}, "event_type"); err != nil {
		log.Printf("Warning: Failed to create index on CheatAuditLog.event_type: %v", err)
	}
	if err := db.Migrator().CreateIndex(&CheatAuditLog{}, "created_at"); err != nil {
		log.Printf("Warning: Failed to create index on CheatAuditLog.created_at: %v", err)
	}

	log.Println("Anticheat migrations completed successfully")
	return nil
}
