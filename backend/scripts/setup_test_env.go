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
	log.Println("🧪 Chemistry UNO - 测试环境初始化")
	log.Println("------------------------------------")

	// 尝试加载 .env 文件
	envPath := ".env"
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		envPath = "../.env"
		if _, err := os.Stat(envPath); os.IsNotExist(err) {
			log.Println("⚠️  未找到 .env 文件，将使用默认配置")
		}
	}

	if err := godotenv.Load(envPath); err == nil {
		absPath, _ := filepath.Abs(envPath)
		log.Printf("✅ 已加载配置文件: %s", absPath)
	}

	// 1. 初始化数据库及迁移
	log.Println("🔄 正在初始化数据库并执行迁移...")
	if err := database.InitDB(""); err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}

	// 初始化仓库
	repository.InitRepositories()

	// 2. 创建管理员测试账号
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
			log.Printf("⚠️  创建管理员账号失败: %v", err)
		} else {
			log.Printf("✅ 已创建管理员账号: admin@chemistryuno.com (Password: %s)", adminPass)
		}
	} else {
		log.Printf("ℹ️  管理员账号已存在")
	}

	// 3. 创建普通测试账号 test@example.com
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
			log.Printf("⚠️  创建测试账号 test 失败: %v", err)
		} else {
			log.Printf("✅ 已创建测试账号: test@example.com (Password: %s)", testPass)
		}
	} else {
		log.Printf("ℹ️  测试账号 test@example.com 已存在")
	}

	// 4. 创建测试账号 test1, test2, test3, test4
	pass := "123456"
	hashedPassword, err := utils.HashPassword(pass)
	if err != nil {
		log.Fatalf("❌ 密码加密失败: %v", err)
	}

	for i := 1; i <= 4; i++ {
		username := fmt.Sprintf("test%d", i)
		email := fmt.Sprintf("test%d", i) // 邮箱字段简化为 test1, test2, test3, test4
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
				log.Printf("⚠️  创建测试账号 %s 失败: %v", username, err)
			} else {
				log.Printf("✅ 已创建测试账号: %s (Password: %s)", username, pass)
			}
		} else {
			log.Printf("ℹ️  测试账号 %s 已存在", username)
		}
	}

	log.Println("------------------------------------")
	log.Println("✅ 测试环境准备完毕！")
	log.Println("------------------------------------")
}
