package model

import "github.com/jinzhu/gorm"

type History struct {
	gorm.Model
	BookId uint
	UserId uint
}
