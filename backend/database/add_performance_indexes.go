package database

import (
	"log"
	"strings"

	"gorm.io/gorm"
)

func ensureIndex(db *gorm.DB, model any, indexName string, fieldName string, label string) {
	if db.Migrator().HasIndex(model, indexName) {
		return
	}

	if err := db.Migrator().CreateIndex(model, fieldName); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "failed to create index with name") {
			log.Printf("ℹ️  索引 %s 已存在或名称冲突，跳过", label)
			return
		}
		log.Printf("⚠️  创建索引 %s 失败: %v", label, err)
		return
	}

	log.Printf("✅ 创建索引 %s", label)
}

// AddPerformanceIndexes 添加必要的性能索引
func AddPerformanceIndexes(db *gorm.DB) error {
	log.Println("📊 添加性能优化索引...")

	// 1. Sessions 表索引 - 高频查询
	ensureIndex(db, &UserSession{}, "idx_user_sessions_user_uid", "user_uid", "user_sessions(user_uid)")
	ensureIndex(db, &UserSession{}, "idx_user_sessions_last_active", "last_active", "user_sessions(last_active)")

	// 2. Users 表索引 - 邮箱和用户名查询
	ensureIndex(db, &User{}, "idx_users_email", "email", "users(email)")
	ensureIndex(db, &User{}, "idx_users_username", "username", "users(username)")

	// 3. Users 表 - 封禁和冻结查询
	ensureIndex(db, &User{}, "idx_users_banned_until", "banned_until", "users(banned_until)")
	ensureIndex(db, &User{}, "idx_users_frozen_until", "frozen_until", "users(frozen_until)")

	log.Println("✅ 性能索引添加完成")
	return nil
}
