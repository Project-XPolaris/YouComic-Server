package controller

import (
	"github.com/allentom/youcomic-api/auth"
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/permission"
	"github.com/allentom/youcomic-api/serializer"
	"github.com/allentom/youcomic-api/services"
	"github.com/allentom/youcomic-api/validate"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
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

type AddUserToUserGroupRequestBody struct {
	userIds []uint
}

// add user to usergroup handler
//
// put: /usergroup/:id/users
//
// method: post
var AddUserToUserGroupHandler gin.HandlerFunc = func(context *gin.Context) {

	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}

	claims, err := auth.ParseAuthHeader(context)
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.UserAuthFailError, nil)
		return
	}

	//check permission
	if hasPermission := permission.CheckPermissionAndServerError(context,
		&permission.StandardPermissionChecker{PermissionName: permission.AddUserToUserGroupPermissionName, UserId: claims.UserId},
	); !hasPermission {
		return
	}

	requestBody := AddUserToUserGroupRequestBody{}
	err = DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}

	users := make([]*model.User, 0)
	for _, userId := range requestBody.userIds {
		users = append(users, &model.User{Model: gorm.Model{ID: userId}})
	}
	err = services.AddUsersToUserGroup(&model.UserGroup{Model: gorm.Model{ID: uint(id)}}, users...)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}
