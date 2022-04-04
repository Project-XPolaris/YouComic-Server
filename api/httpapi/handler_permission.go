package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youcomic/api/httpapi/serializer"
	"github.com/projectxpolaris/youcomic/auth"
	ApiError "github.com/projectxpolaris/youcomic/error"
	"github.com/projectxpolaris/youcomic/permission"
	"github.com/projectxpolaris/youcomic/services"
	"github.com/projectxpolaris/youcomic/utils"
	"net/http"
)

var GetPermissionListHandler haruka.RequestHandler = func(context *haruka.Context) {
	claims, err := auth.ParseAuthHeader(context)
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.UserAuthFailError, nil)
		return
	}

	if hasPermission := permission.CheckPermissionAndServerError(context,
		&permission.StandardPermissionChecker{PermissionName: permission.GetPermissionListPermissionName, UserId: claims.UserId},
	); !hasPermission {
		return
	}

	permissionQueryBuilder := services.PermissionQueryBuilder{}
	//get page
	pagination := DefaultPagination{}
	pagination.Read(context)
	permissionQueryBuilder.SetPageFilter(pagination.Page, pagination.PageSize)

	//query filter
	filterMapping := []FilterMapping{
		{
			Lookup: "id",
			Method: "InId",
			Many:   true,
		},
		{
			Lookup: "name",
			Method: "SetNameFilter",
			Many:   true,
		}, {
			Lookup: "usergroup",
			Method: "SetUserGroupQueryFilter",
			Many:   true,
		}, {
			Lookup: "nameSearch",
			Method: "SetNameSearchQueryFilter",
			Many:   false,
		}, {
			Lookup: "user",
			Method: "SetUserFilter",
			Many:   true,
		},
	}
	for _, filter := range filterMapping {
		utils.FilterByParam(context, filter.Lookup, &permissionQueryBuilder, filter.Method, filter.Many)
	}

	count, permissions, err := permissionQueryBuilder.ReadModels()

	result := serializer.SerializeMultipleTemplate(permissions, &serializer.BasePermissionTemplate{}, nil)
	responseBody := serializer.DefaultListContainer{}
	responseBody.SerializeList(result, map[string]interface{}{
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
		"count":    count,
		"url":      context.Request.URL,
	})
	context.JSONWithStatus(responseBody, http.StatusOK)
}
