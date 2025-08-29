package error

import (
	"errors"
	"net/http"
)

//error + error code => render => api error => response body

type ErrorResponseBody struct {
	Success bool   `json:"success"`
	Reason  string `json:"reason"`
	Code    string `json:"code"`
}

var (
	JsonParseError             = errors.New("parse json error")
	UserAuthFailError          = errors.New("user auth fail")
	PermissionError            = errors.New("permission denied")
	RequestPathError           = errors.New("request path error")
	LLMPluginNotAvailableError = errors.New("LLM plugin not available")
	LLMConfigNotFoundError     = errors.New("LLM config not found")
	LLMConfigInvalidError      = errors.New("LLM config invalid")
)

var (
	DefaultErrorCode               = "9999"
	UserAuthFailErrorCode          = "1001"
	ParseJsonErrorCode             = "1002"
	PermissionErrorCode            = "1003"
	RequestDataInvalidateErrorCode = "1004"
	RequestPathInvalidateErrorCode = "1005"
	InvalidatePasswordErrorCode    = "1006"
	RecordNotFoundErrorCode        = "1007"
	LLMPluginNotAvailableErrorCode = "2001"
	LLMConfigNotFoundErrorCode     = "2002"
	LLMConfigInvalidErrorCode      = "2003"
)

type ApiError struct {
	Status int
	Render func(err error, context map[string]interface{}) string
	Code   string
}

var (
	defaultApiError = ApiError{
		Status: http.StatusInternalServerError,
		Render: func(err error, context map[string]interface{}) string {
			return err.Error()
		},
		Code: DefaultErrorCode,
	}
	parseJsonApiError = ApiError{
		Status: http.StatusBadRequest,
		Render: func(err error, context map[string]interface{}) string {
			return err.Error()
		},
		Code: ParseJsonErrorCode,
	}
	userAuthFailedApiError = ApiError{
		Status: http.StatusForbidden,
		Render: func(err error, context map[string]interface{}) string {
			return err.Error()
		},
		Code: UserAuthFailErrorCode,
	}
	permissionDeniedApiError = ApiError{
		Status: http.StatusForbidden,
		Render: func(err error, context map[string]interface{}) string {
			return err.Error()
		},
		Code: PermissionErrorCode,
	}
	requestPathApiError = ApiError{
		Status: http.StatusBadRequest,
		Render: func(err error, context map[string]interface{}) string {
			return err.Error()
		},
		Code: RequestPathInvalidateErrorCode,
	}
	invalidatePasswordApiError = ApiError{
		Status: http.StatusUnauthorized,
		Render: func(err error, context map[string]interface{}) string {
			return "username or password invalidate"
		},
		Code: InvalidatePasswordErrorCode,
	}
	recordNotFoundApiError = ApiError{
		Status: http.StatusNotFound,
		Render: func(err error, context map[string]interface{}) string {
			return "record not found"
		},
		Code: RecordNotFoundErrorCode,
	}
	llmPluginNotAvailableApiError = ApiError{
		Status: http.StatusServiceUnavailable,
		Render: func(err error, context map[string]interface{}) string {
			return "LLM plugin not available"
		},
		Code: LLMPluginNotAvailableErrorCode,
	}
	llmConfigNotFoundApiError = ApiError{
		Status: http.StatusNotFound,
		Render: func(err error, context map[string]interface{}) string {
			return "LLM config not found"
		},
		Code: LLMConfigNotFoundErrorCode,
	}
	llmConfigInvalidApiError = ApiError{
		Status: http.StatusBadRequest,
		Render: func(err error, context map[string]interface{}) string {
			return "LLM config invalid"
		},
		Code: LLMConfigInvalidErrorCode,
	}
)
