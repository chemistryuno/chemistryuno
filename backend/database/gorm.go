package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite" // 纯Go的SQLite实现
)

var (
	DB          *gorm.DB
	LegacyDB    *LegacyDBWrapper // 旧代码兼容层
	RedisClient *redis.Client
	ctx         = context.Background()
)

// InitDB 初始化GORM数据库连接和Redis
func InitDB(dbPath string) error {
	// 从环境变量获取数据库类型（默认为sqlite）
	dbType := strings.ToLower(os.Getenv("DB_TYPE"))
	if dbType == "" {
		dbType = "sqlite"
	}

	// 配置GORM
	var err error
	var dialector gorm.Dialector

	switch dbType {
	case "mysql":
		// MySQL配置
		dsn := os.Getenv("MYSQL_DSN")
		if dsn == "" {
			dsn = "root:password@tcp(localhost:3306)/chemistryuno?charset=utf8mb4&parseTime=True&loc=Local"
		}
		dialector = mysql.Open(dsn)
		log.Println("📊 使用 MySQL 数据库")

	case "sqlite":
		// SQLite配置 - 使用纯Go实现（modernc.org/sqlite）
		sqlitePath := os.Getenv("SQLITE_PATH")
		if sqlitePath == "" {
			sqlitePath = "./chemistryuno.db"
		}
		// 指定使用 modernc.org/sqlite 纯Go驱动
		dialector = sqlite.Dialector{
			DriverName: "sqlite",
			DSN:        sqlitePath,
		}
		log.Printf("📊 使用 SQLite 数据库 (纯Go): %s\n", sqlitePath)

	default:
		return fmt.Errorf("不支持的数据库类型: %s（支持: mysql, sqlite）", dbType)
	}

	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 生产环境使用Silent
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt:            true, // 预编译SQL语句，提高性能
		SkipDefaultTransaction: true, // 跳过默认事务，提高写入性能
	})

	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}

	// 配置连接池以提高并发性能（仅对MySQL/PostgreSQL等网络数据库有效）
	if dbType == "mysql" {
		sqlDB, err := DB.DB()
		if err != nil {
			return fmt.Errorf("获取database实例失败: %v", err)
		}

		// 连接池配置
		sqlDB.SetMaxIdleConns(50)           // 最大空闲连接数
		sqlDB.SetMaxOpenConns(200)          // 最大打开连接数
		sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生命周期
	}

	// 初始化旧代码兼容层
	LegacyDB = GetLegacyDB()

	// 初始化Redis
	initRedis()

	// 自动迁移数据表
	if err := autoMigrate(); err != nil {
		return fmt.Errorf("数据库迁移失败: %v", err)
	}

	// 初始化默认数据
	if err := initDefaultData(); err != nil {
		return fmt.Errorf("初始化默认数据失败: %v", err)
	}

	log.Println("数据库初始化成功")
	return nil
}

// initRedis 初始化Redis客户端
func initRedis() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		// 如果没有配置Redis地址，跳过Redis初始化
		log.Println("⚠️ 未配置Redis，缓存功能已禁用")
		RedisClient = nil
		return
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0 // 使用默认数据库0

	RedisClient = redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		Password:     redisPassword,
		DB:           redisDB,
		PoolSize:     100,             // 连接池大小
		MinIdleConns: 10,              // 最小空闲连接
		MaxRetries:   1,               // 最大重试次数（减少重试）
		DialTimeout:  1 * time.Second, // 连接超时（减少超时时间）
		ReadTimeout:  1 * time.Second, // 读取超时
		WriteTimeout: 1 * time.Second, // 写入超时
		PoolTimeout:  2 * time.Second, // 连接池超时
	})

	// 测试连接（使用带超时的context）
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := RedisClient.Ping(pingCtx).Err(); err != nil {
		log.Printf("⚠️ Redis连接失败: %v (将继续运行，但缓存功能不可用)", err)
		RedisClient.Close()
		RedisClient = nil
	} else {
		log.Println("✅ Redis连接成功")
	}
}

// Close 关闭数据库连接
func Close() {
	if sqlDB, err := DB.DB(); err == nil {
		sqlDB.Close()
	}
	if RedisClient != nil {
		RedisClient.Close()
	}
}
