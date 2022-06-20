package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youcomic/auth"
	ApiError "github.com/projectxpolaris/youcomic/error"
	"github.com/sirupsen/logrus"
)

type AuthMiddleware struct {
}

var NoAuthPaths = []string{
	"/user/auth",
	"/user/auth2",
	"/oauth/youauth",
	"/oauth/token",
	"/info",
	"/oauth/youplus",
}

func (m AuthMiddleware) OnRequest(c *haruka.Context) {
	for _, path := range NoAuthPaths {
		if c.Request.URL.Path == path {
			return
		}
	}
	claim, err := auth.ParseAuthHeader(c)
	if err != nil {
		c.Abort()
		logrus.Error(err)
		ApiError.RaiseApiError(c, ApiError.UserAuthFailError, nil)
		return
	}
	c.Param["claim"] = claim
}
