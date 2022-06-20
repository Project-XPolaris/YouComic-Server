package main

import (
	"fmt"
	"github.com/allentom/harukap"
	"github.com/allentom/harukap/cli"
	thumbnail2 "github.com/allentom/harukap/thumbnail"
	"github.com/projectxpolaris/youcomic/api/httpapi"
	"github.com/projectxpolaris/youcomic/config"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/module"
	"github.com/projectxpolaris/youcomic/plugin"
	"github.com/projectxpolaris/youcomic/thumbnail"
	"github.com/projectxpolaris/youcomic/youauthplugin"
	"github.com/projectxpolaris/youcomic/youlog"
	"github.com/projectxpolaris/youcomic/youplus"
	"github.com/sirupsen/logrus"
)

func main() {
	err := config.InitConfigProvider()
	if err != nil {
		logrus.Fatal(err)
	}
	err = youlog.DefaultYouLogPlugin.OnInit(config.DefaultConfigProvider)
	if err != nil {
		logrus.Fatal(err)
	}
	appEngine := harukap.NewHarukaAppEngine()
	appEngine.ConfigProvider = config.DefaultConfigProvider
	appEngine.LoggerPlugin = youlog.DefaultYouLogPlugin
	youplus.CreateDefaultYouPlusPlugin()
	appEngine.UsePlugin(youplus.DefaultYouPlusPlugin)
	appEngine.UsePlugin(database.DefaultPlugin)
	appEngine.UsePlugin(&thumbnail.DefaultThumbnailServicePlugin)
	if config.Instance.Thumbnail.Type == "thumbnailservice" {
		thumbnail.DefaultThumbnailServicePlugin.SetConfig(&thumbnail2.ThumbnailServiceConfig{
			Enable:     true,
			ServiceUrl: config.Instance.Thumbnail.ServiceUrl,
		})
	}
	rawAuth := config.DefaultConfigProvider.Manager.GetStringMap("auth")
	for key, _ := range rawAuth {
		rawAuthContent := config.DefaultConfigProvider.Manager.GetString(fmt.Sprintf("auth.%s.type", key))
		if rawAuthContent == "youauth" {
			youauthplugin.CreateYouAuthPlugin()
			youauthplugin.DefaultYouAuthOauthPlugin.ConfigPrefix = fmt.Sprintf("auth.%s", key)
			appEngine.UsePlugin(youauthplugin.DefaultYouAuthOauthPlugin)
		}
	}
	appEngine.UsePlugin(&plugin.DefaultRegisterPlugin)
	module.CreateAuthModule()
	appEngine.UsePlugin(&plugin.InitPlugin{})
	plugin.CreateBaseAuthPlugin()
	appEngine.HttpService = httpapi.GetEngine()
	if err != nil {
		logrus.Fatal(err)
	}
	appWrap, err := cli.NewWrapper(appEngine)
	if err != nil {
		logrus.Fatal(err)
	}
	appWrap.RunApp()
}
