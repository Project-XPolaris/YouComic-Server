package model

import "gorm.io/gorm"

type Tag struct {
	gorm.Model
	Name              string
	Books             []*Book `gorm:"many2many:book_tags;"`
	Type              string
	Subscriptions []*User `gorm:"many2many:user_subscriptions;"`
}
