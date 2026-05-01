package utils

import (
	"crypto/rand"
	"fmt"
)

// GenerateRandomNickname 返回一个基于前缀或默认的随机昵称候选，符合 nicknameRegex 允许的字符集
// 例如: base="研究员" -> "研究员123456"；base=="" -> "研究员123456"
func GenerateRandomNickname(base string) string {
    if base == "" {
        base = "研究员"
    }
    // 生成 6 位数字后缀
    b := make([]byte, 4)
    if _, err := rand.Read(b); err != nil {
        // fallback
        return fmt.Sprintf("%s%d", base, 100000+int64(b[0])%900000)
    }
    // 将随机字节转成 6 位数字
    num := int64(b[0])<<16 | int64(b[1])<<8 | int64(b[2])
    suffix := 100000 + (num % 900000)
    return fmt.Sprintf("%s%d", base, suffix)
}
