package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Generate 签发 HS256 Token
func Generate(secret string, userID uint, expiresIn time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(expiresIn).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// Parse 解析并校验 Token，返回 user_id
func Parse(secret, tokenStr string) (uint, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		// 仅接受 HMAC 签名算法，防止算法混淆攻击
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("token 无效或已过期")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("token 声明解析失败")
	}

	// 安全类型断言，非法 token 直接拒绝而非 panic
	userID, ok := claims["user_id"].(float64)
	if !ok || userID <= 0 {
		return 0, errors.New("token 缺少 user_id")
	}
	return uint(userID), nil
}
