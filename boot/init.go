package boot

import (
	"fmt"
	"github.com/allentom/harukap"
	youlog2 "github.com/project-xpolaris/youplustoolkit/youlog"
	"github.com/projectxpolaris/youcomic/application"
	appconfig "github.com/projectxpolaris/youcomic/config"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/services"
	"os"
)

var permissionNames = []string{
	"CreateBook",
	"UpdateBook",
	"DeleteBook",
	"CreateTag",
	"UpdateTag",
	"DeleteTag",
	"AddTagToBook",
	"CreateCollection",
	"UpdateCollection",
	"DeleteCollection",
	"GetUserList",
	"GetPermissionList",
	"GetUserGroupList",
	"CreateUserGroup",
	"AddUserToUserGroup",
	"AddPermissionToUserGroup",
	"RemoveUserFromUserGroup",
	"PermanentlyDeleteBook",
	"CreateLibrary",
	"UpdateLibrary",
	"DeleteLibrary",
	"ScanLibrary",
}
var Logger *youlog2.Scope

type InitPlugin struct {
}

func (p *InitPlugin) OnInit(e *harukap.HarukaAppEngine) error {
	Logger = e.LoggerPlugin.Logger.NewScope("boot")
	err := SetupApplication()
	if err != nil {
		return err
	}
	return nil
}

func InitApplication() (err error) {
	Logger.Info("initial...")
	var superUserUsername, superUserPassword string
	superUserUsername = os.Getenv("YOUCOMIC_SUPERUSER_USERNAME")
	superUserPassword = os.Getenv("YOUCOMIC_SUPERUSER_PASSWORD")
	// create user group
	Logger.Info(fmt.Sprintf("create user group with name = %s", services.DefaultSuperUserGroupName))
	superUserGroup, err := CreateUserGroupIfNotExist(services.DefaultSuperUserGroupName)
	if err != nil {
		return
	}

	// create default user group
	Logger.Info(fmt.Sprintf("create user group with name = %s", services.DefaultUserGroupName))
	_, err = CreateUserGroupIfNotExist(services.DefaultUserGroupName)
	if err != nil {
		return err
	}

	// create superuser account
	userQueryBuilder := services.UserQueryBuilder{}
	userQueryBuilder.SetPageFilter(1, 1)
	userQueryBuilder.SetUserNameFilter(superUserUsername)
	userQueryBuilder.WithPreload("UserGroups")

	existUser, err := userQueryBuilder.FirstOrNil()
	if err != nil {
		return
	}
	if existUser == nil {
		existUser = &model.User{Username: superUserUsername, Password: superUserPassword}
		err = services.RegisterUser(existUser)
		if err != nil {
			return
		}
	}
	inSuperuserGroup := false
	for _, group := range existUser.UserGroups {
		if group.Name == superUserGroup.Name {
			inSuperuserGroup = true
			break
		}
	}
	if !inSuperuserGroup {
		err = services.AddUsersToUserGroup(superUserGroup, existUser)
		if err != nil {
			return err
		}
	}
	//boot done close
	return
}

func CreateUserGroupIfNotExist(userGroupName string) (userGroup *model.UserGroup, err error) {
	userGroupQueryBuilder := services.UserGroupQueryBuilder{}
	userGroupQueryBuilder.SetPageFilter(1, 1)
	userGroupQueryBuilder.SetNameFilter(userGroupName)
	count, groups, err := userGroupQueryBuilder.ReadModels()
	if err != nil {
		return
	}
	userGroup = &model.UserGroup{Name: userGroupName}
	if count == 0 {
		err = services.CreateModel(userGroup)
		if err != nil {
			return
		}
	} else {
		userGroup = &(groups.([]model.UserGroup))[0]
	}
	return userGroup, nil
}
func SetupApplication() error {
	// boot service (only first start)
	err := InitApplication()
	if err != nil {
		return err
	}
	// setup service
	Logger.Info("boot service,please wait")
	err = initPermissions()
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
	err = services.InitAiTaggerService()
	if err != nil {
		return err
	}
	return nil
}

// boot permission
func initPermissions() error {
	Logger.Info("boot permissions")
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

// boot superuser group permission
//
// superuser group will granted all permission
func initSuperuserPermission() error {
	Logger.Info("boot super user permission")
	superUserGroupName := services.DefaultSuperUserGroupName
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
	return nil
}

func initStorePath() error {
	// boot app store
	err := os.MkdirAll(appconfig.Instance.Store.Root, os.ModePerm)
	if err != nil {
		return err
	}

	// boot default library path
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
