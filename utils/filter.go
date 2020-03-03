package utils

import (
	"github.com/gin-gonic/gin"
	"reflect"
)

func FilterByParam(controller *gin.Context, name string, queryBuilder interface{}, methodName string, many bool) {
	params := controller.QueryArray(name)
	if params == nil  || len(params) == 0{
		return
	}
	builderRef := reflect.ValueOf(queryBuilder)
	filterMethodRef := builderRef.MethodByName(methodName)
	inputs := make([]reflect.Value, len(params))
	if !many {
		inputs[0] = reflect.ValueOf(params[0])
	} else {
		for i := range params {
			inputs[i] = reflect.ValueOf(params[i])
		}
	}
	filterMethodRef.Call(inputs)
}