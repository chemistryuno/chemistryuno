package database

import (
	"chemistryuno/models"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(filepath string) error {
	var err error
	DB, err = sql.Open("sqlite", filepath)
	if err != nil {
		return err
	}

	// 测试连接
	if err = DB.Ping(); err != nil {
		return err
	}

	// 创建表
	if err = createTables(); err != nil {
		return err
	}

	log.Println("数据库初始化成功")
	return nil
}

func createTables() error {
	// 用户表
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		UID INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		avatar TEXT DEFAULT '',
		is_admin BOOLEAN DEFAULT 0,
		role TEXT DEFAULT 'user',
		two_factor_enabled BOOLEAN DEFAULT 0,
		two_factor_secret TEXT DEFAULT '',
		points INTEGER DEFAULT 1000,
		negative_play_count INTEGER DEFAULT 0,
		banned_until DATETIME DEFAULT NULL,
		last_decay_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// 悬赏表
	bountyTable := `
	CREATE TABLE IF NOT EXISTS bounties (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		target_uid INTEGER NOT NULL,
		amount INTEGER NOT NULL,
		created_by INTEGER NOT NULL,
		status TEXT DEFAULT 'active',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (target_uid) REFERENCES users(UID),
		FOREIGN KEY (created_by) REFERENCES users(UID)
	);`

	// 牌组配置表
	deckConfigTable := `
	CREATE TABLE IF NOT EXISTS deck_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		is_global BOOLEAN DEFAULT 0,
		cards TEXT NOT NULL,
		created_by INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (created_by) REFERENCES users(UID)
	);`

	// 游戏记录表
	gameHistoryTable := `
	CREATE TABLE IF NOT EXISTS game_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		room_id TEXT NOT NULL,
		winner_uid INTEGER,
		players TEXT NOT NULL,
		started_at DATETIME,
		finished_at DATETIME,
		FOREIGN KEY (winner_uid) REFERENCES users(UID)
	);`

	// 化学反应表
	reactionsTable := `
	CREATE TABLE IF NOT EXISTS reactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		r1 TEXT NOT NULL,
		r2 TEXT NOT NULL,
		display TEXT NOT NULL,
		status TEXT DEFAULT 'pending_coworker',
		group_id TEXT,
		created_by INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (created_by) REFERENCES users(UID)
	);`

	// 反馈表
	feedbackTable := `
	CREATE TABLE IF NOT EXISTS feedbacks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		content TEXT NOT NULL,
		type TEXT DEFAULT 'general',
		status TEXT DEFAULT 'unread',
		processed_by INTEGER DEFAULT NULL,
		processed_at DATETIME DEFAULT NULL,
		last_urged_at DATETIME DEFAULT NULL,
		urge_count INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(UID)
	);`

	// 系统配置表
	systemConfigTable := `
	CREATE TABLE IF NOT EXISTS system_configs (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		description TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	tables := []string{userTable, bountyTable, deckConfigTable, gameHistoryTable, reactionsTable, feedbackTable, systemConfigTable}

	for _, table := range tables {
		if _, err := DB.Exec(table); err != nil {
			return err
		}
	}

	// 初始化默认系统配置
	initSystemConfigs()

	// 增量更新表结构（针对已存在的数据库）——按需添加列以避免错误
	columnExists := func(table, column string) bool {
		rows, err := DB.Query("PRAGMA table_info(" + table + ")")
		if err != nil {
			return false
		}
		defer rows.Close()

		for rows.Next() {
			var cid int
			var colname, ctype string
			var notnull, dflt_value, pk sql.NullString
			if err := rows.Scan(&cid, &colname, &ctype, &notnull, &dflt_value, &pk); err == nil {
				if colname == column {
					return true
				}
			}
		}
		return false
	}

	// users
	if !columnExists("users", "two_factor_enabled") {
		_, _ = DB.Exec("ALTER TABLE users ADD COLUMN two_factor_enabled BOOLEAN DEFAULT 0")
	}
	if !columnExists("users", "two_factor_secret") {
		_, _ = DB.Exec("ALTER TABLE users ADD COLUMN two_factor_secret TEXT DEFAULT ''")
	}
	if !columnExists("users", "points") {
		_, _ = DB.Exec("ALTER TABLE users ADD COLUMN points INTEGER DEFAULT 1000")
	}
	if !columnExists("users", "negative_play_count") {
		_, _ = DB.Exec("ALTER TABLE users ADD COLUMN negative_play_count INTEGER DEFAULT 0")
	}
	if !columnExists("users", "banned_until") {
		_, _ = DB.Exec("ALTER TABLE users ADD COLUMN banned_until DATETIME DEFAULT NULL")
	}
	if !columnExists("users", "last_decay_at") {
		_, _ = DB.Exec("ALTER TABLE users ADD COLUMN last_decay_at DATETIME DEFAULT CURRENT_TIMESTAMP")
	}

	// reactions
	if !columnExists("reactions", "r1") {
		_, _ = DB.Exec("ALTER TABLE reactions ADD COLUMN r1 TEXT DEFAULT ''")
	}
	if !columnExists("reactions", "r2") {
		_, _ = DB.Exec("ALTER TABLE reactions ADD COLUMN r2 TEXT DEFAULT ''")
	}
	if !columnExists("reactions", "display") {
		_, _ = DB.Exec("ALTER TABLE reactions ADD COLUMN display TEXT DEFAULT ''")
	}
	if !columnExists("reactions", "status") {
		_, _ = DB.Exec("ALTER TABLE reactions ADD COLUMN status TEXT DEFAULT 'approved'")
	}
	if !columnExists("reactions", "group_id") {
		_, _ = DB.Exec("ALTER TABLE reactions ADD COLUMN group_id TEXT DEFAULT ''")
	}

	// feedbacks
	if !columnExists("feedbacks", "processed_by") {
		_, _ = DB.Exec("ALTER TABLE feedbacks ADD COLUMN processed_by INTEGER DEFAULT NULL")
	}
	if !columnExists("feedbacks", "processed_at") {
		_, _ = DB.Exec("ALTER TABLE feedbacks ADD COLUMN processed_at DATETIME DEFAULT NULL")
	}
	if !columnExists("feedbacks", "last_urged_at") {
		_, _ = DB.Exec("ALTER TABLE feedbacks ADD COLUMN last_urged_at DATETIME DEFAULT NULL")
	}
	if !columnExists("feedbacks", "urge_count") {
		_, _ = DB.Exec("ALTER TABLE feedbacks ADD COLUMN urge_count INTEGER DEFAULT 0")
	}
	if !columnExists("feedbacks", "resolution_note") {
		_, _ = DB.Exec("ALTER TABLE feedbacks ADD COLUMN resolution_note TEXT DEFAULT NULL")
	}
	if !columnExists("feedbacks", "remove_at") {
		_, _ = DB.Exec("ALTER TABLE feedbacks ADD COLUMN remove_at DATETIME DEFAULT NULL")
	}

	// 检查并创建默认管理员账号
	if err := createDefaultAdmin(); err != nil {
		log.Printf("创建管理员账户时出错: %v", err)
	}

	// 插入默认化学反应数据
	equations := map[string]string{
		"H2+O2": "2H₂ + O₂ = 2H₂O", "C+O2": "C + O₂ = CO₂", "S+O2": "S + O₂ = SO₂", "P+O2": "4P + 5O₂ = 2P₂O₅",
		"Fe+O2": "3Fe + 2O₂ = Fe₃O₄", "Mg+O2": "2Mg + O₂ = 2MgO", "CO+O2": "2CO + O₂ = 2CO₂",
		"HCl+NaOH": "HCl + NaOH = NaCl + H₂O", "H2SO4+NaOH": "H₂SO₄ + 2NaOH = Na₂SO₄ + 2H₂O", "HNO3+NaOH": "HNO₃ + NaOH = NaNO₃ + H₂O",
		"HCl+Ca(OH)2": "2HCl + Ca(OH)₂ = CaCl₂ + 2H₂O", "H2SO4+BaCl2": "H₂SO₄ + BaCl₂ = BaSO₄↓ + 2HCl", "HCl+AgNO3": "HCl + AgNO₃ = AgCl↓ + HNO₃",
		"NaCl+AgNO3": "NaCl + AgNO₃ = AgCl↓ + NaNO₃", "Fe+HCl": "Fe + 2HCl = FeCl₂ + H₂↑", "Zn+HCl": "Zn + 2HCl = ZnCl₂ + H₂↑",
		"Mg+HCl": "Mg + 2HCl = MgCl₂ + H₂↑", "Al+HCl": "2Al + 6HCl = 2AlCl₃ + 3H₂↑", "Fe+CuSO4": "Fe + CuSO₄ = FeSO₄ + Cu",
		"Zn+CuSO4": "Zn + CuSO₄ = ZnSO₄ + Cu", "CO2+Ca(OH)2": "CO₂ + Ca(OH)₂ = CaCO₃↓ + H₂O", "CO2+NaOH": "CO₂ + 2NaOH = Na₂CO₃ + H₂O",
		"CO2+H2O": "CO₂ + H₂O = H₂CO₃", "CaO+H2O": "CaO + H₂O = Ca(OH)₂", "MgO+H2O": "MgO + H₂O = Mg(OH)₂",
		"CaCO3+HCl": "CaCO₃ + 2HCl = CaCl₂ + H₂O + CO₂↑", "Na2CO3+HCl": "Na₂CO₃ + 2HCl = 2NaCl + H₂O + CO₂↑", "NaHCO3+HCl": "NaHCO₃ + HCl = NaCl + H₂O + CO₂↑",
		"NH3+HCl": "NH₃ + HCl = NH₄Cl", "Cu+AgNO3": "Cu + 2AgNO₃ = Cu(NO₃)₂ + 2Ag", "CuO+H2": "CuO + H₂ = Cu + H₂O",
		"Fe2O3+CO": "Fe₂O₃ + 3CO = 2Fe + 3CO₂", "CuO+C": "2CuO + C = 2Cu + CO₂↑", "BaCl2+Na2SO4": "BaCl₂ + Na₂SO₄ = BaSO₄↓ + 2NaCl",
		"MgCl2+NaOH": "MgCl₂ + 2NaOH = Mg(OH)₂↓ + 2NaCl", "N2+H2": "N₂ + 3H₂ = 2NH₃", "NO+O2": "2NO + O₂ = 2NO₂",
		"NO2+H2O": "3NO₂ + H₂O = 2HNO₃ + NO", "Na2O+H2O": "Na₂O + H₂O = 2NaOH", "K2O+H2O": "K₂O + H₂O = 2KOH",
		"SO3+H2O": "SO₃ + H₂O = H₂SO₄", "BaO+H2O": "BaO + H₂O = Ba(OH)₂", "CuCl2+NaOH": "CuCl₂ + 2NaOH = Cu(OH)₂↓ + 2NaCl",
		"FeCl3+NaOH": "FeCl₃ + 3NaOH = Fe(OH)₃↓ + 3NaCl", "AlCl3+NaOH": "AlCl₃ + 3NaOH = Al(OH)₃↓ + 3NaCl",
		"Cu(OH)2+HCl": "Cu(OH)₂ + 2HCl = CuCl₂ + 2H₂O", "Fe(OH)3+H2SO4": "2Fe(OH)₃ + 3H₂SO₄ = Fe₂(SO₄)₃ + 6H₂O",
		"AgNO3+BaCl2": "2AgNO₃ + BaCl₂ = 2AgCl↓ + Ba(NO₃)₂", "AgNO3+Cu": "2AgNO₃ + Cu = Cu(NO₃)₂ + 2Ag",
		"Zn+AgNO3": "Zn + 2AgNO₃ = Zn(NO₃)₂ + 2Ag", "Fe+AgNO3": "Fe + 2AgNO₃ = Fe(NO₃)₂ + 2Ag",
		"NH4Cl+NaOH": "NH₄Cl + NaOH = NaCl + NH₃↑ + H₂O", "Ca(OH)2+Na2CO3": "Ca(OH)2 + Na2CO3 = CaCO3↓ + 2NaOH",
		"Ba(OH)2+Na2SO4": "Ba(OH)₂ + Na₂SO₄ = BaSO₄↓ + 2NaOH", "Na+H2O": "2Na + 2H₂O = 2NaOH + H₂↑",
		"K+H2O": "2K + 2H₂O = 2KOH + H₂↑", "Ca+H2O": "Ca + 2H₂O = Ca(OH)₂ + H₂↑", "Na+Cl2": "2Na + Cl₂ = 2NaCl",
		"Mg+Cl2": "Mg + Cl₂ = MgCl₂", "Fe+Cl2": "2Fe + 3Cl₂ = 2FeCl₃", "Al+NaOH": "2Al + 2NaOH + 2H₂O = 2NaAlO₂ + 3H₂↑",
		"HgO+C": "2HgO + C = 2Hg + CO₂↑", "Hg+O2": "2Hg + O₂ = 2HgO",
		"NaH+H2O": "NaH + H₂O = NaOH + H₂↑", "CaH2+H2O": "CaH₂ + 2H₂O = Ca(OH)₂ + 2H₂↑",
		"Na2O2+H2O": "2Na₂O₂ + 2H₂O = 4NaOH + O₂↑", "Na2O2+CO2": "2Na₂O₂ + 2CO₂ = 2Na₂CO₃ + O₂",
		"Al+Fe2O3": "2Al + Fe₂O₃ = Al₂O₃ + 2Fe", "F2+H2O": "2F₂ + 2H₂O = 4HF + O₂",
		"H2S+O2": "2H₂S + 3O₂ = 2SO₂ + 2H₂O", "H2S+CuSO4": "H₂S + CuSO₄ = CuS↓ + H₂SO₄",
		"H2S+AgNO3": "H₂S + 2AgNO₃ = Ag₂S↓ + 2HNO₃", "SO2+NaOH": "SO₂ + 2NaOH = Na₂SO₃ + H₂O",
		"SO2+Ca(OH)2": "SO₂ + Ca(OH)₂ = CaSO₃↓ + H₂O", "SO3+NaOH": "SO₃ + 2NaOH = Na₂SO₄ + H₂O",
		"SO3+Ca(OH)2": "SO₃ + Ca(OH)₂ = CaSO₄ + H₂O", "P2O5+NaOH": "P₂O₅ + 6NaOH = 2Na₃PO₄ + 3H₂O",
		"P2O5+Ca(OH)2": "P₂O₅ + 3Ca(OH)₂ = Ca₃(PO₄)₂↓ + 3H₂O", "CuO+H2SO4": "CuO + H₂SO₄ = CuSO₄ + H₂O",
		"CuO+HCl": "CuO + 2HCl = CuCl₂ + H₂O", "Fe2O3+H2SO4": "Fe₂O₃ + 3H₂SO₄ = Fe₂(SO₄)₃ + 3H₂O",
		"CaO+CO2": "CaO + CO₂ = CaCO₃", "CaO+SO2": "CaO + SO₂ = CaSO₃", "MgO+CO2": "MgO + CO₂ = MgCO₃",
		"Cl2+NaBr": "Cl₂ + 2NaBr = 2NaCl + Br₂", "Cl2+KI": "Cl₂ + 2KI = 2KCl + I₂", "Br2+NaI": "Br₂ + 2NaI = 2NaBr + I₂",
		"F2+NaCl": "F₂ + 2NaCl = 2NaF + Cl₂", "HF+SiO2": "4HF + SiO₂ = SiF₄↑ + 2H₂O", "HF+NaOH": "HF + NaOH = NaF + H₂O",
		"Fe+S": "Fe + S = FeS", "Cu+S": "2Cu + S = Cu₂S", "H2+S": "H₂ + S = H₂S",
		"Na2O2+H2S": "Na₂O₂ + H₂S = 2NaOH + S↓", "Na2O2+SO2": "Na₂O₂ + SO₂ = Na₂SO₄",
		"Mg+CO2":    "2Mg + CO₂ = 2MgO + C",
		"NH3+H2SO4": "2NH₃ + H₂SO₄ = (NH₄)₂SO₄", "H3PO4+NaOH": "H₃PO₄ + 3NaOH = Na₃PO₄ + 3H₂O", "Na3PO4+BaCl2": "2Na₃PO₄ + 3BaCl₂ = Ba₃(PO₄)₂↓ + 6NaCl",
		"HI+NaOH": "HI + NaOH = NaI + H₂O", "HBr+KOH": "HBr + KOH = KBr + H₂O", "KClO3+S": "2KClO₃ + 3S = 2KCl + 3SO₂↑",
		"NH4NO3+NaOH": "NH₄NO₃ + NaOH = NaNO₃ + NH₃↑ + H₂O", "ZnSO4+BaCl2": "ZnSO₄ + BaCl₂ = BaSO₄↓ + ZnCl₂",
		"MgSO4+Ba(OH)2": "MgSO₄ + Ba(OH)₂ = BaSO₄↓ + Mg(OH)₂↓", "Al2(SO4)3+KOH": "Al₂(SO₄)₃ + 6KOH = 2Al(OH)₃↓ + 3K₂SO₄",
		"Al+Cl2": "2Al + 3Cl₂ = 2AlCl₃", "Al2O3+NaOH": "Al₂O₃ + 2NaOH = 2NaAlO₂ + H₂O",
		"Na+H2": "2Na + H₂ = 2NaH", "K+H2": "2K + H₂ = 2KH", "Mg+H2": "Mg + H₂ = MgH₂", "Ca+H2": "Ca + H₂ = CaH₂", "Ba+H2": "Ba + H₂ = BaH₂",
		"KH+H2O": "KH + H₂O = KOH + H₂↑", "MgH2+H2O": "MgH₂ + 2H₂O = Mg(OH)₂ + 2H₂↑", "BaH2+H2O": "BaH₂ + 2H₂O = Ba(OH)₂ + 2H₂↑",
		"NaH+HCl": "NaH + HCl = NaCl + H₂↑", "NaH+H2SO4": "2NaH + H₂SO₄ = Na₂SO₄ + 2H₂↑", "CaH2+HCl": "CaH₂ + 2HCl = CaCl₂ + 2H₂↑",
		"CaH2+H2SO4": "CaH₂ + H₂SO₄ = CaSO₄ + 2H₂↑", "KH+HCl": "KH + HCl = KCl + H₂↑", "MgH2+HCl": "MgH₂ + 2HCl = MgCl₂ + 2H₂↑",
		"K+O2": "4K + O₂ = 2K₂O", "Na+O2": "4Na + O₂ = 2Na₂O", "Ca+O2": "2Ca + O₂ = 2CaO",
		"Na+O2->Na2O2": "2Na + O₂ = Na₂O₂", "K+O2->KO2": "K + O₂ = KO₂",
		"Fe+O2->Fe2O3": "4Fe + 3O₂ = 2Fe₂O₃", "Fe+O2->FeO": "2Fe + O₂ = 2FeO",
		"Cu+O2->Cu2O": "4Cu + O₂ = 2Cu₂O",
		"SO2+KOH":     "SO₂ + 2KOH = K₂SO₃ + H₂O", "SO3+KOH": "SO₃ + 2KOH = K₂SO₄ + H₂O", "P2O5+KOH": "P₂O₅ + 6KOH = 2K₃PO₄ + 3H₂O",
		"CuO+HNO3": "CuO + 2HNO₃ = Cu(NO₃)₂ + H₂O", "Fe2O3+HCl": "Fe₂O₃ + 6HCl = 2FeCl₃ + 3H₂O", "Fe2O3+HNO3": "Fe₂O₃ + 6HNO₃ = 2Fe(NO₃)₃ + 3H₂O",
		"MgO+HCl": "MgO + 2HCl = MgCl₂ + H₂O", "MgO+H2SO4": "MgO + H₂SO₄ = MgSO₄ + H₂O", "MgO+HNO3": "MgO + 2HNO₃ = Mg(NO₃)₂ + H₂O",
		"CaO+HCl": "CaO + 2HCl = CaCl₂ + H₂O", "CaO+HNO3": "CaO + 2HNO₃ = Ca(NO₃)₂ + H₂O", "Na2O+HCl": "Na₂O + 2HCl = 2NaCl + H₂O",
		"Na2O+H2SO4": "Na₂O + H₂SO₄ = Na₂SO₄ + H₂O", "CaO+SO3": "CaO + SO₃ = CaSO₄", "MgO+SO2": "MgO + SO₂ = MgSO₃",
		"BaO+CO2": "BaO + CO₂ = BaCO₃", "BaO+SO2": "BaO + SO₂ = BaSO₃", "BaO+SO3": "BaO + SO₃ = BaSO₄",
		"Na2O+CO2": "Na₂O + CO₂ = Na₂CO₃", "Na2O+SO2": "Na₂O + SO₂ = Na₂SO₃", "H3PO4+KOH": "H₃PO₄ + 3KOH = K₃PO₄ + 3H₂O",
		"H3PO4+Ca(OH)2": "2H₃PO₄ + 3Ca(OH)₂ = Ca₃(PO₄)₂↓ + 6H₂O", "Cl2+NaI": "Cl₂ + 2NaI = 2NaCl + I₂",
		"Cl2+KBr": "Cl₂ + 2KBr = 2KCl + Br₂", "Br2+KI": "Br₂ + 2KI = 2KBr + I₂", "Cl2+H2O": "Cl₂ + H₂O = HCl + HClO",
		"Br2+H2O": "Br₂ + H₂O = HBr + HBrO", "Cl2+NaOH": "Cl₂ + 2NaOH = NaCl + NaClO + H₂O", "Cl2+Ca(OH)2": "2Cl₂ + 2Ca(OH)₂ = CaCl₂ + Ca(ClO)₂ + 2H₂O",
		"F2+H2": "F₂ + H₂ = 2HF", "Cl2+H2": "Cl₂ + H₂ = 2HCl", "Br2+H2": "Br₂ + H₂ = 2HBr", "I2+H2": "I₂ + H₂ = 2HI",
		"Na+Br2": "2Na + Br₂ = 2NaBr", "Na+I2": "2Na + I₂ = 2NaI", "Fe+Br2": "2Fe + 3Br₂ = 2FeBr₃", "Cu+Cl2": "Cu + Cl₂ = CuCl₂",
		"AgNO3+NaBr": "AgNO₃ + NaBr = AgBr↓ + NaNO₃", "AgNO3+NaI": "AgNO₃ + NaI = AgI↓ + NaNO₃", "AgNO3+KI": "AgNO₃ + KI = AgI↓ + KNO₃",
		"NaF+CaCl2": "2NaF + CaCl₂ = CaF₂↓ + 2NaCl", "HF+CaO": "2HF + CaO = CaF₂ + H₂O",
		"SO2+O2": "2SO₂ + O₂ = 2SO₃", "SO2+H2O": "SO₂ + H₂O = H₂SO₃", "Na2S+HCl": "Na₂S + 2HCl = 2NaCl + H₂S↑",
		"H2SO3+O2": "2H₂SO₃ + O₂ = 2H₂SO₄", "H2SO3+NaOH": "H₂SO₃ + 2NaOH = Na₂SO₃ + 2H₂O", "Na2SO3+S": "Na₂SO₃ + S = Na₂S₂O₃",
		"H2S+SO2": "2H₂S + SO₂ = 3S↓ + 2H₂O", "Mg+Br2": "Mg + Br₂ = MgBr₂", "Mg+I2": "Mg + I₂ = MgI₂",
		"Al+Br2": "2Al + 3Br₂ = 2AlBr₃", "Al+I2": "2Al + 3I₂ = 2AlI₃", "Fe+I2": "Fe + I₂ = FeI₂", "Zn+Br2": "Zn + Br₂ = ZnBr₂",
		"Zn+I2": "Zn + I₂ = ZnI₂", "Cu+Br2": "Cu + Br₂ = CuBr₂", "Ca+Br2": "Ca + Br₂ = CaBr₂", "Ca+I2": "Ca + I₂ = CaI₂",
		"Fe+H2SO4": "Fe + H₂SO₄ = FeSO₄ + H₂↑", "Zn+H2SO4": "Zn + H₂SO₄ = ZnSO₄ + H₂↑", "Mg+H2SO4": "Mg + H₂SO₄ = MgSO₄ + H₂↑",
		"Al+H2SO4": "2Al + 3H₂SO₄ = Al₂(SO₄)₃ + 3H₂↑", "Na+HCl": "2Na + 2HCl = 2NaCl + H₂↑", "Na+H2SO4": "2Na + H₂SO₄ = Na₂SO₄ + 2H₂↑",
		"Na+HNO3": "8Na + 10HNO₃ = 8NaNO₃ + NH₄NO₃ + 3H₂O", "K+HCl": "2K + 2HCl = 2KCl + H₂↑", "K+H2SO4": "2K + H₂SO₄ = K₂SO₄ + H₂↑",
		"K+HNO3": "8K + 10HNO₃ = 8KNO₃ + NH₄NO₃ + 3H₂O",
		"Ca+HCl": "Ca + 2HCl = CaCl₂ + H₂↑", "Ca+H2SO4": "Ca + H₂SO₄ = CaSO₄ + H₂↑",
		"Ca+HNO3": "4Ca + 10HNO₃ = 4Ca(NO₃)₂ + NH₄NO₃ + 3H₂O", "Ba+HCl": "Ba + 2HCl = BaCl₂ + H₂↑", "Ba+H2SO4": "Ba + H₂SO₄ = BaSO₄ + H₂↑",
		"Ba+HNO3": "4Ba + 10HNO₃ = 4Ba(NO₃)₂ + NH₄NO₃ + 3H₂O", "Cu+HNO3": "Cu + 4HNO₃(浓) = Cu(NO₃)₂ + 2NO₂↑ + 2H₂O",
		"Ag+HNO3": "Ag + 2HNO₃(浓) = AgNO₃ + NO₂↑ + H₂O", "Hg+HNO3": "Hg + 4HNO₃(浓) = Hg(NO₃)₂ + 2NO₂↑ + 2H₂O",
		"Fe+HNO3": "Fe + 4HNO₃(稀) = Fe(NO₃)₃ + NO↑ + 2H₂O", "Zn+HNO3": "4Zn + 10HNO₃(稀) = 4Zn(NO₃)₂ + NH₄NO₃ + 3H₂O",
		"Al+HNO3":  "8Al + 30HNO₃(稀) = 8Al(NO₃)₃ + 3NH₄NO₃ + 9H₂O",
		"Mg+H3PO4": "3Mg + 2H₃PO₄ = Mg₃(PO₄)₂ + 3H₂↑", "Zn+H3PO4": "3Zn + 2H₃PO₄ = Zn₃(PO₄)₂ + 3H₂↑", "Fe+H3PO4": "3Fe + 2H₃PO₄ = Fe₃(PO₄)₂ + 3H₂↑",
		"Na+H3PO4": "6Na + 2H₃PO₄ = 2Na₃PO₄ + 3H₂↑", "K+H3PO4": "6K + 2H₃PO₄ = 2K₃PO₄ + 3H₂↑",
		"Mg+HI": "Mg + 2HI = MgI₂ + H₂↑", "Zn+HI": "Zn + 2HI = ZnI₂ + H₂↑", "Fe+HI": "Fe + 2HI = FeI₂ + H₂↑",
		"Mg+HBr": "Mg + 2HBr = MgBr₂ + H₂↑", "Zn+HBr": "Zn + 2HBr = ZnBr₂ + H₂↑", "Fe+HBr": "Fe + 2HBr = FeBr₂ + H₂↑",
		"Mg+HF": "Mg + 2HF = MgF₂ + H₂↑", "Zn+HF": "Zn + 2HF = ZnF₂ + H₂↑", "Al+HF": "2Al + 6HF = 2AlF₃ + 3H₂↑",
		"K+HI": "2K + 2HI = 2KI + H₂↑", "K+HBr": "2K + 2HBr = 2KBr + H₂↑", "K+HF": "2K + 2HF = 2KF + H₂↑",
		"Na+HI": "2Na + 2HI = 2NaI + H₂↑", "Na+HBr": "2Na + 2HBr = 2NaBr + H₂↑", "Na+HF": "2Na + 2HF = 2NaF + H₂↑",
		"Ca+HI": "Ca + 2HI = CaI₂ + H₂↑", "Ca+HBr": "Ca + 2HBr = CaBr₂ + H₂↑", "Ca+HF": "Ca + 2HF = CaF₂ + H₂↑",
		"Ba+HI": "Ba + 2HI = BaI₂ + H₂↑", "Ba+HBr": "Ba + 2HBr = BaBr₂ + H₂↑", "Ba+HF": "Ba + 2HF = BaF₂ + H₂↑",
		"K+H2S": "2K + H₂S = K₂S + H₂↑", "Na+H2S": "2Na + H₂S = Na₂S + H₂↑", "Mg+H2S": "Mg + H₂S = MgS + H₂↑",
		"Ca+H2S": "Ca + H₂S = CaS + H₂↑", "Ba+H2S": "Ba + H₂S = BaS + H₂↑", "Fe+H2S": "Fe + H₂S = FeS + H₂↑",
		"K+H2SO3": "2K + H₂SO₃ = K₂SO₃ + H₂↑", "Na+H2SO3": "2Na + H₂SO₃ = Na₂SO₃ + H₂↑", "Mg+H2SO3": "Mg + H₂SO₃ = MgSO₃ + H₂↑",
		"Ca+H2SO3": "Ca + H₂SO₃ = CaSO₃ + H₂↑", "Ba+H2SO3": "Ba + H₂SO₃ = BaSO₃ + H₂↑",
		"H2O2+FeSO4": "H₂O₂ + 2FeSO₄ + H₂SO₄ = Fe₂(SO₄)₃ + 2H₂O", "H2O2+Na2SO3": "H₂O₂ + Na₂SO₃ = Na₂SO₄ + H₂O",
		"H2O2+KI": "H₂O₂ + 2KI = 2KOH + I₂", "H2O2+H2S": "H₂O₂ + H₂S = S↓ + 2H₂O", "H2O2+SO2": "H₂O₂ + SO₂ = H₂SO₄",
		"FeO+O2": "4FeO + O₂ = 2Fe₂O₃", "Cu2O+O2": "2Cu₂O + O₂ = 4CuO",
		"Cl2+FeSO4": "3Cl₂ + 6FeSO₄ = 2Fe₂(SO₄)₃ + 2FeCl₃", "Cl2+Na2SO3": "Cl₂ + Na₂SO₃ + H₂O = Na₂SO₄ + 2HCl",
		"Cl2+H2S": "Cl₂ + H₂S = S↓ + 2HCl", "Cl2+FeCl2": "Cl₂ + 2FeCl₂ = 2FeCl₃", "Br2+FeSO4": "3Br₂ + 6FeSO₄ = 2Fe₂(SO₄)₃ + 2FeBr₃",
		"Br2+Na2SO3": "Br₂ + Na₂SO₃ + H₂O = Na₂SO₄ + 2HBr", "Br2+H2S": "Br₂ + H₂S = S↓ + 2HBr",
		"FeCl3+Cu": "2FeCl₃ + Cu = 2FeCl₂ + CuCl₂", "FeCl3+Fe": "2FeCl₃ + Fe = 3FeCl₂", "FeCl3+KI": "2FeCl₃ + 2KI = 2FeCl₂ + 2KCl + I₂",
		"FeCl3+H2S": "2FeCl₃ + H₂S = 2FeCl₂ + S↓ + 2HCl", "FeCl3+Na2SO3": "2FeCl₃ + Na₂SO₃ + H₂O = 2FeCl₂ + Na₂SO₄ + 2HCl",
		"HNO3+C": "4HNO₃(浓) + C = CO₂↑ + 4NO₂↑ + 2H₂O", "HNO3+S": "6HNO₃(浓) + S = H₂SO₄ + 6NO₂↑ + 2H₂O",
		"HNO3+P": "5HNO₃(浓) + P = H₃PO₄ + 5NO₂↑ + H₂O", "HNO3+FeO": "FeO + 4HNO₃(浓) = Fe(NO₃)₃ + NO₂↑ + 2H₂O",
		"H2SO4+Cu": "Cu + 2H₂SO₄(浓) = CuSO₄ + SO₂↑ + 2H₂O", "H2SO4+C": "C + 2H₂SO₄(浓) = CO₂↑ + 2SO₂↑ + 2H₂O",
		"H2SO4+S": "S + 2H₂SO₄(浓) = 3SO₂↑ + 2H₂O", "Na2O2+FeSO4": "3Na₂O₂ + 6FeSO₄ + 6H₂O = 4Fe(OH)₃↓ + 2Fe₂(SO₄)₃ + 6Na⁺",
		"O2+Fe(OH)2": "4Fe(OH)₂ + O₂ + 2H₂O = 4Fe(OH)₃", "S+NaOH": "3S + 6NaOH = 2Na₂S + Na₂SO₃ + 3H₂O",
		"NO2+NaOH": "2NO₂ + 2NaOH = NaNO₂ + NaNO₃ + H₂O", "Al+Fe3O4": "8Al + 3Fe₃O₄ = 4Al₂O₃ + 9Fe",
		"Al+CuO": "2Al + 3CuO = Al₂O₃ + 3Cu", "C+CuO": "C + 2CuO = 2Cu + CO₂↑", "CO+CuO": "CO + CuO = Cu + CO₂",
		"F2+NaBr": "F₂ + 2NaBr = 2NaF + Br₂", "F2+NaI": "F₂ + 2NaI = 2NaF + I₂", "F2+KCl": "F₂ + 2KCl = 2KF + Cl₂",
		"F2+KBr": "F₂ + 2KBr = 2KF + Br₂", "F2+KI": "F₂ + 2KI = 2KF + I₂", "F2+Na": "F₂ + 2Na = 2NaF",
		"F2+K": "F₂ + 2K = 2KF", "F2+Mg": "F₂ + Mg = MgF₂", "F2+Ca": "F₂ + Ca = CaF₂", "F2+Ba": "F₂ + Ba = BaF₂",
		"F2+Al": "3F₂ + 2Al = 2AlF₃", "F2+Fe": "3F₂ + 2Fe = 2FeF₃", "F2+Cu": "F₂ + Cu = CuF₂", "F2+Ag": "F₂ + 2Ag = 2AgF",
		"F2+Hg": "F₂ + Hg = HgF₂", "F2+Zn": "F₂ + Zn = ZnF₂", "F2+NH3": "3F₂ + 2NH₃ = 6HF + N₂",
		"F2+S": "3F₂ + S = SF₆", "F2+P": "5F₂ + 2P = 2PF₅", "F2+C": "2F₂ + C = CF₄", "HF+KOH": "HF + KOH = KF + H₂O",
		"HF+Ba(OH)2": "2HF + Ba(OH)₂ = BaF₂ + 2H₂O", "HF+MgO": "2HF + MgO = MgF₂ + H₂O", "HF+Al2O3": "6HF + Al₂O₃ = 2AlF₃ + 3H₂O",
		"NaF+BaCl2": "2NaF + BaCl₂ = BaF₂↓ + 2NaCl", "KF+CaCl2": "2KF + CaCl₂ = CaF₂↓ + 2KCl", "Zn+O2": "2Zn + O₂ = 2ZnO",
		"Ag+O2": "4Ag + O₂ = 2Ag₂O", "Si+O2": "Si + O₂ = SiO₂", "Cu+O2": "2Cu + O₂ = 2CuO", "Al+O2": "4Al + 3O₂ = 2Al₂O₃",
		"Ba+O2": "2Ba + O₂ = 2BaO", "N2+O2": "N₂ + O₂ = 2NO", "Cl2+O2": "2Cl₂ + 7O₂ = 2Cl₂O₇", "K+Cl2": "2K + Cl₂ = 2KCl",
		"K+Br2": "2K + Br₂ = 2KBr", "K+I2": "2K + I₂ = 2KI", "K+S": "2K + S = K₂S", "Na+S": "2Na + S = Na₂S",
		"Mg+S": "Mg + S = MgS", "Ca+S": "Ca + S = CaS", "Ba+S": "Ba + S = BaS", "Al+S": "2Al + 3S = Al₂S₃",
		"Zn+S": "Zn + S = ZnS", "Ca+Cl2": "Ca + Cl₂ = CaCl₂", "Ba+Cl2": "Ba + Cl₂ = BaCl₂", "Zn+Cl2": "Zn + Cl₂ = ZnCl₂",
		"Ag+Cl2": "2Ag + Cl₂ = 2AgCl", "Hg+Cl2": "Hg + Cl₂ = HgCl₂", "Ba+Br2": "Ba + Br₂ = BaBr₂", "Ba+I2": "Ba + I₂ = BaI₂",
		"Hg+Br2": "Hg + Br₂ = HgBr₂", "Hg+I2": "Hg + I₂ = HgI₂", "Ag+Br2": "2Ag + Br₂ = 2AgBr", "Ag+I2": "2Ag + I₂ = 2AgI",
		"Cu+I2": "2Cu + I₂ = 2CuI", "P+Cl2": "2P + 3Cl₂ = 2PCl₃", "P+S": "2P + 5S = P₂S₅", "Mg+ZnCl2": "Mg + ZnCl₂ = MgCl₂ + Zn",
		"Mg+ZnSO4": "Mg + ZnSO₄ = MgSO₄ + Zn", "Mg+FeCl2": "Mg + FeCl₂ = MgCl₂ + Fe", "Mg+FeSO4": "Mg + FeSO₄ = MgSO₄ + Fe",
		"Mg+CuCl2": "Mg + CuCl₂ = MgCl₂ + Cu", "Mg+AlCl3": "3Mg + 2AlCl₃ = 3MgCl₂ + 2Al", "Mg+Al2(SO4)3": "3Mg + Al₂(SO₄)₃ = 3MgSO₄ + 2Al",
		"Zn+FeCl2": "Zn + FeCl₂ = ZnCl₂ + Fe", "Zn+FeSO4": "Zn + FeSO₄ = ZnSO₄ + Fe", "Zn+CuCl2": "Zn + CuCl₂ = ZnCl₂ + Cu",
		"Fe+CuCl2": "Fe + CuCl₂ = FeCl₂ + Cu", "Al+ZnCl2": "2Al + 3ZnCl₂ = 2AlCl₃ + 3Zn", "Al+FeCl2": "2Al + 3FeCl₂ = 2AlCl₃ + 3Fe",
		"Al+CuCl2": "2Al + 3CuCl₂ = 2AlCl₃ + 3Cu",
		"Mg+HNO3":  "Mg + 2HNO₃ = Mg(NO₃)₂ + H₂↑", "H2+Fe2O3": "3H₂ + Fe₂O₃ = 2Fe + 3H₂O", "Ca+H3PO4": "3Ca + 2H₃PO₄ = Ca₃(PO₄)₂↓ + 3H₂↑",
		"Al+H3PO4": "2Al + 2H₃PO₄ = 2AlPO₄↓ + 3H₂↑", "Al+HI": "2Al + 6HI = 2AlI₃ + 3H₂↑", "Al+HBr": "2Al + 6HBr = 2AlBr₃ + 3H₂↑",
		"Al+H2S": "2Al + 3H₂S = Al₂S₃ + 3H₂↑", "Al+H2SO3": "2Al + 3H₂SO₃ = Al₂(SO₃)₃ + 3H₂↑", "Zn+H2S": "Zn + H₂S = ZnS↓ + H₂↑",
		"Fe+H2SO3": "Fe + H₂SO₃ = FeSO₃ + H₂↑", "Mg+CuSO4": "Mg + CuSO₄ = MgSO₄ + Cu",
	}

	i := 0
	for reactants, display := range equations {
		i++
		parts := strings.Split(reactants, "->")
		var rList []string
		if len(parts) > 1 {
			rList = strings.Split(parts[0], "+")
		} else {
			rList = strings.Split(reactants, "+")
		}

		if len(rList) == 2 {
			r1, r2 := rList[0], rList[1]
			groupID := fmt.Sprintf("system-%d", i)
			_, _ = DB.Exec(`INSERT OR IGNORE INTO reactions (r1, r2, display, status, group_id, created_by) 
				VALUES (?, ?, ?, 'approved', ?, 100000000)`, r1, r2, display, groupID)
			_, _ = DB.Exec(`INSERT OR IGNORE INTO reactions (r1, r2, display, status, group_id, created_by) 
				VALUES (?, ?, ?, 'approved', ?, 100000000)`, r2, r1, display, groupID)
		}
	}

	// 创建默认全局牌组配置
	createDefaultDeck := `
	INSERT OR IGNORE INTO deck_configs (id, name, is_global, cards, created_by) 
	VALUES (1, '默认牌组', 1, '{"H":12,"O":12,"C":4,"N":4,"F":4,"Na":4,"Mg":4,"Al":4,"Si":4,"P":4,"S":4,"Cl":4,"K":4,"Ca":4,"Mn":4,"Fe":4,"Cu":4,"Zn":4,"Br":4,"I":4,"Ag":4,"+2":8,"+4":4,"He":1,"Ne":1,"Ar":1,"Kr":1,"Au":4}', 100000000);`

	_, _ = DB.Exec(createDefaultDeck)

	return nil
}

// createDefaultAdmin 创建默认管理员账户，只在用户不存在时创建
func createDefaultAdmin() error {
	// 检查admin用户是否已存在
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&count)
	if err != nil {
		return err
	}

	// 如果admin用户已存在，跳过创建
	if count > 0 {
		log.Println("管理员账户已存在，跳过创建")
		return nil
	}

	// 生成admin123的密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 创建新的admin用户
	createAdmin := `
	INSERT INTO users (UID, username, password, is_admin, role, avatar) 
	VALUES (100000000, 'admin', ?, 1, 'admin', '👑');`

	_, err = DB.Exec(createAdmin, string(hashedPassword))
	if err != nil {
		return err
	}

	log.Println("✅ 创建默认管理员账户: admin / admin123")
	return nil
}
func initSystemConfigs() {
	configs := []struct {
		key         string
		value       string
		description string
	}{
		{"points_decay_rate", "50", "每周积分衰减量"},
	}

	for _, cfg := range configs {
		_, _ = DB.Exec("INSERT OR IGNORE INTO system_configs (key, value, description) VALUES (?, ?, ?)", cfg.key, cfg.value, cfg.description)
	}
}

func GetConfig(key string, defaultValue string) string {
	var value string
	err := DB.QueryRow("SELECT value FROM system_configs WHERE key = ?", key).Scan(&value)
	if err != nil {
		return defaultValue
	}
	return value
}

func SetConfig(key string, value string) error {
	_, err := DB.Exec("UPDATE system_configs SET value = ?, updated_at = CURRENT_TIMESTAMP WHERE key = ?", value, key)
	return err
}

func GetAllConfigs() (map[string]interface{}, error) {
	rows, err := DB.Query("SELECT key, value, description FROM system_configs")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	configs := make(map[string]interface{})
	for rows.Next() {
		var key, value, description string
		if err := rows.Scan(&key, &value, &description); err != nil {
			return nil, err
		}
		configs[key] = map[string]string{
			"value":       value,
			"description": description,
		}
	}
	return configs, nil
}

func GetAllBounties() ([]models.Bounty, error) {
	rows, err := DB.Query("SELECT id, target_uid, amount, created_by, status, created_at FROM bounties WHERE status = 'active'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bounties []models.Bounty
	for rows.Next() {
		var b models.Bounty
		if err := rows.Scan(&b.ID, &b.TargetUID, &b.Amount, &b.CreatedBy, &b.Status, &b.CreatedAt); err != nil {
			return nil, err
		}
		bounties = append(bounties, b)
	}
	return bounties, nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
