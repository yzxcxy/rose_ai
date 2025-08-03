package jwt

import "github.com/golang-jwt/jwt/v4"

func GetJwtToken(secret string, iat, seconds int64, uid int64, username string) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims["uid"] = uid
	claims["username"] = username // 自定义载荷
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secret))
}
