package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youcomic/api/httpapi/serializer"
	ApiError "github.com/projectxpolaris/youcomic/error"
	"github.com/projectxpolaris/youcomic/services"
	"os"
	"path/filepath"
)

var ReadDirectoryHandler haruka.RequestHandler = func(context *haruka.Context) {
	rootPath := context.GetQueryString("path")
	if len(rootPath) == 0 {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			ApiError.RaiseApiError(context, err, nil)
			return
		}
		rootPath = homeDir
	}
	infos, err := services.ReadDirectory(rootPath)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	data := make([]serializer.BaseFileItemTemplate, 0)
	for _, info := range infos {
		template := serializer.BaseFileItemTemplate{}
		template.Assign(info, rootPath)
		data = append(data, template)
	}
	context.JSON(
		haruka.JSON{
			"success": true,
			"data": haruka.JSON{
				"path":     rootPath,
				"sep":      string(os.PathSeparator),
				"files":    data,
				"backPath": filepath.Dir(rootPath),
			},
		},
	)
}
