package auth

import (
	"errors"
	"github.com/allentom/haruka"
	"github.com/dgrijalva/jwt-go"
	"github.com/projectxpolaris/youcomic/config"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/youauthplugin"
	"github.com/projectxpolaris/youcomic/youplus"
)

type JwtClaims interface {
	GetUserId() uint
}
type UserClaims struct {
	jwt.StandardClaims
	UserId uint `json:"user_id"`
}
type YouAuthClaims struct {
	UserId uint `json:"user_id"`
}

func (c *YouAuthClaims) GetUserId() uint {
	return c.UserId
}

func (c *UserClaims) GetUserId() uint {
	return c.UserId
}

func ParseAuthHeader(c *haruka.Context) (JwtClaims, error) {
	jwtToken := c.Request.Header.Get("Authorization")
	if len(jwtToken) == 0 {
		jwtToken = c.GetQueryString("a")
	}
	if len(jwtToken) == 0 {
		return nil, errors.New("jwt token error")
	}
	return ParseToken(jwtToken)
}
func ParseToken(jwtToken string) (JwtClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(jwtToken, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}
	mapClaims := token.Claims.(jwt.MapClaims)
	isu := mapClaims["iss"].(string)
	switch isu {
	case "youauth":
		user, err := GetUserByYouAuthToken(jwtToken)
		if err != nil {
			return nil, err
		}
		return &YouAuthClaims{
			UserId: user.ID,
		}, nil
	case "YouPlusService":
		user, err := GetUserByYouPlusToken(jwtToken)
		if err != nil {
			return nil, err
		}
		return &UserClaims{
			UserId: user.ID,
		}, nil
	default:
		var claims UserClaims
		_, err := jwt.ParseWithClaims(jwtToken, &claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.Instance.Security.AppSecret), nil
		})
		if err != nil {
			return nil, err
		}
		return &claims, nil
	}
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

func GetUserClaimsFromContext(context *haruka.Context) *model.User {
	contextClaims, _ := context.Param["claim"]
	claims := contextClaims.(*model.User)
	return claims
}
func GetUserByYouPlusToken(accessToken string) (*model.User, error) {
	var oauthRecord model.Oauth
	err := database.Instance.Model(&model.Oauth{}).Preload("User").Where("access_token = ?", accessToken).
		Where("provider = ?", "YouPlusServer").
		Find(&oauthRecord).Error
	if err != nil {
		return nil, err
	}
	_, err = youplus.DefaultYouPlusPlugin.Client.CheckAuth(accessToken)
	if err != nil {
		return nil, err
	}
	return oauthRecord.User, nil
}

func GetUserByYouAuthToken(accessToken string) (*model.User, error) {
	var oauthRecord model.Oauth
	err := database.Instance.Model(&model.Oauth{}).Preload("User").Where("access_token = ?", accessToken).
		Where("provider = ?", "youauth").
		Find(&oauthRecord).Error
	if err != nil {
		return nil, err
	}
	_, err = youauthplugin.DefaultYouAuthOauthPlugin.Client.GetCurrentUser(accessToken)
	if err != nil {
		return nil, err
	}
	return oauthRecord.User, nil
}
func GetUserByYouComicToken(accessToken string) (*model.User, error) {
	var oauthRecord model.Oauth
	err := database.Instance.Model(&model.Oauth{}).Preload("User").Where("access_token = ?", accessToken).
		Where("provider = ?", "self").
		Find(&oauthRecord).Error
	if err != nil {
		return nil, err
	}
	return oauthRecord.User, nil
}
