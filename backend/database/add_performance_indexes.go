package database

import (
	"log"

	"gorm.io/gorm"
)

// AddPerformanceIndexes 添加必要的性能索引
func AddPerformanceIndexes(db *gorm.DB) error {
	log.Println("📊 添加性能优化索引...")

	// 1. Sessions 表索引 - 高频查询
	if !db.Migrator().HasIndex(&UserSession{}, "idx_user_sessions_user_id") {
		if err := db.Migrator().CreateIndex(&UserSession{}, "user_uid"); err != nil {
			log.Printf("⚠️  创建索引 user_sessions(user_uid) 失败: %v", err)
		} else {
			log.Println("✅ 创建索引 user_sessions(user_uid)")
		}
	}

	if !db.Migrator().HasIndex(&UserSession{}, "idx_user_sessions_last_active") {
		if err := db.Migrator().CreateIndex(&UserSession{}, "last_active"); err != nil {
			log.Printf("⚠️  创建索引 user_sessions(last_active) 失败: %v", err)
		} else {
			log.Println("✅ 创建索引 user_sessions(last_active)")
		}
	}

	// 2. Users 表索引 - 邮箱和用户名查询
	if !db.Migrator().HasIndex(&User{}, "idx_users_email") {
		if err := db.Migrator().CreateIndex(&User{}, "email"); err != nil {
			log.Printf("⚠️  创建索引 users(email) 失败: %v", err)
		} else {
			log.Println("✅ 创建索引 users(email)")
		}
	}

	if !db.Migrator().HasIndex(&User{}, "idx_users_username") {
		if err := db.Migrator().CreateIndex(&User{}, "username"); err != nil {
			log.Printf("⚠️  创建索引 users(username) 失败: %v", err)
		} else {
			log.Println("✅ 创建索引 users(username)")
		}
	}

	// 3. Users 表 - 封禁和冻结查询
	if !db.Migrator().HasIndex(&User{}, "idx_users_banned_until") {
		if err := db.Migrator().CreateIndex(&User{}, "banned_until"); err != nil {
			log.Printf("⚠️  创建索引 users(banned_until) 失败: %v", err)
		} else {
			log.Println("✅ 创建索引 users(banned_until)")
		}
	}

	if !db.Migrator().HasIndex(&User{}, "idx_users_frozen_until") {
		if err := db.Migrator().CreateIndex(&User{}, "frozen_until"); err != nil {
			log.Printf("⚠️  创建索引 users(frozen_until) 失败: %v", err)
		} else {
			log.Println("✅ 创建索引 users(frozen_until)")
		}
	}

	log.Println("✅ 性能索引添加完成")
	return nil
}
