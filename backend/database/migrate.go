package database

import (
	"log"
	"os"
	"strings"
)

// autoMigrate 自动迁移所有表结构
func autoMigrate() error {
	log.Println("开始数据库迁移...")

	// 迁移所有模型
	err := DB.AutoMigrate(
		&User{},
		&UserSession{},
		&GlobalChat{},
		&WebAuthnCredential{},
		&Friendship{},
		&Reaction{},
		&Substance{},
		&Feedback{},
		&DeckConfig{},
		&GameHistory{},
		&Bounty{},
		&Announcement{},
		&SystemConfig{},
		&VerificationCode{},
	)

	if err != nil {
		return err
	}

	log.Println("数据库迁移完成")
	return nil
}

// initDefaultData 初始化默认数据
func initDefaultData() error {
	// 设置 UID 起始值
	setInitialUID()

	// 检查是否已有管理员账户
	var count int64
	DB.Model(&User{}).Where("username = ?", "admin@chemistryuno.com").Count(&count)

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

	// 检查并初始化默认系统配置
	if err := initDefaultConfigs(); err != nil {
		log.Printf("初始化默认系统配置失败: %v", err)
	}

	return nil
}

// initDefaultConfigs 初始化系统配置
func initDefaultConfigs() error {
	configs := []SystemConfig{
		{Key: "game_turn_timeout", Value: "30"},
		{Key: "reconnect_grace_period", Value: "30"},
		{Key: "points_scaling_enabled", Value: "true"},
	}

	for _, cfg := range configs {
		var existing SystemConfig
		err := DB.Where("`key` = ?", cfg.Key).First(&existing).Error
		if err != nil {
			// 如果不存在，则创建
			DB.Create(&cfg)
		}
	}
	return nil
}

// createDefaultAdmin 创建默认管理员账户
func createDefaultAdmin() error {
	// bcrypt hash of "admin123"
	hashedPassword := "$2a$10$BTDLnKl4G7Z26XzUU0VLouw1yxATdub5i2HHj0iVcW0cofNNXkMQe"

	admin := User{
		Username:      "admin@chemistryuno.com",
		Email:         "admin@chemistryuno.com",
		Nickname:      "系统管理员",
		Password:      hashedPassword,
		Avatar:        "⚗️",
		IsAdmin:       true,
		Role:          "admin",
		Points:        1000,
		MonthlyPoints: 1000,
	}

	if err := DB.Create(&admin).Error; err != nil {
		return err
	}

	log.Println("✅ 管理员账户创建成功 (用户名: admin@chemistryuno.com, 邮箱: admin@chemistryuno.com, 密码: admin123)")
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
		Name:         "默认牌组",
		Cards:        []byte(defaultCards),
		InitialCards: 10,
		CreatedByUID: 100000000, // admin用户 (UID起始值为100000000)
		IsGlobal:     true,
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
		{Name: "H2", Formula: "H2", Elements: "H", Description: "氢气", CreatedByUID: 100000000, Status: "approved"},
		{Name: "O2", Formula: "O2", Elements: "O", Description: "氧气", CreatedByUID: 100000000, Status: "approved"},
		{Name: "H2O", Formula: "H2O", Elements: "H,O", Description: "水", CreatedByUID: 100000000, Status: "approved"},
		{Name: "CO2", Formula: "CO2", Elements: "C,O", Description: "二氧化碳", CreatedByUID: 100000000, Status: "approved"},
		{Name: "NaCl", Formula: "NaCl", Elements: "Na,Cl", Description: "氯化钠", CreatedByUID: 100000000, Status: "approved"},
		{Name: "HCl", Formula: "HCl", Elements: "H,Cl", Description: "盐酸", CreatedByUID: 100000000, Status: "approved"},
		{Name: "NaOH", Formula: "NaOH", Elements: "Na,O,H", Description: "氢氧化钠", CreatedByUID: 100000000, Status: "approved"},
		{Name: "H2SO4", Formula: "H2SO4", Elements: "H,S,O", Description: "硫酸", CreatedByUID: 100000000, Status: "approved"},
		{Name: "CaCO3", Formula: "CaCO3", Elements: "Ca,C,O", Description: "碳酸钙", CreatedByUID: 100000000, Status: "approved"},
		{Name: "Fe2O3", Formula: "Fe2O3", Elements: "Fe,O", Description: "氧化铁", CreatedByUID: 100000000, Status: "approved"},
	}

	return DB.Create(&substances).Error
}

// initDefaultReactionsGORM 初始化默认反应数据
func initDefaultReactionsGORM() error {
	// 获取第一个可用的group_id
	groupIDBase := uint(1000)

	reactions := []Reaction{
		// 反应1: H2 + O2 -> H2O (燃烧反应) - 双向排列
		{Reactants: "H2", Products: "O2", Display: "2H2 + O2 → 2H2O", GroupID: &groupIDBase, CreatedByUID: 100000000, Status: "approved"},
		{Reactants: "O2", Products: "H2", Display: "2H2 + O2 → 2H2O", GroupID: &groupIDBase, CreatedByUID: 100000000, Status: "approved"},

		// 反应2: HCl + NaOH -> NaCl + H2O (中和反应)
		{Reactants: "HCl", Products: "NaOH", Display: "HCl + NaOH → NaCl + H2O", GroupID: func() *uint { id := groupIDBase + 1; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{Reactants: "NaOH", Products: "HCl", Display: "HCl + NaOH → NaCl + H2O", GroupID: func() *uint { id := groupIDBase + 1; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应3: CaCO3 + HCl -> CaCl2 + H2O + CO2 (复分解反应)
		{Reactants: "CaCO3", Products: "HCl", Display: "CaCO3 + 2HCl → CaCl2 + H2O + CO2", GroupID: func() *uint { id := groupIDBase + 2; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{Reactants: "HCl", Products: "CaCO3", Display: "CaCO3 + 2HCl → CaCl2 + H2O + CO2", GroupID: func() *uint { id := groupIDBase + 2; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应4: Fe + O2 -> Fe2O3 (氧化反应)
		{Reactants: "Fe", Products: "O2", Display: "4Fe + 3O2 → 2Fe2O3", GroupID: func() *uint { id := groupIDBase + 3; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{Reactants: "O2", Products: "Fe", Display: "4Fe + 3O2 → 2Fe2O3", GroupID: func() *uint { id := groupIDBase + 3; return &id }(), CreatedByUID: 100000000, Status: "approved"},
	}

	return DB.Create(&reactions).Error
}

// setInitialUID 设置用户 UID 的起始值
func setInitialUID() {
	var count int64
	// 使用 Unscoped 以确保即便有软删除的用户也计算在内
	DB.Model(&User{}).Unscoped().Count(&count)
	if count == 0 {
		dbType := strings.ToLower(os.Getenv("DB_TYPE"))
		if dbType == "" {
			dbType = "sqlite"
		}

		if dbType == "mysql" {
			DB.Exec("ALTER TABLE users AUTO_INCREMENT = 100000000")
		} else {
			// SQLite: 设置 sqlite_sequence 表
			// 如果表不存在，设置操作不会报错但也不会生效
			// 所以先尝试插入，再尝试更新
			DB.Exec("INSERT OR IGNORE INTO sqlite_sequence (name, seq) VALUES ('users', 99999999)")
			DB.Exec("UPDATE sqlite_sequence SET seq = 99999999 WHERE name = 'users'")
		}
		log.Println("已将用户 UID 起始值设置为 100000000 模式")
	}
}
