package controller

import (
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/serializer"
	"github.com/allentom/youcomic-api/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type RegisterUserResponseBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

var RegisterUserHandler gin.HandlerFunc = func(context *gin.Context) {
	var err error
	responseBody := RegisterUserResponseBody{}
	err = context.ShouldBindJSON(&responseBody)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	user := model.User{Username: responseBody.Username, Password: responseBody.Password, Email: responseBody.Email}
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

var LoginUserHandler gin.HandlerFunc = func(context *gin.Context) {
	var err error
	requestBody := LoginUserRequestBody{}
	err = context.ShouldBindJSON(&requestBody)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
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

var GetUserHandler gin.HandlerFunc = func(context *gin.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
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


