package main

import (
	"fmt"
	"github.com/allentom/youcomic-api/config"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/setup"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"time"
)

func main() {
	initConfig()
	database.ConnectDatabase()
	r := gin.Default()
	r.Use(location.Default())
	corsConfig := cors.Config{
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type","Access-Control-Allow-Origin"},
		ExposeHeaders:    []string{"Content-Length", "Authorization", "Content-Type","Access-Control-Allow-Origin"},
		AllowAllOrigins:  true,
		AllowCredentials: true,
		AllowFiles:       true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(corsConfig))
	r.Static("/assets", config.Config.Store.Root)
	SetRouter(r)
	setup.InitApplication()
	r.Run(fmt.Sprintf("%s:%s", config.Config.Application.Host, config.Config.Application.Port))
}

func initConfig() {
	viper.AutomaticEnv()
	viper.SetDefault("APPLICATION_DEVELOP", true)
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
	err = config.InitApplicationConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	viper.WatchConfig()

}
