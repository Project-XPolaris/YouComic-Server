package validate

import (
	ApplicationError "github.com/allentom/youcomic-api/error"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Validator interface {
	Check() (string, bool)
}

// run validators and raise api error if invalidate (first error of these)
// return false if is invalidate
//
// [validator] => error => response
func RunValidatorsAndRaiseApiError(context *gin.Context, validators ...Validator) bool{
	for _, validator := range validators {
		info, isValidate := validator.Check()
		if !isValidate {
			validateError := ApplicationError.ApiError{
				Code:   ApplicationError.RequestDataInvalidateErrorCode,
				Status: http.StatusBadRequest,
				Render: func(err error, context map[string]interface{}) string {
					return info
				},
			}
			ApplicationError.SendApiError(context, nil, validateError, nil)
			return false
		}
	}
	return true
}
