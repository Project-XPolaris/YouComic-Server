package model

import "gorm.io/gorm"

type History struct {
	gorm.Model
	BookId  uint
	UserId  uint
	PagePos uint
}
