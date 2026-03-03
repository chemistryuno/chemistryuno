package utils

import (
	"errors"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtSecret     []byte
	jwtSecretOnce sync.Once
)

// initJWTSecret 初始化JWT密钥（延迟初始化，确保.env已加载）
func initJWTSecret() {
	jwtSecretOnce.Do(func() {
		secretKey := strings.TrimSpace(os.Getenv("JWT_SECRET"))
		if secretKey == "" {
			if err := EnsureJWTSecret(); err != nil {
				log.Fatalf("JWT_SECRET 未配置且自动生成失败: %v", err)
			}
			secretKey = strings.TrimSpace(os.Getenv("JWT_SECRET"))
		}
		if secretKey == "" {
			log.Fatal("JWT_SECRET 未配置，拒绝启动")
		} else if len(secretKey) < 32 {
			log.Println("警告: JWT_SECRET 长度过短，建议至少32个字符")
		}
		jwtSecret = []byte(secretKey)
		log.Printf("✓ JWT密钥已初始化（长度: %d）", len(secretKey))
	})
}

type Claims struct {
	UID     int    `json:"uid"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
	Role    string `json:"role"`
	SID     string `json:"sid,omitempty"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT Token
func GenerateToken(uid int, email string, isAdmin bool, role string, sid string) (string, error) {
	initJWTSecret() // 确保密钥已初始化

	claims := Claims{
		UID:     uid,
		Email:   email,
		IsAdmin: isAdmin,
		Role:    role,
		SID:     sid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken 解析JWT Token
func ParseToken(tokenString string) (*Claims, error) {
	initJWTSecret() // 确保密钥已初始化

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// StateClaims 用于 OAuth 状态校验
type StateClaims struct {
	Intent string `json:"intent"` // login or bind
	UID    int    `json:"uid,omitempty"`
	jwt.RegisteredClaims
}

// GenerateOAuthState 生成加密的状态字符串
func GenerateOAuthState(intent string, uid int) (string, error) {
	initJWTSecret()
	claims := StateClaims{
		Intent: intent,
		UID:    uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)), // 10分钟有效期
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// VerifyOAuthState 校验加密的状态字符串
func VerifyOAuthState(state string) (*StateClaims, error) {
	initJWTSecret()
	token, err := jwt.ParseWithClaims(state, &StateClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		log.Printf("[OAuthState校验失败] %v", err)
		return nil, err
	}
	if claims, ok := token.Claims.(*StateClaims); ok && token.Valid {
		return claims, nil
	}
	log.Printf("[OAuthState无效] Valid=%v", token.Valid)
	return nil, errors.New("invalid state")
}
