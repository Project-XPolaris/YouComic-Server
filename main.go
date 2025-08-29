package main

import (
	"fmt"

	"github.com/allentom/harukap"
	"github.com/allentom/harukap/cli"
	"github.com/allentom/harukap/plugins/nacos"
	"github.com/projectxpolaris/youcomic/api/httpapi"
	"github.com/projectxpolaris/youcomic/boot"
	"github.com/projectxpolaris/youcomic/config"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/module"
	"github.com/projectxpolaris/youcomic/plugin"
	"github.com/sirupsen/logrus"
)

func main() {
	err := config.InitConfigProvider()
	if err != nil {
		logrus.Fatal(err)
	}
	err = plugin.DefaultYouLogPlugin.OnInit(config.DefaultConfigProvider)
	if err != nil {
		logrus.Fatal(err)
	}
	appEngine := harukap.NewHarukaAppEngine()
	appEngine.ConfigProvider = config.DefaultConfigProvider
	appEngine.LoggerPlugin = plugin.DefaultYouLogPlugin
	plugin.CreateDefaultYouPlusPlugin()
	appEngine.UsePlugin(plugin.DefaultYouPlusPlugin)
	appEngine.UsePlugin(database.DefaultPlugin)
	// init nacos (optional) BEFORE youauth plugins, so they can discover it
	logger := plugin.DefaultYouLogPlugin.Logger.NewScope("main")
	nacosPlugin, err := nacos.NewNacosPluginFromYAML(appEngine.ConfigProvider, appEngine.ConfigProvider.Manager.GetString("application"), 7600)
	if err == nil && nacosPlugin != nil {
		plugin.DefaultNacosPlugin = nacosPlugin
		appEngine.UsePlugin(nacosPlugin)
	} else if err != nil {
		logger.Info(fmt.Sprintf("init nacos plugin failed: %v", err))
	}
	rawAuth := config.DefaultConfigProvider.Manager.GetStringMap("auth")
	for key, _ := range rawAuth {
		rawAuthContent := config.DefaultConfigProvider.Manager.GetString(fmt.Sprintf("auth.%s.type", key))
		if rawAuthContent == "youauth" {
			plugin.CreateYouAuthPlugin()
			plugin.DefaultYouAuthOauthPlugin.ConfigPrefix = fmt.Sprintf("auth.%s", key)
			appEngine.UsePlugin(plugin.DefaultYouAuthOauthPlugin)
		}
	}
	appEngine.UsePlugin(&plugin.DefaultRegisterPlugin)
	module.CreateAuthModule()
	appEngine.UsePlugin(&boot.InitPlugin{})
	appEngine.UsePlugin(plugin.LLM)
	appEngine.UsePlugin(plugin.StorageEngine)
	appEngine.UsePlugin(plugin.ThumbnailEngine)
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
