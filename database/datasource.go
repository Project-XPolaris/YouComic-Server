package database

import (
	"github.com/allentom/harukap/datasource"
	"github.com/projectxpolaris/youcomic/model"
	"gorm.io/gorm"
)

var DefaultPlugin = &datasource.Plugin{
	OnConnected: func(db *gorm.DB) {
		Instance = db
		Instance.AutoMigrate(
			&model.Library{},
			&model.Book{},
			&model.Tag{},
			&model.Page{},
			&model.User{},
			&model.Collection{},
			&model.UserGroup{},
			&model.Permission{},
			&model.History{},
			&model.Oauth{},
		)
	},
}
