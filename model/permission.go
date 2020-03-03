package model

import "github.com/jinzhu/gorm"

type Permission struct {
	gorm.Model
	Name      string
	Groups []*UserGroup `gorm:"many2many:usergroup_permissions;"`
}