package serializer

import (
	"github.com/jinzhu/copier"
	"github.com/projectxpolaris/youcomic/model"
)

type ScanHistoryTemplate struct {
	ID         uint   `json:"id"`
	LibraryId  uint   `json:"library_id"`
	Total      int64  `json:"total"`
	ErrorCount int    `json:"error_count"`
	Status     string `json:"status"`
	StartedAt  int64  `json:"started_at"`
	FinishedAt int64  `json:"finished_at"`
}

func (t *ScanHistoryTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	var md model.ScanHistory
	md = dataModel.(model.ScanHistory)
	return copier.Copy(t, md)
}
