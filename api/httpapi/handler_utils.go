package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youcomic/services"
	"net/http"
)

var GetThumbnailGeneratorStatus haruka.RequestHandler = func(context *haruka.Context) {
	status := services.DefaultThumbnailService.GetQueueStatus()
	context.JSONWithStatus(map[string]interface{}{
		"success":    true,
		"total":      status.Total,
		"maxQueue":   status.MaxQueue,
		"inQueue":    status.InQueue,
		"inProgress": status.InProgress,
	}, http.StatusOK)
}
