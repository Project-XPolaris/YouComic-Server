package module

import (
	"github.com/allentom/harukap"
	"github.com/allentom/harukap/module/auth"
	"github.com/projectxpolaris/youcomic/config"
)

var Auth = &auth.AuthModule{
	Plugins: []harukap.AuthPlugin{},
}

func CreateAuthModule() {
	Auth.ConfigProvider = config.DefaultConfigProvider
	Auth.NoAuthPath = []string{
		"/user/auth",
		"/user/auth2",
		"/oauth/youauth",
		"/oauth/token",
		"/info",
		"/oauth/youplus",
	}
	Auth.InitModule()
}
