package services

import (
	"database/sql"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"gorm.io/gorm"
)

var DefaultUserGroupName = "Default"
var DefaultSuperUserGroupName = "admin"

type UserGroupQueryBuilder struct {
	IdQueryFilter
	UserGroupUserFilter
	NameQueryFilter
	DefaultPageFilter
}

func (b *UserGroupQueryBuilder) ReadModels() (int64, interface{}, error) {
	query := database.Instance
	query = ApplyFilters(b, query)
	var count int64 = 0
	md := make([]model.UserGroup, 0)
	err := query.Limit(b.getLimit()).Offset(b.getOffset()).Find(&md).Offset(-1).Count(&count).Error
	if err == sql.ErrNoRows {
		return 0, query, nil
	}
	return count, md, err
}

func AddPermissionsToUserGroup(userGroup *model.UserGroup, permissions ...*model.Permission) error {
	permissionInterfaces := make([]interface{}, 0)
	for _, permission := range permissions {
		permissionInterfaces = append(permissionInterfaces, permission)
	}
	return database.Instance.Model(userGroup).Association("Permissions").Append(permissionInterfaces...)
}

type UserGroupUserFilter struct {
	UserId interface{}
}

func (f UserGroupUserFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.UserId != nil {
		return db.Joins(
			"inner join usergroup_users on usergroup_users.user_group_id = user_groups.id",
		).Where("usergroup_users.user_id = ?", f.UserId)
	}
	return db
}

func (f *UserGroupUserFilter) SetUserGroupUser(userId interface{}) {
	if userId != nil {
		f.UserId = userId
	}
}

func AddUsersToUserGroup(userGroup *model.UserGroup, users ...*model.User) error {
	userInterfaces := make([]interface{}, 0)
	for _, user := range users {
		userInterfaces = append(userInterfaces, user)
	}
	return database.Instance.Model(userGroup).Association("Users").Append(userInterfaces...)
}

func RemoveUsersFromUserGroup(userGroup *model.UserGroup, users ...*model.User) error {
	userInterfaces := make([]interface{}, 0)
	for _, user := range users {
		userInterfaces = append(userInterfaces, user)
	}
	return database.Instance.Model(userGroup).Association("Users").Delete(userInterfaces...)
}

func RemovePermissionsFromUserGroup(userGroup *model.UserGroup, permissions ...*model.Permission) error {
	permissionInterfaces := make([]interface{}, 0)
	for _, permission := range permissions {
		permissionInterfaces = append(permissionInterfaces, permission)
	}
	return database.Instance.Model(userGroup).Association("Permissions").Delete(permissionInterfaces...)
}
