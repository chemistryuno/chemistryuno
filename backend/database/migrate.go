package database

import (
	"log"
)

// autoMigrate 自动迁移所有表结构
func autoMigrate() error {
	log.Println("开始数据库迁移...")

	// 迁移所有模型
	err := DB.AutoMigrate(
		&User{},
		&UserSession{},
		&WebAuthnCredential{},
		&Reaction{},
		&Substance{},
		&Feedback{},
		&DeckConfig{},
		&GameHistory{},
		&Bounty{},
		&Announcement{},
		&SystemConfig{},
	)

	if err != nil {
		return err
	}

	log.Println("数据库迁移完成")
	return nil
}

// initDefaultData 初始化默认数据
func initDefaultData() error {
	// 检查是否已有管理员账户
	var count int64
	DB.Model(&User{}).Where("username = ?", "admin").Count(&count)

	if count == 0 {
		// 创建默认管理员账户
		if err := createDefaultAdmin(); err != nil {
			return err
		}
	}

	// 检查是否已有全局牌组配置
	DB.Model(&DeckConfig{}).Where("is_global = ?", true).Count(&count)

	if count == 0 {
		// 创建默认全局牌组
		if err := createDefaultDeckConfig(); err != nil {
			return err
		}
	}

	// 初始化默认物质数据（如果需要）
	DB.Model(&Substance{}).Count(&count)
	if count == 0 {
		if err := initDefaultSubstancesGORM(); err != nil {
			log.Printf("初始化默认物质数据失败: %v", err)
		}
	}

	// 初始化默认反应数据（如果需要）
	DB.Model(&Reaction{}).Count(&count)
	if count == 0 {
		if err := initDefaultReactionsGORM(); err != nil {
			log.Printf("初始化默认反应数据失败: %v", err)
		}
	}

	return nil
}

// createDefaultAdmin 创建默认管理员账户
func createDefaultAdmin() error {
	// bcrypt hash of "admin123"
	hashedPassword := "$2a$10$BTDLnKl4G7Z26XzUU0VLouw1yxATdub5i2HHj0iVcW0cofNNXkMQe"

	admin := User{
		Username: "admin",
		Password: hashedPassword,
		Avatar:   "⚗️",
		IsAdmin:  true,
		Role:     "admin",
	}

	if err := DB.Create(&admin).Error; err != nil {
		return err
	}

	log.Println("✅ 管理员账户创建成功 (用户名: admin, 密码: admin123)")
	return nil
}

// createDefaultDeckConfig 创建默认全局牌组配置
func createDefaultDeckConfig() error {
	defaultCards := `{
		"H": 12, "O": 12,
		"C": 4, "N": 4, "F": 4, "Na": 4, "Mg": 4, "Al": 4,
		"Si": 4, "P": 4, "S": 4, "Cl": 4, "K": 4, "Ca": 4,
		"Mn": 4, "Fe": 4, "Cu": 4, "Zn": 4, "Br": 4, "I": 4, "Ag": 4,
		"+2": 8, "+4": 4,
		"He": 1, "Ne": 1, "Ar": 1, "Kr": 1,
		"Au": 4
	}`

	deckConfig := DeckConfig{
		Name:      "默认牌组",
		Cards:     defaultCards,
		CreatedBy: 1, // admin用户
		IsGlobal:  true,
	}

	if err := DB.Create(&deckConfig).Error; err != nil {
		return err
	}

	log.Println("✅ 默认牌组配置创建成功")
	return nil
}

// initDefaultSubstancesGORM 初始化默认物质数据
func initDefaultSubstancesGORM() error {
	substances := []Substance{
		{Name: "H2", Description: "氢气", CreatedBy: 1, Status: "approved"},
		{Name: "O2", Description: "氧气", CreatedBy: 1, Status: "approved"},
		{Name: "H2O", Description: "水", CreatedBy: 1, Status: "approved"},
		{Name: "CO2", Description: "二氧化碳", CreatedBy: 1, Status: "approved"},
		{Name: "NaCl", Description: "氯化钠", CreatedBy: 1, Status: "approved"},
	}

	return DB.Create(&substances).Error
}

// initDefaultReactionsGORM 初始化默认反应数据
func initDefaultReactionsGORM() error {
	reactions := []Reaction{
		{Reactants: "H2,O2", Products: "H2O", CreatedBy: 1, Status: "approved", Bidirection: false},
		{Reactants: "C,O2", Products: "CO2", CreatedBy: 1, Status: "approved", Bidirection: false},
		{Reactants: "Na,Cl2", Products: "NaCl", CreatedBy: 1, Status: "approved", Bidirection: false},
	}

	return DB.Create(&reactions).Error
}
