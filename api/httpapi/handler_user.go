package httpapi

import (
	"github.com/allentom/haruka"
	serializer2 "github.com/allentom/youcomic-api/api/serializer"
	"github.com/allentom/youcomic-api/auth"
	"github.com/allentom/youcomic-api/config"
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/permission"
	"github.com/allentom/youcomic-api/services"
	"github.com/allentom/youcomic-api/utils"
	"github.com/allentom/youcomic-api/validate"
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
	if config.Instance.YouPlus.Auth {
		user, sign, err = services.YouPlusLogin(requestBody.Username, requestBody.Password)
	} else {
		user, sign, err = services.UserLogin(requestBody.Username, requestBody.Password)
	}

	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	context.JSONWithStatus(UserAuthResponse{
		Id:   user.ID,
		Sign: sign,
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

	template := serializer2.BaseUserTemplate{}
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
	result := serializer2.SerializeMultipleTemplate(usergroups, &serializer2.BaseUserGroupTemplate{}, nil)
	responseBody := serializer2.DefaultListContainer{}
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
		&permission.StandardPermissionChecker{PermissionName: permission.GetUserListPermissionName, UserId: claims.UserId},
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

	result := serializer2.SerializeMultipleTemplate(users, &serializer2.ManagerUserTemplate{}, nil)
	responseBody := serializer2.DefaultListContainer{}
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

	err = services.ChangeUserPassword(claims.UserId, requestBody.OldPassword, requestBody.NewPassword)
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

	err = services.ChangeUserNickname(claims.UserId, requestBody.Nickname)
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
		GetContainer: func() serializer2.ListContainerSerializer {
			return &serializer2.DefaultListContainer{}
		},
		GetTemplate: func() serializer2.TemplateSerializer {
			return &serializer2.BaseHistoryTemplate{}
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
