package setup

import (
	appconfig "github.com/allentom/youcomic-api/config"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/services"
)

func InitApplication() (err error) {
	config, err := appconfig.ReadConfig("init")
	if err != nil {
		return err
	}
	needInit := config.GetBool("init")
	if !needInit {
		return
	}
	LogField.Info("initial...")
	superUserGroupName := config.GetString("superUserGroupName")
	defaultUserGroupName := config.GetString("defaultUserGroupName")
	superUserUsername := config.GetString("adminAccount.username")
	superUserPassword := config.GetString("adminAccount.password")

	// create user group
	superUserGroup, err := CreateUserGroupIfNotExist(superUserGroupName)
	if err != nil {
		return
	}

	// create default user group
	_, err = CreateUserGroupIfNotExist(defaultUserGroupName)
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
