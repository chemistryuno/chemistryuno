package utils

import (
	"chemistryuno/database"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/smtp"
	"strconv"
	"time"
)

// 发送验证码（模拟发送，实际项目中需要配置 SMTP）
func SendEmailCode(email string, code string, codeType string) error {
	mockMode := database.GetConfig("email_mock_mode", "true")

	if mockMode == "true" {
		// 这里目前只模拟输出到控制台
		fmt.Printf("\n--- [EMAIL GATEWAY MOCK] ---\nTO: %s\nCODE: %s\nTYPE: %s\nEXPIRY: %s minutes\n---------------------------\n\n",
			email, code, codeType, database.GetConfig("email_verification_expiry", "10"))
		return nil
	}

	// 真实 SMTP 发送逻辑
	host := database.GetConfig("smtp_host", "")
	port := database.GetConfig("smtp_port", "587")
	user := database.GetConfig("smtp_user", "")
	pass := database.GetConfig("smtp_pass", "")
	from := database.GetConfig("smtp_from", "")

	if host == "" || user == "" || from == "" {
		return fmt.Errorf("SMTP 配置不完整，请在后台配置 SMTP 服务器")
	}

	subject := "Chemistry UNO 验证码"
	var body string
	switch codeType {
	case "register":
		body = fmt.Sprintf("欢迎加入化学 UNO 实验室。您的注册验证码为：%s。请在 %s 分钟内使用。",
			code, database.GetConfig("email_verification_expiry", "10"))
	case "login":
		body = fmt.Sprintf("您正在尝试登录化学 UNO 实验室。验证码为：%s。请在 %s 分钟内使用。",
			code, database.GetConfig("email_verification_expiry", "10"))
	case "reset":
		body = fmt.Sprintf("您正在重置化学 UNO 实验室的通行密钥。验证码为：%s。请在 %s 分钟内使用。如果不是本人操作，请忽略此邮件。",
			code, database.GetConfig("email_verification_expiry", "10"))
	default:
		body = fmt.Sprintf("您的验证码为：%s", code)
	}

	header := make(map[string]string)
	header["From"] = from
	header["To"] = email
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	auth := smtp.PlainAuth("", user, pass, host)
	err := smtp.SendMail(host+":"+port, auth, from, []string{email}, []byte(message))
	if err != nil {
		fmt.Printf("SMTP 发送失败: %v\n", err)
		return fmt.Errorf("发送邮件失败: %v", err)
	}

	return nil
}

// 生成验证码
func GenerateCode() string {
	lengthStr := database.GetConfig("email_verification_length", "6")
	length, _ := strconv.Atoi(lengthStr)
	if length <= 0 {
		length = 6
	}

	code := ""
	for i := 0; i < length; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		code += fmt.Sprintf("%d", n)
	}
	return code
}

// 保存验证码到数据库
func SaveVerificationCode(email string, code string, codeType string) error {
	expiryStr := database.GetConfig("email_verification_expiry", "10")
	expiry, _ := strconv.Atoi(expiryStr)
	if expiry <= 0 {
		expiry = 10
	}

	expiresAt := time.Now().Add(time.Duration(expiry) * time.Minute)
	_, err := database.DB.Exec(
		"INSERT INTO verification_codes (email, code, type, expires_at) VALUES (?, ?, ?, ?)",
		email, code, codeType, expiresAt,
	)
	return err
}

// 校验验证码
func VerifyEmailCode(email string, code string, codeType string) bool {
	var dbCode string
	err := database.DB.QueryRow(
		"SELECT code FROM verification_codes WHERE email = ? AND type = ? AND code = ? AND expires_at > ? ORDER BY id DESC LIMIT 1",
		email, codeType, code, time.Now(),
	).Scan(&dbCode)

	if err != nil {
		return false
	}

	// 消费验证码，防止重复使用
	database.DB.Exec("DELETE FROM verification_codes WHERE email = ? AND type = ?", email, codeType)

	return dbCode == code
}
