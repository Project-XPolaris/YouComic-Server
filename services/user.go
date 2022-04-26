package services

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/projectxpolaris/youcomic/auth"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/utils"
	"github.com/projectxpolaris/youcomic/youplus"
	"github.com/rs/xid"
	"gorm.io/gorm"
)

const (
	ProviderYouAuth = "youauth"
	ProviderYouPlus = "YouPlusService"
	ProviderSelf    = "self"
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
	err = database.Instance.Where(&model.User{Username: username, Password: password}).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, "", UserPasswordInvalidate
	}
	if err != nil {
		return nil, "", err
	}
	sign, err := auth.GenerateJWTSign(&user)
	oauth := model.Oauth{
		UserId:   user.ID,
		Uid:      fmt.Sprintf("%d", user.ID),
		Provider: ProviderSelf,
	}
	err = database.Instance.Create(&oauth).Error
	if err != nil {
		return nil, "", err
	}
	return &user, sign, nil
}
func YouPlusLogin(username string, rawPassword string) (*model.User, string, error) {
	authResult, err := youplus.DefaultYouPlusPlugin.Client.FetchUserAuth(username, rawPassword)
	if err != nil {
		return nil, "", err
	}
	if !authResult.Success {
		return nil, "", errors.New("user auth failed")
	}
	var oauthRecord model.Oauth
	err = database.Instance.Preload("User").Where("uid = ?", authResult.Uid).
		Where("provider = ?", "YouPlusServer").
		First(&oauthRecord).Error
	var user *model.User
	if oauthRecord.User != nil {
		user = oauthRecord.User
	}
	if err == gorm.ErrRecordNotFound {
		// create new user
		uid := xid.New().String()
		user = &model.User{
			Username: uid,
			Nickname: username,
		}
		err = database.Instance.Create(&user).Error
		if err != nil {
			return nil, "", err
		}
	} else {
		if err != nil {
			return nil, "", err
		}
	}
	newOauth := model.Oauth{
		Uid:         authResult.Uid,
		Provider:    "YouPlusServer",
		AccessToken: authResult.Token,
		UserId:      user.ID,
	}
	err = database.Instance.Create(&newOauth).Error
	if err != nil {
		return nil, "", err
	}
	return oauthRecord.User, authResult.Token, nil
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
	Preload []string
}

func (b *UserQueryBuilder) WithPreload(preload ...string) *UserQueryBuilder {
	b.Preload = append(b.Preload, preload...)
	return b
}
func (b *UserQueryBuilder) build() *gorm.DB {
	query := database.Instance
	query = ApplyFilters(b, query)
	if b.Preload != nil && len(b.Preload) > 0 {
		for _, key := range b.Preload {
			query = query.Preload(key)
		}
	}
	return query
}
func (b *UserQueryBuilder) ReadModels() (int64, interface{}, error) {
	query := b.build()
	var count int64 = 0
	md := make([]model.User, 0)
	err := query.Limit(b.getLimit()).Offset(b.getOffset()).Find(&md).Offset(-1).Count(&count).Error
	if err == sql.ErrNoRows {
		return 0, query, nil
	}
	return count, md, err
}

func (b *UserQueryBuilder) FirstOrNil() (*model.User, error) {
	query := b.build()
	var user model.User
	err := query.First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
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
