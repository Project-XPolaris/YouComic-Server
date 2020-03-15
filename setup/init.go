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
	err = CreateUserGroupIfNotExist(superUserGroupName)
	if err != nil {
		return
	}

	// create default user group
	err = CreateUserGroupIfNotExist(defaultUserGroupName)
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
	if count == 0 {
		err = services.RegisterUser(&model.User{Username: superUserUsername, Password: superUserPassword})
		if err != nil {
			return
		}
	}

	//init done close
	config.Set("init", false)
	err = config.WriteConfig()
	if err != nil {
		return
	}
	return
}

func CreateUserGroupIfNotExist(userGroupName string) (err error) {
	userGroupQueryBuilder := services.UserGroupQueryBuilder{}
	userGroupQueryBuilder.SetPageFilter(1, 1)
	userGroupQueryBuilder.SetNameFilter(userGroupName)
	count, _, err := userGroupQueryBuilder.ReadModels()
	if err != nil {
		return
	}
	if count != 0 {
		err = services.CreateModel(&model.UserGroup{Name: userGroupName})
		if err != nil {
			return
		}
	}
	return nil
}
