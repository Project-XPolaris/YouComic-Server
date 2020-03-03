package services

import (
	"fmt"
	"github.com/allentom/youcomic-api/auth"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/utils"
)

func RegisterUser(user *model.User) error {
	password, err := utils.EncryptSha1WithSalt(user.Password)
	user.Password = password
	if err != nil {
		return err
	}
	err = database.DB.Save(user).Error
	if err != nil {
		return err
	}
	err = database.DB.Model(user).Update("nickname", fmt.Sprintf("user_%d", user.ID)).Error
	if err != nil {
		return err
	}
	return nil
}

func UserLogin(username string, rawPassword string) (*model.User, string, error) {
	var err error
	password, err := utils.EncryptSha1WithSalt(rawPassword)
	if err != nil {
		return nil, "", err
	}
	var user model.User
	err = database.DB.Where(&model.User{Username: username, Password: password}).Find(&user).Error
	if err != nil {
		return nil, "", err
	}
	sign, err := auth.GenerateJWTSign(&user)
	if err != nil {
		return nil, "", err
	}
	return &user, sign, nil
}

type UserQueryBuilder struct {
	IdQueryFilter
	DefaultPageFilter
}

func (b *UserQueryBuilder) ReadModels() (int, interface{}, error) {
	query := database.DB
	query = ApplyFilters(b, query)
	var count = 0
	md := make([]model.User, 0)
	err := query.Limit(b.PageSize).Offset(b.getOffset()).Find(&md).Offset(-1).Count(&count).Error
	return count, md, err
}

