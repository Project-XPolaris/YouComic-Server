package model

import "github.com/jinzhu/gorm"

type Book struct {
	gorm.Model
	Name  string
	Cover string
	Page  []Page `gorm:"foreignkey:BookId"`
	Tags  []*Tag `gorm:"many2many:book_tags;"`
}
