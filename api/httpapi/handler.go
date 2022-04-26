package httpapi

import (
	"fmt"
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youcomic/config"
	ApiError "github.com/projectxpolaris/youcomic/error"
	"github.com/projectxpolaris/youcomic/youauthplugin"
)

var serviceInfoHandler haruka.RequestHandler = func(context *haruka.Context) {
	// get oauth addr
	oauthUrl, err := youauthplugin.DefaultYouAuthOauthPlugin.GetOauthUrl()
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	authMaps := []haruka.JSON{}
	configManager := config.DefaultConfigProvider.Manager
	for key := range configManager.GetStringMap("auth") {
		authType := configManager.GetString(fmt.Sprintf("auth.%s.type", key))
		enable := configManager.GetBool(fmt.Sprintf("auth.%s.enable", key))
		if !enable {
			continue
		}
		switch authType {
		case "youauth":
			oauthUrl, err = youauthplugin.DefaultYouAuthOauthPlugin.GetOauthUrl()
			if err != nil {
				ApiError.RaiseApiError(context, err, nil)
				return
			}
			authMaps = append(authMaps, haruka.JSON{
				"name": "YouAuth",
				"type": "weboauth",
				"url":  oauthUrl,
			})
		case "youplus":
			authMaps = append(authMaps, haruka.JSON{
				"type": "base",
				"url":  "/oauth/youplus",
				"name": "YouPlus",
			})
		case "origin":
			authMaps = append(authMaps, haruka.JSON{
				"type": "base",
				"url":  "/user/auth",
				"name": "Account",
			})
		}
	}
	context.JSON(haruka.JSON{
		"success": true,
		"name":    "YouComic service",
		//"authEnable":  config.Instance.EnableAuth,
		//"authUrl":     fmt.Sprintf("%s/%s", config.Instance.YouPlusUrl, "user/auth"),
		"allowPublic": false,
		"auth":        authMaps,
	})
}
