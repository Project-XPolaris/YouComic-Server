package serializer

import "github.com/allentom/youcomic-api/services"

type ScanTaskSerializer struct {
	ID        string `json:"id"`
	TargetDir string `json:"targetDir"`
	LibraryId uint   `json:"libraryId"`
	Name      string `json:"name"`
	Total     int64  `json:"total"`
	Current   int64  `json:"current"`
	Status    string `json:"status"`
	Created   string `json:"created"`
	CurrentDir string `json:"currentDir"`
}

func (t *ScanTaskSerializer) Serializer(dataModel interface{}, context map[string]interface{}) error {
	model := dataModel.(*services.ScanTask)
	t.ID = model.ID
	t.TargetDir = model.TargetDir
	t.LibraryId = model.LibraryId
	t.Name = model.Name
	t.Total = model.Total
	t.Current = model.Current
	t.Status = model.Status
	t.Created = model.Created.Format(timeFormat)
	t.CurrentDir = model.CurrentDir
	return nil
}
