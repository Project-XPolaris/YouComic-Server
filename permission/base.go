package permission

import (
	"github.com/allentom/youcomic-api/database"
	ApplicationError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/model"
	"github.com/gin-gonic/gin"
)

func CheckUserHasPermission(userId uint, permissionName string) (error, bool) {
	permission := model.Permission{Name: permissionName}
	err := database.DB.Model(
		&permission,
	).Select(
		"permissions.*",
	).Joins(
		"inner join usergroup_permissions on permissions.id = usergroup_permissions.permission_id",
	).Joins(
		"inner join user_groups on user_groups.id = usergroup_permissions.user_group_id",
	).Joins(
		"inner join usergroup_users on usergroup_users.user_group_id = user_groups.id",
	).Where(
		"user_id = ?", userId,
	).Where(
		"permissions.name = ?", permissionName,
	).Find(&permission).Error
	if err != nil {
		return err, false
	}
	return nil, permission.ID != 0
}

func ChePermissionAndServerError(context *gin.Context, permissions ...PermissionChecker) {
	for _, permission := range permissions {
		isValidate := permission.CheckPermission()
		if !isValidate {
			ApplicationError.RaiseApiError(context, ApplicationError.PermissionError, nil)
			return
		}
	}
}
