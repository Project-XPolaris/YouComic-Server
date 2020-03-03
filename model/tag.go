package model

import "github.com/jinzhu/gorm"

type Tag struct {
	gorm.Model
	Name string
	Books []*Book `gorm:"many2many:book_tags;"`
	Type string
}
