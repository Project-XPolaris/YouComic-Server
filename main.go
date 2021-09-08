package main

import (
	"context"
	"fmt"
	"github.com/allentom/youcomic-api/config"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/install"
	applogger "github.com/allentom/youcomic-api/log"
	"github.com/allentom/youcomic-api/middleware"
	"github.com/allentom/youcomic-api/router"
	"github.com/allentom/youcomic-api/setup"
	util "github.com/allentom/youcomic-api/utils"
	"github.com/allentom/youcomic-api/youplus"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	srv "github.com/kardianos/service"
	entry "github.com/project-xpolaris/youplustoolkit/youplus/entity"
	"github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"path/filepath"
	"time"
)

var MainLogger = applogger.Logger.WithField("scope", "main")
var svcConfig *srv.Config

func initService() error {
	workPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}
	svcConfig = &srv.Config{
		Name:             "YouComicCoreService",
		DisplayName:      "YouComic Core Service",
		WorkingDirectory: workPath,
		Arguments:        []string{"run"},
	}
	return nil
}
func Program() {
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

	// register rpc
	if len(config.Config.YouPlus.RPCUrl) > 0 {
		err = youplus.LoadYouPlusRPCClient()
		if err != nil {
			MainLogger.WithFields(logrus.Fields{
				"signal": "rpc_connect_failed",
				"url":    config.Config.YouPlus.RPCUrl,
			}).Fatal("RPC connect success")
		}
		MainLogger.WithFields(logrus.Fields{
			"signal": "rpc_connect_success",
			"port":   config.Config.YouPlus.RPCUrl,
		}).Info("RPC connect success")
	}
	// youplus entity
	if config.Config.YouPlus.EntityConfig.Enable {
		youplus.InitEntity()
		err = youplus.DefaultEntry.Register()
		if err != nil {
			MainLogger.WithFields(logrus.Fields{
				"signal": "entity_register_failed",
			}).Fatal("entity register failed")
		}

		addrs, err := util.GetHostIpList()
		urls := make([]string, 0)
		for _, addr := range addrs {
			urls = append(urls, fmt.Sprintf("http://%s:%s", addr, config.Config.Application.Port))
		}
		if err != nil {
			MainLogger.Fatal(err.Error())
		}
		err = youplus.DefaultEntry.UpdateExport(entry.EntityExport{Urls: urls, Extra: map[string]interface{}{}})
		if err != nil {
			MainLogger.Fatal(err.Error())
		}

		err = youplus.DefaultEntry.StartHeartbeat(context.Background())
		if err != nil {
			MainLogger.Fatal(err.Error())
		}
		MainLogger.Info("success register entity")

	}
	MainLogger.WithFields(logrus.Fields{
		"signal": "start_success",
		"port":   config.Config.Application.Port,
	}).Info("Service start success!")
	err = r.Run(fmt.Sprintf("%s:%s", config.Config.Application.Host, config.Config.Application.Port))
	if err != nil {
		MainLogger.Fatalf("start gin service with error of %s", err.Error())
	}
}

type program struct{}

func (p *program) Start(s srv.Service) error {
	// Start should not block. Do the actual work async.
	go Program()
	return nil
}

func (p *program) Stop(s srv.Service) error {
	// Stop should not block. Return with a few seconds.
	return nil
}

func InstallAsService() {
	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	s.Uninstall()

	err = s.Install()
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("successful install service")
}

func UnInstall() {

	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	s.Uninstall()
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("successful uninstall service")
}
func StartService() {
	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	err = s.Start()
	if err != nil {
		logrus.Fatal(err)
	}
}
func StopService() {
	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	err = s.Stop()
	if err != nil {
		logrus.Fatal(err)
	}
}
func RestartService() {
	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	err = s.Restart()
	if err != nil {
		logrus.Fatal(err)
	}
}
func RunApp() {
	app := &cli.App{
		Flags: []cli.Flag{},
		Commands: []*cli.Command{
			&cli.Command{
				Name:  "service",
				Usage: "service manager",
				Subcommands: []*cli.Command{
					{
						Name:  "install",
						Usage: "install service",
						Action: func(context *cli.Context) error {
							InstallAsService()
							return nil
						},
					},
					{
						Name:  "uninstall",
						Usage: "uninstall service",
						Action: func(context *cli.Context) error {
							UnInstall()
							return nil
						},
					},
					{
						Name:  "start",
						Usage: "start service",
						Action: func(context *cli.Context) error {
							StartService()
							return nil
						},
					},
					{
						Name:  "stop",
						Usage: "stop service",
						Action: func(context *cli.Context) error {
							StopService()
							return nil
						},
					},
					{
						Name:  "restart",
						Usage: "restart service",
						Action: func(context *cli.Context) error {
							RestartService()
							return nil
						},
					},
				},
				Description: "YouComic service controller",
			},
			{
				Name:  "run",
				Usage: "run app",
				Action: func(context *cli.Context) error {
					Program()
					return nil
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	err := initService()
	if err != nil {
		logrus.Fatal(err)
	}
	RunApp()
}
