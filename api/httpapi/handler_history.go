package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youcomic/api/httpapi/serializer"
	ApiError "github.com/projectxpolaris/youcomic/error"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/services"
	"github.com/sirupsen/logrus"
)

var HistoryListHandler haruka.RequestHandler = func(context *haruka.Context) {
	queryBuilder := &services.HistoryQueryBuilder{}
	userClaimsInterface, _ := context.Param["claim"]
	userClaim := userClaimsInterface.(*model.User)
	queryBuilder.SetUserIdFilter(userClaim.GetUserId())

	withBook := context.GetQueryString("withBook")
	err := context.BindingInput(queryBuilder)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	view := ListView{
		Context:      context,
		Pagination:   &DefaultPagination{},
		QueryBuilder: queryBuilder,
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
		},
		GetContainer: func() serializer.ListContainerSerializer {
			return &serializer.DefaultListContainer{}
		},
		GetTemplate: func() serializer.TemplateSerializer {
			if withBook == "True" {
				return &serializer.HistoryWithBookTemplate{}
			}
			return &serializer.BaseHistoryTemplate{}
		},
	}
	view.Run()
}

// delete history handler
//
// path: /history/:id
//
// method: delete
var DeleteHistoryHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	// read id look up
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	// get user id
	userClaimsInterface, _ := context.Param["claim"]
	userClaim := userClaimsInterface.(*model.User)

	//setup query builder
	queryBuilder := &services.HistoryQueryBuilder{}
	queryBuilder.SetUserIdFilter(userClaim.GetUserId())
	queryBuilder.InId(id)

	err = queryBuilder.DeleteModels(true)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

type CreateHistoryRequestBody struct {
	BookId  int `json:"bookId"`
	PagePos int `json:"pagePos"`
}

var CreateHistoryHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	// get user id
	userClaimsInterface, _ := context.Param["claim"]
	userClaim := userClaimsInterface.(*model.User)

	// read request body
	requestBody := &CreateHistoryRequestBody{}
	err = context.ParseJson(requestBody)
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.JsonParseError, nil)
		return
	}

	history, err := services.CreateHistoryByBook(uint(requestBody.BookId), uint(requestBody.PagePos), userClaim.GetUserId())
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	serializer := serializer.HistoryWithBookTemplate{}
	serializer.Serializer(*history, map[string]interface{}{})
	context.JSON(serializer)

}
