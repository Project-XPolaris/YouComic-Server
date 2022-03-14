package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/allentom/youcomic-api/auth"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/utils"
	"github.com/allentom/youcomic-api/youplus"
	"github.com/project-xpolaris/youplustoolkit/youplus/rpc"
	"gorm.io/gorm"
)

var (
	UserPasswordInvalidate = errors.New("invalidate user password")
	UserNotFoundError      = errors.New("user not found")
)

func RegisterUser(user *model.User) error {
	password, err := utils.EncryptSha1WithSalt(user.Password)
	user.Password = password
	if err != nil {
		return err
	}
	err = database.Instance.Save(user).Error
	if err != nil {
		return err
	}
	err = database.Instance.Model(user).Update("nickname", fmt.Sprintf("user_%d", user.ID)).Error
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
	err = database.Instance.Where(&model.User{Username: username, Password: password}).Find(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, "", UserPasswordInvalidate
	}
	if err != nil {
		return nil, "", err
	}
	sign, err := auth.GenerateJWTSign(&user)
	if err != nil {
		return nil, "", err
	}
	return &user, sign, nil
}
func YouPlusLogin(username string, rawPassword string) (*model.User, string, error) {
	client, err := youplus.DefaultYouPlusPlugin.RPCClient.GetClient()
	if err != nil {
		return nil, "", err
	}
	result, err := client.GenerateToken(context.Background(), &rpc.GenerateTokenRequest{
		Username: &username,
		Password: &rawPassword,
	})
	if err != nil {
		return nil, "", err
	}
	if !*result.Success {
		return nil, "", errors.New(*result.Reason)
	}
	var accountCount int64
	err = database.Instance.Model(&model.User{}).Where("username = ?", username).Count(&accountCount).Error
	if err != nil {
		return nil, "", err
	}
	var account model.User
	if accountCount == 0 {
		var defaultUserGroup model.UserGroup
		err = database.Instance.Where("name = ?", DefaultUserGroupName).First(&defaultUserGroup).Error
		if err != nil {
			return nil, "", err
		}
		account = model.User{
			Username:       username,
			Nickname:       username,
			YouPlusAccount: true,
			UserGroups: []*model.UserGroup{
				&defaultUserGroup,
			},
		}
		err = database.Instance.Create(&account).Error
		if err != nil {
			return nil, "", err
		}
	} else {
		err = database.Instance.Where("username = ?", username).First(&account).Error
		if err != nil {
			return nil, "", err
		}
	}
	sign, err := auth.GenerateJWTSign(&account)
	if err != nil {
		return nil, "", err
	}
	return &account, sign, nil

}

type UserQueryBuilder struct {
	IdQueryFilter
	DefaultPageFilter
	NameQueryFilter
	UserToUserGroupQueryFilter
	UserNameSearchQueryFilter
	NicknameSearchQueryFilter
	UserNameQueryFilter
	OrderQueryFilter
}

func (b *UserQueryBuilder) ReadModels() (int64, interface{}, error) {
	query := database.Instance
	query = ApplyFilters(b, query)
	var count int64 = 0
	md := make([]model.User, 0)
	err := query.Limit(b.getLimit()).Offset(b.getOffset()).Find(&md).Offset(-1).Count(&count).Error
	if err == sql.ErrNoRows {
		return 0, query, nil
	}
	return count, md, err
}

type UserToUserGroupQueryFilter struct {
	userGroups []interface{}
}

func (f *UserToUserGroupQueryFilter) SetUserGroupQueryFilter(userGroups ...interface{}) {
	for _, userGroupId := range userGroups {
		if len(userGroupId.(string)) > 0 {
			f.userGroups = append(f.userGroups, userGroupId)
		}
	}
}
func (f UserToUserGroupQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.userGroups != nil && len(f.userGroups) != 0 {
		return db.Joins(
			"inner join usergroup_users on user_id = id",
		).Where("usergroup_users.user_group_id in (?)", f.userGroups)
	}
	return db
}

type UserNameSearchQueryFilter struct {
	nameSearch interface{}
}

func (f UserNameSearchQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.nameSearch != nil && len(f.nameSearch.(string)) != 0 {
		return db.Where("username like ?", fmt.Sprintf("%%%s%%", f.nameSearch))
	}
	return db
}

func (f *UserNameSearchQueryFilter) SetNameSearchQueryFilter(nameSearch interface{}) {
	if len(nameSearch.(string)) > 0 {
		f.nameSearch = nameSearch
	}
}

type NicknameSearchQueryFilter struct {
	nicknameSearch interface{}
}

func (f NicknameSearchQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.nicknameSearch != nil && len(f.nicknameSearch.(string)) != 0 {
		return db.Where("nickname like ?", fmt.Sprintf("%%%s%%", f.nicknameSearch))
	}
	return db
}

func (f *NicknameSearchQueryFilter) SetNicknameSearchQueryFilter(nameSearch interface{}) {
	if len(nameSearch.(string)) > 0 {
		f.nicknameSearch = nameSearch
	}
}

type UserNameQueryFilter struct {
	Names []interface{}
}

func (f *UserNameQueryFilter) SetUserNameFilter(names ...interface{}) {
	for _, name := range names {
		if len(name.(string)) != 0 {
			f.Names = append(f.Names, name)
		}
	}

}
func (f UserNameQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.Names != nil && len(f.Names) != 0 {
		return db.Where("username in (?)", f.Names)
	}
	return db
}

//change user password
func ChangeUserPassword(userId uint, oldRawPassword string, newRawPassword string) error {
	oldPassword, err := utils.EncryptSha1WithSalt(oldRawPassword)
	var user model.User
	err = database.Instance.Where(&model.User{Model: gorm.Model{ID: userId}, Password: oldPassword}).Find(&user).Error
	if err == gorm.ErrRecordNotFound {
		return UserPasswordInvalidate
	}
	newPassword, err := utils.EncryptSha1WithSalt(newRawPassword)
	err = database.Instance.Model(&user).Update("password", newPassword).Error
	return err
}

//change user nickname
func ChangeUserNickname(userId uint, nickname string) error {
	var user model.User
	err := database.Instance.Where(&model.User{Model: gorm.Model{ID: userId}}).Find(&user).Error
	if err == gorm.ErrRecordNotFound {
		return UserNotFoundError
	}
	err = database.Instance.Model(&user).Update("nickname", nickname).Error
	return err
}
