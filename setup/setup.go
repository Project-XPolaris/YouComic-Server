package setup

import (
	"database/sql"
	"fmt"
	"github.com/allentom/youcomic-api/application"
	appconfig "github.com/allentom/youcomic-api/config"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/log"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/services"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"os"
)

var LogField = log.Logger.WithField("scope", "setup")

func SetupApplication() error {
	// init service (only first start)
	err := InitApplication()
	if err != nil {
		return err
	}
	// setup service
	LogField.Info("init service,please wait")
	LogField.Info("read setup file")
	config, err := appconfig.ReadConfig("setup")
	if err != nil {
		return err
	}
	// init permission
	err = initPermissions(config)
	if err != nil {
		return err
	}
	err = initSuperuserPermission()
	if err != nil {
		return err
	}
	err = initStorePath()
	if err != nil {
		return err
	}
	err = initDefaultLibrary()
	if err != nil {
		return err
	}
	return nil
}

// create database if not exist
func CheckDatabase() {
	LogField.Info("check database")
	if appconfig.Config.Database.Type == "sqlite" {
		return
	}
	db, err := sql.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/",
		appconfig.Config.Mysql.Username,
		appconfig.Config.Mysql.Password,
		appconfig.Config.Mysql.Host,
		appconfig.Config.Mysql.Port,
	))
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + appconfig.Config.Mysql.Database)
	if err != nil {
		panic(err)
	}
}

// init permission
func initPermissions(config *viper.Viper) error {
	LogField.Info("init permissions")
	permissionNames := config.GetStringSlice("permissions")
	//create permission if is NOT exist
	for _, permissionName := range permissionNames {
		builder := services.PermissionQueryBuilder{}
		builder.SetNameFilter(permissionName)
		builder.SetPageFilter(1, 1)
		count, _, err := builder.ReadModels()
		if err != nil {
			return err
		}
		if count == 0 {
			LogField.Info(fmt.Sprintf("permission [%s] is not exist,will created", permissionName))
			err = services.CreateModel(&model.Permission{Name: permissionName})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// init superuser group permission
//
// superuser group will granted all permission
func initSuperuserPermission() error {
	LogField.Info("init super user permission")
	initConfig, err := appconfig.ReadConfig("init")
	if err != nil {
		return err
	}

	setupConfig, err := appconfig.ReadConfig("setup")
	if err != nil {
		return err
	}
	permissionNames := setupConfig.GetStringSlice("permissions")

	superUserGroupName := initConfig.GetString("superusergroupname")
	if len(superUserGroupName) > 0 {
		userGroupQueryBuilder := services.UserGroupQueryBuilder{}
		userGroupQueryBuilder.SetPageFilter(1, 1)
		userGroupQueryBuilder.SetNameFilter(superUserGroupName)
		count, userResult, err := userGroupQueryBuilder.ReadModels()
		if err != nil {
			return err
		}
		if count == 0 {
			return nil
		}
		userGroups := userResult.([]model.UserGroup)
		userGroup := userGroups[0]

		//queryPermission
		permissionQueryBuilder := services.PermissionQueryBuilder{}
		permissionQueryBuilder.SetPageFilter(1, len(permissionNames))
		for _, permissionName := range permissionNames {
			permissionQueryBuilder.SetNameFilter(permissionName)
		}
		count, permissionResult, err := permissionQueryBuilder.ReadModels()
		if err != nil {
			return err
		}
		permissionPtrs := make([]*model.Permission, 0)
		permissions := permissionResult.([]model.Permission)
		for idx := range permissions {
			permissionPtrs = append(permissionPtrs, &permissions[idx])
		}
		err = services.AddPermissionsToUserGroup(&userGroup, permissionPtrs...)
		if err != nil {
			return err
		}
	}
	return nil
}

func initStorePath() error {
	// init app store
	err := os.MkdirAll(appconfig.Config.Store.Root, os.ModePerm)
	if err != nil {
		return err
	}

	// init default library path
	err = os.MkdirAll(appconfig.Config.Store.Books, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func initDefaultLibrary() error {
	count := 0
	err := database.DB.Model(&model.Library{}).Where("name = ?", application.DEFAULT_LIBRARY_NAME).Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		err = database.DB.Create(&model.Library{Path: appconfig.Config.Store.Books, Name: application.DEFAULT_LIBRARY_NAME}).Error
		if err != nil {
			return err
		}
	}
	return nil
}
