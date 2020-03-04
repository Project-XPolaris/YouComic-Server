package validate

import (
	"fmt"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/services"
)

type UniqBookNameValidator struct {
	Value string
}

func (v *UniqBookNameValidator) Check() (string,bool) {
	bookQueryBuilder := services.BooksQueryBuilder{}
	bookQueryBuilder.SetNameFilter(v.Value)
	var books []model.Book
	count, err := bookQueryBuilder.ReadModels(&books)
	if err != nil || count != 0 {
		return  fmt.Sprintf("name [%s] is already exist!",v.Value),false
	}
	return "",true
}
