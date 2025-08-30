package httpapi

import (
	"net/http"
	"time"

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
	Text           string   `json:"text"`
	Pattern        string   `json:"pattern"`
	Texts          []string `json:"texts"`
	UseLLM         bool     `json:"useLLM"`
	CustomPrompt   string   `json:"customPrompt,omitempty"`   // 可选的自定义prompt
	ForceReprocess bool     `json:"forceReprocess,omitempty"` // 强制重新处理，忽略历史记录
}

var AnalyzeTagFromTextHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody AnalyzeTagFromTextRequestBody
	err := DecodeJsonBody(context, &requestBody)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	tags := services.MatchTagWithHistory(requestBody.Text, requestBody.Pattern, requestBody.UseLLM, requestBody.ForceReprocess, requestBody.CustomPrompt)
	context.JSONWithStatus(tags, http.StatusOK)
}

var BatchAnalyzeTagFromTextHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody AnalyzeTagFromTextRequestBody
	err := DecodeJsonBody(context, &requestBody)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	tags := services.BatchMatchTagWithHistory(requestBody.Texts, requestBody.Pattern, requestBody.UseLLM, requestBody.ForceReprocess, requestBody.CustomPrompt)
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

// GetLLMTagHistoryRequestBody LLM标签历史记录查询请求体
type GetLLMTagHistoryRequestBody struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Search   string `json:"search,omitempty"` // 搜索原始文本
}

// LLMTagHistoryResponseItem LLM标签历史记录响应项
type LLMTagHistoryResponseItem struct {
	ID           uint   `json:"id"`
	OriginalText string `json:"originalText"`
	ModelName    string `json:"modelName"`
	ModelVersion string `json:"modelVersion"`
	CustomPrompt string `json:"customPrompt"`
	Results      []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"results"`
	ProcessingTimeMs int64      `json:"processingTimeMs"`
	Success          bool       `json:"success"`
	UsageCount       int        `json:"usageCount"`
	CreatedAt        time.Time  `json:"createdAt"`
	LastUsedAt       *time.Time `json:"lastUsedAt,omitempty"`
}

// LLMTagHistoryListResponse LLM标签历史记录列表响应
type LLMTagHistoryListResponse struct {
	Count int                         `json:"count"`
	Data  []LLMTagHistoryResponseItem `json:"data"`
}

var GetLLMTagHistoryHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody GetLLMTagHistoryRequestBody
	err := DecodeJsonBody(context, &requestBody)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	// 默认值设置
	if requestBody.Page <= 0 {
		requestBody.Page = 1
	}
	if requestBody.PageSize <= 0 {
		requestBody.PageSize = 20
	}
	if requestBody.PageSize > 100 {
		requestBody.PageSize = 100 // 限制最大页面大小
	}

	count, histories, err := services.GetLLMTagHistoryList(requestBody.Page, requestBody.PageSize, requestBody.Search)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	// 转换为响应格式
	responseData := make([]LLMTagHistoryResponseItem, len(histories))
	for i, history := range histories {
		results := make([]struct {
			Name string `json:"name"`
			Type string `json:"type"`
		}, len(history.Results))

		for j, result := range history.Results {
			results[j] = struct {
				Name string `json:"name"`
				Type string `json:"type"`
			}{
				Name: result.Name,
				Type: result.Type,
			}
		}

		responseData[i] = LLMTagHistoryResponseItem{
			ID:               history.ID,
			OriginalText:     history.OriginalText,
			ModelName:        history.ModelName,
			ModelVersion:     history.ModelVersion,
			CustomPrompt:     history.CustomPrompt,
			Results:          results,
			ProcessingTimeMs: history.ProcessingTimeMs,
			Success:          history.Success,
			UsageCount:       history.UsageCount,
			CreatedAt:        history.CreatedAt,
			LastUsedAt:       history.LastUsedAt,
		}
	}

	response := LLMTagHistoryListResponse{
		Count: int(count),
		Data:  responseData,
	}

	context.JSONWithStatus(response, http.StatusOK)
}
