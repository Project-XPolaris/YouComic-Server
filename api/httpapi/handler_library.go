package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youcomic/api/httpapi/serializer"
	"github.com/projectxpolaris/youcomic/auth"
	ApiError "github.com/projectxpolaris/youcomic/error"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/permission"
	"github.com/projectxpolaris/youcomic/services"
	"net/http"
)

type CreateLibraryRequestBody struct {
	Name string `form:"name" json:"name" xml:"name"  binding:"required"`
	Path string `form:"path" json:"path" xml:"path"  binding:"required"`
}

var CreateLibraryHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody CreateLibraryRequestBody

	rawUserClaims, _ := context.Param["claim"]
	createLibraryPermission := permission.StandardPermissionChecker{
		PermissionName: permission.CreateLibraryPermissionName,
		UserId:         (rawUserClaims.(auth.JwtClaims)).GetUserId(),
	}
	if hasPermission := permission.CheckPermissionAndServerError(context, &createLibraryPermission); !hasPermission {
		return
	}
	err := DecodeJsonBody(context, &requestBody)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	library, err := services.CreateLibrary(requestBody.Name, requestBody.Path)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	template := serializer.BaseLibraryTemplate{}
	err = template.Serializer(*library, nil)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	context.JSONWithStatus(template, http.StatusOK)
}

var DeleteLibraryHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}
	rawUserClaims, _ := context.Param["claim"]
	deleteLibraryPermission := permission.StandardPermissionChecker{
		PermissionName: permission.DeleteLibraryPermissionName,
		UserId:         (rawUserClaims.(auth.JwtClaims)).GetUserId(),
	}
	if hasPermission := permission.CheckPermissionAndServerError(context, &deleteLibraryPermission); !hasPermission {
		return
	}
	_, err = services.DefaultTaskPool.NewRemoveLibraryTask(services.RemoveLibraryTaskOption{
		LibraryId: id,
		OnError: func(task *services.RemoveLibraryTask, taskError error) {
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": EventRemoveLibraryTaskError,
				"data":  serializer.NewTaskTemplate(task),
			})
		},
		OnDone: func(task *services.RemoveLibraryTask) {
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": EventRemoveLibraryTaskDone,
				"data":  serializer.NewTaskTemplate(task),
			})
		},
	})
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}
	ServerSuccessResponse(context)
}

var LibraryObjectHandler haruka.RequestHandler = func(context *haruka.Context) {
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}

	library, err := services.GetLibraryById(uint(id))
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}

	template := serializer.BaseLibraryTemplate{}
	err = template.Serializer(library, nil)
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}
	context.JSONWithStatus(template, http.StatusOK)
}

var LibraryListHandler haruka.RequestHandler = func(context *haruka.Context) {
	queryBuilder := &services.LibraryQueryBuilder{}
	view := ListView{
		Context:      context,
		Pagination:   &DefaultPagination{},
		QueryBuilder: queryBuilder,
		FilterMapping: []FilterMapping{
			{
				Lookup: "id",
				Method: "InId",
				Many:   true,
			},
			{
				Lookup: "order",
				Method: "SetOrderFilter",
				Many:   false,
			},
			{
				Lookup: "name",
				Method: "SetNameFilter",
				Many:   true,
			},
		},
		GetContainer: func() serializer.ListContainerSerializer {
			return &serializer.DefaultListContainer{}
		},
		GetTemplate: func() serializer.TemplateSerializer {
			return &serializer.BaseLibraryTemplate{}
		},
	}
	view.Run()
}

var LibraryBatchHandler haruka.RequestHandler = func(context *haruka.Context) {
	view := ModelsBatchView{
		Context: context,
		AllowUpdateField: []string{
			"path",
		},
		AllowOperations: []BatchOperation{
			Create, Update,
		},
		CreateModel: func() interface{} {
			return &model.Library{}
		},
		Permissions: map[BatchOperation]func(v *ModelsBatchView) []permission.PermissionChecker{
			Create: func(v *ModelsBatchView) []permission.PermissionChecker {
				return []permission.PermissionChecker{
					&permission.StandardPermissionChecker{UserId: v.Claims.GetUserId(), PermissionName: permission.CreateLibraryPermissionName},
				}
			},
			Update: func(v *ModelsBatchView) []permission.PermissionChecker {
				return []permission.PermissionChecker{
					&permission.StandardPermissionChecker{UserId: v.Claims.GetUserId(), PermissionName: permission.UpdateLibraryPermissionName},
				}
			},
			Delete: func(v *ModelsBatchView) []permission.PermissionChecker {
				return []permission.PermissionChecker{
					&permission.StandardPermissionChecker{UserId: v.Claims.GetUserId(), PermissionName: permission.DeleteLibraryPermissionName},
				}
			},
		},
	}
	view.Run()
}

var ScanLibraryHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}
	rawUserClaims, _ := context.Param["claim"]
	scanLibraryPermission := permission.StandardPermissionChecker{
		PermissionName: permission.ScanLibraryPermissionName,
		UserId:         (rawUserClaims.(auth.JwtClaims)).GetUserId(),
	}
	if hasPermission := permission.CheckPermissionAndServerError(context, &scanLibraryPermission); !hasPermission {
		return
	}
	task, err := services.ScanLibrary(uint(id), services.ScanLibraryOption{
		OnDone: func(task *services.ScanTask) {
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": EventScanTaskDone,
				"data":  serializer.NewTaskTemplate(task),
			})
		},
		OnError: func(task *services.ScanTask, err error) {
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": EventScanTaskError,
				"data":  serializer.NewTaskTemplate(task),
			})
		},
		OnStop: func(task *services.ScanTask) {
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": EventScanTaskStop,
				"data":  serializer.NewTaskTemplate(task),
			})
		},
		//OnDirError: func(task *services.ScanTask, syncErr services.SyncError) {
		//	DefaultNotificationManager.sendJSONToAll(haruka.JSON{
		//		"event": EventScanTaskFileError,
		//		"data":  serializer.NewTaskTemplate(task),
		//	})
		//},
	})
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}
	context.JSONWithStatus(task, http.StatusOK)
}

var StopLibraryScanHandler haruka.RequestHandler = func(context *haruka.Context) {
	id := context.GetQueryString("id")
	rawUserClaims, _ := context.Param["claim"]
	scanLibraryPermission := permission.StandardPermissionChecker{
		PermissionName: permission.ScanLibraryPermissionName,
		UserId:         (rawUserClaims.(auth.JwtClaims)).GetUserId(),
	}
	if hasPermission := permission.CheckPermissionAndServerError(context, &scanLibraryPermission); !hasPermission {
		return
	}
	services.DefaultTaskPool.StopTask(id)
	ServerSuccessResponse(context)
}

var NewLibraryMatchTagHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}
	strategy := context.GetQueryString("strategy")
	if len(strategy) == 0 {
		strategy = "fillEmpty"
	}
	rawUserClaims, _ := context.Param["claim"]
	scanLibraryPermission := permission.StandardPermissionChecker{
		PermissionName: permission.ScanLibraryPermissionName,
		UserId:         (rawUserClaims.(auth.JwtClaims)).GetUserId(),
	}
	if hasPermission := permission.CheckPermissionAndServerError(context, &scanLibraryPermission); !hasPermission {
		return
	}
	task, err := services.DefaultTaskPool.NewMatchLibraryTagTask(uint(id), strategy)
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}
	context.JSONWithStatus(task, http.StatusOK)
}

var NewLibraryGenerateThumbnailsHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}
	rawUserClaims, _ := context.Param["claim"]
	scanLibraryPermission := permission.StandardPermissionChecker{
		PermissionName: permission.ScanLibraryPermissionName,
		UserId:         (rawUserClaims.(auth.JwtClaims)).GetUserId(),
	}
	if hasPermission := permission.CheckPermissionAndServerError(context, &scanLibraryPermission); !hasPermission {
		return
	}
	task, err := services.DefaultTaskPool.NewGenerateThumbnailTask(services.GenerateThumbnailTaskOption{
		LibraryId: id,
		OnError: func(task *services.GenerateThumbnailTask, err error) {
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": EventGenerateThumbnailTaskError,
				"data":  serializer.NewTaskTemplate(task),
			})
		},
		OnBookError: func(task *services.GenerateThumbnailTask, err services.GenerateError) {
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": EventGenerateThumbnailTaskFileError,
				"data":  serializer.NewTaskTemplate(task),
			})
		},
		OnDone: func(task *services.GenerateThumbnailTask) {
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": EventGenerateThumbnailTaskDone,
				"data":  serializer.NewTaskTemplate(task),
			})
		},
	})
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}
	context.JSONWithStatus(task, http.StatusOK)
}
