package middleware

import (
	"github.com/allentom/youcomic-api/auth"
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/user/auth"{
			c.Next()
			return
		}
		claim,err := auth.ParseAuthHeader(c)
		if err != nil {
			logrus.Error(err)
			ApiError.RaiseApiError(c, ApiError.UserAuthFailError, nil)
			return
		}
		c.Set("claim",claim)
		c.Next()
	}
}