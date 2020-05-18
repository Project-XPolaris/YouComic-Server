package controller

import (
	"github.com/allentom/youcomic-api/auth"
	"github.com/allentom/youcomic-api/serializer"
	"github.com/allentom/youcomic-api/services"
	"github.com/gin-gonic/gin"
)

var HistoryListHandler gin.HandlerFunc = func(context *gin.Context) {
	queryBuilder := &services.HistoryQueryBuilder{}
	userClaimsInterface, _ := context.Get("claim")
	userClaim := userClaimsInterface.(*auth.UserClaims)
	queryBuilder.SetUserIdFilter(userClaim.UserId)

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
			return &serializer.BaseHistoryTemplate{}
		},
	}
	view.Run()
}
