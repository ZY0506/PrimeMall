package jwt

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

type Claims struct {
	UserID   uint64 `json:"user_id"`
	UserName string `json:"user_name"`
}

type MyClaims struct {
	Claims
	jwt.StandardClaims
}

// GenToken 生成token
func GenToken(data Claims, mySecret []byte, expire int64, issuer string) (token string, err error) {
	// 创建一个我们自己的声明
	c := MyClaims{
		data,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(expire) * time.Second).Unix(), // 过期时间
			Issuer:    issuer,                                                     // 签发人
		},
	}
	// 加密并获得完整的编码后的字符串token
	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(mySecret)
	return
}
