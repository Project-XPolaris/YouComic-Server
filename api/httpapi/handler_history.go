package httpapi

import (
	"github.com/allentom/haruka"
	serializer2 "github.com/allentom/youcomic-api/api/serializer"
	"github.com/allentom/youcomic-api/auth"
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/services"
	"github.com/sirupsen/logrus"
)

var HistoryListHandler haruka.RequestHandler = func(context *haruka.Context) {
	queryBuilder := &services.HistoryQueryBuilder{}
	userClaimsInterface, _ := context.Param["claim"]
	userClaim := userClaimsInterface.(*auth.UserClaims)
	queryBuilder.SetUserIdFilter(userClaim.UserId)

	withBook := context.GetQueryString("withBook")

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
		GetContainer: func() serializer2.ListContainerSerializer {
			return &serializer2.DefaultListContainer{}
		},
		GetTemplate: func() serializer2.TemplateSerializer {
			if withBook == "True" {
				return &serializer2.HistoryWithBookTemplate{}
			}
			return &serializer2.BaseHistoryTemplate{}
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
	userClaim := userClaimsInterface.(*auth.UserClaims)

	//setup query builder
	queryBuilder := &services.HistoryQueryBuilder{}
	queryBuilder.SetUserIdFilter(userClaim.UserId)
	queryBuilder.InId(id)

	err = queryBuilder.DeleteModels(true)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}
