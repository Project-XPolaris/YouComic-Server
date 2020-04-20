package middleware

import (
	"fmt"
	"github.com/allentom/youcomic-api/auth"
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println(c.GetHeader("Authorization"))
		if c.Request.URL.Path == "/user/auth"{
			c.Next()
			return
		}
		_,err := auth.ParseAuthHeader(c)
		if err != nil {
			logrus.Error(err)
			ApiError.RaiseApiError(c, ApiError.UserAuthFailError, nil)
			return
		}
		c.Next()
	}
}