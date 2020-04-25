package model

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model
	Username      string
	Password      string
	Email         string
	Nickname      string
	Avatar        string
	UserGroups     UserGroup    `gorm:"many2many:usergroup_users;"`
	OwnCollection []Collection `gorm:"foreignkey:Owner"`
	SubscriptionTags []*Tag `gorm:"many2many:user_subscriptions;"`
}
