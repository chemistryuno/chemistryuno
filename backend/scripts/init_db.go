package main

import (
	"chemistryuno/database"
	"chemistryuno/repository"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("🧪 Chemistry UNO - 数据库初始化工具")
	log.Println("------------------------------------")

	// 尝试加载 .env 文件
	// 脚本可能在 backend 目录下运行，也可能在项目根目录下运行
	envPath := ".env"
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		// 尝试上级目录
		envPath = "../.env"
		if _, err := os.Stat(envPath); os.IsNotExist(err) {
			log.Println("⚠️  未找到 .env 文件，将使用默认配置")
		}
	}

	if err := godotenv.Load(envPath); err == nil {
		absPath, _ := filepath.Abs(envPath)
		log.Printf("✅ 已加载配置文件: %s", absPath)
	}

	// 初始化数据库连接
	// InitDB 会自动执行 autoMigrate 和 initDefaultData
	log.Println("🔄 正在初始化数据库连接并执行迁移...")
	if err := database.InitDB(""); err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}

	// 初始化所有仓库（虽然在这个脚本里不一定全用到，但保持一致性）
	repository.InitRepositories()

	log.Println("------------------------------------")
	log.Println("✅ 数据库初始化成功！")
	log.Println("👤 默认管理员: admin / admin123")
	log.Println("🧪 基础物质与化学反应数据已加载")
	log.Println("------------------------------------")
}
