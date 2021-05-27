package model

import "gorm.io/gorm"

type Page struct {
	gorm.Model
	PageOrder int
	Path      string
	BookId    int
}
