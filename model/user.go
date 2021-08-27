package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username         string
	Password         string
	Email            string
	Nickname         string
	Avatar           string
	YouPlusAccount   bool         `gorm:"default:false"`
	UserGroups       []*UserGroup    `gorm:"many2many:usergroup_users;"`
	History          []*History   `gorm:"foreignkey:UserId"`
	OwnCollection    []Collection `gorm:"foreignkey:Owner"`
	SubscriptionTags []*Tag       `gorm:"many2many:user_subscriptions;"`
}
