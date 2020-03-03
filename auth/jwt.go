package auth

import (
	"errors"
	"github.com/allentom/youcomic-api/config"
	"github.com/allentom/youcomic-api/model"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type UserClaims struct {
	jwt.StandardClaims
	UserId uint `json:"user_id"`
}

func ParseAuthHeader(c *gin.Context) (*UserClaims, error) {
	jwtToken := c.GetHeader("Authorization")
	if len(jwtToken) == 0 {
		return nil, errors.New("jwt token error")
	}
	var claims UserClaims
	_, err := jwt.ParseWithClaims(jwtToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Config.Security.AppSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return &claims, nil
}
func GenerateJWTSign(user *model.User) (string, error) {
	now := jwt.TimeFunc()
	expire := now.AddDate(0, 0, 15)
	claims := &UserClaims{
		StandardClaims: jwt.StandardClaims{
			Audience:  user.Username,
			ExpiresAt: expire.Unix(),
			NotBefore: now.Unix(),
			IssuedAt:  now.Unix(),
			Issuer:    "YouComic",
			Subject:   "All",
		},
		UserId: user.ID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.Config.Security.AppSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
