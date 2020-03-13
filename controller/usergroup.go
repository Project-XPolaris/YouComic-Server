package controller

import (
	"github.com/allentom/youcomic-api/permission"
	"github.com/allentom/youcomic-api/serializer"
	"github.com/allentom/youcomic-api/services"
	"github.com/gin-gonic/gin"
)

var GetUserGroupListHandler gin.HandlerFunc = func(context *gin.Context) {
	view := ListView{
		Context:      context,
		Pagination:   &DefaultPagination{},
		QueryBuilder: &services.UserGroupQueryBuilder{},
		FilterMapping: []FilterMapping{
			{
				Lookup: "id",
				Method: "InId",
				Many:   true,
			},
			{
				Lookup: "order",
				Method: "SetOrderFilter",
				Many:   false,
			}, {
				Lookup: "user",
				Method: "SetUserGroupUser",
				Many:   false,
			},
		},
		GetTemplate: func() serializer.TemplateSerializer {
			return &serializer.BaseUserGroupTemplate{}
		},
		GetContainer: func() serializer.ListContainerSerializer {
			return &serializer.DefaultListContainer{}
		},
		GetPermissions: func(v *ListView) []permission.PermissionChecker {
			return []permission.PermissionChecker{
				&permission.StandardPermissionChecker{UserId: v.Claims.UserId,PermissionName: permission.GetUserGroupListPermissionName},
			}
		},

	}
	view.Run()
}
