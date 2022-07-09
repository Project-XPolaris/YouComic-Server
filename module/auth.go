package module

import (
	"github.com/allentom/haruka"
	"github.com/allentom/harukap"
	"github.com/allentom/harukap/module/auth"
	"github.com/projectxpolaris/youcomic/config"
)

var Auth = &auth.AuthModule{
	Plugins: []harukap.AuthPlugin{},
}

func CreateAuthModule() {
	Auth.ConfigProvider = config.DefaultConfigProvider

	Auth.InitModule()
	Auth.AuthMiddleware.RequestFilter = func(c *haruka.Context) bool {
		noAuthPattern := []string{
			"/user/auth",
			"/user/auth2",
			"/oauth/youauth",
			"/oauth/token",
			"/info",
			"/oauth/youplus",
		}
		for _, pattern := range noAuthPattern {
			if c.Pattern == pattern {
				return false
			}
		}
		return true
	}
}
