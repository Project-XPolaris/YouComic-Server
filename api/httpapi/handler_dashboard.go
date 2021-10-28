package httpapi

import (
	"github.com/allentom/haruka"
	serializer2 "github.com/allentom/youcomic-api/api/serializer"
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/services"
	"github.com/allentom/youcomic-api/utils"
	"net/http"
)

var BookCountDailySummaryHandler haruka.RequestHandler = func(context *haruka.Context) {
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

	serializerTemplates := serializer2.SerializeMultipleTemplate(result, &serializer2.BookDailySummaryTemplate{}, nil)
	responseBody := serializer2.DefaultListContainer{}
	responseBody.SerializeList(serializerTemplates, map[string]interface{}{
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
		"count":    count,
		"url":      context.Request.URL,
	})
	context.JSONWithStatus(responseBody, http.StatusOK)
}

var TagBooksCountHandler haruka.RequestHandler = func(context *haruka.Context) {
	pagination := DefaultPagination{}
	pagination.Read(context)

	queryBuilder := services.TagQueryBuilder{}
	queryBuilder.SetPageFilter(pagination.Page, pagination.PageSize)

	filterMapping := []FilterMapping{
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
		{
			Lookup: "nameSearch",
			Method: "SetNameSearchQueryFilter",
			Many:   false,
		},
		{
			Lookup: "type",
			Method: "SetTagTypeQueryFilter",
			Many:   true,
		},
		{
			Lookup: "subscription",
			Method: "SetTagSubscriptionQueryFilter",
			Many:   true,
		},
	}
	for _, filter := range filterMapping {
		utils.FilterByParam(context, filter.Lookup, &queryBuilder, filter.Method, filter.Many)
	}

	count, result, err := queryBuilder.GetTagCount()
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	serializerTemplates := serializer2.SerializeMultipleTemplate(result, &serializer2.TagCountTemplate{}, nil)
	responseBody := serializer2.DefaultListContainer{}
	responseBody.SerializeList(serializerTemplates, map[string]interface{}{
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
		"count":    count,
		"url":      context.Request.URL,
	})
	context.JSONWithStatus(responseBody, http.StatusOK)
}

var TagTypeCountHandler haruka.RequestHandler = func(context *haruka.Context) {
	pagination := DefaultPagination{}
	pagination.Read(context)

	queryBuilder := services.TagQueryBuilder{}
	queryBuilder.SetPageFilter(pagination.Page, pagination.PageSize)

	filterMapping := []FilterMapping{
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
		{
			Lookup: "nameSearch",
			Method: "SetNameSearchQueryFilter",
			Many:   false,
		},
		{
			Lookup: "type",
			Method: "SetTagTypeQueryFilter",
			Many:   true,
		},
		{
			Lookup: "subscription",
			Method: "SetTagSubscriptionQueryFilter",
			Many:   true,
		},
	}
	for _, filter := range filterMapping {
		utils.FilterByParam(context, filter.Lookup, &queryBuilder, filter.Method, filter.Many)
	}

	count, result, err := queryBuilder.GetTagTypeCount()
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	serializerTemplates := serializer2.SerializeMultipleTemplate(result, &serializer2.TagTypeCountTemplate{}, nil)
	responseBody := serializer2.DefaultListContainer{}
	responseBody.SerializeList(serializerTemplates, map[string]interface{}{
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
		"count":    count,
		"url":      context.Request.URL,
	})
	context.JSONWithStatus(responseBody, http.StatusOK)

}
