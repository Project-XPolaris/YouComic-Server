package controller

import (
	"github.com/allentom/youcomic-api/api/auth"
	serializer2 "github.com/allentom/youcomic-api/api/serializer"
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/permission"
	"github.com/allentom/youcomic-api/services"
	"github.com/allentom/youcomic-api/validate"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
		GetTemplate: func() serializer2.TemplateSerializer {
			return &serializer2.BaseUserGroupTemplate{}
		},
		GetContainer: func() serializer2.ListContainerSerializer {
			return &serializer2.DefaultListContainer{}
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
		ResponseTemplate: &serializer2.BaseUserGroupTemplate{},
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
	UserIds []uint `json:"userIds"`
}

// add user to usergroup handler
//
// put: /usergroup/:id/users
//
// method: put
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
	for _, userId := range requestBody.UserIds {
		users = append(users, &model.User{Model: gorm.Model{ID: userId}})
	}
	err = services.AddUsersToUserGroup(&model.UserGroup{Model: gorm.Model{ID: uint(id)}}, users...)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

type AddPermissionToUserGroupRequestBody struct {
	PermissionIds []uint `json:"permissionIds"`
}

// add user to usergroup handler
//
// put: /usergroup/:id/permissions
//
// method: put
var AddPermissionToUserGroupHandler gin.HandlerFunc = func(context *gin.Context) {

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
		&permission.StandardPermissionChecker{PermissionName: permission.AddPermissionToUserGroupPermissionName, UserId: claims.UserId},
	); !hasPermission {
		return
	}

	requestBody := AddPermissionToUserGroupRequestBody{}
	err = DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}

	permissions := make([]*model.Permission, 0)
	for _, permissionId := range requestBody.PermissionIds {
		permissions = append(permissions, &model.Permission{Model: gorm.Model{ID: permissionId}})
	}
	err = services.AddPermissionsToUserGroup(&model.UserGroup{Model: gorm.Model{ID: uint(id)}}, permissions...)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

type RemoveUserToUserGroupRequestBody struct {
	UserIds []uint `json:"userIds"`
}

// remove user from usergroup handler
//
// path: /usergroup/:id/users
//
// method: delete
var RemoveUserFromUserGroupHandler gin.HandlerFunc = func(context *gin.Context) {

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
		&permission.StandardPermissionChecker{PermissionName: permission.RemoveUserFromUserGroupPermissionName, UserId: claims.UserId},
	); !hasPermission {
		return
	}

	requestBody := RemoveUserToUserGroupRequestBody{}
	err = DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}

	users := make([]*model.User, 0)
	for _, userId := range requestBody.UserIds {
		users = append(users, &model.User{Model: gorm.Model{ID: userId}})
	}
	err = services.RemoveUsersFromUserGroup(&model.UserGroup{Model: gorm.Model{ID: uint(id)}}, users...)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

type RemovePermissionFromUserGroupRequestBody struct {
	PermissionIds []uint `json:"permissionIds"`
}

// remove permission from usergroup handler
//
// path: /usergroup/:id/permissions
//
// method: delete
var RemovePermissionFromUserGroupHandler gin.HandlerFunc = func(context *gin.Context) {

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
		&permission.StandardPermissionChecker{PermissionName: permission.AddPermissionToUserGroupPermissionName, UserId: claims.UserId},
	); !hasPermission {
		return
	}

	requestBody := RemovePermissionFromUserGroupRequestBody{}
	err = DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}

	permissions := make([]*model.Permission, 0)
	for _, permissionId := range requestBody.PermissionIds {
		permissions = append(permissions, &model.Permission{Model: gorm.Model{ID: permissionId}})
	}
	err = services.RemovePermissionsFromUserGroup(&model.UserGroup{Model: gorm.Model{ID: uint(id)}}, permissions...)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}
