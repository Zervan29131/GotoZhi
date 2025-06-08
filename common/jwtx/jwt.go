package jwtx

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// GenToken 生成JWT令牌
// 参数:
//
//	uid: 用户唯一标识
//	exp: 令牌过期时间
//	signKey: 签名密钥
//
// 返回值:
//
//	生成的令牌字符串
//	错误对象，如果生成令牌过程中出现错误
func GenToken(uid int64, exp time.Time, signKey string) (string, error) {
	claims := jwt.MapClaims{
		"uid": uid, // 用户id
		"exp": exp.Unix(),
		"iat": time.Now().Unix(), // 签发时间
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(signKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
