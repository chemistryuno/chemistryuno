package main

import (
	"chemistryuno/database"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 迁移脚本：将现有的二进制 credential ID 转换为 base64 URL-safe 编码
func move() {
	// 初始化数据库连接
	sqlitePath := os.Getenv("SQLITE_PATH")
	if sqlitePath == "" {
		sqlitePath = "./chemistryuno.db"
	}

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        sqlitePath,
	}, &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 查询所有 WebAuthn credentials
	var credentials []database.WebAuthnCredential
	if err := db.Find(&credentials).Error; err != nil {
		log.Fatalf("查询凭证失败: %v", err)
	}

	if len(credentials) == 0 {
		log.Println("✅ 没有需要迁移的凭证")
		return
	}

	log.Printf("📦 找到 %d 个凭证，开始迁移...\n", len(credentials))

	migrated := 0
	skipped := 0

	for _, cred := range credentials {
		// 检查是否已经是 base64 编码
		if _, err := base64.RawURLEncoding.DecodeString(cred.ID); err == nil {
			// 已经是有效的 base64，跳过
			skipped++
			continue
		}

		// 将原始字符串（可能包含二进制数据）转换为 base64
		originalID := cred.ID
		encodedID := base64.RawURLEncoding.EncodeToString([]byte(originalID))

		// 更新数据库
		if err := db.Model(&database.WebAuthnCredential{}).
			Where("id = ?", originalID).
			Update("id", encodedID).Error; err != nil {
			log.Printf("⚠️  迁移失败 (ID: %s): %v", originalID, err)
			continue
		}

		migrated++
		log.Printf("✓ 迁移成功: %s -> %s", originalID, encodedID)
	}

	fmt.Println()
	log.Printf("🎉 迁移完成！")
	log.Printf("   - 已迁移: %d", migrated)
	log.Printf("   - 已跳过: %d", skipped)
	log.Printf("   - 总计: %d", len(credentials))
}
