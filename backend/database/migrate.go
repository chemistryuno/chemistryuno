package database

import (
	"log"
	"os"
	"strings"
)

// autoMigrate 自动迁移所有表结构
func autoMigrate() error {
	log.Println("🔄 开始数据库迁移和初始化...")

	// 检查是否为首次初始化（User 表是否为空）
	var userCount int64
	DB.Model(&User{}).Count(&userCount)
	isFirstInit := userCount == 0

	if isFirstInit {
		log.Println("📊 检测到首次启动，正在初始化数据库表结构...")
	}

	// 迁移所有模型
	err := DB.AutoMigrate(
		&User{},
		&UserSession{},
		&GlobalChat{},
		&PrivateChat{},
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

	if isFirstInit {
		log.Println("✅ 数据库表结构初始化成功")
	}

	// 执行数据迁移逻辑 (特别是针对 Reaction 表的 R1/R2 结构升级)
	if err := MigrateReactionsToR1R2(); err != nil {
		log.Printf("⚠️  Reaction 数据迁移失败: %v", err)
	}

	// 修复可能存在的 R1/R2 顺序问题
	if err := fixReactionOrdering(); err != nil {
		log.Printf("⚠️  修复反应顺序失败: %v", err)
	}

	log.Println("✅ 数据库迁移完成")
	return nil
}

// fixReactionOrdering 确保数据库中所有反应都遵循 Canonical ordering (r1 <= r2)
func fixReactionOrdering() error {
	var reactions []Reaction
	// 查找所有 r1 > r2 的记录
	err := DB.Where("r1 > r2").Find(&reactions).Error
	if err != nil {
		return err
	}

	if len(reactions) == 0 {
		return nil
	}

	log.Printf("发现 %d 条顺序不正确的反应记录，正在修复...", len(reactions))

	for _, r := range reactions {
		err := DB.Model(&r).Updates(map[string]interface{}{
			"r1": r.R2,
			"r2": r.R1,
		}).Error
		if err != nil {
			log.Printf("修复记录 %d 失败: %v", r.ID, err)
		}
	}

	log.Println("反应顺序修复完成")
	return nil
}

// initDefaultData 初始化默认数据
func initDefaultData() error {
	log.Println("🔧 初始化默认数据...")

	// 设置 UID 起始值
	setInitialUID()

	// 检查是否已有管理员账户
	var count int64
	DB.Model(&User{}).Where("email = ? OR username = ? OR is_admin = ?", "admin@chemistryuno.com", "admin", true).Count(&count)

	if count == 0 {
		log.Println("👤 创建默认管理员账户...")
		// 创建默认管理员账户
		if err := createDefaultAdmin(); err != nil {
			return err
		}
		log.Println("✅ 默认管理员账户创建成功 (admin@chemistryuno.com / 123456)")
	}

	// 检查是否已有全局牌组配置
	DB.Model(&DeckConfig{}).Where("is_global = ?", true).Count(&count)

	if count == 0 {
		log.Println("🃏 创建默认全局牌组配置...")
		// 创建默认全局牌组
		if err := createDefaultDeckConfig(); err != nil {
			return err
		}
		log.Println("✅ 默认牌组配置创建成功")
	}

	// 初始化默认物质数据（如果需要）
	DB.Model(&Substance{}).Count(&count)
	if count == 0 {
		log.Println("⚗️  初始化默认物质数据...")
		if err := initDefaultSubstancesGORM(); err != nil {
			log.Printf("⚠️  初始化默认物质数据失败: %v", err)
		} else {
			log.Println("✅ 默认物质数据初始化成功")
		}
	}

	// 初始化默认反应数据（如果需要）
	DB.Model(&Reaction{}).Count(&count)
	if count == 0 {
		log.Println("🧪 初始化默认化学反应数据...")
		if err := initDefaultReactionsGORM(); err != nil {
			log.Printf("⚠️  初始化默认反应数据失败: %v", err)
		} else {
			log.Println("✅ 默认反应数据初始化成功")
		}
	}

	// 检查并初始化默认系统配置
	if err := initDefaultConfigs(); err != nil {
		log.Printf("⚠️  初始化默认系统配置失败: %v", err)
	}

	// 初始化默认提示数据
	if err := initDefaultHints(); err != nil {
		log.Printf("⚠️  初始化默认提示数据失败: %v", err)
	}

	log.Println("✅ 默认数据初始化完成")

	return nil
}

// initDefaultHints 初始化实验情报提示
func initDefaultHints() error {
	var count int64
	DB.Model(&Announcement{}).Where("type = ?", "hint").Count(&count)
	if count > 0 {
		return nil
	}

	hints := []Announcement{
		{Title: "合成技巧", Content: "在 Uno 中，你可以通过合理的卡牌组合来通过化学方程式合成更高级的物质。", Type: "hint", Active: true},
		{Title: "稀有元素", Content: "金(Au)是极其稀有的元素，如果能成功合成金相关的化合物，通常会获得高额分数或成就。", Type: "hint", Active: true},
		{Title: "加牌机制", Content: "使用 +2 或 +4 类型的反应卡可以强迫下一位研究员摸取对应的卡牌，除非他能打出另一张叠加卡。", Type: "hint", Active: true},
		{Title: "双重反应", Content: "在某些模式下，你可以一次打出两张底物来尝试更复杂的双重置换反应。", Type: "hint", Active: true},
		{Title: "非法操作", Content: "尝试不符合真实化学逻辑的反应会导致爆炸并强制罚牌，请务必保证你的实验计划符合科学规律。", Type: "hint", Active: true},
	}

	for _, hint := range hints {
		if err := DB.Create(&hint).Error; err != nil {
			log.Printf("创建默认提示失败: %v", err)
		}
	}
	log.Println("✅ 默认实验情报提示初始化成功")
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
		// 基础单质和常见物质
		{Name: "水", Formula: "H2O", Elements: "H,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "二氧化碳", Formula: "CO2", Elements: "C,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "盐酸", Formula: "HCl", Elements: "H,Cl", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氢氧化钠", Formula: "NaOH", Elements: "Na,O,H", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氯化钠", Formula: "NaCl", Elements: "Na,Cl", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫酸", Formula: "H2SO4", Elements: "H,S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氧气", Formula: "O2", Elements: "O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氢气", Formula: "H2", Elements: "H", CreatedByUID: 100000000, Status: "approved"},
		{Name: "铁", Formula: "Fe", Elements: "Fe", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫酸铜", Formula: "CuSO4", Elements: "Cu,S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "铜", Formula: "Cu", Elements: "Cu", CreatedByUID: 100000000, Status: "approved"},
		{Name: "锌", Formula: "Zn", Elements: "Zn", CreatedByUID: 100000000, Status: "approved"},
		{Name: "一氧化碳", Formula: "CO", Elements: "C,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "碳酸钙", Formula: "CaCO3", Elements: "Ca,C,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氧化钙", Formula: "CaO", Elements: "Ca,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氢氧化钙", Formula: "Ca(OH)2", Elements: "Ca,O,H", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氨气", Formula: "NH3", Elements: "N,H", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硝酸", Formula: "HNO3", Elements: "H,N,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硝酸银", Formula: "AgNO3", Elements: "Ag,N,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氯化银", Formula: "AgCl", Elements: "Ag,Cl", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氧化镁", Formula: "MgO", Elements: "Mg,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "镁", Formula: "Mg", Elements: "Mg", CreatedByUID: 100000000, Status: "approved"},
		{Name: "铝", Formula: "Al", Elements: "Al", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氧化铝", Formula: "Al2O3", Elements: "Al,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氧化铁", Formula: "Fe2O3", Elements: "Fe,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "四氧化三铁", Formula: "Fe3O4", Elements: "Fe,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "碳酸钠", Formula: "Na2CO3", Elements: "Na,C,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "碳酸氢钠", Formula: "NaHCO3", Elements: "Na,H,C,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氯酸钾", Formula: "KClO3", Elements: "K,Cl,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氯化钾", Formula: "KCl", Elements: "K,Cl", CreatedByUID: 100000000, Status: "approved"},
		{Name: "过氧化氢", Formula: "H2O2", Elements: "H,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "二氧化硫", Formula: "SO2", Elements: "S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氯化钡", Formula: "BaCl2", Elements: "Ba,Cl", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫酸钡", Formula: "BaSO4", Elements: "Ba,S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "碳", Formula: "C", Elements: "C", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫", Formula: "S", Elements: "S", CreatedByUID: 100000000, Status: "approved"},
		{Name: "磷", Formula: "P", Elements: "P", CreatedByUID: 100000000, Status: "approved"},
		{Name: "五氧化二磷", Formula: "P2O5", Elements: "P,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氧化铜", Formula: "CuO", Elements: "Cu,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氯化铁", Formula: "FeCl3", Elements: "Fe,Cl", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫酸亚铁", Formula: "FeSO4", Elements: "Fe,S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氢氧化钾", Formula: "KOH", Elements: "K,O,H", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氯化镁", Formula: "MgCl2", Elements: "Mg,Cl", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氯化钙", Formula: "CaCl2", Elements: "Ca,Cl", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氯化铵", Formula: "NH4Cl", Elements: "N,H,Cl", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫酸钠", Formula: "Na2SO4", Elements: "Na,S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "一氧化氮", Formula: "NO", Elements: "N,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "二氧化氮", Formula: "NO2", Elements: "N,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氮气", Formula: "N2", Elements: "N", CreatedByUID: 100000000, Status: "approved"},
		{Name: "汞", Formula: "Hg", Elements: "Hg", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氧化汞", Formula: "HgO", Elements: "Hg,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氧化钠", Formula: "Na2O", Elements: "Na,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氧化钾", Formula: "K2O", Elements: "K,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "三氧化硫", Formula: "SO3", Elements: "S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氧化钡", Formula: "BaO", Elements: "Ba,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氯化铜", Formula: "CuCl2", Elements: "Cu,Cl", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氢氧化铜", Formula: "Cu(OH)2", Elements: "Cu,O,H", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氢氧化铁", Formula: "Fe(OH)3", Elements: "Fe,O,H", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硝酸铵", Formula: "NH4NO3", Elements: "N,H,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氯化锌", Formula: "ZnCl2", Elements: "Zn,Cl", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫酸锌", Formula: "ZnSO4", Elements: "Zn,S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氯化铝", Formula: "AlCl3", Elements: "Al,Cl", CreatedByUID: 100000000, Status: "approved"},
		{Name: "银", Formula: "Ag", Elements: "Ag", CreatedByUID: 100000000, Status: "approved"},
		{Name: "磷酸", Formula: "H3PO4", Elements: "H,P,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "磷酸钠", Formula: "Na3PO4", Elements: "Na,P,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氯气", Formula: "Cl2", Elements: "Cl", CreatedByUID: 100000000, Status: "approved"},
		{Name: "溴单质", Formula: "Br2", Elements: "Br", CreatedByUID: 100000000, Status: "approved"},
		{Name: "碘单质", Formula: "I2", Elements: "I", CreatedByUID: 100000000, Status: "approved"},
		{Name: "碘化氢", Formula: "HI", Elements: "H,I", CreatedByUID: 100000000, Status: "approved"},
		{Name: "溴化氢", Formula: "HBr", Elements: "H,Br", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫酸钾", Formula: "K2SO4", Elements: "K,S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫酸镁", Formula: "MgSO4", Elements: "Mg,S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫酸钙", Formula: "CaSO4", Elements: "Ca,S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "过氧化钠", Formula: "Na2O2", Elements: "Na,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "钠", Formula: "Na", Elements: "Na", CreatedByUID: 100000000, Status: "approved"},
		{Name: "钾", Formula: "K", Elements: "K", CreatedByUID: 100000000, Status: "approved"},
		{Name: "钙", Formula: "Ca", Elements: "Ca", CreatedByUID: 100000000, Status: "approved"},
		{Name: "偏铝酸钠", Formula: "NaAlO2", Elements: "Na,Al,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "碳酸钾", Formula: "K2CO3", Elements: "K,C,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氢化钠", Formula: "NaH", Elements: "Na,H", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氢化钾", Formula: "KH", Elements: "K,H", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氢化镁", Formula: "MgH2", Elements: "Mg,H", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氢化钡", Formula: "BaH2", Elements: "Ba,H", CreatedByUID: 100000000, Status: "approved"},
		{Name: "超氧化钾", Formula: "KO2", Elements: "K,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氢化钙", Formula: "CaH2", Elements: "Ca,H", CreatedByUID: 100000000, Status: "approved"},
		{Name: "亚硫酸钠", Formula: "Na2SO3", Elements: "Na,S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "亚硫酸钾", Formula: "K2SO3", Elements: "K,S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫酸氢钠", Formula: "NaHSO4", Elements: "Na,H,S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "亚硫酸氢钠", Formula: "NaHSO3", Elements: "Na,H,S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "亚硫酸钙", Formula: "CaSO3", Elements: "Ca,S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "碳酸钡", Formula: "BaCO3", Elements: "Ba,C,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "溴化钠", Formula: "NaBr", Elements: "Na,Br", CreatedByUID: 100000000, Status: "approved"},
		{Name: "碘化钠", Formula: "NaI", Elements: "Na,I", CreatedByUID: 100000000, Status: "approved"},
		{Name: "溴化钾", Formula: "KBr", Elements: "K,Br", CreatedByUID: 100000000, Status: "approved"},
		{Name: "溴化银", Formula: "AgBr", Elements: "Ag,Br", CreatedByUID: 100000000, Status: "approved"},
		{Name: "碘化银", Formula: "AgI", Elements: "Ag,I", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氟气", Formula: "F2", Elements: "F", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氢氟酸", Formula: "HF", Elements: "H,F", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氟化钠", Formula: "NaF", Elements: "Na,F", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氟化钙", Formula: "CaF2", Elements: "Ca,F", CreatedByUID: 100000000, Status: "approved"},
		{Name: "次氯酸", Formula: "HClO", Elements: "H,Cl,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "次氯酸钠", Formula: "NaClO", Elements: "Na,Cl,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫化氢", Formula: "H2S", Elements: "H,S", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫化钠", Formula: "Na2S", Elements: "Na,S", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫化亚铁", Formula: "FeS", Elements: "Fe,S", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫化铜", Formula: "CuS", Elements: "Cu,S", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫化锌", Formula: "ZnS", Elements: "Zn,S", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫化银", Formula: "Ag2S", Elements: "Ag,S", CreatedByUID: 100000000, Status: "approved"},
		{Name: "亚硫酸", Formula: "H2SO3", Elements: "H,S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "亚硫酸钡", Formula: "BaSO3", Elements: "Ba,S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "溴化镁", Formula: "MgBr2", Elements: "Mg,Br", CreatedByUID: 100000000, Status: "approved"},
		{Name: "碘化镁", Formula: "MgI2", Elements: "Mg,I", CreatedByUID: 100000000, Status: "approved"},
		{Name: "溴化铝", Formula: "AlBr3", Elements: "Al,Br", CreatedByUID: 100000000, Status: "approved"},
		{Name: "碘化铝", Formula: "AlI3", Elements: "Al,I", CreatedByUID: 100000000, Status: "approved"},
		{Name: "溴化锌", Formula: "ZnBr2", Elements: "Zn,Br", CreatedByUID: 100000000, Status: "approved"},
		{Name: "碘化锌", Formula: "ZnI2", Elements: "Zn,I", CreatedByUID: 100000000, Status: "approved"},
		{Name: "溴化铜", Formula: "CuBr2", Elements: "Cu,Br", CreatedByUID: 100000000, Status: "approved"},
		{Name: "溴化铁", Formula: "FeBr3", Elements: "Fe,Br", CreatedByUID: 100000000, Status: "approved"},
		{Name: "碘化亚铁", Formula: "FeI2", Elements: "Fe,I", CreatedByUID: 100000000, Status: "approved"},
		{Name: "溴化钙", Formula: "CaBr2", Elements: "Ca,Br", CreatedByUID: 100000000, Status: "approved"},
		{Name: "碘化钙", Formula: "CaI2", Elements: "Ca,I", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氟化钾", Formula: "KF", Elements: "K,F", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氟化钡", Formula: "BaF2", Elements: "Ba,F", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氟化镁", Formula: "MgF2", Elements: "Mg,F", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氟化铝", Formula: "AlF3", Elements: "Al,F", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氟化铁", Formula: "FeF3", Elements: "Fe,F", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氟化铜", Formula: "CuF2", Elements: "Cu,F", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氟化银", Formula: "AgF", Elements: "Ag,F", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氟化汞", Formula: "HgF2", Elements: "Hg,F", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氟化锌", Formula: "ZnF2", Elements: "Zn,F", CreatedByUID: 100000000, Status: "approved"},
		{Name: "二氧化硅", Formula: "SiO2", Elements: "Si,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "四氟化硅", Formula: "SiF4", Elements: "Si,F", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氧化锌", Formula: "ZnO", Elements: "Zn,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氧化亚铜", Formula: "Cu2O", Elements: "Cu,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氧化亚铁", Formula: "FeO", Elements: "Fe,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氧化银", Formula: "Ag2O", Elements: "Ag,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "一氧化二氮", Formula: "N2O", Elements: "N,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "七氧化二氯", Formula: "Cl2O7", Elements: "Cl,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫化钾", Formula: "K2S", Elements: "K,S", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫化镁", Formula: "MgS", Elements: "Mg,S", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫化钙", Formula: "CaS", Elements: "Ca,S", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫化钡", Formula: "BaS", Elements: "Ba,S", CreatedByUID: 100000000, Status: "approved"},
		{Name: "亚硫酸镁", Formula: "MgSO3", Elements: "Mg,S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫化铝", Formula: "Al2S3", Elements: "Al,S", CreatedByUID: 100000000, Status: "approved"},
		{Name: "亚硫酸铝", Formula: "Al2(SO3)3", Elements: "Al,S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "亚硫酸亚铁", Formula: "FeSO3", Elements: "Fe,S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "碘化钾", Formula: "KI", Elements: "K,I", CreatedByUID: 100000000, Status: "approved"},
		{Name: "溴化钡", Formula: "BaBr2", Elements: "Ba,Br", CreatedByUID: 100000000, Status: "approved"},
		{Name: "碘化钡", Formula: "BaI2", Elements: "Ba,I", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氯化亚铁", Formula: "FeCl2", Elements: "Fe,Cl", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫酸铝", Formula: "Al2(SO4)3", Elements: "Al,S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氢氧化钡", Formula: "Ba(OH)2", Elements: "Ba,O,H", CreatedByUID: 100000000, Status: "approved"},
		{Name: "钡", Formula: "Ba", Elements: "Ba", CreatedByUID: 100000000, Status: "approved"},
		{Name: "碳酸氢钾", Formula: "KHCO3", Elements: "K,H,C,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "碳酸氢钙", Formula: "Ca(HCO3)2", Elements: "Ca,H,C,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氢氧化镁", Formula: "Mg(OH)2", Elements: "Mg,O,H", CreatedByUID: 100000000, Status: "approved"},
		{Name: "三氯化磷", Formula: "PCl3", Elements: "P,Cl", CreatedByUID: 100000000, Status: "approved"},
		{Name: "五硫化二磷", Formula: "P2S5", Elements: "P,S", CreatedByUID: 100000000, Status: "approved"},
		{Name: "五氟化磷", Formula: "PF5", Elements: "P,F", CreatedByUID: 100000000, Status: "approved"},
		{Name: "四氟化碳", Formula: "CF4", Elements: "C,F", CreatedByUID: 100000000, Status: "approved"},
		{Name: "六氟化硫", Formula: "SF6", Elements: "S,F", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫代硫酸钠", Formula: "Na2S2O3", Elements: "Na,S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硫代硫酸钾", Formula: "K2S2O3", Elements: "K,S,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氧化亚汞", Formula: "Hg2O", Elements: "Hg,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氯化汞", Formula: "HgCl2", Elements: "Hg,Cl", CreatedByUID: 100000000, Status: "approved"},
		{Name: "溴化汞", Formula: "HgBr2", Elements: "Hg,Br", CreatedByUID: 100000000, Status: "approved"},
		{Name: "碘化汞", Formula: "HgI2", Elements: "Hg,I", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硝酸汞", Formula: "Hg(NO3)2", Elements: "Hg,N,O", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氢氧化铝", Formula: "Al(OH)3", Elements: "Al,O,H", CreatedByUID: 100000000, Status: "approved"},
		{Name: "氢氧化锌", Formula: "Zn(OH)2", Elements: "Zn,O,H", CreatedByUID: 100000000, Status: "approved"},
		{Name: "硝酸钾", Formula: "KNO3", Elements: "K,N,O", CreatedByUID: 100000000, Status: "approved"},
	}

	log.Printf("开始批量导入 %d 种化学物质...", len(substances))

	// 使用分批插入以避免单次插入过多数据
	batchSize := 50
	for i := 0; i < len(substances); i += batchSize {
		end := i + batchSize
		if end > len(substances) {
			end = len(substances)
		}
		batch := substances[i:end]
		if err := DB.Create(&batch).Error; err != nil {
			log.Printf("批次 %d-%d 插入失败: %v", i, end, err)
			return err
		}
		log.Printf("已导入 %d/%d 种物质", end, len(substances))
	}

	log.Printf("✅ 成功导入 %d 种化学物质", len(substances))
	return nil
}

// initDefaultReactionsGORM 初始化默认反应数据
func initDefaultReactionsGORM() error {
	// 获取第一个可用的group_id
	groupIDBase := uint(1000)

	// 统计添加的反应数量
	reactionCount := 0

	reactions := []Reaction{
		{R1: "H2", R2: "O2", Display: "2H₂ + O₂ = 2H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C", R2: "O2", Display: "C + O₂ = CO₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "O2", R2: "S", Display: "S + O₂ = SO₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "O2", R2: "P", Display: "4P + 5O₂ = 2P₂O₅", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe", R2: "O2", Display: "3Fe + 2O₂ = Fe₃O₄", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Mg", R2: "O2", Display: "2Mg + O₂ = 2MgO", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CO", R2: "O2", Display: "2CO + O₂ = 2CO₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "NaOH", Display: "HCl + NaOH = NaCl + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2SO4", R2: "NaOH", Display: "H₂SO₄ + 2NaOH = Na₂SO₄ + 2H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HNO3", R2: "NaOH", Display: "HNO₃ + NaOH = NaNO₃ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca(OH)2", R2: "HCl", Display: "2HCl + Ca(OH)₂ = CaCl₂ + 2H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "BaCl2", R2: "H2SO4", Display: "H₂SO₄ + BaCl₂ = BaSO₄↓ + 2HCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "AgNO3", R2: "HCl", Display: "HCl + AgNO₃ = AgCl↓ + HNO₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "AgNO3", R2: "NaCl", Display: "NaCl + AgNO₃ = AgCl↓ + NaNO₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe", R2: "HCl", Display: "Fe + 2HCl = FeCl₂ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "Zn", Display: "Zn + 2HCl = ZnCl₂ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "Mg", Display: "Mg + 2HCl = MgCl₂ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "HCl", Display: "2Al + 6HCl = 2AlCl₃ + 3H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CuSO4", R2: "Fe", Display: "Fe + CuSO₄ = FeSO₄ + Cu", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CuSO4", R2: "Zn", Display: "Zn + CuSO₄ = ZnSO₄ + Cu", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CO2", R2: "Ca(OH)2", Display: "CO₂ + Ca(OH)₂ = CaCO₃↓ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CO2", R2: "NaOH", Display: "CO₂ + 2NaOH = Na₂CO₃ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CO2", R2: "H2O", Display: "CO₂ + H₂O = H₂CO₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CaO", R2: "H2O", Display: "CaO + H₂O = Ca(OH)₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O", R2: "MgO", Display: "MgO + H₂O = Mg(OH)₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CaCO3", R2: "HCl", Display: "CaCO₃ + 2HCl = CaCl₂ + H₂O + CO₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "Na2CO3", Display: "Na₂CO₃ + 2HCl = 2NaCl + H₂O + CO₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "NaHCO3", Display: "NaHCO₃ + HCl = NaCl + H₂O + CO₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "NH3", Display: "NH₃ + HCl = NH₄Cl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "AgNO3", R2: "Cu", Display: "Cu + 2AgNO₃ = Cu(NO₃)₂ + 2Ag", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CuO", R2: "H2", Display: "CuO + H₂ = Cu + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CO", R2: "Fe2O3", Display: "Fe₂O₃ + 3CO = 2Fe + 3CO₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C", R2: "CuO", Display: "2CuO + C = 2Cu + CO₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "BaCl2", R2: "Na2SO4", Display: "BaCl₂ + Na₂SO₄ = BaSO₄↓ + 2NaCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "MgCl2", R2: "NaOH", Display: "MgCl₂ + 2NaOH = Mg(OH)₂↓ + 2NaCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2", R2: "N2", Display: "N₂ + 3H₂ = 2NH₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "NO", R2: "O2", Display: "2NO + O₂ = 2NO₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O", R2: "NO2", Display: "3NO₂ + H₂O = 2HNO₃ + NO", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O", R2: "Na2O", Display: "Na₂O + H₂O = 2NaOH", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O", R2: "K2O", Display: "K₂O + H₂O = 2KOH", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O", R2: "SO3", Display: "SO₃ + H₂O = H₂SO₄", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "BaO", R2: "H2O", Display: "BaO + H₂O = Ba(OH)₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CuCl2", R2: "NaOH", Display: "CuCl₂ + 2NaOH = Cu(OH)₂↓ + 2NaCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "FeCl3", R2: "NaOH", Display: "FeCl₃ + 3NaOH = Fe(OH)₃↓ + 3NaCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "AlCl3", R2: "NaOH", Display: "AlCl₃ + 3NaOH = Al(OH)₃↓ + 3NaCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cu(OH)2", R2: "HCl", Display: "Cu(OH)₂ + 2HCl = CuCl₂ + 2H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe(OH)3", R2: "H2SO4", Display: "2Fe(OH)₃ + 3H₂SO₄ = Fe₂(SO₄)₃ + 6H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "AgNO3", R2: "BaCl2", Display: "2AgNO₃ + BaCl₂ = 2AgCl↓ + Ba(NO₃)₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "AgNO3", R2: "Cu", Display: "2AgNO₃ + Cu = Cu(NO₃)₂ + 2Ag", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "AgNO3", R2: "Zn", Display: "Zn + 2AgNO₃ = Zn(NO₃)₂ + 2Ag", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "AgNO3", R2: "Fe", Display: "Fe + 2AgNO₃ = Fe(NO₃)₂ + 2Ag", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "NH4Cl", R2: "NaOH", Display: "NH₄Cl + NaOH = NaCl + NH₃↑ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca(OH)2", R2: "Na2CO3", Display: "Ca(OH)₂ + Na₂CO₃ = CaCO₃↓ + 2NaOH", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ba(OH)2", R2: "Na2SO4", Display: "Ba(OH)₂ + Na₂SO₄ = BaSO₄↓ + 2NaOH", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O", R2: "Na", Display: "2Na + 2H₂O = 2NaOH + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O", R2: "K", Display: "2K + 2H₂O = 2KOH + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca", R2: "H2O", Display: "Ca + 2H₂O = Ca(OH)₂ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "Na", Display: "2Na + Cl₂ = 2NaCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "Mg", Display: "Mg + Cl₂ = MgCl₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "Fe", Display: "2Fe + 3Cl₂ = 2FeCl₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "NaOH", Display: "2Al + 2NaOH + 2H₂O = 2NaAlO₂ + 3H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C", R2: "HgO", Display: "2HgO + C = 2Hg + CO₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Hg", R2: "O2", Display: "2Hg + O₂ = 2HgO", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O", R2: "NaH", Display: "NaH + H₂O = NaOH + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CaH2", R2: "H2O", Display: "CaH₂ + 2H₂O = Ca(OH)₂ + 2H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O", R2: "Na2O2", Display: "2Na₂O₂ + 2H₂O = 4NaOH + O₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CO2", R2: "Na2O2", Display: "2Na₂O₂ + 2CO₂ = 2Na₂CO₃ + O₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "Fe2O3", Display: "2Al + Fe₂O₃ = Al₂O₃ + 2Fe", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "F2", R2: "H2O", Display: "2F₂ + 2H₂O = 4HF + O₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2S", R2: "O2", Display: "2H₂S + 3O₂ = 2SO₂ + 2H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CuSO4", R2: "H2S", Display: "H₂S + CuSO₄ = CuS↓ + H₂SO₄", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "AgNO3", R2: "H2S", Display: "H₂S + 2AgNO₃ = Ag₂S↓ + 2HNO₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "NaOH", R2: "SO2", Display: "SO₂ + 2NaOH = Na₂SO₃ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca(OH)2", R2: "SO2", Display: "SO₂ + Ca(OH)₂ = CaSO₃↓ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "NaOH", R2: "SO3", Display: "SO₃ + 2NaOH = Na₂SO₄ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca(OH)2", R2: "SO3", Display: "SO₃ + Ca(OH)₂ = CaSO₄ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "NaOH", R2: "P2O5", Display: "P₂O₅ + 6NaOH = 2Na₃PO₄ + 3H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca(OH)2", R2: "P2O5", Display: "P₂O₅ + 3Ca(OH)₂ = Ca₃(PO₄)₂↓ + 3H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CuO", R2: "H2SO4", Display: "CuO + H₂SO₄ = CuSO₄ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CuO", R2: "HCl", Display: "CuO + 2HCl = CuCl₂ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe2O3", R2: "H2SO4", Display: "Fe₂O₃ + 3H₂SO₄ = Fe₂(SO₄)₃ + 3H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CO2", R2: "CaO", Display: "CaO + CO₂ = CaCO₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CaO", R2: "SO2", Display: "CaO + SO₂ = CaSO₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CO2", R2: "MgO", Display: "MgO + CO₂ = MgCO₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "NaBr", Display: "Cl₂ + 2NaBr = 2NaCl + Br₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "KI", Display: "Cl₂ + 2KI = 2KCl + I₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Br2", R2: "NaI", Display: "Br₂ + 2NaI = 2NaBr + I₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "F2", R2: "NaCl", Display: "F₂ + 2NaCl = 2NaF + Cl₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HF", R2: "SiO2", Display: "4HF + SiO₂ = SiF₄↑ + 2H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HF", R2: "NaOH", Display: "HF + NaOH = NaF + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe", R2: "S", Display: "Fe + S = FeS", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cu", R2: "S", Display: "2Cu + S = Cu₂S", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2", R2: "S", Display: "H₂ + S = H₂S", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2S", R2: "Na2O2", Display: "Na₂O₂ + H₂S = 2NaOH + S↓", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Na2O2", R2: "SO2", Display: "Na₂O₂ + SO₂ = Na₂SO₄", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CO2", R2: "Mg", Display: "2Mg + CO₂ = 2MgO + C", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2SO4", R2: "NH3", Display: "2NH₃ + H₂SO₄ = (NH₄)₂SO₄", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H3PO4", R2: "NaOH", Display: "H₃PO₄ + 3NaOH = Na₃PO₄ + 3H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "BaCl2", R2: "Na3PO4", Display: "2Na₃PO₄ + 3BaCl₂ = Ba₃(PO₄)₂↓ + 6NaCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HI", R2: "NaOH", Display: "HI + NaOH = NaI + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HBr", R2: "KOH", Display: "HBr + KOH = KBr + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "KClO3", R2: "S", Display: "2KClO₃ + 3S = 2KCl + 3SO₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "NH4NO3", R2: "NaOH", Display: "NH₄NO₃ + NaOH = NaNO₃ + NH₃↑ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "BaCl2", R2: "ZnSO4", Display: "ZnSO₄ + BaCl₂ = BaSO₄↓ + ZnCl₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ba(OH)2", R2: "MgSO4", Display: "MgSO₄ + Ba(OH)₂ = BaSO₄↓ + Mg(OH)₂↓", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al2(SO4)3", R2: "KOH", Display: "Al₂(SO₄)₃ + 6KOH = 2Al(OH)₃↓ + 3K₂SO₄", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "Cl2", Display: "2Al + 3Cl₂ = 2AlCl₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al2O3", R2: "NaOH", Display: "Al₂O₃ + 2NaOH = 2NaAlO₂ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2", R2: "Na", Display: "2Na + H₂ = 2NaH", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2", R2: "K", Display: "2K + H₂ = 2KH", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2", R2: "Mg", Display: "Mg + H₂ = MgH₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca", R2: "H2", Display: "Ca + H₂ = CaH₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ba", R2: "H2", Display: "Ba + H₂ = BaH₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O", R2: "KH", Display: "KH + H₂O = KOH + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O", R2: "MgH2", Display: "MgH₂ + 2H₂O = Mg(OH)₂ + 2H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "BaH2", R2: "H2O", Display: "BaH₂ + 2H₂O = Ba(OH)₂ + 2H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "NaH", Display: "NaH + HCl = NaCl + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2SO4", R2: "NaH", Display: "2NaH + H₂SO₄ = Na₂SO₄ + 2H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CaH2", R2: "HCl", Display: "CaH₂ + 2HCl = CaCl₂ + 2H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CaH2", R2: "H2SO4", Display: "CaH₂ + H₂SO₄ = CaSO₄ + 2H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "KH", Display: "KH + HCl = KCl + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "MgH2", Display: "MgH₂ + 2HCl = MgCl₂ + 2H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "K", R2: "O2", Display: "4K + O₂ = 2K₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Na", R2: "O2", Display: "4Na + O₂ = 2Na₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca", R2: "O2", Display: "2Ca + O₂ = 2CaO", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Na", R2: "O2", Display: "2Na + O₂ = Na₂O₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "K", R2: "O2", Display: "K + O₂ = KO₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe", R2: "O2", Display: "4Fe + 3O₂ = 2Fe₂O₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe", R2: "O2", Display: "2Fe + O₂ = 2FeO", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cu", R2: "O2", Display: "4Cu + O₂ = 2Cu₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "KOH", R2: "SO2", Display: "SO₂ + 2KOH = K₂SO₃ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "KOH", R2: "SO3", Display: "SO₃ + 2KOH = K₂SO₄ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "KOH", R2: "P2O5", Display: "P₂O₅ + 6KOH = 2K₃PO₄ + 3H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CuO", R2: "HNO3", Display: "CuO + 2HNO₃ = Cu(NO₃)₂ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe2O3", R2: "HCl", Display: "Fe₂O₃ + 6HCl = 2FeCl₃ + 3H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe2O3", R2: "HNO3", Display: "Fe₂O₃ + 6HNO₃ = 2Fe(NO₃)₃ + 3H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "MgO", Display: "MgO + 2HCl = MgCl₂ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2SO4", R2: "MgO", Display: "MgO + H₂SO₄ = MgSO₄ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HNO3", R2: "MgO", Display: "MgO + 2HNO₃ = Mg(NO₃)₂ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CaO", R2: "HCl", Display: "CaO + 2HCl = CaCl₂ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CaO", R2: "HNO3", Display: "CaO + 2HNO₃ = Ca(NO₃)₂ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "Na2O", Display: "Na₂O + 2HCl = 2NaCl + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2SO4", R2: "Na2O", Display: "Na₂O + H₂SO₄ = Na₂SO₄ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CaO", R2: "SO3", Display: "CaO + SO₃ = CaSO₄", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "MgO", R2: "SO2", Display: "MgO + SO₂ = MgSO₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "BaO", R2: "CO2", Display: "BaO + CO₂ = BaCO₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "BaO", R2: "SO2", Display: "BaO + SO₂ = BaSO₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "BaO", R2: "SO3", Display: "BaO + SO₃ = BaSO₄", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CO2", R2: "Na2O", Display: "Na₂O + CO₂ = Na₂CO₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Na2O", R2: "SO2", Display: "Na₂O + SO₂ = Na₂SO₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H3PO4", R2: "KOH", Display: "H₃PO₄ + 3KOH = K₃PO₄ + 3H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca(OH)2", R2: "H3PO4", Display: "2H₃PO₄ + 3Ca(OH)₂ = Ca₃(PO₄)₂↓ + 6H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "NaI", Display: "Cl₂ + 2NaI = 2NaCl + I₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "KBr", Display: "Cl₂ + 2KBr = 2KCl + Br₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Br2", R2: "KI", Display: "Br₂ + 2KI = 2KBr + I₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "H2O", Display: "Cl₂ + H₂O = HCl + HClO", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Br2", R2: "H2O", Display: "Br₂ + H₂O = HBr + HBrO", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "NaOH", Display: "Cl₂ + 2NaOH = NaCl + NaClO + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca(OH)2", R2: "Cl2", Display: "2Cl₂ + 2Ca(OH)₂ = CaCl₂ + Ca(ClO)₂ + 2H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "F2", R2: "H2", Display: "F₂ + H₂ = 2HF", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "H2", Display: "Cl₂ + H₂ = 2HCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Br2", R2: "H2", Display: "Br₂ + H₂ = 2HBr", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2", R2: "I2", Display: "I₂ + H₂ = 2HI", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Br2", R2: "Na", Display: "2Na + Br₂ = 2NaBr", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "I2", R2: "Na", Display: "2Na + I₂ = 2NaI", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Br2", R2: "Fe", Display: "2Fe + 3Br₂ = 2FeBr₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "Cu", Display: "Cu + Cl₂ = CuCl₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "AgNO3", R2: "NaBr", Display: "AgNO₃ + NaBr = AgBr↓ + NaNO₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "AgNO3", R2: "NaI", Display: "AgNO₃ + NaI = AgI↓ + NaNO₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "AgNO3", R2: "KI", Display: "AgNO₃ + KI = AgI↓ + KNO₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CaCl2", R2: "NaF", Display: "2NaF + CaCl₂ = CaF₂↓ + 2NaCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CaO", R2: "HF", Display: "2HF + CaO = CaF₂ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "O2", R2: "SO2", Display: "2SO₂ + O₂ = 2SO₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O", R2: "SO2", Display: "SO₂ + H₂O = H₂SO₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "Na2S", Display: "Na₂S + 2HCl = 2NaCl + H₂S↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2SO3", R2: "O2", Display: "2H₂SO₃ + O₂ = 2H₂SO₄", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2SO3", R2: "NaOH", Display: "H₂SO₃ + 2NaOH = Na₂SO₃ + 2H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Na2SO3", R2: "S", Display: "Na₂SO₃ + S = Na₂S₂O₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2S", R2: "SO2", Display: "2H₂S + SO₂ = 3S↓ + 2H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Br2", R2: "Mg", Display: "Mg + Br₂ = MgBr₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "I2", R2: "Mg", Display: "Mg + I₂ = MgI₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "Br2", Display: "2Al + 3Br₂ = 2AlBr₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "I2", Display: "2Al + 3I₂ = 2AlI₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe", R2: "I2", Display: "Fe + I₂ = FeI₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Br2", R2: "Zn", Display: "Zn + Br₂ = ZnBr₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "I2", R2: "Zn", Display: "Zn + I₂ = ZnI₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Br2", R2: "Cu", Display: "Cu + Br₂ = CuBr₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Br2", R2: "Ca", Display: "Ca + Br₂ = CaBr₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca", R2: "I2", Display: "Ca + I₂ = CaI₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe", R2: "H2SO4", Display: "Fe + H₂SO₄ = FeSO₄ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2SO4", R2: "Zn", Display: "Zn + H₂SO₄ = ZnSO₄ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2SO4", R2: "Mg", Display: "Mg + H₂SO₄ = MgSO₄ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "H2SO4", Display: "2Al + 3H₂SO₄ = Al₂(SO₄)₃ + 3H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "Na", Display: "2Na + 2HCl = 2NaCl + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2SO4", R2: "Na", Display: "2Na + H₂SO₄ = Na₂SO₄ + 2H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HNO3", R2: "Na", Display: "8Na + 10HNO₃ = 8NaNO₃ + NH₄NO₃ + 3H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "K", Display: "2K + 2HCl = 2KCl + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2SO4", R2: "K", Display: "2K + H₂SO₄ = K₂SO₄ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HNO3", R2: "K", Display: "8K + 10HNO₃ = 8KNO₃ + NH₄NO₃ + 3H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca", R2: "HCl", Display: "Ca + 2HCl = CaCl₂ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca", R2: "H2SO4", Display: "Ca + H₂SO₄ = CaSO₄ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca", R2: "HNO3", Display: "4Ca + 10HNO₃ = 4Ca(NO₃)₂ + NH₄NO₃ + 3H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ba", R2: "HCl", Display: "Ba + 2HCl = BaCl₂ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ba", R2: "H2SO4", Display: "Ba + H₂SO₄ = BaSO₄ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ba", R2: "HNO3", Display: "4Ba + 10HNO₃ = 4Ba(NO₃)₂ + NH₄NO₃ + 3H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cu", R2: "HNO3", Display: "Cu + 4HNO₃(浓) = Cu(NO₃)₂ + 2NO₂↑ + 2H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ag", R2: "HNO3", Display: "Ag + 2HNO₃(浓) = AgNO₃ + NO₂↑ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HNO3", R2: "Hg", Display: "Hg + 4HNO₃(浓) = Hg(NO₃)₂ + 2NO₂↑ + 2H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe", R2: "HNO3", Display: "Fe + 4HNO₃(稀) = Fe(NO₃)₃ + NO↑ + 2H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HNO3", R2: "Zn", Display: "4Zn + 10HNO₃(稀) = 4Zn(NO₃)₂ + NH₄NO₃ + 3H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "HNO3", Display: "8Al + 30HNO₃(稀) = 8Al(NO₃)₃ + 3NH₄NO₃ + 9H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H3PO4", R2: "Mg", Display: "3Mg + 2H₃PO₄ = Mg₃(PO₄)₂ + 3H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H3PO4", R2: "Zn", Display: "3Zn + 2H₃PO₄ = Zn₃(PO₄)₂ + 3H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe", R2: "H3PO4", Display: "3Fe + 2H₃PO₄ = Fe₃(PO₄)₂ + 3H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H3PO4", R2: "Na", Display: "6Na + 2H₃PO₄ = 2Na₃PO₄ + 3H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H3PO4", R2: "K", Display: "6K + 2H₃PO₄ = 2K₃PO₄ + 3H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HI", R2: "Mg", Display: "Mg + 2HI = MgI₂ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HI", R2: "Zn", Display: "Zn + 2HI = ZnI₂ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe", R2: "HI", Display: "Fe + 2HI = FeI₂ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HBr", R2: "Mg", Display: "Mg + 2HBr = MgBr₂ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HBr", R2: "Zn", Display: "Zn + 2HBr = ZnBr₂ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe", R2: "HBr", Display: "Fe + 2HBr = FeBr₂ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HF", R2: "Mg", Display: "Mg + 2HF = MgF₂ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HF", R2: "Zn", Display: "Zn + 2HF = ZnF₂ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "HF", Display: "2Al + 6HF = 2AlF₃ + 3H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HI", R2: "K", Display: "2K + 2HI = 2KI + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HBr", R2: "K", Display: "2K + 2HBr = 2KBr + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HF", R2: "K", Display: "2K + 2HF = 2KF + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HI", R2: "Na", Display: "2Na + 2HI = 2NaI + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HBr", R2: "Na", Display: "2Na + 2HBr = 2NaBr + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HF", R2: "Na", Display: "2Na + 2HF = 2NaF + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca", R2: "HI", Display: "Ca + 2HI = CaI₂ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca", R2: "HBr", Display: "Ca + 2HBr = CaBr₂ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca", R2: "HF", Display: "Ca + 2HF = CaF₂ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ba", R2: "HI", Display: "Ba + 2HI = BaI₂ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ba", R2: "HBr", Display: "Ba + 2HBr = BaBr₂ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ba", R2: "HF", Display: "Ba + 2HF = BaF₂ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2S", R2: "K", Display: "2K + H₂S = K₂S + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2S", R2: "Na", Display: "2Na + H₂S = Na₂S + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2S", R2: "Mg", Display: "Mg + H₂S = MgS + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca", R2: "H2S", Display: "Ca + H₂S = CaS + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ba", R2: "H2S", Display: "Ba + H₂S = BaS + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe", R2: "H2S", Display: "Fe + H₂S = FeS + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2SO3", R2: "K", Display: "2K + H₂SO₃ = K₂SO₃ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2SO3", R2: "Na", Display: "2Na + H₂SO₃ = Na₂SO₃ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2SO3", R2: "Mg", Display: "Mg + H₂SO₃ = MgSO₃ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca", R2: "H2SO3", Display: "Ca + H₂SO₃ = CaSO₃ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ba", R2: "H2SO3", Display: "Ba + H₂SO₃ = BaSO₃ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "FeSO4", R2: "H2O2", Display: "H₂O₂ + 2FeSO₄ + H₂SO₄ = Fe₂(SO₄)₃ + 2H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O2", R2: "Na2SO3", Display: "H₂O₂ + Na₂SO₃ = Na₂SO₄ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O2", R2: "KI", Display: "H₂O₂ + 2KI = 2KOH + I₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O2", R2: "H2S", Display: "H₂O₂ + H₂S = S↓ + 2H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O2", R2: "SO2", Display: "H₂O₂ + SO₂ = H₂SO₄", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "FeO", R2: "O2", Display: "4FeO + O₂ = 2Fe₂O₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cu2O", R2: "O2", Display: "2Cu₂O + O₂ = 4CuO", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "FeSO4", Display: "3Cl₂ + 6FeSO₄ = 2Fe₂(SO₄)₃ + 2FeCl₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "Na2SO3", Display: "Cl₂ + Na₂SO₃ + H₂O = Na₂SO₄ + 2HCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "H2S", Display: "Cl₂ + H₂S = S↓ + 2HCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "FeCl2", Display: "Cl₂ + 2FeCl₂ = 2FeCl₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Br2", R2: "FeSO4", Display: "3Br₂ + 6FeSO₄ = 2Fe₂(SO₄)₃ + 2FeBr₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Br2", R2: "Na2SO3", Display: "Br₂ + Na₂SO₃ + H₂O = Na₂SO₄ + 2HBr", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Br2", R2: "H2S", Display: "Br₂ + H₂S = S↓ + 2HBr", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cu", R2: "FeCl3", Display: "2FeCl₃ + Cu = 2FeCl₂ + CuCl₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe", R2: "FeCl3", Display: "2FeCl₃ + Fe = 3FeCl₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "FeCl3", R2: "KI", Display: "2FeCl₃ + 2KI = 2FeCl₂ + 2KCl + I₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "FeCl3", R2: "H2S", Display: "2FeCl₃ + H₂S = 2FeCl₂ + S↓ + 2HCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "FeCl3", R2: "Na2SO3", Display: "2FeCl₃ + Na₂SO₃ + H₂O = 2FeCl₂ + Na₂SO₄ + 2HCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C", R2: "HNO3", Display: "4HNO₃(浓) + C = CO₂↑ + 4NO₂↑ + 2H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HNO3", R2: "S", Display: "6HNO₃(浓) + S = H₂SO₄ + 6NO₂↑ + 2H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HNO3", R2: "P", Display: "5HNO₃(浓) + P = H₃PO₄ + 5NO₂↑ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "FeO", R2: "HNO3", Display: "FeO + 4HNO₃(浓) = Fe(NO₃)₃ + NO₂↑ + 2H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cu", R2: "H2SO4", Display: "Cu + 2H₂SO₄(浓) = CuSO₄ + SO₂↑ + 2H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C", R2: "H2SO4", Display: "C + 2H₂SO₄(浓) = CO₂↑ + 2SO₂↑ + 2H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2SO4", R2: "S", Display: "S + 2H₂SO₄(浓) = 3SO₂↑ + 2H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "FeSO4", R2: "Na2O2", Display: "3Na₂O₂ + 6FeSO₄ + 6H₂O = 4Fe(OH)₃↓ + 2Fe₂(SO₄)₃ + 6Na⁺", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe(OH)2", R2: "O2", Display: "4Fe(OH)₂ + O₂ + 2H₂O = 4Fe(OH)₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "NaOH", R2: "S", Display: "3S + 6NaOH = 2Na₂S + Na₂SO₃ + 3H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "NO2", R2: "NaOH", Display: "2NO₂ + 2NaOH = NaNO₂ + NaNO₃ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "Fe3O4", Display: "8Al + 3Fe₃O₄ = 4Al₂O₃ + 9Fe", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "CuO", Display: "2Al + 3CuO = Al₂O₃ + 3Cu", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C", R2: "CuO", Display: "C + 2CuO = 2Cu + CO₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CO", R2: "CuO", Display: "CO + CuO = Cu + CO₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "F2", R2: "NaBr", Display: "F₂ + 2NaBr = 2NaF + Br₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "F2", R2: "NaI", Display: "F₂ + 2NaI = 2NaF + I₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "F2", R2: "KCl", Display: "F₂ + 2KCl = 2KF + Cl₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "F2", R2: "KBr", Display: "F₂ + 2KBr = 2KF + Br₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "F2", R2: "KI", Display: "F₂ + 2KI = 2KF + I₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "F2", R2: "Na", Display: "F₂ + 2Na = 2NaF", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "F2", R2: "K", Display: "F₂ + 2K = 2KF", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "F2", R2: "Mg", Display: "F₂ + Mg = MgF₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca", R2: "F2", Display: "F₂ + Ca = CaF₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ba", R2: "F2", Display: "F₂ + Ba = BaF₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "F2", Display: "3F₂ + 2Al = 2AlF₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "F2", R2: "Fe", Display: "3F₂ + 2Fe = 2FeF₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cu", R2: "F2", Display: "F₂ + Cu = CuF₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ag", R2: "F2", Display: "F₂ + 2Ag = 2AgF", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "F2", R2: "Hg", Display: "F₂ + Hg = HgF₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "F2", R2: "Zn", Display: "F₂ + Zn = ZnF₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "F2", R2: "NH3", Display: "3F₂ + 2NH₃ = 6HF + N₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "F2", R2: "S", Display: "3F₂ + S = SF₆", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "F2", R2: "P", Display: "5F₂ + 2P = 2PF₅", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C", R2: "F2", Display: "2F₂ + C = CF₄", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HF", R2: "KOH", Display: "HF + KOH = KF + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ba(OH)2", R2: "HF", Display: "2HF + Ba(OH)₂ = BaF₂ + 2H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HF", R2: "MgO", Display: "2HF + MgO = MgF₂ + H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al2O3", R2: "HF", Display: "6HF + Al₂O₃ = 2AlF₃ + 3H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "BaCl2", R2: "NaF", Display: "2NaF + BaCl₂ = BaF₂↓ + 2NaCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CaCl2", R2: "KF", Display: "2KF + CaCl₂ = CaF₂↓ + 2KCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "O2", R2: "Zn", Display: "2Zn + O₂ = 2ZnO", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ag", R2: "O2", Display: "4Ag + O₂ = 2Ag₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "O2", R2: "Si", Display: "Si + O₂ = SiO₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cu", R2: "O2", Display: "2Cu + O₂ = 2CuO", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "O2", Display: "4Al + 3O₂ = 2Al₂O₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ba", R2: "O2", Display: "2Ba + O₂ = 2BaO", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "N2", R2: "O2", Display: "N₂ + O₂ = 2NO", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "O2", Display: "2Cl₂ + 7O₂ = 2Cl₂O₇", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "K", Display: "2K + Cl₂ = 2KCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Br2", R2: "K", Display: "2K + Br₂ = 2KBr", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "I2", R2: "K", Display: "2K + I₂ = 2KI", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "K", R2: "S", Display: "2K + S = K₂S", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Na", R2: "S", Display: "2Na + S = Na₂S", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Mg", R2: "S", Display: "Mg + S = MgS", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca", R2: "S", Display: "Ca + S = CaS", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ba", R2: "S", Display: "Ba + S = BaS", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "S", Display: "2Al + 3S = Al₂S₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "S", R2: "Zn", Display: "Zn + S = ZnS", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca", R2: "Cl2", Display: "Ca + Cl₂ = CaCl₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ba", R2: "Cl2", Display: "Ba + Cl₂ = BaCl₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "Zn", Display: "Zn + Cl₂ = ZnCl₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ag", R2: "Cl2", Display: "2Ag + Cl₂ = 2AgCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "Hg", Display: "Hg + Cl₂ = HgCl₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ba", R2: "Br2", Display: "Ba + Br₂ = BaBr₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ba", R2: "I2", Display: "Ba + I₂ = BaI₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Br2", R2: "Hg", Display: "Hg + Br₂ = HgBr₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Hg", R2: "I2", Display: "Hg + I₂ = HgI₂", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ag", R2: "Br2", Display: "2Ag + Br₂ = 2AgBr", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ag", R2: "I2", Display: "2Ag + I₂ = 2AgI", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cu", R2: "I2", Display: "2Cu + I₂ = 2CuI", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "P", Display: "2P + 3Cl₂ = 2PCl₃", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "P", R2: "S", Display: "2P + 5S = P₂S₅", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Mg", R2: "ZnCl2", Display: "Mg + ZnCl₂ = MgCl₂ + Zn", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Mg", R2: "ZnSO4", Display: "Mg + ZnSO₄ = MgSO₄ + Zn", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "FeCl2", R2: "Mg", Display: "Mg + FeCl₂ = MgCl₂ + Fe", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "FeSO4", R2: "Mg", Display: "Mg + FeSO₄ = MgSO₄ + Fe", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CuCl2", R2: "Mg", Display: "Mg + CuCl₂ = MgCl₂ + Cu", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "AlCl3", R2: "Mg", Display: "3Mg + 2AlCl₃ = 3MgCl₂ + 2Al", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al2(SO4)3", R2: "Mg", Display: "3Mg + Al₂(SO₄)₃ = 3MgSO₄ + 2Al", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "FeCl2", R2: "Zn", Display: "Zn + FeCl₂ = ZnCl₂ + Fe", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "FeSO4", R2: "Zn", Display: "Zn + FeSO₄ = ZnSO₄ + Fe", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CuCl2", R2: "Zn", Display: "Zn + CuCl₂ = ZnCl₂ + Cu", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CuCl2", R2: "Fe", Display: "Fe + CuCl₂ = FeCl₂ + Cu", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "ZnCl2", Display: "2Al + 3ZnCl₂ = 2AlCl₃ + 3Zn", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "FeCl2", Display: "2Al + 3FeCl₂ = 2AlCl₃ + 3Fe", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "CuCl2", Display: "2Al + 3CuCl₂ = 2AlCl₃ + 3Cu", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HNO3", R2: "Mg", Display: "Mg + 2HNO₃ = Mg(NO₃)₂ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe2O3", R2: "H2", Display: "3H₂ + Fe₂O₃ = 2Fe + 3H₂O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca", R2: "H3PO4", Display: "3Ca + 2H₃PO₄ = Ca₃(PO₄)₂↓ + 3H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "H3PO4", Display: "2Al + 2H₃PO₄ = 2AlPO₄↓ + 3H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "HI", Display: "2Al + 6HI = 2AlI₃ + 3H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "HBr", Display: "2Al + 6HBr = 2AlBr₃ + 3H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "H2S", Display: "2Al + 3H₂S = Al₂S₃ + 3H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "H2SO3", Display: "2Al + 3H₂SO₃ = Al₂(SO₃)₃ + 3H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2S", R2: "Zn", Display: "Zn + H₂S = ZnS↓ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe", R2: "H2SO3", Display: "Fe + H₂SO₃ = FeSO₃ + H₂↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CuSO4", R2: "Mg", Display: "Mg + CuSO₄ = MgSO₄ + Cu", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应67: Na2O2 + CO2 -> Na2CO3 + O2
		{R1: "CO2", R2: "Na2O2", Display: "2Na2O2 + 2CO2 = 2Na2CO3 + O2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应68: Al + Fe2O3 -> Al2O3 + Fe (铝热反应)
		{R1: "Al", R2: "Fe2O3", Display: "2Al + Fe2O3 = Al2O3 + 2Fe", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应69: F2 + H2O -> HF + O2
		{R1: "F2", R2: "H2O", Display: "2F2 + 2H2O = 4HF + O2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应70: H2S + O2 -> SO2 + H2O (硫化氢燃烧)
		{R1: "H2S", R2: "O2", Display: "2H2S + 3O2 = 2SO2 + 2H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应71: H2S + CuSO4 -> CuS↓ + H2SO4
		{R1: "CuSO4", R2: "H2S", Display: "H2S + CuSO4 = CuS↓ + H2SO4", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应72: H2S + AgNO3 -> Ag2S↓ + HNO3
		{R1: "AgNO3", R2: "H2S", Display: "H2S + 2AgNO3 = Ag2S↓ + 2HNO3", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应73: SO2 + NaOH -> Na2SO3 + H2O
		{R1: "NaOH", R2: "SO2", Display: "SO2 + 2NaOH = Na2SO3 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应74: SO2 + Ca(OH)2 -> CaSO3↓ + H2O
		{R1: "Ca(OH)2", R2: "SO2", Display: "SO2 + Ca(OH)2 = CaSO3↓ + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应75: SO3 + NaOH -> Na2SO4 + H2O
		{R1: "NaOH", R2: "SO3", Display: "SO3 + 2NaOH = Na2SO4 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应76: SO3 + Ca(OH)2 -> CaSO4 + H2O
		{R1: "Ca(OH)2", R2: "SO3", Display: "SO3 + Ca(OH)2 = CaSO4 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应77: P2O5 + NaOH -> Na3PO4 + H2O
		{R1: "NaOH", R2: "P2O5", Display: "P2O5 + 6NaOH = 2Na3PO4 + 3H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应78: P2O5 + Ca(OH)2 -> Ca3(PO4)2↓ + H2O
		{R1: "Ca(OH)2", R2: "P2O5", Display: "P2O5 + 3Ca(OH)2 = Ca3(PO4)2↓ + 3H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应79: CuO + H2SO4 -> CuSO4 + H2O
		{R1: "CuO", R2: "H2SO4", Display: "CuO + H2SO4 = CuSO4 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应80: CuO + HCl -> CuCl2 + H2O
		{R1: "CuO", R2: "HCl", Display: "CuO + 2HCl = CuCl2 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应81: Fe2O3 + H2SO4 -> Fe2(SO4)3 + H2O
		{R1: "Fe2O3", R2: "H2SO4", Display: "Fe2O3 + 3H2SO4 = Fe2(SO4)3 + 3H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应82: CaO + CO2 -> CaCO3
		{R1: "CO2", R2: "CaO", Display: "CaO + CO2 = CaCO3", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应83: CaO + SO2 -> CaSO3
		{R1: "CaO", R2: "SO2", Display: "CaO + SO2 = CaSO3", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应84: MgO + CO2 -> MgCO3
		{R1: "CO2", R2: "MgO", Display: "MgO + CO2 = MgCO3", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应85: Cl2 + NaBr -> NaCl + Br2
		{R1: "Cl2", R2: "NaBr", Display: "Cl2 + 2NaBr = 2NaCl + Br2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应86: Cl2 + KI -> KCl + I2
		{R1: "Cl2", R2: "KI", Display: "Cl2 + 2KI = 2KCl + I2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应87: Br2 + NaI -> NaBr + I2
		{R1: "Br2", R2: "NaI", Display: "Br2 + 2NaI = 2NaBr + I2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应88: F2 + NaCl -> NaF + Cl2
		{R1: "F2", R2: "NaCl", Display: "F2 + 2NaCl = 2NaF + Cl2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应89: HF + SiO2 -> SiF4↑ + H2O
		{R1: "HF", R2: "SiO2", Display: "4HF + SiO2 = SiF4↑ + 2H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应90: HF + NaOH -> NaF + H2O
		{R1: "HF", R2: "NaOH", Display: "HF + NaOH = NaF + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应91: Fe + S -> FeS
		{R1: "Fe", R2: "S", Display: "Fe + S = FeS", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应92: Cu + S -> Cu2S
		{R1: "Cu", R2: "S", Display: "2Cu + S = Cu2S", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应93: H2 + S -> H2S
		{R1: "H2", R2: "S", Display: "H2 + S = H2S", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应94: Na2O2 + H2S -> NaOH + S↓
		{R1: "H2S", R2: "Na2O2", Display: "Na2O2 + H2S = 2NaOH + S↓", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应95: Na2O2 + SO2 -> Na2SO4
		{R1: "Na2O2", R2: "SO2", Display: "Na2O2 + SO2 = Na2SO4", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应96: Mg + CO2 -> MgO + C
		{R1: "CO2", R2: "Mg", Display: "2Mg + CO2 = 2MgO + C", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应97: NH3 + H2SO4 -> (NH4)2SO4
		{R1: "H2SO4", R2: "NH3", Display: "2NH3 + H2SO4 = (NH4)2SO4", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应98: H3PO4 + NaOH -> Na3PO4 + H2O
		{R1: "H3PO4", R2: "NaOH", Display: "H3PO4 + 3NaOH = Na3PO4 + 3H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应99: Na3PO4 + BaCl2 -> Ba3(PO4)2↓ + NaCl
		{R1: "BaCl2", R2: "Na3PO4", Display: "2Na3PO4 + 3BaCl2 = Ba3(PO4)2↓ + 6NaCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应100: HI + NaOH -> NaI + H2O
		{R1: "HI", R2: "NaOH", Display: "HI + NaOH = NaI + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应101: HBr + KOH -> KBr + H2O
		{R1: "HBr", R2: "KOH", Display: "HBr + KOH = KBr + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应102: KClO3 + S -> KCl + SO2↑
		{R1: "KClO3", R2: "S", Display: "2KClO3 + 3S = 2KCl + 3SO2↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应103: NH4NO3 + NaOH -> NaNO3 + NH3↑ + H2O
		{R1: "NH4NO3", R2: "NaOH", Display: "NH4NO3 + NaOH = NaNO3 + NH3↑ + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应104: ZnSO4 + BaCl2 -> BaSO4↓ + ZnCl2
		{R1: "BaCl2", R2: "ZnSO4", Display: "ZnSO4 + BaCl2 = BaSO4↓ + ZnCl2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应105: MgSO4 + Ba(OH)2 -> BaSO4↓ + Mg(OH)2↓
		{R1: "Ba(OH)2", R2: "MgSO4", Display: "MgSO4 + Ba(OH)2 = BaSO4↓ + Mg(OH)2↓", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应106: Al2(SO4)3 + KOH -> Al(OH)3↓ + K2SO4
		{R1: "Al2(SO4)3", R2: "KOH", Display: "Al2(SO4)3 + 6KOH = 2Al(OH)3↓ + 3K2SO4", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应107: Al + Cl2 -> AlCl3
		{R1: "Al", R2: "Cl2", Display: "2Al + 3Cl2 = 2AlCl3", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应108: Al2O3 + NaOH -> NaAlO2 + H2O
		{R1: "Al2O3", R2: "NaOH", Display: "Al2O3 + 2NaOH = 2NaAlO2 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应109-113: 金属与氢气反应
		{R1: "H2", R2: "Na", Display: "2Na + H2 = 2NaH", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2", R2: "K", Display: "2K + H2 = 2KH", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2", R2: "Mg", Display: "Mg + H2 = MgH2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca", R2: "H2", Display: "Ca + H2 = CaH2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ba", R2: "H2", Display: "Ba + H2 = BaH2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应114-116: 氢化物与水反应
		{R1: "H2O", R2: "KH", Display: "KH + H2O = KOH + H2↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O", R2: "MgH2", Display: "MgH2 + 2H2O = Mg(OH)2 + 2H2↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "BaH2", R2: "H2O", Display: "BaH2 + 2H2O = Ba(OH)2 + 2H2↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应117-121: 氢化物与酸反应
		{R1: "HCl", R2: "NaH", Display: "NaH + HCl = NaCl + H2↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2SO4", R2: "NaH", Display: "2NaH + H2SO4 = Na2SO4 + 2H2↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CaH2", R2: "HCl", Display: "CaH2 + 2HCl = CaCl2 + 2H2↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CaH2", R2: "H2SO4", Display: "CaH2 + H2SO4 = CaSO4 + 2H2↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "KH", Display: "KH + HCl = KCl + H2↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "MgH2", Display: "MgH2 + 2HCl = MgCl2 + 2H2↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应122-127: 金属与氧气反应 (不同产物)
		{R1: "K", R2: "O2", Display: "4K + O2 = 2K2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Na", R2: "O2", Display: "4Na + O2 = 2Na2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca", R2: "O2", Display: "2Ca + O2 = 2CaO", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Na", R2: "O2", Display: "2Na + O2 = Na2O2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "K", R2: "O2", Display: "K + O2 = KO2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe", R2: "O2", Display: "4Fe + 3O2 = 2Fe2O3", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe", R2: "O2", Display: "2Fe + O2 = 2FeO", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cu", R2: "O2", Display: "4Cu + O2 = 2Cu2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 反应128-130: 与钾碱反应
		{R1: "KOH", R2: "SO2", Display: "SO2 + 2KOH = K2SO3 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "KOH", R2: "SO3", Display: "SO3 + 2KOH = K2SO4 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "KOH", R2: "P2O5", Display: "P2O5 + 6KOH = 2K3PO4 + 3H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// ============================================
		// 从 ref.json 补充的反应（缺失的159条）
		// 添加时间: 2026-02-08
		// 生成工具: backend/scripts/check_missing_reactions.js
		// ============================================
		{R1: "H2O", R2: "P2O5", Display: "P2O5 + 3H2O = 2H3PO4", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "I2", R2: "NaOH", Display: "I2 + 2NaOH = NaI + NaIO + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "AgNO3", R2: "KBr", Display: "AgNO3 + KBr = AgBr + KNO3", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HNO3", R2: "KOH", Display: "HNO3 + KOH = KNO3 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CuSO4", R2: "NaOH", Display: "CuSO4 + 2NaOH = Cu(OH)2 + Na2SO4", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "CuSO4", Display: "2Al + 3CuSO4 = Al2(SO4)3 + 3Cu", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CH4", R2: "O2", Display: "CH4 + 2O2 = CO2 + 2H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C2H5OH", R2: "O2", Display: "C2H5OH + 3O2 = 2CO2 + 3H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "NaOH", R2: "SiO2", Display: "SiO2 + 2NaOH = Na2SiO3 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al(OH)3", R2: "HCl", Display: "Al(OH)3 + 3HCl = AlCl3 + 3H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ba(OH)2", R2: "H2SO4", Display: "Ba(OH)2 + H2SO4 = BaSO4 + 2H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HNO3", R2: "NH3", Display: "NH3 + HNO3 = NH4NO3", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "NH4HCO3", Display: "NH4HCO3 + HCl = NH4Cl + CO2 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "NH4Cl", R2: "NaNO2", Display: "NaNO2 + NH4Cl = NaCl + N2 + 2H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2S", R2: "Pb(NO3)2", Display: "H2S + Pb(NO3)2 = PbS + 2HNO3", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "SO2", Display: "SO2 + Cl2 + 2H2O = H2SO4 + 2HCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Na2S", R2: "O2", Display: "2Na2S + O2 + 2H2O = 4NaOH + 2S", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O", R2: "PCl3", Display: "PCl3 + 3H2O = H3PO3 + 3HCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O", R2: "PCl5", Display: "PCl5 + 4H2O = H3PO4 + 5HCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CaCl2", R2: "Na3PO4", Display: "2Na3PO4 + 3CaCl2 = Ca3(PO4)2 + 6NaCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Br2", R2: "SO2", Display: "Br2 + SO2 + 2H2O = H2SO4 + 2HBr", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "I2", R2: "Na2S2O3", Display: "I2 + 2Na2S2O3 = 2NaI + Na2S4O6", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "NaClO", Display: "NaClO + 2HCl = NaCl + Cl2 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "KClO3", Display: "KClO3 + 6HCl = KCl + 3Cl2 + 3H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "MnO2", Display: "MnO2 + 4HCl = MnCl2 + Cl2 + 2H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "AlCl3", R2: "NH3·H2O", Display: "AlCl3 + 3NH3·H2O = Al(OH)3 + 3NH4Cl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "FeS", R2: "HCl", Display: "FeS + 2HCl = FeCl2 + H2S", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2SO4", R2: "Na2S2O3", Display: "Na2S2O3 + H2SO4 = Na2SO4 + S + SO2 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O2", R2: "H2SO3", Display: "H2SO3 + H2O2 = H2SO4 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Na2SO3", R2: "O2", Display: "2Na2SO3 + O2 = 2Na2SO4", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca(H2PO4)2", R2: "Ca(OH)2", Display: "Ca(H2PO4)2 + 2Ca(OH)2 = Ca3(PO4)2 + 4H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "O2", R2: "PH3", Display: "2PH3 + 4O2 = P2O5 + 3H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "NaH2PO4", R2: "NaOH", Display: "NaH2PO4 + 2NaOH = Na3PO4 + 2H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "H2O2", Display: "Cl2 + H2O2 = 2HCl + O2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Br2", R2: "H2O2", Display: "Br2 + H2O2 = 2HBr + O2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "AgNO3", R2: "NH4Cl", Display: "NH4Cl + AgNO3 = AgCl + NH4NO3", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al(OH)3", R2: "NaOH", Display: "Al(OH)3 + NaOH = Na[Al(OH)4]", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "Na2SiO3", Display: "Na2SiO3 + 2HCl = H2SiO3 + 2NaCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CaSiO3", R2: "HCl", Display: "CaSiO3 + 2HCl = CaCl2 + H2SiO3", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "NaOH", R2: "Si", Display: "Si + 2NaOH + H2O = Na2SiO3 + 2H2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O", R2: "Mg3N2", Display: "Mg3N2 + 6H2O = 3Mg(OH)2 + 2NH3", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ca3P2", R2: "H2O", Display: "Ca3P2 + 6H2O = 3Ca(OH)2 + 2PH3", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CaC2", R2: "H2O", Display: "CaC2 + 2H2O = Ca(OH)2 + C2H2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CO2", R2: "KO2", Display: "4KO2 + 2CO2 = 2K2CO3 + 3O2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "KMnO4", Display: "2KMnO4 + 16HCl = 2MnCl2 + 2KCl + 5Cl2 + 8H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O2", R2: "KMnO4", Display: "2KMnO4 + 5H2O2 + 3H2SO4 = 2MnSO4 + K2SO4 + 5O2 + 8H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "FeSO4", R2: "K2Cr2O7", Display: "K2Cr2O7 + 6FeSO4 + 7H2SO4 = Cr2(SO4)3 + 3Fe2(SO4)3 + K2SO4 + 7H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HNO3", R2: "I2", Display: "I2 + 10HNO3 = 2HIO3 + 10NO2 + 4H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe(OH)2", R2: "HNO3", Display: "3Fe(OH)2 + 10HNO3 = 3Fe(NO3)3 + NO + 8H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HNO3", R2: "SO2", Display: "3SO2 + 2HNO3 + 2H2O = 3H2SO4 + 2NO", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2C2O4", R2: "KMnO4", Display: "5H2C2O4 + 2KMnO4 + 3H2SO4 = K2SO4 + 2MnSO4 + 10CO2 + 8H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCOOH", R2: "KMnO4", Display: "5HCOOH + 2KMnO4 + 3H2SO4 = K2SO4 + 2MnSO4 + 5CO2 + 8H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ag(NH3)2OH", R2: "CH3CHO", Display: "CH3CHO + 2Ag(NH3)2OH = CH3COONH4 + 2Ag + 3NH3 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CH3CHO", R2: "Cu(OH)2", Display: "CH3CHO + 2Cu(OH)2 + NaOH = CH3COONa + Cu2O + 3H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ag(NH3)2OH", R2: "HCHO", Display: "HCHO + 4Ag(NH3)2OH = (NH4)2CO3 + 4Ag + 6NH3 + 2H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Br2", R2: "C6H5OH", Display: "C6H5OH + 3Br2 = C6H2Br3OH + 3HBr", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C6H5OH", R2: "NaOH", Display: "C6H5OH + NaOH = C6H5ONa + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C6H5ONa", R2: "CO2", Display: "C6H5ONa + CO2 + H2O = C6H5OH + NaHCO3", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CH3COOC2H5", R2: "NaOH", Display: "CH3COOC2H5 + NaOH = CH3COONa + C2H5OH", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ag(NH3)2OH", R2: "C6H12O6", Display: "C6H12O6 + 2Ag(NH3)2OH = C6H12O7 + 2Ag + 4NH3 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C6H12O6", R2: "Cu(OH)2", Display: "C6H12O6 + 2Cu(OH)2 = C6H12O7 + Cu2O + 2H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Ag(NH3)2OH", R2: "C12H22O11", Display: "C12H22O11 + 2Ag(NH3)2OH = C12H22O12 + 2Ag + 4NH3 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "(C6H10O5)n", R2: "H2O", Display: "(C6H10O5)n + nH2O = nC6H12O6", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C2H2", R2: "HCl", Display: "C2H2 + HCl = CH2=CHCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C2H4", R2: "HCl", Display: "C2H4 + HCl = CH3CH2Cl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Br2", R2: "C6H6", Display: "C6H6 + Br2 = C6H5Br + HBr", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C6H6", R2: "HNO3", Display: "C6H6 + HNO3 = C6H5NO2 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C6H6", R2: "H2", Display: "C6H6 + 3H2 = C6H12", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C2H5OH", R2: "Na", Display: "2C2H5OH + 2Na = 2C2H5ONa + H2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C2H5OH", R2: "HBr", Display: "C2H5OH + HBr = C2H5Br + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CH3COOH", R2: "Na", Display: "2CH3COOH + 2Na = 2CH3COONa + H2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CH3COOH", R2: "NaOH", Display: "CH3COOH + NaOH = CH3COONa + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CH3COOH", R2: "Na2CO3", Display: "2CH3COOH + Na2CO3 = 2CH3COONa + CO2 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCOONa", R2: "NaOH", Display: "HCOONa + NaOH = Na2CO3 + H2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CH3COONa", R2: "NaOH", Display: "CH3COONa + NaOH = Na2CO3 + CH4", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C2H5Br", R2: "NaOH", Display: "C2H5Br + NaOH = C2H5OH + NaBr", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C2H5Cl", R2: "NaOH", Display: "C2H5Cl + NaOH = C2H5OH + NaCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C6H5Cl", R2: "NaOH", Display: "C6H5Cl + 2NaOH = C6H5ONa + NaCl + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "AgOH", R2: "CH3I", Display: "CH3I + AgOH = CH3OH + AgI", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CH3CH2CH2Br", R2: "KOH", Display: "CH3CH2CH2Br + KOH = CH3CH=CH2 + KBr + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "NH3", R2: "O2", Display: "4NH3 + 5O2 = 4NO + 6H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O", R2: "SiCl4", Display: "SiCl4 + 4H2O = H4SiO4 + 4HCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "AlCl3", R2: "Na2S", Display: "2AlCl3 + 3Na2S + 6H2O = 2Al(OH)3 + 3H2S + 6NaCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CO2", R2: "NaAlO2", Display: "2NaAlO2 + CO2 + 3H2O = 2Al(OH)3 + Na2CO3", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "FeCl3", R2: "Na2CO3", Display: "2FeCl3 + 3Na2CO3 + 3H2O = 2Fe(OH)3 + 6NaCl + 3CO2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CuSO4", R2: "Na2CO3", Display: "CuSO4 + Na2CO3 + H2O = Cu(OH)2 + Na2SO4 + CO2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "AgNO3", R2: "Na3PO4", Display: "3AgNO3 + Na3PO4 = Ag3PO4 + 3NaNO3", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Na2HPO4", R2: "NaOH", Display: "Na2HPO4 + NaOH = Na3PO4 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "NaHCO3", R2: "SO2", Display: "SO2 + 2NaHCO3 = Na2SO3 + 2CO2 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CuCl2", R2: "H2S", Display: "H2S + CuCl2 = CuS + 2HCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "I2", R2: "Na2S", Display: "I2 + Na2S = 2NaI + S", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "FeBr2", Display: "3Cl2 + 2FeBr2 = 2FeCl3 + 2Br2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Br2", R2: "FeI2", Display: "2FeI2 + 3Br2 = 2FeBr3 + 2I2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CO2", R2: "Ca(ClO)2", Display: "Ca(ClO)2 + CO2 + H2O = CaCO3 + 2HClO", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cu(NO3)2", R2: "Fe", Display: "Fe + Cu(NO3)2 = Fe(NO3)2 + Cu", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "FeCl3", R2: "Zn", Display: "Zn + 2FeCl3 = ZnCl2 + 2FeCl2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al2(SO4)3", R2: "NaHCO3", Display: "Al2(SO4)3 + 6NaHCO3 = 2Al(OH)3 + 3Na2SO4 + 6CO2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Fe(OH)3", R2: "HI", Display: "2Fe(OH)3 + 6HI = 2FeI2 + I2 + 6H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cu(OH)2", R2: "NH3·H2O", Display: "Cu(OH)2 + 4NH3·H2O = [Cu(NH3)4](OH)2 + 4H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "AgOH", R2: "NH3·H2O", Display: "AgOH + 2NH3·H2O = [Ag(NH3)2]OH + 2H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "NaOH", R2: "Zn(OH)2", Display: "Zn(OH)2 + 2NaOH = Na2[Zn(OH)4]", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "NaOH", R2: "Pb(OH)2", Display: "Pb(OH)2 + 2NaOH = Na2[Pb(OH)4]", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O", R2: "Mg2C3", Display: "Mg2C3 + 4H2O = 2Mg(OH)2 + C3H4", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "FeS", R2: "O2", Display: "4FeS + 7O2 = 2Fe2O3 + 4SO2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cu2S", R2: "O2", Display: "Cu2S + 2O2 = 2CuO + SO2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "O2", R2: "PbS", Display: "2PbS + 3O2 = 2PbO + 2SO2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "O2", R2: "SiH4", Display: "SiH4 + 2O2 = SiO2 + 2H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CO", R2: "PdCl2", Display: "CO + PdCl2 + H2O = CO2 + Pd + 2HCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "FeSO4", R2: "KMnO4", Display: "10FeSO4 + 2KMnO4 + 8H2SO4 = 5Fe2(SO4)3 + 2MnSO4 + K2SO4 + 8H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "KMnO4", R2: "Na2SO3", Display: "5Na2SO3 + 2KMnO4 + 3H2SO4 = 5Na2SO4 + 2MnSO4 + K2SO4 + 3H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "KI", R2: "KMnO4", Display: "10KI + 2KMnO4 + 8H2SO4 = 5I2 + 2MnSO4 + 6K2SO4 + 8H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "FeCl2", R2: "K2Cr2O7", Display: "6FeCl2 + K2Cr2O7 + 14HCl = 6FeCl3 + 2KCl + 2CrCl3 + 7H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CH4", R2: "Cl2", Display: "CH4 + Cl2 = CH3Cl + HCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C2H6", R2: "Cl2", Display: "C2H6 + Cl2 = C2H5Cl + HCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Br2", R2: "C3H8", Display: "C3H8 + Br2 = C3H7Br + HBr", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Br2", R2: "C2H2", Display: "C2H2 + 2Br2 = C2H2Br4", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C6H6", R2: "Cl2", Display: "C6H6 + Cl2 = C6H5Cl + HCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C6H5CH3", R2: "HNO3", Display: "C6H5CH3 + 3HNO3 = C6H2(NO2)3CH3 + 3H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C6H5ONa", R2: "HCl", Display: "C6H5ONa + HCl = C6H5OH + NaCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C2H5OH", R2: "CH3COOH", Display: "CH3COOH + C2H5OH = CH3COOC2H5 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C6H5Br", R2: "NaOH", Display: "C6H5Br + 2NaOH = C6H5ONa + NaBr + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CH3OH", R2: "O2", Display: "2CH3OH + 3O2 = 2CO2 + 4H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 补充遗漏的2条反应
		{R1: "CO2", R2: "NaClO", Display: "NaClO + CO2 + H2O = NaHCO3 + HClO", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "FeCl3", R2: "Na2S", Display: "3Na2S + 2FeCl3 + 6H2O = 2Fe(OH)3 + 3H2S + 6NaCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 补充锰相关反应
		{R1: "H2O2", R2: "KMnO4", Display: "2KMnO4 + 3H2O2 = 2MnO2 + 3O2 + 2KOH + 2H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "(NH4)2S", R2: "MnSO4", Display: "MnSO4 + (NH4)2S = MnS + (NH4)2SO4", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "Mn", Display: "Mn + 2HCl = MnCl2 + H2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HI", R2: "MnO2", Display: "MnO2 + 4HI = MnI2 + I2 + 2H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Mn", R2: "O2", Display: "2Mn + O2 = 2MnO", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "Mn(OH)2", Display: "Mn(OH)2 + 2HCl = MnCl2 + 2H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2SO4", R2: "MnO2", Display: "2MnO2 + 2H2SO4 = 2MnSO4 + O2 + 2H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "KMnO4", R2: "Na2S", Display: "2KMnO4 + 3Na2S + 4H2O = 2MnO2 + 3S + 6NaOH + 2KOH", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "Mn", Display: "Mn + Cl2 = MnCl2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "MnCO3", Display: "MnCO3 + 2HCl = MnCl2 + CO2 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Mn(NO3)2", R2: "NaOH", Display: "Mn(NO3)2 + 2NaOH = Mn(OH)2 + 2NaNO3", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CO", R2: "MnO2", Display: "MnO2 + 2CO = Mn + 2CO2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		// 补充锰相关反应
		{R1: "H2O2", R2: "KMnO4", Display: "2KMnO4 + 3H2O2 = 2MnO2 + 3O2 + 2KOH + 2H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "(NH4)2S", R2: "MnSO4", Display: "MnSO4 + (NH4)2S = MnS + (NH4)2SO4", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "Mn", Display: "Mn + 2HCl = MnCl2 + H2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HI", R2: "MnO2", Display: "MnO2 + 4HI = MnI2 + I2 + 2H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Mn", R2: "O2", Display: "2Mn + O2 = 2MnO", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "Mn(OH)2", Display: "Mn(OH)2 + 2HCl = MnCl2 + 2H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2SO4", R2: "MnO2", Display: "2MnO2 + 2H2SO4 = 2MnSO4 + O2 + 2H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "KMnO4", R2: "Na2S", Display: "2KMnO4 + 3Na2S + 4H2O = 2MnO2 + 3S + 6NaOH + 2KOH", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "Mn", Display: "Mn + Cl2 = MnCl2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "MnCO3", Display: "MnCO3 + 2HCl = MnCl2 + CO2 + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Mn(NO3)2", R2: "NaOH", Display: "Mn(NO3)2 + 2NaOH = Mn(OH)2 + 2NaNO3", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "CO", R2: "MnO2", Display: "MnO2 + 2CO = Mn + 2CO2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},

		{R1: "MnCl2", R2: "NaOH", Display: "MnCl2 + 2NaOH = Mn(OH)2↓ + 2NaCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Mn(OH)2", R2: "O2", Display: "2Mn(OH)2 + O2 = 2MnO2 + 2H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2O2", R2: "MnO2", Display: "MnO2 + H2O2 + H2SO4 = MnSO4 + O2↑ + 2H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "MnO2", Display: "3MnO2 + 4Al = 3Mn + 2Al2O3", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "MnSO4", R2: "NaOH", Display: "MnSO4 + 2NaOH = Mn(OH)2↓ + Na2SO4", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "MnSO4", R2: "Na2CO3", Display: "MnSO4 + Na2CO3 = MnCO3↓ + Na2SO4", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "KMnO4", R2: "SO2", Display: "2KMnO4 + 5SO2 + 2H2O = K2SO4 + 2MnSO4 + 2H2SO4", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "KClO3", R2: "MnO2", Display: "2KClO3 = 2KCl + 3O2↑ (MnO2为催化剂)", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "HCl", R2: "MnS", Display: "MnS + 2HCl = MnCl2 + H2S↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "MnS", R2: "O2", Display: "2MnS + 3O2 = 2MnO + 2SO2", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "C", R2: "MnO", Display: "MnO + C = Mn + CO↑", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2", R2: "MnO", Display: "MnO + H2 = Mn + H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Al", R2: "MnO", Display: "3MnO + 2Al = 3Mn + Al2O3", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "H2C2O4", R2: "MnO2", Display: "MnO2 + H2C2O4 + H2SO4 = MnSO4 + 2CO2↑ + 2H2O", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
		{R1: "Cl2", R2: "K2MnO4", Display: "2K2MnO4 + Cl2 = 2KMnO4 + 2KCl", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},
	}

	log.Printf("开始批量导入 %d 组化学反应...", reactionCount)

	// 在导入前确保所有 R1/R2 都符合 Canonical ordering (R1 <= R2)
	for i := range reactions {
		if reactions[i].R1 > reactions[i].R2 {
			reactions[i].R1, reactions[i].R2 = reactions[i].R2, reactions[i].R1
		}
	}

	// 使用分批插入以避免单次插入过多数据
	batchSize := 50
	for i := 0; i < len(reactions); i += batchSize {
		end := i + batchSize
		if end > len(reactions) {
			end = len(reactions)
		}
		batch := reactions[i:end]
		if err := DB.Create(&batch).Error; err != nil {
			log.Printf("批次 %d-%d 插入失败: %v", i, end, err)
			return err
		}
		log.Printf("已导入 %d/%d 组反应", end, len(reactions))
	}

	log.Printf("✅ 成功导入 %d 组化学反应", len(reactions))
	return nil
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
