package model

import "github.com/jinzhu/gorm"

type Page struct {
	gorm.Model
	Order  int
	Path   string
	BookId int
}
