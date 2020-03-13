package controller

import (
	"github.com/allentom/youcomic-api/auth"
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/permission"
	"github.com/allentom/youcomic-api/serializer"
	"github.com/allentom/youcomic-api/services"
	"github.com/allentom/youcomic-api/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

var GetPermissionListHandler gin.HandlerFunc = func(context *gin.Context) {
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
	permissionQueryBuilder.SetPageFilter(pagination.Page,pagination.PageSize)

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
		},
	}
	for _, filter := range filterMapping {
		utils.FilterByParam(context, filter.Lookup, &permissionQueryBuilder, filter.Method, filter.Many)
	}

	count, permissions,err := permissionQueryBuilder.ReadModels()

	result := serializer.SerializeMultipleTemplate(permissions, &serializer.BasePermissionTemplate{},nil)
	responseBody := serializer.DefaultListContainer{}
	responseBody.SerializeList(result, map[string]interface{}{
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
		"count":    count,
		"url":      context.Request.URL,
	})
	context.JSON(http.StatusOK, responseBody)
}