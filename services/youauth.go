package services

import (
	"fmt"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/plugin"
	"github.com/rs/xid"
)

const YouAuthProvider = "youauth"

func GenerateYouAuthToken(code string) (string, string, uint, error) {
	tokens, err := plugin.DefaultYouAuthOauthPlugin.Client.GetAccessToken(code)
	if err != nil {
		return "", "", 0, err
	}
	currentUserResponse, err := plugin.DefaultYouAuthOauthPlugin.Client.GetCurrentUser(tokens.AccessToken)
	if err != nil {
		return "", "", 0, err
	}
	// check if user exists
	uid := fmt.Sprintf("%d", currentUserResponse.Id)
	historyOauth := make([]model.Oauth, 0)
	err = database.Instance.Where("uid = ?", uid).
		Where("provider = ?", YouAuthProvider).
		Preload("User").
		Find(&historyOauth).Error
	if err != nil {
		return "", "", 0, err
	}
	var user *model.User
	if len(historyOauth) == 0 {
		username := xid.New().String()
		// create new user
		user = &model.User{
			Username: username,
			Nickname: fmt.Sprintf("user_%s", username),
		}
		err = database.Instance.Create(&user).Error
		if err != nil {
			return "", "", 0, err
		}
	} else {
		user = historyOauth[0].User
	}

	oauthRecord := model.Oauth{
		Uid:          fmt.Sprintf("%d", currentUserResponse.Id),
		UserId:       user.ID,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		Provider:     YouAuthProvider,
	}
	err = database.Instance.Create(&oauthRecord).Error
	if err != nil {
		return "", "", 0, err
	}
	return tokens.AccessToken, currentUserResponse.Username, user.ID, nil
}

func refreshToken(accessToken string) (string, error) {
	tokenRecord := model.Oauth{}
	err := database.Instance.Where("access_token = ?", accessToken).First(&tokenRecord).Error
	if err != nil {
		return "", err
	}
	token, err := plugin.DefaultYouAuthOauthPlugin.Client.RefreshAccessToken(tokenRecord.RefreshToken)
	if err != nil {
		return "", err
	}
	err = database.Instance.Delete(&tokenRecord).Error
	if err != nil {
		return "", err
	}
	newOauthRecord := model.Oauth{
		UserId:       tokenRecord.UserId,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}
	err = database.Instance.Create(&newOauthRecord).Error
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}
