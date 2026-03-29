package main

import (
	"chemistryuno/backend/database"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func ensureProjectRoot() {
	if _, err := os.Stat("backend"); err == nil {
		return
	}
	if _, err := os.Stat("../backend"); err == nil {
		_ = os.Chdir("..")
	}
}

func main() {
	ensureProjectRoot()

	log.Println("🧪 Chemistry UNO - 数据库初始化工具")
	log.Println("------------------------------------")

	if err := godotenv.Load("backend/.env"); err != nil {
		_ = godotenv.Load(".env")
	}

	log.Println("🔄 正在初始化数据库连接并执行迁移...")
	if err := database.InitDB(""); err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}

	log.Println("------------------------------------")
	log.Println("✅ 数据库初始化成功！")
	log.Println("👤 默认管理员: admin@chemistryuno.com / admin123")
	log.Println("🧪 基础物质与化学反应数据已加载")
	log.Println("⚙️  系统参数 (重连宽限、回合时间) 已初始化")
	log.Println("------------------------------------")
}
