package controller

import (
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/serializer"
	"github.com/allentom/youcomic-api/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type CreateTagRequestBody struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

var CreateTagHandler gin.HandlerFunc = func(context *gin.Context) {
	view := CreateModelView{
		Context: context,
		CreateModel: func() interface{} {
			return &model.Tag{}
		},
		ResponseTemplate: &serializer.BaseTagTemplate{},
		RequestBody:      &CreateTagRequestBody{},
	}

	view.Run()
}

var BatchTagHandler gin.HandlerFunc = func(context *gin.Context) {
	view := ModelsBatchView{
		Context: context,
		CreateModel: func() interface{} {
			return &model.Tag{}
		},
		AllowOperations: []BatchOperation{
			Create, Update, Delete,
		},
		AllowUpdateField: []string{
			"name",
		},
	}
	view.Run()
}

var TagListHandler gin.HandlerFunc = func(context *gin.Context) {
	view := ListView{
		Context:      context,
		Pagination:   &DefaultPagination{},
		QueryBuilder: &services.TagQueryBuilder{},
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
		},
		GetContainer: func() serializer.ListContainerSerializer {
			return &serializer.DefaultListContainer{}
		},
		GetTemplate: func() serializer.TemplateSerializer {
			return &serializer.BaseTagTemplate{}
		},
	}
	view.Run()
}

var TagBooksHandler gin.HandlerFunc = func(context *gin.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	pagination := DefaultPagination{}
	pagination.Read(context)

	count, books, err := services.GetTagBooks(uint(id), pagination.Page, pagination.PageSize)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	result := serializer.SerializeMultipleTemplate(books, &serializer.BaseBookTemplate{}, nil)
	responseBody := serializer.DefaultListContainer{}
	responseBody.SerializeList(result, map[string]interface{}{
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
		"count":    count,
		"url":      context.Request.URL,
	})
	context.JSON(http.StatusOK, responseBody)
}

type AddBookToTagRequestBody struct {
	Books []int `json:"books"`
}

var AddBooksToTagHandler gin.HandlerFunc = func(context *gin.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	var requestBody AddBookToTagRequestBody
	err = context.BindJSON(&requestBody)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	err = services.AddBooksToTag(id,requestBody.Books)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}


type RemoveBookFromTagRequestBody struct {
	Books []int `json:"books"`
}

var RemoveBooksFromTagHandler gin.HandlerFunc = func(context *gin.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	var requestBody RemoveBookFromTagRequestBody
	err = context.BindJSON(&requestBody)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	err = services.RemoveBooksFromTag(id,requestBody.Books)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}