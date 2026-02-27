package utils

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"math/big"
	"net/smtp"
	"os"
	"strings"
)

// IsSMTPConfigured 检查SMTP是否已配置
func IsSMTPConfigured() bool {
	return strings.TrimSpace(os.Getenv("SMTP_HOST")) != "" &&
		strings.TrimSpace(os.Getenv("SMTP_PORT")) != "" &&
		strings.TrimSpace(os.Getenv("SMTP_USER")) != "" &&
		strings.TrimSpace(os.Getenv("SMTP_PASS")) != ""
}

// SendEmail 发送邮件
func SendEmail(to, subject, body string) error {
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	port := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	user := strings.TrimSpace(os.Getenv("SMTP_USER"))
	pass := strings.TrimSpace(os.Getenv("SMTP_PASS"))
	from := strings.TrimSpace(os.Getenv("SMTP_FROM"))
	if from == "" {
		from = user
	}

	auth := smtp.PlainAuth("", user, pass, host)

	// 构建邮件内容
	header := make(map[string]string)
	header["From"] = from
	header["To"] = to
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// 发送邮件 (处理可能需要的 TLS)
	addr := fmt.Sprintf("%s:%s", host, port)

	// 如果端口是 465，通常需要 SSL/TLS
	if port == "465" {
		tlsconfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         host,
		}
		conn, err := tls.Dial("tcp", addr, tlsconfig)
		if err != nil {
			return err
		}
		client, err := smtp.NewClient(conn, host)
		if err != nil {
			return err
		}
		if err = client.Auth(auth); err != nil {
			return err
		}
		if err = client.Mail(from); err != nil {
			return err
		}
		if err = client.Rcpt(to); err != nil {
			return err
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(message))
		if err != nil {
			return err
		}
		err = w.Close()
		if err != nil {
			return err
		}
		return client.Quit()
	}

	// 否则尝试普通 SMTP 发送（可能包含 STARTTLS）
	err := smtp.SendMail(addr, auth, from, []string{to}, []byte(message))
	return err
}

// GenerateVerificationCode 生成6位数字验证码
func GenerateVerificationCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
