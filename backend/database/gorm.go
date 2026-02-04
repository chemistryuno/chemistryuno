package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
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
		// 指定使用 modernc.org/sqlite 纯Go驱动（SQLite3 默认使用 UTF-8）
		dialector = sqlite.Dialector{
			DriverName: "sqlite",
			DSN:        sqlitePath,
		}
		log.Printf("📊 使用 SQLite 数据库 (纯Go, UTF-8): %s\n", sqlitePath)

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

// initRedis 初始化Redis客户端（可选功能）
func initRedis() {
	redisAddr := os.Getenv("REDIS_ADDR")

	// 默认尝试连接本地Redis
	if redisAddr == "" {
		// 尝试常见的Redis端口
		defaultAddrs := []string{"localhost:6379", "127.0.0.1:6379"}
		for _, addr := range defaultAddrs {
			if tryConnectRedis(addr, "") {
				redisAddr = addr
				log.Printf("✅ 自动连接到本地Redis: %s", addr)
				break
			}
		}

		if redisAddr == "" {
			log.Println("ℹ️  Redis未配置（缓存功能已禁用，不影响核心功能）")
			log.Println("💡 如需启用Redis，请设置环境变量: REDIS_ADDR=localhost:6379")
			RedisClient = nil
			return
		}
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0 // 使用默认数据库0

	// 从环境变量获取连接池配置（可选）
	poolSize := getEnvInt("REDIS_POOL_SIZE", 100)
	minIdleConns := getEnvInt("REDIS_MIN_IDLE_CONNS", 10)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		Password:     redisPassword,
		DB:           redisDB,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		MaxRetries:   1,
		DialTimeout:  1 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		PoolTimeout:  2 * time.Second,
	})

	// 测试连接
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := RedisClient.Ping(pingCtx).Err(); err != nil {
		log.Printf("⚠️  Redis连接失败: %v (缓存功能已禁用)", err)
		RedisClient.Close()
		RedisClient = nil
	} else {
		log.Printf("✅ Redis连接成功: %s", redisAddr)
	}
}

// tryConnectRedis 尝试连接Redis（快速检测）
func tryConnectRedis(addr, password string) bool {
	client := redis.NewClient(&redis.Options{
		Addr:        addr,
		Password:    password,
		DB:          0,
		DialTimeout: 500 * time.Millisecond,
	})
	defer client.Close()

	pingCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := client.Ping(pingCtx).Err()
	return err == nil
}

// getEnvInt 从环境变量获取整数值
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
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
