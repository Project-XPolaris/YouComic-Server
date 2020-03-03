package services

import (
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
)

type UserGroupQueryBuilder struct {
	NameQueryFilter
}

func (b *UserGroupQueryBuilder) ReadModels() (int, interface{}, error) {
	query := database.DB
	query = ApplyFilters(b, query)
	var count = 0
	md := make([]model.UserGroup, 0)
	err := query.Find(&md).Offset(-1).Count(&count).Error
	return count, md, err
}

func AddPermissionsToUserGroup(userGroup *model.UserGroup, permissions ...*model.Permission) error {
	permissionInterfaces := make([]interface{}, 0)
	for _, permission := range permissions {
		permissionInterfaces = append(permissionInterfaces, permission)
	}
	return database.DB.Model(userGroup).Association("Permissions").Append(permissionInterfaces...).Error
}
