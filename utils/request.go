package utils

import "github.com/jinzhu/copier"

func GetUpdateModel(requestBody interface{}, model interface{}) error {
	return copier.Copy(model, requestBody)
}
