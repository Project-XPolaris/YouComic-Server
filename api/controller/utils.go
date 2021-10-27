package controller

import (
	"github.com/allentom/youcomic-api/services"
	"github.com/gin-gonic/gin"
)

var GetThumbnailGeneratorStatus gin.HandlerFunc = func(context *gin.Context) {
	status := services.DefaultThumbnailService.GetQueueStatus()
	context.JSON(200, map[string]interface{}{
		"success":    true,
		"total":      status.Total,
		"maxQueue":   status.MaxQueue,
		"inQueue":    status.InQueue,
		"inProgress": status.InProgress,
	})
}
