//go:build scripts
// +build scripts

package main

import (
	"chemistryuno/backend/database"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("警告: 无法加载 .env 文件: %v\n", err)
	}

	// 初始化数据库
	if err := database.InitDB(""); err != nil {
		log.Fatal("数据库初始化失败:", err)
	}
	defer database.Close()

	// 强制重新初始化等级配置
	log.Println("开始强制重新初始化等级配置...")

	// 清空现有配置
	if err := database.DB.Exec("DELETE FROM level_configs").Error; err != nil {
		log.Fatal("清空等级配置失败:", err)
	}

	// 重新运行迁移
	if err := database.MigrateLevelSystem(database.DB); err != nil {
		log.Fatal("等级系统迁移失败:", err)
	}

	// 验证结果
	var count int64
	if err := database.DB.Table("level_configs").Count(&count).Error; err != nil {
		log.Fatal("查询等级配置失败:", err)
	}

	log.Printf("✅ 成功初始化 %d 个等级配置\n", count)
}
