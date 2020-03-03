package model

import "github.com/jinzhu/gorm"

type UserGroup struct {
	gorm.Model
	Name  string
	Users []*User `gorm:"many2many:usergroup_users;"`
	Permissions []*Permission `gorm:"many2many:usergroup_permissions;"`
}
