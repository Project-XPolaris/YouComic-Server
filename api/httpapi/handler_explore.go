package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youcomic/api/httpapi/serializer"
	ApiError "github.com/projectxpolaris/youcomic/error"
	"github.com/projectxpolaris/youcomic/services"
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
	data := serializer.SerializeMultipleTemplate(items, &serializer.FileItemSerializer{}, map[string]interface{}{"root": absPath})
	context.JSONWithStatus(map[string]interface{}{
		"sep":   filepath.Separator,
		"items": data,
		"back":  filepath.Dir(target),
	}, 200)
}
