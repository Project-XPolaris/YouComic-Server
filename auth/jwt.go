package auth

import (
	"errors"
	"github.com/allentom/haruka"
	"github.com/dgrijalva/jwt-go"
	"github.com/projectxpolaris/youcomic/config"
	"github.com/projectxpolaris/youcomic/model"
)

type UserClaims struct {
	jwt.StandardClaims
	UserId uint `json:"user_id"`
}

func ParseAuthHeader(c *haruka.Context) (*UserClaims, error) {
	jwtToken := c.Request.Header.Get("Authorization")
	if len(jwtToken) == 0 {
		jwtToken = c.GetQueryString("a")
	}
	if len(jwtToken) == 0 {
		return nil, errors.New("jwt token error")
	}
	var claims UserClaims
	_, err := jwt.ParseWithClaims(jwtToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Instance.Security.AppSecret), nil
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
	tokenString, err := token.SignedString([]byte(config.Instance.Security.AppSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func GetUserClaimsFromContext(context *haruka.Context) *UserClaims {
	contextClaims, _ := context.Param["claim"]
	claims := contextClaims.(*UserClaims)
	return claims
}
