package model

import "gorm.io/gorm"

type Page struct {
	gorm.Model
	Order  int
	Path   string
	BookId int
}
