package plugin

import (
	"github.com/allentom/harukap/commons"
	"github.com/allentom/harukap/plugins/youplus"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/module"
)

var DefaultYouPlusPlugin *youplus.Plugin

func CreateDefaultYouPlusPlugin() {
	DefaultYouPlusPlugin = &youplus.Plugin{}
	DefaultYouPlusPlugin.AuthFromToken = func(token string) (commons.AuthUser, error) {
		return GetUserByPlusAuthToken(token)
	}
	DefaultYouPlusPlugin.AuthUrl = "/oauth/youplus"
	module.Auth.Plugins = append(module.Auth.Plugins, DefaultYouPlusPlugin)
}
func GetUserByPlusAuthToken(accessToken string) (commons.AuthUser, error) {
	var oauthRecord model.Oauth
	err := database.Instance.Model(&model.Oauth{}).Preload("User").Where("access_token = ?", accessToken).
		Where("provider = ?", "YouPlusServer").
		Find(&oauthRecord).Error
	if err != nil {
		return nil, err
	}
	_, err = DefaultYouPlusPlugin.Client.CheckAuth(accessToken)
	if err != nil {
		return nil, err
	}
	return oauthRecord.User, nil
}
