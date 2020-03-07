package controller

import (
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/serializer"
	"github.com/allentom/youcomic-api/services"
	"github.com/allentom/youcomic-api/validate"
	"github.com/gin-gonic/gin"
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
var RegisterUserHandler gin.HandlerFunc = func(context *gin.Context) {
	var err error
	requestBody := RegisterUserResponseBody{}
	err = context.ShouldBindJSON(&requestBody)
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.JsonParseError, nil)
		return
	}
	// check validate
	if isValidate := validate.RunValidatorsAndRaiseApiError(context,
		&validate.UniqUserNameValidator{Value: requestBody.Username},
		&validate.StringLengthValidator{Value: requestBody.Username, FieldName: "username", LessThan: 16, GreaterThan: 4},
		&validate.StringLengthValidator{Value: requestBody.Password, FieldName: "password", LessThan: 16, GreaterThan: 4},
		&validate.EmailValidator{Value: requestBody.Email},
	); isValidate {
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
	Username string `json:"username"`
	Password string `json:"password"`
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
var LoginUserHandler gin.HandlerFunc = func(context *gin.Context) {
	var err error
	requestBody := LoginUserRequestBody{}
	err = context.ShouldBindJSON(&requestBody)
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.JsonParseError, nil)
		return
	}

	//validate value
	if isValidate := validate.RunValidatorsAndRaiseApiError(context,
		&validate.StringLengthValidator{Value: requestBody.Username, FieldName: "username", LessThan: 16, GreaterThan: 4},
		&validate.StringLengthValidator{Value: requestBody.Password, FieldName: "password", LessThan: 16, GreaterThan: 4},
	); !isValidate {
		return
	}

	user, sign, err := services.UserLogin(requestBody.Username, requestBody.Password)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	context.JSON(http.StatusOK, UserAuthResponse{
		Id:   user.ID,
		Sign: sign,
	})
}

// get user handler
//
// path: /user/:id
//
// method: get
var GetUserHandler gin.HandlerFunc = func(context *gin.Context) {
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

	context.JSON(http.StatusOK, template)
}
