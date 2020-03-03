package error

import (
	"errors"
	"net/http"
)

type ErrorResponseBody struct {
	Success bool   `json:"success"`
	Reason  string `json:"reason"`
}

var (
	JsonParseError = errors.New("parse json error")
)

type ApiError struct {
	Status int
	Render func(err error, context map[string]interface{}) string
}

var (
	defaultApiError = ApiError{
		Status: http.StatusInternalServerError,
		Render: func(err error, context map[string]interface{}) string {
			return err.Error()
		},
	}
	parseJsonApiError = ApiError{
		Status: http.StatusBadRequest,
		Render: func(err error, context map[string]interface{}) string {
			return err.Error()
		},
	}
)
