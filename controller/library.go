package controller

import (
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/permission"
	"github.com/allentom/youcomic-api/serializer"
	"github.com/allentom/youcomic-api/services"
	"github.com/gin-gonic/gin"
)

type CreateLibraryRequestBody struct {
	Name string `form:"name" json:"name" xml:"name"  binding:"required"`
	Path string `form:"path" json:"path" xml:"path"  binding:"required"`
}

var CreateLibraryHandler gin.HandlerFunc = func(context *gin.Context) {
	var requestBody CreateLibraryRequestBody
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
	context.JSON(200, template)
}

var DeleteLibraryHandler gin.HandlerFunc = func(context *gin.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}
	err = services.DeleteLibrary(uint(id))
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}
	ServerSuccessResponse(context)
}

var LibraryObjectHandler gin.HandlerFunc = func(context *gin.Context) {
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
	context.JSON(200, template)
}

var LibraryListHandler gin.HandlerFunc = func(context *gin.Context) {
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

var LibraryBatchHandler gin.HandlerFunc = func(context *gin.Context) {
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
					&permission.StandardPermissionChecker{UserId: v.Claims.UserId,PermissionName: permission.CreateLibraryPermissionName},
				}
			},
			Update: func(v *ModelsBatchView) []permission.PermissionChecker {
				return []permission.PermissionChecker{
					&permission.StandardPermissionChecker{UserId: v.Claims.UserId,PermissionName: permission.UpdateLibraryPermissionName},
				}
			},
			Delete: func(v *ModelsBatchView) []permission.PermissionChecker {
				return []permission.PermissionChecker{
					&permission.StandardPermissionChecker{UserId: v.Claims.UserId,PermissionName: permission.DeleteLibraryPermissionName},
				}
			},
		},
	}
	view.Run()
}
