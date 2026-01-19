package myjwt

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/mou-he/graduation-design/config"
)

// 自定义声明结构体
type Claims struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// 生成token
func GenerateToken(id int64, username string) (string, error) {
	claims := Claims{
		ID:       id,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			//
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.GetConfig().ExpireDuration) * time.Hour)),
			Issuer:    config.GetConfig().Issuer,
			Subject:   config.GetConfig().Subject,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	// 生成token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 签名token
	return token.SignedString([]byte(config.GetConfig().Key))
}

// 解析token
func ParseToken(token string) (string, bool) {
	claims := new(Claims)
	// 解析token
	t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.GetConfig().Key), nil
	})

	if !t.Valid || err != nil || claims == nil {
		return "", false
	}
	return claims.Username, true
}
