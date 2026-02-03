package database

import (
	"database/sql"
	"fmt"

	"gorm.io/gorm"
)

// 兼容层：为旧代码提供database/sql接口
// 这样可以逐步迁移代码而不需要一次性重写所有SQL查询

// LegacyDB 提供旧的*sql.DB接口
type LegacyDBWrapper struct {
	*sql.DB
}

// Query 执行查询并返回多行结果
func (w *LegacyDBWrapper) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return w.DB.Query(query, args...)
}

// QueryRow 执行查询并返回单行结果  
func (w *LegacyDBWrapper) QueryRow(query string, args ...interface{}) *sql.Row {
	return w.DB.QueryRow(query, args...)
}

// Exec 执行SQL语句（INSERT, UPDATE, DELETE）
func (w *LegacyDBWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	return w.DB.Exec(query, args...)
}

// GetLegacyDB 返回兼容旧代码的数据库包装器
func GetLegacyDB() *LegacyDBWrapper {
	sqlDB, _ := DB.DB()
	return &LegacyDBWrapper{DB: sqlDB}

// GetDB 返回GORM数据库实例
func GetDB() *gorm.DB {
	return DB
}

// GetSQLDB 返回底层*sql.DB实例
func GetSQLDB() (*sql.DB, error) {
	return DB.DB()
}

// Transaction 使用GORM事务
func Transaction(fc func(*gorm.DB) error) error {
	return DB.Transaction(fc)
}

// WithContext 创建带上下文的数据库会话
func WithContext(ctx interface{}) *gorm.DB {
	// 如果需要context支持，可以在这里实现
	return DB
}

// 便捷查询方法

// GetConfigGORM 使用GORM获取系统配置
func GetConfigGORM(key string, defaultValue string) string {
	var config SystemConfig
	if err := DB.Where("`key` = ?", key).First(&config).Error; err != nil {
		return defaultValue
	}
	return config.Value
}

// SetConfigGORM 使用GORM设置系统配置
func SetConfigGORM(key string, value string) error {
	return DB.Where("`key` = ?", key).
		Assign(SystemConfig{Value: value}).
		FirstOrCreate(&SystemConfig{Key: key}).Error
}

// 兼容旧代码的查询函数（逐步迁移）

// GetConfig 兼容旧API
func GetConfig(key string, defaultValue string) string {
	return GetConfigGORM(key, defaultValue)
}

// SetConfig 兼容旧API
func SetConfig(key string, value string) error {
	return SetConfigGORM(key, value)
}

// GetAllConfigs 获取所有配置
func GetAllConfigs() (map[string]interface{}, error) {
	var configs []SystemConfig
	if err := DB.Find(&configs).Error; err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, config := range configs {
		result[config.Key] = config.Value
	}
	return result, nil
}

// columnExists 检查列是否存在（GORM不需要，保留兼容）
func columnExists(tableName, columnName string) bool {
	// GORM会自动处理列的存在性，这个函数主要用于兼容旧代码
	sqlDB, err := DB.DB()
	if err != nil {
		return false
	}

	var count int
	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM information_schema.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = '%s' 
		AND COLUMN_NAME = '%s'
	`, tableName, columnName)

	row := sqlDB.QueryRow(query)
	row.Scan(&count)
	return count > 0
}
