package services

import (
	"database/sql"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"gorm.io/gorm"
)

type PermissionQueryBuilder struct {
	IdQueryFilter
	NameQueryFilter
	DefaultPageFilter
	UserGroupQueryFilter
	NameSearchQueryFilter
	UserFilter
}

func (b *PermissionQueryBuilder) ReadModels() (int64, interface{}, error) {
	query := database.Instance
	query = ApplyFilters(b, query)
	var count int64 = 0
	md := make([]model.Permission, 0)
	err := query.Limit(b.PageSize).Offset(b.getOffset()).Find(&md).Offset(-1).Count(&count).Error
	if err == sql.ErrNoRows {
		return 0, md, nil
	}
	return count, md, err
}

type UserGroupQueryFilter struct {
	userGroups []interface{}
}

func (f *UserGroupQueryFilter) SetUserGroupQueryFilter(userGroups ...interface{}) {
	for _, userGroupId := range userGroups {
		if len(userGroupId.(string)) > 0 {
			f.userGroups = append(f.userGroups, userGroupId)
		}
	}
}
func (f UserGroupQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.userGroups != nil && len(f.userGroups) != 0 {
		return db.Joins(
			"inner join usergroup_permissions on permissions.id = usergroup_permissions.permission_id",
		).Where("usergroup_permissions.user_group_id in (?)", f.userGroups)
	}
	return db
}

type UserFilter struct {
	users []interface{}
}

func (f *UserFilter) SetUserFilter(users ...interface{}) {
	for _, userId := range users {
		if len(userId.(string)) > 0 {
			f.users = append(f.users, userId)
		}
	}
}

func (f UserFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.users != nil && len(f.users) != 0 {
		return db.Joins(
			"inner join usergroup_permissions on permissions.id = usergroup_permissions.permission_id",
		).Joins(
			"inner join usergroup_users on usergroup_permissions.user_group_id = usergroup_users.user_group_id",
		).Where("usergroup_users.user_id in (?)", f.users)
	}
	return db
}
