package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youcomic/api/httpapi/serializer"
	"github.com/projectxpolaris/youcomic/auth"
	ApiError "github.com/projectxpolaris/youcomic/error"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/permission"
	"github.com/projectxpolaris/youcomic/services"
	"github.com/projectxpolaris/youcomic/validate"
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
var CreateTagHandler haruka.RequestHandler = func(context *haruka.Context) {
	view := CreateModelView{
		Context: context,
		CreateModel: func() interface{} {
			return &model.Tag{}
		},
		ResponseTemplate: &serializer.BaseTagTemplate{},
		RequestBody:      &CreateTagRequestBody{},
		GetPermissions: func(v *CreateModelView) []permission.PermissionChecker {
			return []permission.PermissionChecker{
				&permission.StandardPermissionChecker{UserId: v.Claims.GetUserId(), PermissionName: permission.CreateTagPermissionName},
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
var BatchTagHandler haruka.RequestHandler = func(context *haruka.Context) {
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
					&permission.StandardPermissionChecker{UserId: v.Claims.GetUserId(), PermissionName: permission.CreateTagPermissionName},
				}
			},
			Update: func(v *ModelsBatchView) []permission.PermissionChecker {
				return []permission.PermissionChecker{
					&permission.StandardPermissionChecker{UserId: v.Claims.GetUserId(), PermissionName: permission.UpdateTagPermissionName},
				}
			},
			Delete: func(v *ModelsBatchView) []permission.PermissionChecker {
				return []permission.PermissionChecker{
					&permission.StandardPermissionChecker{UserId: v.Claims.GetUserId(), PermissionName: permission.DeleteTagPermissionName},
				}
			},
		},
	}
	view.Run()
}

var TagListHandler haruka.RequestHandler = func(context *haruka.Context) {
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
			{
				Lookup: "random",
				Method: "SetRandomQueryFilter",
				Many:   false,
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

var TagBooksHandler haruka.RequestHandler = func(context *haruka.Context) {
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
	context.JSONWithStatus(responseBody, http.StatusOK)
}

type AddBookToTagRequestBody struct {
	Books []int `json:"books"`
}

var AddBooksToTagHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	var requestBody AddBookToTagRequestBody
	err = DecodeJsonBody(context, &requestBody)
	if err != nil {
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

var RemoveBooksFromTagHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	var requestBody RemoveBookFromTagRequestBody
	err = DecodeJsonBody(context, &requestBody)
	if err != nil {
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
var AddSubscriptionUser haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}
	claims := auth.GetUserClaimsFromContext(context)

	user := &model.User{Model: gorm.Model{ID: claims.ID}}
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
var RemoveSubscriptionUser haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}
	claims := auth.GetUserClaimsFromContext(context)

	user := &model.User{Model: gorm.Model{ID: claims.ID}}
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
var GetTag haruka.RequestHandler = func(context *haruka.Context) {
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
	context.JSONWithStatus(template, http.StatusOK)
}

type AddTagBooksToTagRequestBody struct {
	From uint `json:"from"`
	To   uint `json:"to"`
}

var AddTagBooksToTag haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody AddTagBooksToTagRequestBody
	err := DecodeJsonBody(context, &requestBody)
	if err != nil {
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
	Text    string `json:"text"`
	Pattern string `json:"pattern"`
}

var AnalyzeTagFromTextHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody AnalyzeTagFromTextRequestBody
	err := DecodeJsonBody(context, &requestBody)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	tags := services.MatchTag(requestBody.Text, requestBody.Pattern)
	context.JSONWithStatus(tags, http.StatusOK)
}

var ClearEmptyTagHandler haruka.RequestHandler = func(context *haruka.Context) {
	task, err := services.DefaultTaskPool.NewRemoveEmptyTagTask()
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	go task.Start()
	ServerSuccessResponse(context)
}
