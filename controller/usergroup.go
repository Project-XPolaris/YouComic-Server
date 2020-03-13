package controller

import (
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/permission"
	"github.com/allentom/youcomic-api/serializer"
	"github.com/allentom/youcomic-api/services"
	"github.com/allentom/youcomic-api/validate"
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
				&permission.StandardPermissionChecker{UserId: v.Claims.UserId, PermissionName: permission.GetUserGroupListPermissionName},
			}
		},
	}
	view.Run()
}

type CreateUserGroupRequestBody struct {
	Name string `json:"name"`
}

var CreateUserGroupHandler gin.HandlerFunc = func(context *gin.Context) {
	view := CreateModelView{
		Context: context,
		CreateModel: func() interface{} {
			return &model.UserGroup{}
		},
		ResponseTemplate: &serializer.BaseUserGroupTemplate{},
		RequestBody:      &CreateUserGroupRequestBody{},
		GetPermissions: func(v *CreateModelView) []permission.PermissionChecker {
			return []permission.PermissionChecker{
				&permission.StandardPermissionChecker{UserId: v.Claims.UserId, PermissionName: permission.CreateUserGroupPermissionName},
			}
		},
		GetValidators: func(v *CreateModelView) []validate.Validator {
			requestBody := v.RequestBody.(*CreateUserGroupRequestBody)
			return []validate.Validator{
				&validate.StringLengthValidator{Value: requestBody.Name, FieldName: "Name", GreaterThan: 0, LessThan: 256},
			}
		},
	}
	view.Run()
}
