package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB          *gorm.DB
	RedisClient *redis.Client
	ctx         = context.Background()
)

// InitDB 初始化GORM数据库连接和Redis
func InitDB(dbPath string) error {
	// 从环境变量获取MySQL DSN
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:password@tcp(localhost:3306)/chemistryuno?charset=utf8mb4&parseTime=True&loc=Local"
	}

	// 配置GORM
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
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

	// 配置连接池以提高并发性能
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取database实例失败: %v", err)
	}

	// 连接池配置
	sqlDB.SetMaxIdleConns(50)           // 最大空闲连接数
	sqlDB.SetMaxOpenConns(200)          // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生命周期

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
		redisAddr = "localhost:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0 // 使用默认数据库0

	RedisClient = redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		Password:     redisPassword,
		DB:           redisDB,
		PoolSize:     100,             // 连接池大小
		MinIdleConns: 10,              // 最小空闲连接
		MaxRetries:   3,               // 最大重试次数
		DialTimeout:  5 * time.Second, // 连接超时
		ReadTimeout:  3 * time.Second, // 读取超时
		WriteTimeout: 3 * time.Second, // 写入超时
		PoolTimeout:  4 * time.Second, // 连接池超时
	})

	// 测试连接
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		log.Printf("警告: Redis连接失败: %v (将继续运行，但缓存功能不可用)", err)
		RedisClient = nil
	} else {
		log.Println("Redis连接成功")
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
