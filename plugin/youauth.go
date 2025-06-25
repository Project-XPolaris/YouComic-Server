package plugin

import (
	"errors"
	"strconv"
	"sync"

	"github.com/allentom/harukap/commons"
	"github.com/allentom/harukap/plugins/youauth"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/module"
	"gorm.io/gorm"
)

var DefaultYouAuthOauthPlugin = &youauth.OauthPlugin{}

// 用于保护用户创建操作的互斥锁
var userCreationMutex sync.Map

func CreateYouAuthPlugin() {
	DefaultYouAuthOauthPlugin = &youauth.OauthPlugin{}
	DefaultYouAuthOauthPlugin.AuthFromToken = func(token string) (commons.AuthUser, error) {
		return GetUserByYouAuthToken(token)
	}
	DefaultYouAuthOauthPlugin.PasswordAuthUrl = "/oauth/youauth/password"
	module.Auth.Plugins = append(module.Auth.Plugins, DefaultYouAuthOauthPlugin.GetOauthPlugin(), DefaultYouAuthOauthPlugin.GetPasswordPlugin())
}
func SaveUserByYouAuthToken(accessToken string) (*model.User, error) {
	youAuthUser, err := DefaultYouAuthOauthPlugin.Client.GetCurrentUser(accessToken)
	if err != nil {
		return nil, err
	}
	// 创建用户和oauth记录
	uid := strconv.Itoa(youAuthUser.Id)

	// 获取或创建该uid的互斥锁
	mutex, _ := userCreationMutex.LoadOrStore(uid, &sync.Mutex{})
	userMutex := mutex.(*sync.Mutex)
	userMutex.Lock()
	defer userMutex.Unlock()

	// 使用事务来确保原子性
	var user *model.User
	err = database.Instance.Transaction(func(tx *gorm.DB) error {
		// 尝试查找或创建用户
		err = tx.Model(&model.User{}).Where("uid = ?", uid).First(&user).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				user = &model.User{
					Username: youAuthUser.Username,
					Uid:      uid,
				}
				err = tx.Create(user).Error
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		// 检查是否已存在oauth记录
		var existingOauth model.Oauth
		err = tx.Model(&model.Oauth{}).Where("access_token = ? AND provider = ?", accessToken, "youauth").First(&existingOauth).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// 只有在oauth记录不存在时才创建
		if errors.Is(err, gorm.ErrRecordNotFound) {
			oauthRecord := model.Oauth{
				Provider:    "youauth",
				AccessToken: accessToken,
				Uid:         uid,
				UserId:      user.ID,
			}
			err = tx.Create(&oauthRecord).Error
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByYouAuthToken(accessToken string) (*model.User, error) {
	var oauthRecord model.Oauth
	// 检查是否存在，不存在则创建用户和oauth记录
	err := database.Instance.Model(&model.Oauth{}).Preload("User").Where("access_token = ?", accessToken).
		Where("provider = ?", "youauth").
		Find(&oauthRecord).Error
	if err != nil {
		return nil, err
	}
	if oauthRecord.User == nil {
		user, err := SaveUserByYouAuthToken(accessToken)
		if err != nil {
			return nil, err
		}
		return user, nil
	}
	_, err = DefaultYouAuthOauthPlugin.Client.GetCurrentUser(accessToken)
	if err != nil {
		return nil, err
	}
	return oauthRecord.User, nil
}
