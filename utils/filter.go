package utils

import (
	"github.com/allentom/haruka"
	"reflect"
)

func FilterByParam(controller *haruka.Context, name string, queryBuilder interface{}, methodName string, many bool) {
	params := controller.GetQueryStrings(name)
	if params == nil || len(params) == 0 {
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
