package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/allentom/youcomic-api/auth"
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/sirupsen/logrus"
)

type AuthMiddleware struct {
}

func (m AuthMiddleware) OnRequest(c *haruka.Context) {
	if c.Request.URL.Path == "/user/auth" {
		return
	}
	claim, err := auth.ParseAuthHeader(c)
	if err != nil {
		c.Interrupt()
		logrus.Error(err)
		ApiError.RaiseApiError(c, ApiError.UserAuthFailError, nil)
		return
	}
	c.Param["claim"] = claim
}
