package controller

import (
	"github.com/allentom/youcomic-api/auth"
	ApiError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/permission"
	"github.com/allentom/youcomic-api/services"
	"github.com/gin-gonic/gin"
)

type NewScannerRequestBody struct {
	DirPath string `json:"dir_path"`
}

var NewScannerHandler gin.HandlerFunc = func(context *gin.Context) {
	rawClaim, _ := context.Get("claim")
	claim, _ := rawClaim.(*auth.UserClaims)
	if hasPermission := permission.CheckPermissionAndServerError(context,
		&permission.StandardPermissionChecker{
			PermissionName: permission.CreateLibraryPermissionName, UserId: claim.UserId,
		},
		&permission.StandardPermissionChecker{
			PermissionName: permission.CreateBookPermissionName, UserId: claim.UserId,
		},
	); !hasPermission {
		return
	}

	var requestBody NewScannerRequestBody
	err := context.ShouldBindJSON(&requestBody)
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.JsonParseError, nil)
		return
	}
	services.DefaultScanTaskPool.StartTask(requestBody.DirPath)
	ServerSuccessResponse(context)
}
