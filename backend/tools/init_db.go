package main

import (
	"chemistryuno/database"
	"log"
	"os"
)

func main() {
	log.Println("🗄️ 正在初始化数据库...")

	// 检查是否是初始化模式
	if os.Getenv("INIT_DB") == "true" {
		log.Println("📦 初始化数据库模式")
	}

	// 初始化数据库
	if err := database.InitDB("./data.db"); err != nil {
		log.Fatal("❌ 数据库初始化失败:", err)
	}

	log.Println("✅ 数据库初始化完成")
	log.Println("📊 数据库文件: ./data.db")
	log.Println("👤 默认管理员账户: admin / admin123")
	log.Println("🧪 默认化学反应数据已创建")
}
