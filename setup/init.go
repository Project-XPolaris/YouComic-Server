package setup

import (
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/services"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func InitApplication() error {
	logrus.Info("init service,please wait")
	config := viper.New()
	config.SetConfigType("json")
	config.AddConfigPath("./conf")
	config.SetConfigName("setup")
	err := config.ReadInConfig()
	if err != nil {
		return err
	}
	err = initPermissions(config)
	if err != nil {
		return err
	}
	err = initUserGroups(config)
	if err != nil {
		return err
	}
	return nil
}

func initPermissions(config *viper.Viper) error {
	logrus.Info("init permissions")
	permissionNames := config.GetStringSlice("permissions")
	for _, permissionName := range permissionNames {
		builder := services.PermissionQueryBuilder{}
		builder.SetNameFilter(permissionName)
		count, _, err := builder.ReadModels()
		if err != nil {
			return err
		}
		if count == 0 {
			err = services.CreateModel(&model.Permission{Name: permissionName})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func initUserGroups(config *viper.Viper) error {
	logrus.Info("init user group")
	userGroupNames := config.GetStringSlice("usergroups")
	groupPermissionNames := config.GetStringMapStringSlice("groupPermissions")
	for _, userGroupName := range userGroupNames {
		userGroupQueryBuilder := services.UserGroupQueryBuilder{}
		userGroupQueryBuilder.SetNameFilter(userGroupName)
		count, _, err := userGroupQueryBuilder.ReadModels()
		if err != nil {
			return err
		}
		userGroup := &model.UserGroup{Name: userGroupName}
		if count == 0 {
			err = services.CreateModel(userGroup)
			if err != nil {
				return err
			}
		} else {
			// already exist skip
			return nil
		}
		if permissionNames, isUserPermissionConfigExist := groupPermissionNames[userGroupName]; isUserPermissionConfigExist {
			permissionQueryBuilder := services.PermissionQueryBuilder{}
			for _, permissionName := range permissionNames {
				permissionQueryBuilder.SetNameFilter(permissionName)
			}
			_, queryPermissionResult, err := permissionQueryBuilder.ReadModels()
			if err != nil {
				return err
			}
			permissions := queryPermissionResult.([]model.Permission)
			for _, permission := range permissions {
				err = services.AddPermissionsToUserGroup(userGroup, &permission)
				if err != nil {
					return err
				}
			}

		}

	}
	return nil
}
