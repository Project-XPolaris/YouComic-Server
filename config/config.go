package config

import "github.com/spf13/viper"

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
	Sqlite struct{
		Path string `json:"path"`
	}`json:"sqlite"`
	Mysql struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database"`
	} `json:"mysql"`
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
