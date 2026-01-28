package database

import (
	"database/sql"
	"log"

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
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
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
		reactant TEXT NOT NULL,
		product TEXT NOT NULL,
		type TEXT NOT NULL,
		created_by INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (created_by) REFERENCES users(UID)
	);`

	tables := []string{userTable, deckConfigTable, gameHistoryTable, reactionsTable}

	for _, table := range tables {
		if _, err := DB.Exec(table); err != nil {
			return err
		}
	}

	// 增量更新表结构（针对已存在的数据库）
	_, _ = DB.Exec("ALTER TABLE users ADD COLUMN two_factor_enabled BOOLEAN DEFAULT 0")
	_, _ = DB.Exec("ALTER TABLE users ADD COLUMN two_factor_secret TEXT DEFAULT ''")

	// 检查并创建默认管理员账号
	if err := createDefaultAdmin(); err != nil {
		log.Printf("创建管理员账户时出错: %v", err)
	}

	// 插入默认化学反应数据
	insertReactions := `
	INSERT OR IGNORE INTO reactions (reactant, product, type, created_by) VALUES
	('H2SO4 + 2NaOH', 'Na2SO4 + 2H2O', 'acid-base', 100000000),
	('HCl + NaOH', 'NaCl + H2O', 'acid-base', 100000000),
	('Zn + 2HCl', 'ZnCl2 + H2', 'displacement', 100000000),
	('Cu + 2AgNO3', 'Cu(NO3)2 + 2Ag', 'displacement', 100000000),
	('2H2 + O2', '2H2O', 'synthesis', 100000000),
	('CaCO3 + heat', 'CaO + CO2', 'decomposition', 100000000),
	('AgNO3 + NaCl', 'AgCl + NaNO3', 'precipitation', 100000000),
	('2Mg + O2', '2MgO', 'combustion', 100000000);`

	_, _ = DB.Exec(insertReactions)

	// 创建默认全局牌组配置
	createDefaultDeck := `
	INSERT OR IGNORE INTO deck_configs (id, name, is_global, cards, created_by) 
	VALUES (1, '默认牌组', 1, '{"H":12,"O":12,"C":4,"N":4,"F":4,"Na":4,"Mg":4,"Al":4,"Si":4,"P":4,"S":4,"Cl":4,"K":4,"Ca":4,"Mn":4,"Fe":4,"Cu":4,"Zn":4,"Br":4,"I":4,"Ag":4,"+2":8,"+4":4,"He":1,"Ne":1,"Ar":1,"Kr":1,"Au":4,"Choice":4}', 100000000);`

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

func Close() {
	if DB != nil {
		DB.Close()
	}
}
