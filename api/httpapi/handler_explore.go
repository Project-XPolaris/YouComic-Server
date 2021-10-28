package httpapi

import (
	"github.com/allentom/haruka"
	serializer2 "github.com/allentom/youcomic-api/api/serializer"
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/services"
	"github.com/sirupsen/logrus"
	"path/filepath"
)

var ReadDirectoryHandler haruka.RequestHandler = func(context *haruka.Context) {
	target := context.GetQueryString("target")
	absPath, err := filepath.Abs(target)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	items, err := services.ReadDirectory(target)
	data := serializer2.SerializeMultipleTemplate(items, &serializer2.FileItemSerializer{}, map[string]interface{}{"root": absPath})
	context.JSONWithStatus(map[string]interface{}{
		"sep":   filepath.Separator,
		"items": data,
		"back":  filepath.Dir(target),
	}, 200)
}
