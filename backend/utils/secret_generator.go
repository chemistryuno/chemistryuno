package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// GenerateRandomSecret 生成指定长度的随机密钥
func GenerateRandomSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes)[:length], nil
}

// EnsureJWTSecret 确保JWT密钥存在，如果不存在则生成并保存
func EnsureJWTSecret() error {
	// 查找.env文件路径
	envPath := ".env"

	// 检查.env文件是否存在
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		// 如果.env不存在，尝试从.env.example复制
		examplePath := ".env.example"
		if _, err := os.Stat(examplePath); err == nil {
			content, err := os.ReadFile(examplePath)
			if err != nil {
				return fmt.Errorf("读取 .env.example 失败: %v", err)
			}
			if err := os.WriteFile(envPath, content, 0644); err != nil {
				return fmt.Errorf("创建 .env 文件失败: %v", err)
			}
			log.Println("已从 .env.example 创建 .env 文件")
		}
	}

	// 读取现有.env内容
	var content []byte
	contentExists := false
	if _, err := os.Stat(envPath); err == nil {
		content, err = os.ReadFile(envPath)
		if err != nil {
			return fmt.Errorf("读取 .env 文件失败: %v", err)
		}
		contentExists = true
	}

	contentStr := strings.ReplaceAll(string(content), "\r\n", "\n")
	lines := strings.Split(contentStr, "\n")
	found := false
	hasValidSecret := false

	// 查找JWT_SECRET行
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		lines[i] = trimmed // 顺便清理所有行的首尾空格
		if strings.HasPrefix(trimmed, "JWT_SECRET=") {
			// 检查是否有有效的密钥值
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 && len(strings.TrimSpace(parts[1])) > 0 {
				found = true
				hasValidSecret = true
				log.Println("JWT_SECRET 已在 .env 文件中配置")
				break
			}
		}
	}

	// 如果已有有效密钥，无需生成
	if hasValidSecret {
		return nil
	}

	// 生成50位随机密钥
	secret, err := GenerateRandomSecret(50)
	if err != nil {
		return fmt.Errorf("生成随机密钥失败: %v", err)
	}

	// 查找并替换JWT_SECRET行
	if found {
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "JWT_SECRET=") || strings.HasPrefix(trimmed, "#JWT_SECRET=") {
				lines[i] = fmt.Sprintf("JWT_SECRET=%s", secret)
				break
			}
		}
	} else {
		// 如果没找到，添加到文件末尾
		if contentExists && len(contentStr) > 0 && !strings.HasSuffix(contentStr, "\n") {
			lines = append(lines, "")
		}
		lines = append(lines, "# 自动生成的JWT密钥 - 请勿共享")
		lines = append(lines, fmt.Sprintf("JWT_SECRET=%s", secret))
	}

	// 写回文件
	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(envPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("写入 .env 文件失败: %v", err)
	}

	// 设置到当前环境变量
	os.Setenv("JWT_SECRET", secret)

	absPath, _ := filepath.Abs(envPath)
	log.Printf("✓ 已自动生成并保存50位JWT密钥到 %s", absPath)
	log.Println("⚠ 重要提示: 请妥善保管 .env 文件，不要将其提交到版本控制系统")

	return nil
}
