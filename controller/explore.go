package controller

import (
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/serializer"
	"github.com/allentom/youcomic-api/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"path/filepath"
)

var ReadDirectoryHandler gin.HandlerFunc = func(context *gin.Context) {
	target, _ := context.GetQuery("target")
	absPath, err := filepath.Abs(target)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	items, err := services.ReadDirectory(target)
	data := serializer.SerializeMultipleTemplate(items, &serializer.FileItemSerializer{}, map[string]interface{}{"root": absPath})
	context.JSON(200, map[string]interface{}{
		"sep":   filepath.Separator,
		"items": data,
		"back":  filepath.Dir(target),
	})
}
