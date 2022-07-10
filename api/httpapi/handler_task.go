package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youcomic/api/httpapi/serializer"
	"github.com/projectxpolaris/youcomic/auth"
	ApiError "github.com/projectxpolaris/youcomic/error"
	"github.com/projectxpolaris/youcomic/permission"
	"github.com/projectxpolaris/youcomic/services"
	"net/http"
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
			status := services.DefaultThumbnailService.GetQueueStatus()
			DefaultNotificationManager.sendJSONToAll(map[string]interface{}{
				"event": "GeneratorStatusBeat",
				"data": map[string]interface{}{
					"total":      status.Total,
					"maxQueue":   status.MaxQueue,
					"inQueue":    status.InQueue,
					"inProgress": status.InProgress,
				},
			})
		}
	}()
}

type NewScannerRequestBody struct {
	DirPath string `json:"dir_path"`
}

var NewScannerHandler haruka.RequestHandler = func(context *haruka.Context) {
	claim := auth.GetUserClaimsFromContext(context)
	if hasPermission := permission.CheckPermissionAndServerError(context,
		&permission.StandardPermissionChecker{
			PermissionName: permission.CreateLibraryPermissionName, UserId: claim.ID,
		},
		&permission.StandardPermissionChecker{
			PermissionName: permission.CreateBookPermissionName, UserId: claim.ID,
		},
	); !hasPermission {
		return
	}

	var requestBody NewScannerRequestBody
	err := DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}
	services.DefaultTaskPool.NewLibraryAndScan(requestBody.DirPath, filepath.Base(requestBody.DirPath), services.ScanLibraryOption{})
	ServerSuccessResponse(context)
}

type NewRenameLibraryBookDirectoryRequestBody struct {
	Pattern string                `json:"pattern"`
	Slots   []services.RenameSlot `json:"slots"`
}

var NewRenameLibraryBookDirectoryHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}
	var requestBody NewRenameLibraryBookDirectoryRequestBody
	err = DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}
	userClaims := auth.GetUserClaimsFromContext(context)
	scanLibraryPermission := permission.StandardPermissionChecker{
		PermissionName: permission.ScanLibraryPermissionName,
		UserId:         userClaims.ID,
	}
	if hasPermission := permission.CheckPermissionAndServerError(context, &scanLibraryPermission); !hasPermission {
		return
	}
	task, err := services.DefaultTaskPool.NewRenameBookDirectoryLibraryTask(uint(id), requestBody.Pattern, requestBody.Slots)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	context.JSONWithStatus(task, http.StatusOK)
}

type NewMoveBookTaskRequestBody struct {
	BookIds []int `json:"bookIds"`
	To      int   `json:"to"`
}

var NewMoveBookTaskHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	var requestBody NewMoveBookTaskRequestBody
	err = DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}
	userClaims := auth.GetUserClaimsFromContext(context)
	scanLibraryPermission := permission.StandardPermissionChecker{
		PermissionName: permission.UpdateBookPermissionName,
		UserId:         userClaims.ID,
	}
	if hasPermission := permission.CheckPermissionAndServerError(context, &scanLibraryPermission); !hasPermission {
		return
	}
	task, err := services.DefaultTaskPool.NewMoveBookTask(requestBody.BookIds, requestBody.To)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	context.JSONWithStatus(task, http.StatusOK)
}

var WriteBookMetaTaskHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}
	library, err := services.GetLibraryById(uint(id))
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	task, err := services.DefaultTaskPool.NewWriteBookMetaTask(&library)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	context.JSONWithStatus(task, http.StatusOK)
}
