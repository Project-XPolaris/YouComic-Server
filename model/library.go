package model

import "github.com/jinzhu/gorm"

type Library struct {
	gorm.Model
	Path string
	Name string
	Books []*Book `gorm:"foreignkey:LibraryId"`
}