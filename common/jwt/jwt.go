package jwt

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type Claims struct {
	UserId uint64 `json:"user_id"`
	JTI    string `json:"token_id"`
	Role   string `json:"role"`
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
			Id:        data.JTI,
		},
	}
	// 加密并获得完整的编码后的字符串token
	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(mySecret)
	return
}

func ParseToken(token string, mySecret []byte) (myClaim *MyClaims, err error) {
	var t = new(jwt.Token)
	var c = new(MyClaims)
	t, err = jwt.ParseWithClaims(token, c, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return mySecret, nil
	})
	if err != nil {
		return nil, err
	}
	// 防御性检查：理论上 ParseWithClaims 出错时 err != nil，但为了健壮性保留判断
	if !t.Valid {
		return c, fmt.Errorf("invalid token")
	}
	return c, nil
}
