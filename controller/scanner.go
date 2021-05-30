package controller

import (
	"github.com/allentom/youcomic-api/auth"
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/permission"
	"github.com/allentom/youcomic-api/serializer"
	"github.com/allentom/youcomic-api/services"
	"github.com/gin-gonic/gin"
	"path/filepath"
	"time"
)

func init() {
	go func() {
		for {
			<-time.After(1 * time.Second)
			data := make([]serializer.TaskSerializer, 0)
			for _, task := range services.DefaultTaskPool.Tasks {
				template := serializer.TaskSerializer{}
				err := template.Serializer(task, nil)
				if err != nil {
					continue
				}
				data = append(data, template)
			}
			DefaultNotificationManager.sendJSONToAll(map[string]interface{}{
				"event": "TaskBeat",
				"data":  data,
			})
		}
	}()
}

type NewScannerRequestBody struct {
	DirPath string `json:"dir_path"`
}

var NewScannerHandler gin.HandlerFunc = func(context *gin.Context) {
	rawClaim, _ := context.Get("claim")
	claim, _ := rawClaim.(*auth.UserClaims)
	if hasPermission := permission.CheckPermissionAndServerError(context,
		&permission.StandardPermissionChecker{
			PermissionName: permission.CreateLibraryPermissionName, UserId: claim.UserId,
		},
		&permission.StandardPermissionChecker{
			PermissionName: permission.CreateBookPermissionName, UserId: claim.UserId,
		},
	); !hasPermission {
		return
	}

	var requestBody NewScannerRequestBody
	err := context.ShouldBindJSON(&requestBody)
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.JsonParseError, nil)
		return
	}
	services.DefaultTaskPool.NewLibraryAndScan(requestBody.DirPath, filepath.Base(requestBody.DirPath))
	ServerSuccessResponse(context)
}
