package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your-secret-key-change-this-in-production")

type Claims struct {
	UID      int    `json:"uid"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	Role     string `json:"role"`
	SID      string `json:"sid,omitempty"`
	jwt.RegisteredClaims
}

// 生成JWT Token
func GenerateToken(uid int, username string, isAdmin bool, role string, sid string) (string, error) {
	claims := Claims{
		UID:      uid,
		Username: username,
		IsAdmin:  isAdmin,
		Role:     role,
		SID:      sid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// 解析JWT Token
func ParseToken(tokenString string) (*Claims, error) {
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
