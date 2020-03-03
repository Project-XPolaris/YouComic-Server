package controller

import (
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/serializer"
	"github.com/allentom/youcomic-api/services"
	"github.com/gin-gonic/gin"
)

type CreateCollectionRequestBody struct {
	Name string `json:"name"`
}

var CreateCollectionHandler gin.HandlerFunc = func(context *gin.Context) {
	view := CreateModelView{
		Context: context,
		CreateModel: func() interface{} {
			return &model.Collection{}
		},
		ResponseTemplate: &serializer.BaseCollectionTemplate{},
		RequestBody:      &CreateCollectionRequestBody{},
		OnBeforeCreate: func(v *CreateModelView, modelToCreate interface{}) {
			dataModel := modelToCreate.(*model.Collection)
			dataModel.Owner = int(v.Claims.UserId)
		},
	}
	view.Run()
}

var CollectionsListHandler gin.HandlerFunc = func(context *gin.Context) {
	view := ListView{
		Context:      context,
		Pagination:   &DefaultPagination{},
		QueryBuilder: &services.CollectionQueryBuilder{},
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
				Lookup: "owner",
				Method: "SetOwnerQueryFilter",
				Many:   true,
			}, {
				Lookup: "user",
				Method: "SetUsersQueryFilter",
				Many:   true,
			},
			{
				Lookup: "aboutuser",
				Method: "SetUsersAndOwnerQueryFilter",
				Many:   true,
			},
			{
				Lookup: "nameSearch",
				Method: "SetNameSearchQueryFilter",
				Many:   false,
			},
		},
		GetContainer: func() serializer.ListContainerSerializer {
			return &serializer.DefaultListContainer{}
		},
		GetTemplate: func() serializer.TemplateSerializer {
			return &serializer.BaseCollectionTemplate{}
		},
	}
	view.Run()
}

type AddToCollectionRequestBody struct {
	Books []int `json:"books"`
}

var AddToCollectionHandler gin.HandlerFunc = func(context *gin.Context) {
	var requestBody AddToCollectionRequestBody
	DecodeJsonBody(context, &requestBody)
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	err = services.AddBooksToCollection(uint(id), requestBody.Books...)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

type RemoveFromCollectionRequestBody struct {
	Books []int `json:"books"`
}

var DeleteFromCollectionHandler gin.HandlerFunc = func(context *gin.Context) {
	var requestBody RemoveFromCollectionRequestBody
	DecodeJsonBody(context, &requestBody)
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	err = services.RemoveBooksFromCollection(uint(id), requestBody.Books...)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

type AddUsersToCollectionRequestBody struct {
	Users []int `json:"users"`
}

var AddUsersToCollectionHandler gin.HandlerFunc = func(context *gin.Context) {
	var requestBody AddUsersToCollectionRequestBody
	DecodeJsonBody(context, &requestBody)
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	err = services.AddUsersToCollection(uint(id), requestBody.Users...)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

type RemoveUsersFromCollectionRequestBody struct {
	Users []int `json:"users"`
}

var DeleteUsersFromCollectionHandler gin.HandlerFunc = func(context *gin.Context) {
	var requestBody RemoveUsersFromCollectionRequestBody
	DecodeJsonBody(context, &requestBody)
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	err = services.RemoveUsersFromCollection(uint(id), requestBody.Users...)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

var DeleteCollectionHandler gin.HandlerFunc = func(context *gin.Context) {
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	err = services.DeleteModels(&model.Collection{}, id)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}
