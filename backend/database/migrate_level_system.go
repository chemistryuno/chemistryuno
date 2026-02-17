package database

import (
	"log"
	"math"

	"gorm.io/gorm"
)

// MigrateLevelSystem 迁移等级系统相关表和字段
func MigrateLevelSystem(db *gorm.DB) error {
	log.Println("开始迁移等级系统...")

	// 1. 添加用户等级相关字段
	if err := db.Exec(`
		ALTER TABLE users
		ADD COLUMN IF NOT EXISTS level INTEGER DEFAULT 1,
		ADD COLUMN IF NOT EXISTS xp INTEGER DEFAULT 0,
		ADD COLUMN IF NOT EXISTS total_xp INTEGER DEFAULT 0
	`).Error; err != nil {
		log.Printf("添加用户等级字段失败: %v", err)
		return err
	}

	// 2. 创建等级配置表
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS level_configs (
			level INTEGER PRIMARY KEY,
			required_xp INTEGER NOT NULL,
			tier VARCHAR(20) NOT NULL,
			tier_name VARCHAR(50) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		log.Printf("创建等级配置表失败: %v", err)
		return err
	}

	// 3. 添加游戏历史经验字段
	if err := db.Exec(`
		ALTER TABLE game_history
		ADD COLUMN IF NOT EXISTS xp_rewards JSON
	`).Error; err != nil {
		log.Printf("添加游戏历史经验字段失败: %v", err)
		return err
	}

	// 4. 初始化等级配置数据（1-100级）
	if err := initializeLevelConfigs(db); err != nil {
		log.Printf("初始化等级配置失败: %v", err)
		return err
	}

	// 5. 修复历史数据：确保所有用户等级至少为1
	if err := db.Exec("UPDATE users SET level = 1 WHERE level < 1 OR level IS NULL").Error; err != nil {
		log.Printf("修复用户等级数据失败: %v", err)
		// 不返回错误，因为这可能是非致命的
	}

	log.Println("等级系统迁移完成")
	return nil
}

// initializeLevelConfigs 初始化等级配置数据
func initializeLevelConfigs(db *gorm.DB) error {
	// 检查是否已有配置
	var count int64
	if err := db.Table("level_configs").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		log.Println("等级配置已存在，跳过初始化")
		return nil
	}

	log.Println("初始化等级配置数据...")

	// 等级段定义
	tiers := []struct {
		minLevel int
		maxLevel int
		tier     string
		tierName string
	}{
		{1, 10, "bronze", "青铜"},
		{11, 25, "silver", "白银"},
		{26, 45, "gold", "黄金"},
		{46, 70, "platinum", "铂金"},
		{71, 90, "diamond", "钻石"},
		{91, 100, "master", "大师"},
	}

	// 经验计算参数（增强版 - 升级更难）
	const baseXP = 100.0         // 基础经验
	const growthRate = 0.12      // 增长系数（每级增长12%，原8%）
	const scaleFactor = 2.0      // 高等级缩放因子（原1.5，现2.0更陡峭）

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, tierInfo := range tiers {
		for level := tierInfo.minLevel; level <= tierInfo.maxLevel; level++ {
			// 经验公式：基础经验 * (1 + (等级-1) * 增长系数) ^ 缩放因子
			// 随着等级升高，所需经验呈指数增长
			requiredXP := int(baseXP * math.Pow(1+float64(level-1)*growthRate, scaleFactor))

			// 插入配置
			if err := tx.Exec(`
				INSERT INTO level_configs (level, required_xp, tier, tier_name)
				VALUES (?, ?, ?, ?)
			`, level, requiredXP, tierInfo.tier, tierInfo.tierName).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

// LevelConfig GORM模型 - 等级配置表
type LevelConfig struct {
	Level      int    `gorm:"primaryKey" json:"level"`
	RequiredXP int    `gorm:"not null" json:"required_xp"`
	Tier       string `gorm:"size:20;not null" json:"tier"`
	TierName   string `gorm:"size:50;not null" json:"tier_name"`
}

func (LevelConfig) TableName() string {
	return "level_configs"
}
