package model

import "gorm.io/gorm"

type Oauth struct {
	gorm.Model
	Uid          string
	UserId       uint
	AccessToken  string
	RefreshToken string
	Provider     string
	User         *User
}
