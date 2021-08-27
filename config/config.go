package config

import (
	"fmt"
	"github.com/spf13/viper"
)

var Config ApplicationConfig

type ApplicationConfig struct {
	Application struct {
		Port string `json:"port"`
		Host string `json:"host"`
	} `json:"application"`
	Security struct {
		Salt      string `json:"salt"`
		AppSecret string `json:"app_secret"`
	} `json:"security"`
	Store struct {
		Root  string `json:"root"`
		Books string `json:"books"`
	} `json:"store"`
	Database struct {
		Type string `json:"type"`
	} `json:"database"`
	Sqlite struct {
		Path string `json:"path"`
	} `json:"sqlite"`
	Mysql struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database"`
	} `json:"mysql"`
	Thumbnail struct {
		Type   string `json:"type"`
		Target string `json:"target"`
	} `json:"thumbnail"`
	YouPlus struct {
		RPCUrl       string `json:"rpc"`
		Auth         bool   `json:"auth"`
		EntityConfig struct {
			Enable  bool   `json:"enable"`
			Name    string `json:"name"`
			Version int64  `json:"version"`
		} `json:"entity"`
	} `json:"youplus"`
}

func InitApplicationConfig() error {
	err := viper.Unmarshal(&Config)
	return err
}

type InitConfig struct {
	AdminAccount struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"adminaccount"`
	DefaultUserGroupName string `json:"defaultusergroupname"`
	Init                 bool   `json:"init"`
	SuperuserGroupName   string `json:"superusergroupname"`
}

func LoadConfig() {
	viper.AutomaticEnv()
	viper.SetDefault("APPLICATION_DEVELOP", false)
	developMode := viper.GetBool("APPLICATION_DEVELOP")
	if developMode {
		viper.SetConfigName("config.develop")
	} else {
		viper.SetConfigName("config")
	}
	viper.SetConfigType("json")
	viper.AddConfigPath("./conf")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	err = InitApplicationConfig()
	Config.YouPlus.RPCUrl = viper.GetString("youplus.rpc")
	Config.YouPlus.EntityConfig.Name = viper.GetString("youplus.entity.name")
	Config.YouPlus.EntityConfig.Enable = viper.GetBool("youplus.entity.enable")
	Config.YouPlus.EntityConfig.Version = viper.GetInt64("youplus.entity.version")
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	viper.WatchConfig()

}
