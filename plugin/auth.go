package plugin

import (
	"github.com/allentom/harukap/commons"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/module"
)

type BaseAuthPlugin struct {
}

func CreateBaseAuthPlugin() {
	module.Auth.Plugins = append(module.Auth.Plugins, &BaseAuthPlugin{})
}

func (p *BaseAuthPlugin) GetAuthInfo() (*commons.AuthInfo, error) {
	return &commons.AuthInfo{
		Name: "Account",
		Type: "base",
		Url:  "/user/auth",
	}, nil
}

func (p *BaseAuthPlugin) AuthName() string {
	return "Account"
}

func (p *BaseAuthPlugin) GetAuthUserByToken(accessToken string) (commons.AuthUser, error) {
	var oauthRecord model.Oauth
	err := database.Instance.Model(&model.Oauth{}).Preload("User").Where("access_token = ?", accessToken).
		Where("provider = ?", "self").
		Find(&oauthRecord).Error
	if err != nil {
		return nil, err
	}
	return oauthRecord.User, nil
}

func (p *BaseAuthPlugin) TokenTypeName() string {
	return "YouComic"
}
