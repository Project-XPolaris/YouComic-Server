package error

import (
	"github.com/gin-gonic/gin"
)

var errorMapping = map[error]ApiError{
	JsonParseError: parseJsonApiError,
}

func RegisterApiError(err error, apiError ApiError) {
	errorMapping[err] = apiError
}

func RaiseApiError(ctx *gin.Context, err error, context map[string]interface{}) {
	apiError, exists := errorMapping[err]
	if !exists {
		apiError = defaultApiError
	}
	reason := apiError.Render(err, context)
	ctx.JSON(apiError.Status, ErrorResponseBody{
		Success: false,
		Reason:  reason,
	})
}
