package main

import (
	"github.com/allentom/harukap"
	"github.com/allentom/harukap/cli"
	thumbnail2 "github.com/allentom/harukap/thumbnail"
	"github.com/projectxpolaris/youcomic/api/httpapi"
	"github.com/projectxpolaris/youcomic/config"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/plugin"
	"github.com/projectxpolaris/youcomic/thumbnail"
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
	appEngine.UsePlugin(&youplus.DefaultYouPlusPlugin)
	appEngine.UsePlugin(database.DefaultPlugin)
	appEngine.UsePlugin(&thumbnail.DefaultThumbnailServicePlugin)
	if config.Instance.Thumbnail.Type == "thumbnailservice" {
		thumbnail.DefaultThumbnailServicePlugin.SetConfig(&thumbnail2.ThumbnailServiceConfig{
			Enable:     true,
			ServiceUrl: config.Instance.Thumbnail.ServiceUrl,
		})
	}
	appEngine.UsePlugin(&plugin.DefaultRegisterPlugin)
	appEngine.UsePlugin(&plugin.InitPlugin{})
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
