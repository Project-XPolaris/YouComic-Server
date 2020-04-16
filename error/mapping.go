package error

import (
	"github.com/allentom/youcomic-api/services"
	"github.com/gin-gonic/gin"
)

//error => ApiError
var errorMapping = map[error]ApiError{
	JsonParseError:    parseJsonApiError,
	UserAuthFailError: userAuthFailedApiError,
	PermissionError:   permissionDeniedApiError,
	RequestPathError:  requestPathApiError,
	services.UserPasswordInvalidate:invalidatePasswordApiError,
}

//add error => ApiError mapping
func RegisterApiError(err error, apiError ApiError) {
	errorMapping[err] = apiError
}

//error + context => ApiError
//
//generate api error and server response
func RaiseApiError(ctx *gin.Context, err error, context map[string]interface{}) {
	apiError, exists := errorMapping[err]
	if !exists {
		apiError = defaultApiError
	}
	reason := apiError.Render(err, context)
	ctx.AbortWithStatusJSON(apiError.Status,ErrorResponseBody{
		Success: false,
		Reason:  reason,
		Code:    apiError.Code,
	})
}

func SendApiError(ctx *gin.Context, err error, apiError ApiError, context map[string]interface{}) {
	reason := apiError.Render(err, context)
	ctx.AbortWithStatusJSON(apiError.Status,ErrorResponseBody{
		Success: false,
		Reason:  reason,
		Code:    apiError.Code,
	})
}
