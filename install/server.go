package install

import (
	"encoding/json"
	"github.com/allentom/haruka"
	"github.com/allentom/youcomic-api/config"
	"github.com/allentom/youcomic-api/log"
	"github.com/allentom/youcomic-api/utils"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

var LogField = log.Logger.WithField("scope", "install")

var PreinstallConfig config.ApplicationConfig
var InitUsername string
var InitPassword string

func checkNeedInstall() bool {
	_, err := os.Stat("conf/config.json")
	return err != nil
}
func createNewConfig() {
	PreinstallConfig = config.ApplicationConfig{}
}
func RunInstallServer() {
	skipInstall := os.Getenv("SKIP_INSTALL")
	if skipInstall == "True" {
		return
	}
	if !checkNeedInstall() {
		return
	}
	createNewConfig()
	r := haruka.NewEngine()
	r.Router.Static("/assets", "assets/install/static")
	r.Router.GET("/", IndexController)
	r.Router.GET("/database", SettingDatabaseController)
	r.Router.GET("/mysql", SettingMysqlController)
	r.Router.GET("/sqlite", SettingSqliteController)
	r.Router.POST("/sqlite", SettingSqliteSubmitController)
	r.Router.POST("/mysql", SettingMysqlSubmitController)
	r.Router.GET("/store", SettingStoreController)
	r.Router.POST("/store", SettingStoreSubmitController)
	r.Router.GET("/security", SettingSecurityController)
	r.Router.POST("/security", SettingSecuritySubmitController)
	r.Router.GET("/application", SettingApplicationController)
	r.Router.POST("/application", SettingApplicationSubmitController)
	r.Router.GET("/complete", SettingCompleteController)
	LogField.WithFields(logrus.Fields{
		"signal": "need_install",
		"port":   8880,
	}).Info("need install")
	open := os.Getenv("Open")
	if len(open) != 0 {
		utils.OpenBrowserWithURL("http://localhost:8880")
	}
	r.RunAndListen("localhost:8880")
}

func generateSettingFile() {
	PreinstallConfig.Application.Host = "0.0.0.0"
	PreinstallConfig.Security.AppSecret = utils.RandomString(32)
	PreinstallConfig.Security.Salt = utils.RandomString(32)
	file, _ := json.MarshalIndent(PreinstallConfig, "", " ")
	_ = ioutil.WriteFile("conf/config.json", file, 0644)

	initConfig := config.InitConfig{
		DefaultUserGroupName: "Default",
		Init:                 true,
		SuperuserGroupName:   "admin",
	}
	initConfig.AdminAccount.Username = InitUsername
	initConfig.AdminAccount.Password = InitPassword
	initConfigJson, _ := json.MarshalIndent(initConfig, "", " ")
	_ = ioutil.WriteFile("conf/init.json", initConfigJson, 0644)
}
