package httpapi

import (
	"fmt"
	"github.com/allentom/haruka"
	"github.com/project-xpolaris/youplustoolkit/youlink"
	"github.com/projectxpolaris/youcomic/api/httpapi/serializer"
	"github.com/projectxpolaris/youcomic/auth"
	"github.com/projectxpolaris/youcomic/config"
	ApiError "github.com/projectxpolaris/youcomic/error"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/permission"
	"github.com/projectxpolaris/youcomic/services"
	"github.com/projectxpolaris/youcomic/utils"
	"github.com/projectxpolaris/youcomic/validate"
	"github.com/projectxpolaris/youcomic/youauthplugin"
	"net/http"
)

type RegisterUserResponseBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

// register user handler
//
// path: /user/register
//
// method: post
var RegisterUserHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	requestBody := RegisterUserResponseBody{}
	err = DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}
	// check validate
	if isValidate := validate.RunValidatorsAndRaiseApiError(context,
		&validate.UniqUserNameValidator{Value: requestBody.Username},
		&validate.StringLengthValidator{Value: requestBody.Username, FieldName: "username", LessThan: 16, GreaterThan: 4},
		&validate.StringLengthValidator{Value: requestBody.Password, FieldName: "password", LessThan: 16, GreaterThan: 4},
		&validate.EmailValidator{Value: requestBody.Email},
	); !isValidate {
		return
	}

	user := model.User{Username: requestBody.Username, Password: requestBody.Password, Email: requestBody.Email}
	err = services.RegisterUser(&user)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

type LoginUserRequestBody struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	WithYouPlus bool   `json:"withYouPlus"`
}
type UserAuthResponse struct {
	Id   uint   `json:"id"`
	Sign string `json:"sign"`
}

// login user handler
//
// path: /user/auth
//
// method: post
var LoginUserHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	requestBody := LoginUserRequestBody{}
	err = DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}
	//validate value
	if isValidate := validate.RunValidatorsAndRaiseApiError(context,
		&validate.StringLengthValidator{Value: requestBody.Username, FieldName: "username", LessThan: 16, GreaterThan: 4},
		&validate.StringLengthValidator{Value: requestBody.Password, FieldName: "password", LessThan: 16, GreaterThan: 4},
	); !isValidate {
		return
	}

	var user *model.User
	var sign string
	user, sign, err = services.UserLogin(requestBody.Username, requestBody.Password)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	accessType := context.GetQueryString("type")
	switch accessType {
	case "accessToken":
		context.JSONWithStatus(haruka.JSON{
			"success": true,
			"data": haruka.JSON{
				"accessToken": sign,
				"username":    user.Username,
			},
		}, http.StatusOK)
	default:
		context.JSONWithStatus(haruka.JSON{
			"success": true,
			"uid":     fmt.Sprintf("%d", user.ID),
			"token":   sign,
		}, http.StatusOK)
	}
}
var YouPlusLoginHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	requestBody := LoginUserRequestBody{}
	err = DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}

	//validate value
	if isValidate := validate.RunValidatorsAndRaiseApiError(context,
		&validate.StringLengthValidator{Value: requestBody.Username, FieldName: "username", LessThan: 16, GreaterThan: 4},
		&validate.StringLengthValidator{Value: requestBody.Password, FieldName: "password", LessThan: 16, GreaterThan: 4},
	); !isValidate {
		return
	}
	var user *model.User
	var sign string
	if config.Instance.AuthEnable {
		user, sign, err = services.YouPlusLogin(requestBody.Username, requestBody.Password)
	}

	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	accessType := context.GetQueryString("type")
	switch accessType {
	case "accessToken":
		context.JSON(haruka.JSON{
			"success": true,
			"data": haruka.JSON{
				"accessToken": sign,
				"username":    user.Username,
			},
		})
	default:
		context.JSONWithStatus(haruka.JSON{
			"success": true,
			"uid":     fmt.Sprintf("%d", user.ID),
			"token":   sign,
		}, http.StatusOK)
	}
}
var GetCurrentHandler2 haruka.RequestHandler = func(context *haruka.Context) {
	tokenString := context.GetQueryString("token")
	token, err := auth.ParseToken(tokenString)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	context.JSONWithStatus(haruka.JSON{
		"success": true,
		"uid":     fmt.Sprintf("%d", token.GetUserId()),
	}, http.StatusOK)
}

// get user handler
//
// path: /user/:id
//
// method: get
var GetUserHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}
	var user model.User
	err = services.GetModelById(&user, id)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	template := serializer.BaseUserTemplate{}
	err = template.Serializer(user, nil)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	context.JSONWithStatus(template, http.StatusOK)
}

// get user groups handler
//
// path: /user/:id/groups
//
// method: get
var GetUserUserGroupsHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}
	queryBuilder := services.UserGroupQueryBuilder{}
	queryBuilder.SetUserGroupUser(id)
	count, usergroups, err := queryBuilder.ReadModels()
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	result := serializer.SerializeMultipleTemplate(usergroups, &serializer.BaseUserGroupTemplate{}, nil)
	responseBody := serializer.DefaultListContainer{}
	responseBody.SerializeList(result, map[string]interface{}{
		"page":     1,
		"pageSize": 10,
		"count":    count,
		"url":      context.Request.URL,
	})
	context.JSONWithStatus(responseBody, http.StatusOK)
}

// get user list handler
//
// path: /users
//
// method: get
var GetUserUserListHandler haruka.RequestHandler = func(context *haruka.Context) {
	claims, err := auth.ParseAuthHeader(context)
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.UserAuthFailError, nil)
		return
	}

	if hasPermission := permission.CheckPermissionAndServerError(context,
		&permission.StandardPermissionChecker{PermissionName: permission.GetUserListPermissionName, UserId: claims.GetUserId()},
	); !hasPermission {
		return
	}
	userQueryBuilder := services.UserQueryBuilder{}
	//get page
	pagination := DefaultPagination{}
	pagination.Read(context)
	userQueryBuilder.SetPageFilter(pagination.Page, pagination.PageSize)

	//query filter
	filterMapping := []FilterMapping{
		{
			Lookup: "id",
			Method: "InId",
			Many:   true,
		},
		{
			Lookup: "name",
			Method: "SetUserNameFilter",
			Many:   true,
		},
		{
			Lookup: "nameSearch",
			Method: "SetNameSearchQueryFilter",
			Many:   true,
		},
		{
			Lookup: "nicknameSearch",
			Method: "SetNicknameSearchQueryFilter",
			Many:   true,
		},
		{
			Lookup: "usergroup",
			Method: "SetUserGroupQueryFilter",
			Many:   true,
		}, {
			Lookup: "order",
			Method: "SetOrderFilter",
			Many:   true,
		},
	}
	for _, filter := range filterMapping {
		utils.FilterByParam(context, filter.Lookup, &userQueryBuilder, filter.Method, filter.Many)
	}

	count, users, err := userQueryBuilder.ReadModels()

	result := serializer.SerializeMultipleTemplate(users, &serializer.ManagerUserTemplate{}, nil)
	responseBody := serializer.DefaultListContainer{}
	responseBody.SerializeList(result, map[string]interface{}{
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
		"count":    count,
		"url":      context.Request.URL,
	})
	context.JSONWithStatus(responseBody, http.StatusOK)
}

type ChangeUserPasswordRequestBody struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// change password handler
//
// path: /user/password
//
// method: put
var ChangeUserPasswordHandler haruka.RequestHandler = func(context *haruka.Context) {
	claims, err := auth.ParseAuthHeader(context)
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.UserAuthFailError, nil)
		return
	}

	requestBody := ChangeUserPasswordRequestBody{}
	err = DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}

	isValidate := validate.RunValidatorsAndRaiseApiError(
		context,
		&validate.StringLengthValidator{Value: requestBody.OldPassword, GreaterThan: 4, LessThan: 256, FieldName: "oldPassword"},
		&validate.StringLengthValidator{Value: requestBody.NewPassword, GreaterThan: 4, LessThan: 256, FieldName: "newPassword"},
	)
	if !isValidate {
		return
	}

	err = services.ChangeUserPassword(claims.GetUserId(), requestBody.OldPassword, requestBody.NewPassword)
	if err != nil {
		if err == services.UserPasswordInvalidate {
			ApiError.RaiseApiError(context, ApiError.UserAuthFailError, nil)
			return
		}
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

type ChangeUserNicknameRequestBody struct {
	Nickname string `json:"nickname"`
}

// change nickname handler
//
// path: /user/nickname
//
// method: put
var ChangeUserNicknameHandler haruka.RequestHandler = func(context *haruka.Context) {
	claims, err := auth.ParseAuthHeader(context)
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.UserAuthFailError, nil)
		return
	}

	requestBody := ChangeUserNicknameRequestBody{}
	err = DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}

	isValidate := validate.RunValidatorsAndRaiseApiError(
		context,
		&validate.StringLengthValidator{Value: requestBody.Nickname, GreaterThan: 4, LessThan: 256, FieldName: "nickname"},
	)
	if !isValidate {
		return
	}

	err = services.ChangeUserNickname(claims.GetUserId(), requestBody.Nickname)
	if err != nil {
		if err == services.UserNotFoundError {
			ApiError.RaiseApiError(context, ApiError.UserAuthFailError, nil)
			return
		}
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

// get account histories
//
// path: /account/histories
//
// method: get
var UserHistoryHandler haruka.RequestHandler = func(context *haruka.Context) {
	queryBuilder := &services.HistoryQueryBuilder{}
	userClaim := auth.GetUserClaimsFromContext(context)
	queryBuilder.SetUserIdFilter(userClaim.UserId)

	view := ListView{
		Context:      context,
		Pagination:   &DefaultPagination{},
		QueryBuilder: queryBuilder,
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
			},
		},
		GetContainer: func() serializer.ListContainerSerializer {
			return &serializer.DefaultListContainer{}
		},
		GetTemplate: func() serializer.TemplateSerializer {
			return &serializer.BaseHistoryTemplate{}
		},
	}
	view.Run()
}

// clear account histories
//
// path: /account/histories
//
// method: delete
var DeleteUserHistoryHandler haruka.RequestHandler = func(context *haruka.Context) {
	userClaim := auth.GetUserClaimsFromContext(context)
	queryBuilder := services.HistoryQueryBuilder{}
	queryBuilder.SetUserIdFilter(userClaim.UserId)
	err := queryBuilder.DeleteModels(true)
	if err == services.UserNotFoundError {
		ApiError.RaiseApiError(context, ApiError.UserAuthFailError, nil)
		return
	}
	ServerSuccessResponse(context)
}

var generateAccessCodeWithYouAuthHandler haruka.RequestHandler = func(context *haruka.Context) {
	code := context.GetQueryString("code")
	accessToken, username, err := services.GenerateYouAuthToken(code)
	if err != nil {
		youlink.AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
		"data": haruka.JSON{
			"accessToken": accessToken,
			"username":    username,
		},
	})
}

var youAuthTokenHandler haruka.RequestHandler = func(context *haruka.Context) {
	// check token is valid
	token := context.GetQueryString("token")
	_, err := youauthplugin.DefaultYouAuthOauthPlugin.Client.GetCurrentUser(token)
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.UserAuthFailError, nil)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}
