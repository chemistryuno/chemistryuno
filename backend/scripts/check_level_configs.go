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

	// 查询等级配置数量
	var count int64
	if err := database.DB.Table("level_configs").Count(&count).Error; err != nil {
		log.Fatal("查询等级配置失败:", err)
	}

	log.Printf("数据库中的等级配置数量: %d\n", count)

	// 查询几个示例配置
	var configs []database.LevelConfig
	if err := database.DB.Limit(5).Find(&configs).Error; err != nil {
		log.Fatal("查询等级配置失败:", err)
	}

	log.Println("前5个等级配置:")
	for _, config := range configs {
		log.Printf("  等级 %d: %s(%s), 所需经验: %d\n",
			config.Level, config.TierName, config.Tier, config.RequiredXP)
	}
}
