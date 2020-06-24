package controller

import (
	"fmt"
	"github.com/allentom/youcomic-api/auth"
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/permission"
	"github.com/allentom/youcomic-api/serializer"
	"github.com/allentom/youcomic-api/services"
	"github.com/allentom/youcomic-api/validate"
	"github.com/gin-gonic/gin"
	"net/http"
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
		GetValidators: func(v *CreateModelView) []validate.Validator {
			responseBody := v.RequestBody.(*CreateCollectionRequestBody)
			return []validate.Validator{
				&validate.StringLengthValidator{Value: responseBody.Name, FieldName: "Name", GreaterThan: 0, LessThan: 256},
			}
		},
		GetPermissions: func(v *CreateModelView) []permission.PermissionChecker {
			return []permission.PermissionChecker{
				&permission.StandardPermissionChecker{UserId: v.Claims.UserId, PermissionName: permission.CreateCollectionPermissionName},
			}
		},
	}
	view.Run()
}

var CollectionsListHandler gin.HandlerFunc = func(context *gin.Context) {
	getSerializerContext := func(v *ListView, result interface{}) map[string]interface{} {
		serializerContext := map[string]interface{}{}

		withBookContain := context.Query("withBookContain")
		if len(withBookContain) > 0 {
			queryBuilder := services.CollectionQueryBuilder{}
			containBookIdsQuery := context.QueryArray("withBookContain")
			containBookIds := make([]interface{}, 0)
			for _, bookId := range containBookIdsQuery {
				containBookIds = append(containBookIds, bookId)
			}
			queryBuilder.SetHasBookQueryFilter(containBookIds...)


			collections := result.([]model.Collection)
			collectionIds := make([]interface{}, 0)
			for _, collection := range collections {
				collectionIds = append(collectionIds, collection.ID)
			}
			queryBuilder.InId(collectionIds...)
			queryBuilder.SetPageFilter(1,len(collections))

			_, bookInCollections, _ := queryBuilder.ReadModels()
			fmt.Println(bookInCollections)
			serializerContext["bookCollections"] = bookInCollections
		}

		return serializerContext

	}
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
			{
				Lookup: "hasBook",
				Method: "SetHasBookQueryFilter",
				Many:   true,
			},
		},
		GetContainer: func() serializer.ListContainerSerializer {
			return &serializer.DefaultListContainer{}
		},
		GetTemplate: func() serializer.TemplateSerializer {
			if _,exist := context.GetQuery("withBookContain");exist {
				return &serializer.CollectionWithBookContainTemplate{}
			}
			return &serializer.BaseCollectionTemplate{}
		},
		GetSerializerContext: getSerializerContext,
	}
	view.Run()
}

type AddToCollectionRequestBody struct {
	Books []int `json:"books"`
}

var AddToCollectionHandler gin.HandlerFunc = func(context *gin.Context) {
	var requestBody AddToCollectionRequestBody
	err := DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}
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
	err := DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}
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
	err := DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}
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
	err := DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}
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

type UpdateCollectionRequestBody struct {
	Name string `form:"name" json:"name" xml:"name"  binding:"required"`
}

// update collection handler
//
// path: /collection/:id
//
// method: patch
var UpdateCollectionHandler gin.HandlerFunc = func(context *gin.Context) {

	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}

	claims, err := auth.ParseAuthHeader(context)
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.UserAuthFailError, nil)
		return
	}

	requestBody := UpdateCollectionRequestBody{}
	err = DecodeJsonBody(context, &requestBody)
	if err != nil {
		return
	}

	//validate
	if isValidate := validate.RunValidatorsAndRaiseApiError(context,
		&validate.StringLengthValidator{Value: requestBody.Name, LessThan: 256, GreaterThan: 0, FieldName: "CollectionName"},
	); !isValidate {
		return
	}

	err, collection := services.GetCollectionById(uint(id))
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}

	if collection.Owner != int(claims.UserId) {
		ApiError.RaiseApiError(context, ApiError.PermissionError, nil)
		return
	}

	err = AssignUpdateModel(&requestBody, collection)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	err = services.UpdateModel(collection, "Name")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	template := &serializer.BaseCollectionTemplate{}
	RenderTemplate(context, template, *collection)
	context.JSON(http.StatusOK, template)
}
