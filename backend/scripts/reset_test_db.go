//go:build scripts
// +build scripts

package main

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/utils"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("🔄 Chemistry UNO - 测试数据库重置工具")
	log.Println("==========================================")

	// 加载 .env 文件
	envPath := ".env"
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		envPath = "../.env"
	}

	if err := godotenv.Load(envPath); err == nil {
		absPath, _ := filepath.Abs(envPath)
		log.Printf("✅ 已加载配置文件: %s\n", absPath)
	}

	// 1. 删除旧数据库文件以确保完全重置
	dbPath := os.Getenv("SQLITE_PATH")
	if dbPath == "" {
		dbPath = "./chemistryuno.db"
	}

	log.Printf("🗑️  删除旧数据库文件: %s", dbPath)
	if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
		log.Printf("   ⚠️  删除数据库文件失败: %v (继续执行)", err)
	} else if !os.IsNotExist(err) {
		log.Println("   ✅ 旧数据库文件已删除")
	}

	// 删除 WAL 和 SHM 文件
	os.Remove(dbPath + "-wal")
	os.Remove(dbPath + "-shm")

	// 2. 初始化数据库（会创建默认数据）
	log.Println("\n🔄 正在初始化全新数据库...")
	if err := database.InitDB(""); err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}

	// 3. 初始化Repository
	repository.InitRepositories()

	// 4. 清空默认创建的用户，准备创建测试用户
	log.Println("\n🗑️  清空默认用户...")
	database.DB.Exec("DELETE FROM users")
	database.DB.Exec("DELETE FROM sessions")
	database.DB.Exec("DELETE FROM game_history")
	log.Println("✅ 默认用户已清空")

	// 4. 创建测试用户
	log.Println("\n👥 创建测试用户...")

	// 管理员
	log.Println("   👨‍💼 管理员账号...")
	adminPass := "admin123"
	adminHash, _ := utils.HashPassword(adminPass)
	admin := database.User{
		Username: "admin", Email: "admin@chemistryuno.com",
		Nickname: "系统管理员", Password: adminHash,
		Avatar: "👨‍💼", IsAdmin: true, Role: "admin",
		Points: 10000, MonthlyPoints: 10000,
	}
	if err := database.DB.Create(&admin).Error; err != nil {
		log.Printf("      ❌ admin 创建失败: %v", err)
	} else {
		log.Printf("      ✅ admin (密码: %s)", adminPass)
	}

	// test 用户
	log.Println("   🧑‍🔬 测试账号...")
	testPass := "test123"
	testHash, _ := utils.HashPassword(testPass)
	test := database.User{
		Username: "test", Email: "test@example.com",
		Nickname: "测试用户", Password: testHash,
		Avatar: "🧑‍🔬", IsAdmin: false, Role: "user",
		Points: 1000, MonthlyPoints: 1000,
	}
	if err := database.DB.Create(&test).Error; err != nil {
		log.Printf("      ❌ test 创建失败: %v", err)
	} else {
		log.Printf("      ✅ test (密码: %s)", testPass)
	}

	// test1-4 批量用户
	log.Println("   🧪 批量测试账号...")
	batchPass := "123456"
	batchHash, _ := utils.HashPassword(batchPass)
	successCount := 0

	for i := 1; i <= 4; i++ {
		u := database.User{
			Username: fmt.Sprintf("test%d", i),
			Email: fmt.Sprintf("test%d@example.com", i),
			Nickname: fmt.Sprintf("测试用户%d", i),
			Password: batchHash,
			Avatar: "🧪", IsAdmin: false, Role: "user",
			Points: 1000, MonthlyPoints: 1000,
		}
		if err := database.DB.Create(&u).Error; err != nil {
			log.Printf("      ❌ test%d 创建失败: %v", i, err)
		} else {
			log.Printf("      ✅ test%d (密码: %s)", i, batchPass)
			successCount++
		}
	}
	log.Printf("      📊 成功创建 %d/4 个批量用户\n", successCount)

	// 5. 显示结果
	log.Println("==========================================")
	log.Println("✅ 测试数据库重置完成！")
	log.Println("==========================================")

	// 列出所有用户
	var users []database.User
	database.DB.Order("uid").Find(&users)

	log.Println("📋 测试账号列表:")
	log.Println("------------------------------------------")
	log.Printf("%-6s %-12s %-25s %-10s", "UID", "用户名", "邮箱", "密码")
	log.Println("\n------------------------------------------")

	passMap := map[string]string{
		"admin": "admin123", "test": "test123",
		"test1": "123456", "test2": "123456",
		"test3": "123456", "test4": "123456",
	}

	for _, u := range users {
		pass := passMap[u.Username]
		if pass == "" {
			pass = "-"
		}
		log.Printf("%-6d %-12s %-25s %-10s", u.UID, u.Username, u.Email, pass)
	}

	log.Println("------------------------------------------")
	log.Printf("共 %d 个测试账号\n", len(users))
}
