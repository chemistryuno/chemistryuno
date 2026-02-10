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
	log.Println("🧪 Chemistry UNO - 测试用户创建工具")
	log.Println("====================================")

	// 尝试加载 .env 文件
	envPath := "../.env"
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		envPath = "../../.env"
	}

	if err := godotenv.Load(envPath); err == nil {
		absPath, _ := filepath.Abs(envPath)
		log.Printf("✅ 已加载配置文件: %s", absPath)
	} else {
		log.Println("⚠️  未找到 .env 文件，将使用默认配置")
	}

	// 初始化数据库
	log.Println("\n🔄 正在连接数据库...")
	if err := database.InitDB(""); err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}
	log.Println("✅ 数据库连接成功")

	// 初始化仓库
	repository.InitRepositories()

	// 统计当前用户数
	var totalUsers int64
	database.DB.Model(&database.User{}).Count(&totalUsers)
	log.Printf("\n📊 当前数据库中共有 %d 个用户", totalUsers)

	// 创建管理员账号
	log.Println("\n👨‍💼 正在创建管理员账号...")
	createAdmin()

	// 创建 test 账号
	log.Println("\n🧑‍🔬 正在创建 test 账号...")
	createTestUser()

	// 创建 test1-4 账号
	log.Println("\n🧪 正在创建 test1-4 批量测试账号...")
	createBatchUsers()

	// 显示最终结果
	database.DB.Model(&database.User{}).Count(&totalUsers)
	log.Println("\n====================================")
	log.Printf("✅ 操作完成！当前共有 %d 个用户", totalUsers)
	log.Println("====================================")

	// 列出所有用户
	log.Println("\n📋 用户列表:")
	var users []database.User
	database.DB.Order("created_at desc").Find(&users)
	for _, u := range users {
		role := "普通用户"
		if u.IsAdmin {
			role = "管理员"
		} else if u.Role == "co-worker" {
			role = "协作者"
		}
		log.Printf("   - %s (UID: %d, 角色: %s, 用户名: %s)", u.Nickname, u.UID, role, u.Username)
	}
	log.Println("")
}

func createAdmin() {
	adminPass := "admin123"
	adminHashedPassword, err := utils.HashPassword(adminPass)
	if err != nil {
		log.Fatalf("❌ 管理员密码加密失败: %v", err)
	}

	var adminCount int64
	database.DB.Model(&database.User{}).Where("username = ? OR email = ?", "admin", "admin@chemistryuno.com").Count(&adminCount)

	if adminCount == 0 {
		adminUser := database.User{
			Username:      "admin",
			Email:         "admin@chemistryuno.com",
			Nickname:      "系统管理员",
			Password:      adminHashedPassword,
			Avatar:        "👨‍💼",
			IsAdmin:       true,
			Role:          "admin",
			Points:        10000,
			MonthlyPoints: 10000,
		}

		if err := database.DB.Create(&adminUser).Error; err != nil {
			log.Printf("   ❌ 创建失败: %v", err)
		} else {
			log.Printf("   ✅ 创建成功: admin@chemistryuno.com (密码: %s)", adminPass)
		}
	} else {
		log.Printf("   ℹ️  账号已存在，跳过创建")
	}
}

func createTestUser() {
	testPass := "test123"
	testHashedPassword, err := utils.HashPassword(testPass)
	if err != nil {
		log.Fatalf("❌ 测试用户密码加密失败: %v", err)
	}

	var testCount int64
	database.DB.Model(&database.User{}).Where("username = ? OR email = ?", "test", "test@example.com").Count(&testCount)

	if testCount == 0 {
		testUser := database.User{
			Username:      "test",
			Email:         "test@example.com",
			Nickname:      "测试用户",
			Password:      testHashedPassword,
			Avatar:        "🧑‍🔬",
			IsAdmin:       false,
			Role:          "user",
			Points:        1000,
			MonthlyPoints: 1000,
		}

		if err := database.DB.Create(&testUser).Error; err != nil {
			log.Printf("   ❌ 创建失败: %v", err)
		} else {
			log.Printf("   ✅ 创建成功: test@example.com (密码: %s)", testPass)
		}
	} else {
		log.Printf("   ℹ️  账号已存在，跳过创建")
	}
}

func createBatchUsers() {
	pass := "123456"
	hashedPassword, err := utils.HashPassword(pass)
	if err != nil {
		log.Fatalf("❌ 密码加密失败: %v", err)
	}

	successCount := 0
	existCount := 0

	for i := 1; i <= 4; i++ {
		username := fmt.Sprintf("test%d", i)
		email := fmt.Sprintf("test%d@example.com", i)
		nickname := fmt.Sprintf("测试用户%d", i)

		var count int64
		database.DB.Model(&database.User{}).Where("username = ? OR email = ?", username, email).Count(&count)

		if count == 0 {
			user := database.User{
				Username:      username,
				Email:         email,
				Nickname:      nickname,
				Password:      hashedPassword,
				Avatar:        "🧪",
				IsAdmin:       false,
				Role:          "user",
				Points:        1000,
				MonthlyPoints: 1000,
			}

			if err := database.DB.Create(&user).Error; err != nil {
				log.Printf("   ❌ %s 创建失败: %v", username, err)
			} else {
				log.Printf("   ✅ %s 创建成功 (密码: %s)", username, pass)
				successCount++
			}
		} else {
			log.Printf("   ℹ️  %s 已存在，跳过创建", username)
			existCount++
		}
	}

	log.Printf("\n   📊 创建统计: 成功 %d 个，已存在 %d 个", successCount, existCount)
}
