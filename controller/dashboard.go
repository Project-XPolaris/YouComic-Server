package controller

import (
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/serializer"
	"github.com/allentom/youcomic-api/services"
	"github.com/allentom/youcomic-api/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

var BookCountDailySummaryHandler gin.HandlerFunc = func(context *gin.Context) {
	//get page
	pagination := DefaultPagination{}
	pagination.Read(context)

	queryBuilder := services.BooksQueryBuilder{}
	queryBuilder.SetPageFilter(pagination.Page, pagination.PageSize)

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
		{
			Lookup: "order",
			Method: "SetOrderFilter",
			Many:   false,
		},
		{
			Lookup: "collection",
			Method: "SetCollectionQueryFilter",
			Many:   true,
		},
		{
			Lookup: "tag",
			Method: "SetTagQueryFilter",
			Many:   true,
		},
		{
			Lookup: "startTime",
			Method: "SetStartTimeQueryFilter",
			Many:   false,
		},
		{
			Lookup: "endTime",
			Method: "SetEndTimeQueryFilter",
			Many:   false,
		},
		{
			Lookup: "nameSearch",
			Method: "SetNameSearchQueryFilter",
			Many:   false,
		},
		{
			Lookup: "library",
			Method: "SetLibraryQueryFilter",
			Many:   true,
		},
	}
	for _, filter := range filterMapping {
		utils.FilterByParam(context, filter.Lookup, &queryBuilder, filter.Method, filter.Many)
	}

	result, count, err := queryBuilder.GetDailyCount()
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	serializerTemplates := serializer.SerializeMultipleTemplate(result, &serializer.BookDailySummaryTemplate{}, nil)
	responseBody := serializer.DefaultListContainer{}
	responseBody.SerializeList(serializerTemplates, map[string]interface{}{
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
		"count":    count,
		"url":      context.Request.URL,
	})
	context.JSON(http.StatusOK, responseBody)
}
