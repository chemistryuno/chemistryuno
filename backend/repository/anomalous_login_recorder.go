package repository

import (
	"chemistryuno/backend/database"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// AnomalousLoginRecorder 异常登录记录器
type AnomalousLoginRecorder struct {
	db *gorm.DB
}

// NewAnomalousLoginRecorder 创建新的异常登录记录器
func NewAnomalousLoginRecorder(db *gorm.DB) *AnomalousLoginRecorder {
	return &AnomalousLoginRecorder{db: db}
}

// RecordAnomalousLogin 记录异常登录并自动创建反馈
func (r *AnomalousLoginRecorder) RecordAnomalousLogin(uid uint, ip string, anomalies []string) error {
	if len(anomalies) == 0 {
		return nil
	}

	// 生成异常登录的详细描述
	content := "🔐 检测到异常登录尝试：\n\n"
	for i, anomaly := range anomalies {
		content += fmt.Sprintf("%d. %s\n", i+1, anomaly)
	}
	content += fmt.Sprintf("\n🌐 登录IP: %s\n", ip)
	content += fmt.Sprintf("⏰ 登录时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	content += "\n如果这不是您的登录，请立即修改密码！"

	// 创建系统反馈记录
	feedback := database.Feedback{
		UserUID:        uid,
		Content:        content,
		Type:           "system_alert",
		Status:         "unread",
		ProcessedByUID: nil,
		ProcessedAt:    nil,
		CreatedAt:      time.Now(),
	}

	if err := r.db.Create(&feedback).Error; err != nil {
		log.Printf("❌ 创建异常登录反馈失败 UID=%d: %v", uid, err)
		return err
	}

	log.Printf("📧 异常登录反馈已创建 - UID=%d, 异常类型: %v", uid, anomalies)
	return nil
}

// RecordLoginFailureAlert 记录多次登录失败的警告
func (r *AnomalousLoginRecorder) RecordLoginFailureAlert(uid uint, failureCount int, ip string) error {
	content := fmt.Sprintf("⚠️ 登录安全警告\n\n"+
		"您的账户在最近短时间内尝试登录失败了 %d 次。\n"+
		"来源IP: %s\n"+
		"时间: %s\n\n"+
		"如果这不是您的行为，请立即修改密码以保护账户安全！",
		failureCount, ip, time.Now().Format("2006-01-02 15:04:05"))

	feedback := database.Feedback{
		UserUID:        uid,
		Content:        content,
		Type:           "security_alert",
		Status:         "unread",
		ProcessedByUID: nil,
		ProcessedAt:    nil,
		CreatedAt:      time.Now(),
	}

	if err := r.db.Create(&feedback).Error; err != nil {
		log.Printf("❌ 创建登录失败警告反馈失败 UID=%d: %v", uid, err)
		return err
	}

	log.Printf("📧 登录失败警告反馈已创建 - UID=%d, 失败次数: %d", uid, failureCount)
	return nil
}

// CleanupOldAnomalousLoginRecords 清理旧的异常登录记录（超过30天）
func (r *AnomalousLoginRecorder) CleanupOldAnomalousLoginRecords() error {
	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)

	result := r.db.
		Where("type IN (?, ?) AND created_at < ?", "system_alert", "security_alert", thirtyDaysAgo).
		Delete(&database.Feedback{})

	if result.Error != nil {
		log.Printf("❌ 清理旧的异常登录记录失败: %v", result.Error)
		return result.Error
	}

	if result.RowsAffected > 0 {
		log.Printf("✅ 清理了 %d 条旧的异常登录反馈记录", result.RowsAffected)
	}

	return nil
}

// CheckRecentAnomalousLogins 检查用户最近是否有异常登录提醒
func (r *AnomalousLoginRecorder) CheckRecentAnomalousLogins(uid uint, withinHours int) (bool, error) {
	var count int64
	err := r.db.
		Where("user_uid = ? AND type IN (?, ?) AND created_at > ?",
			uid, "system_alert", "security_alert",
			time.Now().Add(-time.Duration(withinHours)*time.Hour)).
		Model(&database.Feedback{}).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
