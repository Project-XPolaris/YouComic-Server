package controller

import (
	"github.com/allentom/youcomic-api/auth"
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/permission"
	"github.com/allentom/youcomic-api/serializer"
	"github.com/allentom/youcomic-api/services"
	"github.com/allentom/youcomic-api/validate"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
)

type CreateTagRequestBody struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// create book handler
//
// path: /tags
//
// method: post
var CreateTagHandler gin.HandlerFunc = func(context *gin.Context) {
	view := CreateModelView{
		Context: context,
		CreateModel: func() interface{} {
			return &model.Tag{}
		},
		ResponseTemplate: &serializer.BaseTagTemplate{},
		RequestBody:      &CreateTagRequestBody{},
		GetPermissions: func(v *CreateModelView) []permission.PermissionChecker {
			return []permission.PermissionChecker{
				&permission.StandardPermissionChecker{UserId: v.Claims.UserId, PermissionName: permission.CreateTagPermissionName},
			}
		},
		GetValidators: func(v *CreateModelView) []validate.Validator {
			requestBody := v.RequestBody.(*CreateTagRequestBody)
			return []validate.Validator{
				&validate.StringLengthValidator{Value: requestBody.Name, FieldName: "Name", GreaterThan: 0, LessThan: 256},
				&validate.StringLengthValidator{Value: requestBody.Type, FieldName: "Type", GreaterThan: 0, LessThan: 256},
			}
		},
	}
	view.Run()
}

// tag batch handler
//
// path: /tags/batch
//
// method: post
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
		Permissions: map[BatchOperation]func(v *ModelsBatchView) []permission.PermissionChecker{
			Create: func(v *ModelsBatchView) []permission.PermissionChecker {
				return []permission.PermissionChecker{
					&permission.StandardPermissionChecker{UserId: v.Claims.UserId, PermissionName: permission.CreateTagPermissionName},
				}
			},
			Update: func(v *ModelsBatchView) []permission.PermissionChecker {
				return []permission.PermissionChecker{
					&permission.StandardPermissionChecker{UserId: v.Claims.UserId, PermissionName: permission.UpdateTagPermissionName},
				}
			},
			Delete: func(v *ModelsBatchView) []permission.PermissionChecker {
				return []permission.PermissionChecker{
					&permission.StandardPermissionChecker{UserId: v.Claims.UserId, PermissionName: permission.DeleteTagPermissionName},
				}
			},
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
			{
				Lookup: "subscription",
				Method: "SetTagSubscriptionQueryFilter",
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
	err = services.AddBooksToTag(id, requestBody.Books)
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
	err = services.RemoveBooksFromTag(id, requestBody.Books)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

type AddSubscriptionRequestBody struct {
}

// add user to tag handler
//
// path: /tag/:id/subscription
//
// method: put
var AddSubscriptionUser gin.HandlerFunc = func(context *gin.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}
	claims := auth.GetUserClaimsFromContext(context)

	user := &model.User{Model: gorm.Model{ID: claims.UserId}}
	err = services.AddTagSubscription(uint(id), user)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

// remove user from tag handler
//
// path: /tag/:id/subscription
//
// method: delete
var RemoveSubscriptionUser gin.HandlerFunc = func(context *gin.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}
	claims := auth.GetUserClaimsFromContext(context)

	user := &model.User{Model: gorm.Model{ID: claims.UserId}}
	err = services.RemoveTagSubscription(uint(id), user)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

// get tag handler
//
// path: /tag/:id
//
// method: get
var GetTag gin.HandlerFunc = func(context *gin.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}

	tag, err := services.GetTagById(uint(id))
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	template := &serializer.BaseTagTemplate{}
	RenderTemplate(context, template, tag)
	context.JSON(http.StatusOK, template)
}

type AddTagBooksToTagRequestBody struct {
	From uint `json:"from"`
	To   uint `json:"to"`
}

var AddTagBooksToTag gin.HandlerFunc = func(context *gin.Context) {
	var requestBody AddTagBooksToTagRequestBody
	err := context.BindJSON(&requestBody)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	err = services.TagAdd(requestBody.From, requestBody.To)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

type AnalyzeTagFromTextRequestBody struct {
	Text string `json:"text"`
}
var AnalyzeTagFromTextHandler gin.HandlerFunc = func(context *gin.Context) {
	var requestBody AnalyzeTagFromTextRequestBody
	err := context.BindJSON(&requestBody)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	tags := services.MatchTag(requestBody.Text)
	context.JSON(200,tags)
}