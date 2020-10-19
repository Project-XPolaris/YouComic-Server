package main

import (
	"fmt"
	"github.com/allentom/youcomic-api/config"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/install"
	applogger "github.com/allentom/youcomic-api/log"
	"github.com/allentom/youcomic-api/middleware"
	"github.com/allentom/youcomic-api/router"
	"github.com/allentom/youcomic-api/setup"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
	"os"
	"time"
)

var MainLogger = applogger.Logger.WithField("scope", "main")

func main() {
	applogger.Logger.SetOutput(os.Stdout)
	applogger.Logger.SetFormatter(&logrus.JSONFormatter{})
	// run installer
	install.RunInstallServer()
	//load global application config
	config.LoadConfig()
	//prepare database
	setup.CheckDatabase()
	//connect to database
	database.ConnectDatabase()

	// set up application
	err := setup.SetupApplication()
	if err != nil {
		MainLogger.Fatalf("setup application with error of %s", err.Error())
	}

	//init gin
	r := gin.New()
	r.Use(location.Default())
	r.Use(gin.Recovery())
	gin.SetMode(gin.ReleaseMode)

	r.Use(ginlogrus.Logger(applogger.Logger), gin.Recovery())
	corsConfig := cors.Config{
		AllowMethods:           []string{"PUT", "PATCH", "GET", "POST", "PATCH", "DELETE", "OPTION"},
		AllowHeaders:           []string{"Origin", "Authorization", "Content-Type", "Access-Control-Allow-Origin"},
		ExposeHeaders:          []string{"Content-Length", "Authorization", "Content-Type", "Access-Control-Allow-Origin"},
		AllowAllOrigins:        true,
		AllowCredentials:       true,
		AllowWebSockets:        true,
		AllowBrowserExtensions: true,
		AllowWildcard:          true,
		AllowFiles:             true,
		MaxAge:                 12 * time.Hour,
	}
	r.Use(cors.New(corsConfig))
	r.Use(middleware.JWTAuth())
	r.Use(middleware.StaticRouter())
	r.Static("/assets", config.Config.Store.Root)
	router.SetRouter(r)

	MainLogger.WithFields(logrus.Fields{
		"signal": "start_success",
		"port":   config.Config.Application.Port,
	}).Info("Service start success!")
	err = r.Run(fmt.Sprintf("%s:%s", config.Config.Application.Host, config.Config.Application.Port))
	if err != nil {
		MainLogger.Fatalf("start gin service with error of %s", err.Error())
	}
}


