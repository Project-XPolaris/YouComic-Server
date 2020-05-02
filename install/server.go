package install

import (
	"encoding/json"
	"github.com/allentom/youcomic-api/config"
	"github.com/allentom/youcomic-api/log"
	"github.com/allentom/youcomic-api/utils"
	"github.com/gin-gonic/gin"
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
	r := gin.New()
	r.Use(gin.Recovery())
	gin.SetMode(gin.ReleaseMode)
	r.Static("/assets", "assets/install/static")
	r.LoadHTMLGlob("assets/install/templates/*")
	r.GET("/", IndexController)
	r.GET("/database", SettingDatabaseController)
	r.GET("/mysql", SettingMysqlController)
	r.GET("/sqlite", SettingSqliteController)
	r.POST("/sqlite", SettingSqliteSubmitController)
	r.POST("/mysql", SettingMysqlSubmitController)
	r.GET("/store", SettingStoreController)
	r.POST("/store", SettingStoreSubmitController)
	r.GET("/security", SettingSecurityController)
	r.POST("/security", SettingSecuritySubmitController)
	r.GET("/application", SettingApplicationController)
	r.POST("/application", SettingApplicationSubmitController)
	r.GET("/complete", SettingCompleteController)
	LogField.WithFields(logrus.Fields{
		"signal": "need_install",
		"port":8880,
	}).Info("need install")
	utils.OpenBrowserWithURL("http://localhost:8880")
	r.Run(":8880")
}

func generateSettingFile() {
	PreinstallConfig.Application.Host = "localhost"
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
