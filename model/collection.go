package model

import "gorm.io/gorm"

type Collection struct {
	gorm.Model
	Name  string
	Books []*Book `gorm:"many2many:collection_books;"`
	Users []*User `gorm:"many2many:collection_users;"`
	Owner int
}
