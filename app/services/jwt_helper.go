package services

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/goravel/framework/facades"
)

// GeneratePermanentToken 为永久token用户生成一个永不过期的JWT token
// 注意：为了兼容框架的guard，我们设置一个很远的过期时间（100年后）
// 这样框架的guard能够识别token，同时逻辑上视为永久有效
func GeneratePermanentToken(userID uint) (string, error) {
	secret := facades.Config().GetString("jwt.secret")
	if secret == "" {
		return "", jwt.ErrInvalidKey
	}

	// 创建token claims
	// 设置一个很远的过期时间（100年后），这样框架的guard能够识别token
	// 同时逻辑上视为永久有效（因为100年足够长）
	now := time.Now()
	expiresAt := now.AddDate(100, 0, 0) // 100年后过期

	claims := jwt.MapClaims{
		"key": userID,
		"sub": "admin",
		"iat": now.Unix(),
		"exp": expiresAt.Unix(), // 设置一个很远的过期时间，让框架能够识别
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名token
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

