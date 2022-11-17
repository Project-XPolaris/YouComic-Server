package plugin

import (
	"github.com/allentom/harukap/commons"
	"github.com/allentom/harukap/plugins/youauth"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/module"
)

var DefaultYouAuthOauthPlugin = &youauth.OauthPlugin{}

func CreateYouAuthPlugin() {
	DefaultYouAuthOauthPlugin = &youauth.OauthPlugin{}
	DefaultYouAuthOauthPlugin.AuthFromToken = func(token string) (commons.AuthUser, error) {
		return GetUserByYouAuthToken(token)
	}
	DefaultYouAuthOauthPlugin.PasswordAuthUrl = "/oauth/youauth/password"
	module.Auth.Plugins = append(module.Auth.Plugins, DefaultYouAuthOauthPlugin.GetOauthPlugin(), DefaultYouAuthOauthPlugin.GetPasswordPlugin())
}
func GetUserByYouAuthToken(accessToken string) (commons.AuthUser, error) {
	var oauthRecord model.Oauth
	err := database.Instance.Model(&model.Oauth{}).Preload("User").Where("access_token = ?", accessToken).
		Where("provider = ?", "youauth").
		Find(&oauthRecord).Error
	if err != nil {
		return nil, err
	}
	_, err = DefaultYouAuthOauthPlugin.Client.GetCurrentUser(accessToken)
	if err != nil {
		return nil, err
	}
	return oauthRecord.User, nil
}
