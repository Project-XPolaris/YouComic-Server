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
	JsonParseError    = errors.New("parse json error")
	UserAuthFailError = errors.New("user auth fail")
	PermissionError = errors.New("permission denied")
	RequestPathError  = errors.New("request path error")
)

var (
	DefaultErrorCode      = "9999"
	UserAuthFailErrorCode = "1001"
	ParseJsonErrorCode    = "1002"
	PermissionErrorCode = "1003"
	RequestDataInvalidateErrorCode = "1004"
	RequestPathInvalidateErrorCode = "1005"
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
)
