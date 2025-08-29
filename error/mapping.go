package error

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youcomic/services"
)

// error => ApiError
var errorMapping = map[error]ApiError{
	JsonParseError:                  parseJsonApiError,
	UserAuthFailError:               userAuthFailedApiError,
	PermissionError:                 permissionDeniedApiError,
	RequestPathError:                requestPathApiError,
	services.UserPasswordInvalidate: invalidatePasswordApiError,
	services.RecordNotFoundError:    recordNotFoundApiError,
	LLMPluginNotAvailableError:      llmPluginNotAvailableApiError,
	LLMConfigNotFoundError:          llmConfigNotFoundApiError,
	LLMConfigInvalidError:           llmConfigInvalidApiError,
}

// add error => ApiError mapping
func RegisterApiError(err error, apiError ApiError) {
	errorMapping[err] = apiError
}

// error + context => ApiError
//
// generate api error and server response
func RaiseApiError(ctx *haruka.Context, err error, context map[string]interface{}) {
	apiError, exists := errorMapping[err]
	if !exists {
		apiError = defaultApiError
	}
	reason := apiError.Render(err, context)
	ctx.JSONWithStatus(ErrorResponseBody{
		Success: false,
		Reason:  reason,
		Code:    apiError.Code,
	}, apiError.Status)
}

func SendApiError(ctx *haruka.Context, err error, apiError ApiError, context map[string]interface{}) {
	reason := apiError.Render(err, context)
	ctx.JSONWithStatus(ErrorResponseBody{
		Success: false,
		Reason:  reason,
		Code:    apiError.Code,
	}, apiError.Status)
}
