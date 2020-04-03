package install

import (
	"encoding/json"
	"github.com/allentom/youcomic-api/config"
	"github.com/allentom/youcomic-api/utils"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"os"
)

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
	r.Static("/assets", "install/static")
	r.LoadHTMLGlob("install/templates/*")
	r.GET("/", IndexController)
	r.GET("/database", SettingDatabaseController)
	r.POST("/database", SettingDatabaseSubmitController)
	r.GET("/store", SettingStoreController)
	r.POST("/store", SettingStoreSubmitController)
	r.GET("/security", SettingSecurityController)
	r.POST("/security", SettingSecuritySubmitController)
	r.GET("/application", SettingApplicationController)
	r.POST("/application", SettingApplicationSubmitController)
	r.GET("/complete", SettingCompleteController)
	r.Run(":3004")
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
