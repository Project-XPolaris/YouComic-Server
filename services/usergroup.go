package services

import (
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"github.com/jinzhu/gorm"
)

type UserGroupQueryBuilder struct {
	IdQueryFilter
	UserGroupUserFilter
	NameQueryFilter
	DefaultPageFilter
}

func (b *UserGroupQueryBuilder) ReadModels() (int, interface{}, error) {
	query := database.DB
	query = ApplyFilters(b, query)
	var count = 0
	md := make([]model.UserGroup, 0)
	err := query.Limit(b.PageSize).Offset(b.getOffset()).Find(&md).Offset(-1).Count(&count).Error
	return count, md, err
}

func AddPermissionsToUserGroup(userGroup *model.UserGroup, permissions ...*model.Permission) error {
	permissionInterfaces := make([]interface{}, 0)
	for _, permission := range permissions {
		permissionInterfaces = append(permissionInterfaces, permission)
	}
	return database.DB.Model(userGroup).Association("Permissions").Append(permissionInterfaces...).Error
}

type UserGroupUserFilter struct {
	UserId interface{}
}

func (f *UserGroupUserFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.UserId != nil {
		return db.Joins(
			"inner join usergroup_users on usergroup_users.user_group_id = user_groups.id",
		).Where("usergroup_users.user_id = ?", f.UserId)
	}
	return db
}

func (f *UserGroupUserFilter) SetUserGroupUser(userId interface{})  {
	if f.UserId != nil {
		f.UserId = userId
	}
}

func AddUsersToUserGroup(userGroup *model.UserGroup, users ...*model.User) error {
	userInterfaces := make([]interface{}, 0)
	for _, user := range users {
		userInterfaces = append(userInterfaces, user)
	}
	return database.DB.Model(userGroup).Association("Users").Append(userInterfaces...).Error
}

func RemoveUsersFromUserGroup(userGroup *model.UserGroup, users ...*model.User) error {
	userInterfaces := make([]interface{}, 0)
	for _, user := range users {
		userInterfaces = append(userInterfaces, user)
	}
	return database.DB.Model(userGroup).Association("Users").Delete(userInterfaces...).Error
}

func RemovePermissionsFromUserGroup(userGroup *model.UserGroup, permissions ...*model.Permission) error {
	permissionInterfaces := make([]interface{}, 0)
	for _, permission := range permissions {
		permissionInterfaces = append(permissionInterfaces, permission)
	}
	return database.DB.Model(userGroup).Association("Permissions").Delete(permissionInterfaces...).Error
}