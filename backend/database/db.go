package database

import (
	"database/sql"
	"log"

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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		avatar TEXT DEFAULT '',
		is_admin BOOLEAN DEFAULT 0,
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
		FOREIGN KEY (created_by) REFERENCES users(id)
	);`

	// 游戏记录表
	gameHistoryTable := `
	CREATE TABLE IF NOT EXISTS game_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		room_id TEXT NOT NULL,
		winner_id INTEGER,
		players TEXT NOT NULL,
		started_at DATETIME,
		finished_at DATETIME,
		FOREIGN KEY (winner_id) REFERENCES users(id)
	);`

	tables := []string{userTable, deckConfigTable, gameHistoryTable}

	for _, table := range tables {
		if _, err := DB.Exec(table); err != nil {
			return err
		}
	}

	// 创建默认管理员账号 (用户名: admin, 密码: admin123)
	// 密码hash: admin123
	createAdmin := `
	INSERT OR IGNORE INTO users (id, username, password, is_admin, avatar) 
	VALUES (1, 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 1, '👑');`

	_, _ = DB.Exec(createAdmin)

	// 创建默认全局牌组配置
	createDefaultDeck := `
	INSERT OR IGNORE INTO deck_configs (id, name, is_global, cards, created_by) 
	VALUES (1, '默认牌组', 1, '{"H":12,"O":12,"C":4,"N":4,"F":4,"Na":4,"Mg":4,"Al":4,"Si":4,"P":4,"S":4,"Cl":4,"K":4,"Ca":4,"Mn":4,"Fe":4,"Cu":4,"Zn":4,"Br":4,"I":4,"Ag":4,"+2":8,"+4":4,"He":1,"Ne":1,"Ar":1,"Kr":1,"Au":4}', 1);`

	_, _ = DB.Exec(createDefaultDeck)

	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
