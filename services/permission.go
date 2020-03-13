package services

import (
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"github.com/jinzhu/gorm"
)

type PermissionQueryBuilder struct {
	IdQueryFilter
	NameQueryFilter
	DefaultPageFilter
	UserGroupQueryFilter
	NameSearchQueryFilter
}

func (b *PermissionQueryBuilder) ReadModels() (int, interface{}, error) {
	query := database.DB
	query = ApplyFilters(b, query)
	var count = 0
	md := make([]model.Permission, 0)
	err := query.Limit(b.PageSize).Offset(b.getOffset()).Find(&md).Offset(-1).Count(&count).Error
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
