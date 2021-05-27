package model

import "gorm.io/gorm"

type Library struct {
	gorm.Model
	Path string
	Name string
	Books []*Book `gorm:"foreignkey:LibraryId"`
}