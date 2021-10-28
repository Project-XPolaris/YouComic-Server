package httpapi

import (
	"github.com/allentom/haruka"
	serializer2 "github.com/allentom/youcomic-api/api/serializer"
	"github.com/allentom/youcomic-api/auth"
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/permission"
	"github.com/allentom/youcomic-api/services"
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
		UserId:         (rawUserClaims.(*auth.UserClaims)).UserId,
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
	template := serializer2.BaseLibraryTemplate{}
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
		UserId:         (rawUserClaims.(*auth.UserClaims)).UserId,
	}
	if hasPermission := permission.CheckPermissionAndServerError(context, &deleteLibraryPermission); !hasPermission {
		return
	}
	_, err = services.DefaultTaskPool.NewRemoveLibraryTask(id)
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

	template := serializer2.BaseLibraryTemplate{}
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
		GetContainer: func() serializer2.ListContainerSerializer {
			return &serializer2.DefaultListContainer{}
		},
		GetTemplate: func() serializer2.TemplateSerializer {
			return &serializer2.BaseLibraryTemplate{}
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
					&permission.StandardPermissionChecker{UserId: v.Claims.UserId, PermissionName: permission.CreateLibraryPermissionName},
				}
			},
			Update: func(v *ModelsBatchView) []permission.PermissionChecker {
				return []permission.PermissionChecker{
					&permission.StandardPermissionChecker{UserId: v.Claims.UserId, PermissionName: permission.UpdateLibraryPermissionName},
				}
			},
			Delete: func(v *ModelsBatchView) []permission.PermissionChecker {
				return []permission.PermissionChecker{
					&permission.StandardPermissionChecker{UserId: v.Claims.UserId, PermissionName: permission.DeleteLibraryPermissionName},
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
		UserId:         (rawUserClaims.(*auth.UserClaims)).UserId,
	}
	if hasPermission := permission.CheckPermissionAndServerError(context, &scanLibraryPermission); !hasPermission {
		return
	}
	task, err := services.ScanLibrary(uint(id))
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
		UserId:         (rawUserClaims.(*auth.UserClaims)).UserId,
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
		UserId:         (rawUserClaims.(*auth.UserClaims)).UserId,
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
