package model

import (
	"gorm.io/gorm"
)

type Book struct {
	gorm.Model
	Name         string
	Cover        string
	History      []*History `gorm:"foreignkey:BookId"`
	Page         []Page     `gorm:"foreignkey:BookId"`
	Tags         []*Tag     `gorm:"many2many:book_tags;"`
	Path         string
	LibraryId    uint
	OriginalName string
}
