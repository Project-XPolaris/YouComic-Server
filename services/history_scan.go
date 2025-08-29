package services

import (
	"encoding/json"
	"time"

	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
)

func SaveScanHistory(t *ScanTask) {
	// serialize error list
	errStr := ""
	if len(t.TaskOutput.SyncError) > 0 {
		b, _ := json.Marshal(t.TaskOutput.SyncError)
		errStr = string(b)
	}
	history := &model.ScanHistory{
		LibraryId:  t.TaskOutput.LibraryId,
		Total:      t.TaskOutput.Total,
		ErrorCount: len(t.TaskOutput.SyncError),
		Status:     t.Status,
		FinishedAt: time.Now().Unix(),
		Errors:     errStr,
	}
	_ = database.Instance.Create(history).Error
}
