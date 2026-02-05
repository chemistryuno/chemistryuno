package main

import (
	"chemistryuno/database"
	"chemistryuno/repository"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("🧪 Chemistry UNO - 数据库初始化工具")
	log.Println("------------------------------------")

	// 尝试加载 .env 文件 (从 tools 运行通常在项目根目录运行或 tools 目录下)
	if err := godotenv.Load("backend/.env"); err != nil {
		godotenv.Load("../backend/.env")
	}

	// 初始化数据库连接
	log.Println("🔄 正在初始化数据库连接并执行迁移...")
	if err := database.InitDB(""); err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}

	// 初始化所有仓库
	repository.InitRepositories()

	log.Println("------------------------------------")
	log.Println("✅ 数据库初始化成功！")
	log.Println("👤 默认管理员: admin / admin@chemistryuno.com / admin123")
	log.Println("🧪 基础物质与化学反应数据已加载")
	log.Println("------------------------------------")
}
