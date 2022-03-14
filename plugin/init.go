package plugin

import (
	"fmt"
	"github.com/allentom/harukap"
	"github.com/allentom/youcomic-api/application"
	appconfig "github.com/allentom/youcomic-api/config"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/services"
	"github.com/allentom/youcomic-api/utils"
	youlog2 "github.com/project-xpolaris/youplustoolkit/youlog"
	"github.com/spf13/viper"
	"os"
)

var Logger *youlog2.Scope

type InitPlugin struct {
}

func (p *InitPlugin) OnInit(e *harukap.HarukaAppEngine) error {
	Logger = e.LoggerPlugin.Logger.NewScope("init")
	err := SetupApplication()
	if err != nil {
		return err
	}
	return nil
}

func InitApplication() (err error) {
	config, err := utils.ReadConfig("init")
	if err != nil {
		return err
	}
	needInit := config.GetBool("init")
	if !needInit {
		return
	}

	Logger.Info("initial...")
	superUserGroupName := config.GetString("superUserGroupName")
	superUserUsername := config.GetString("adminAccount.username")
	superUserPassword := config.GetString("adminAccount.password")
	// create user group
	superUserGroup, err := CreateUserGroupIfNotExist(superUserGroupName)
	if err != nil {
		return
	}

	// create default user group
	_, err = CreateUserGroupIfNotExist(services.DefaultUserGroupName)
	if err != nil {
		return err
	}

	// create superuser account
	userQueryBuilder := services.UserQueryBuilder{}
	userQueryBuilder.SetPageFilter(1, 1)
	userQueryBuilder.SetUserNameFilter(superUserUsername)
	count, _, err := userQueryBuilder.ReadModels()
	if err != nil {
		return
	}
	superuser := &model.User{Username: superUserUsername, Password: superUserPassword}
	if count == 0 {
		err = services.RegisterUser(superuser)
		if err != nil {
			return
		}
	}
	err = services.AddUsersToUserGroup(superUserGroup, superuser)
	if err != nil {
		return err
	}

	//init done close
	config.Set("init", false)
	err = config.WriteConfig()
	if err != nil {
		return
	}
	return
}

func CreateUserGroupIfNotExist(userGroupName string) (userGroup *model.UserGroup, err error) {
	userGroupQueryBuilder := services.UserGroupQueryBuilder{}
	userGroupQueryBuilder.SetPageFilter(1, 1)
	userGroupQueryBuilder.SetNameFilter(userGroupName)
	count, _, err := userGroupQueryBuilder.ReadModels()
	if err != nil {
		return
	}
	userGroup = &model.UserGroup{Name: userGroupName}
	if count == 0 {
		err = services.CreateModel(userGroup)
		if err != nil {
			return
		}
	}
	return userGroup, nil
}
func SetupApplication() error {
	// init service (only first start)
	err := InitApplication()
	if err != nil {
		return err
	}
	// setup service
	Logger.Info("init service,please wait")
	Logger.Info("read setup file")
	config, err := utils.ReadConfig("setup")
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
	// check thumbnails
	return nil
}

// init permission
func initPermissions(config *viper.Viper) error {
	Logger.Info("init permissions")
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
			Logger.Info(fmt.Sprintf("permission [%s] is not exist,will created", permissionName))
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
	Logger.Info("init super user permission")
	initConfig, err := utils.ReadConfig("init")
	if err != nil {
		return err
	}

	setupConfig, err := utils.ReadConfig("setup")
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
	err := os.MkdirAll(appconfig.Instance.Store.Root, os.ModePerm)
	if err != nil {
		return err
	}

	// init default library path
	err = os.MkdirAll(appconfig.Instance.Store.Books, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func initDefaultLibrary() error {
	var count int64 = 0
	err := database.Instance.Model(&model.Library{}).Where("name = ?", application.DEFAULT_LIBRARY_NAME).Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		err = database.Instance.Create(&model.Library{Path: appconfig.Instance.Store.Books, Name: application.DEFAULT_LIBRARY_NAME}).Error
		if err != nil {
			return err
		}
	}
	return nil
}
