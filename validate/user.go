package validate

import (
	"fmt"
	"github.com/allentom/youcomic-api/services"
)

type UniqUserNameValidator struct {
	Value string
}

func (v *UniqUserNameValidator) Check() (string,bool) {
	userQueryBuilder := services.UserQueryBuilder{}
	userQueryBuilder.SetUserNameFilter(v.Value)
	count, _, err := userQueryBuilder.ReadModels()
	if err != nil {
		return "",false
	}
	if count == 0{
		return "",true
	}else{
		return fmt.Sprintf("username [%s] is already exist!",v.Value),false
	}
}
