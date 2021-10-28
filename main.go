package main

import (
	"context"
	"fmt"
	"github.com/allentom/youcomic-api/api/httpapi"
	"github.com/allentom/youcomic-api/config"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/install"
	applogger "github.com/allentom/youcomic-api/log"
	"github.com/allentom/youcomic-api/setup"
	util "github.com/allentom/youcomic-api/utils"
	"github.com/allentom/youcomic-api/youplus"
	srv "github.com/kardianos/service"
	entry "github.com/project-xpolaris/youplustoolkit/youplus/entity"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"path/filepath"
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

	// register rpc
	if len(config.Instance.YouPlus.RPCUrl) > 0 {
		err = youplus.LoadYouPlusRPCClient()
		if err != nil {
			MainLogger.WithFields(logrus.Fields{
				"signal": "rpc_connect_failed",
				"url":    config.Instance.YouPlus.RPCUrl,
			}).Fatal("RPC connect success")
		}
		MainLogger.WithFields(logrus.Fields{
			"signal": "rpc_connect_success",
			"port":   config.Instance.YouPlus.RPCUrl,
		}).Info("RPC connect success")
	}
	// youplus entity
	if config.Instance.YouPlus.EntityConfig.Enable {
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
			urls = append(urls, fmt.Sprintf("http://%s:%s", addr, config.Instance.Application.Port))
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
		"port":   config.Instance.Application.Port,
	}).Info("Service start success!")
	httpapi.RunApiService()
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
