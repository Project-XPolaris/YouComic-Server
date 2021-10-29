package serializer

import (
	"github.com/allentom/youcomic-api/services"
)

type TaskSerializer struct {
	ID      string      `json:"id"`
	Status  string      `json:"status"`
	Created string      `json:"created"`
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
}

func NewTaskTemplate(dataModel interface{}) *TaskSerializer {
	template := &TaskSerializer{}
	template.Serializer(dataModel, map[string]interface{}{})
	return template
}

func (t *TaskSerializer) Serializer(dataModel interface{}, context map[string]interface{}) error {
	task := dataModel.(services.Task)
	t.ID = task.GetBaseInfo().ID
	t.Created = task.GetBaseInfo().Created.Format(timeFormat)
	t.Status = task.GetBaseInfo().Status
	switch dataModel.(type) {
	case *services.ScanTask:
		t.Data = SerializeScanTask(dataModel)
		t.Type = "ScanLibrary"
	case *services.MatchLibraryTagTask:
		t.Data = SerializeMatchTask(dataModel)
		t.Type = "MatchLibrary"
	case *services.RenameBookDirectoryTask:
		t.Data = SerializeRenameTask(dataModel)
		t.Type = "RenameLibraryBookDirectory"
	case *services.MoveBookTask:
		t.Data = SerializeMoveBookTask(dataModel)
		t.Type = "MoveBook"
	case *services.RemoveEmptyTagTask:
		t.Data = SerializeRemoveEmptyTagTask(dataModel)
		t.Type = "RemoveEmptyTag"
	case *services.WriteBookMetaTask:
		t.Data = SerializeWriteBookMetaTask(dataModel)
		t.Type = "WriteBookMeta"
	case *services.RemoveLibraryTask:
		t.Data = SerializeRemoveLibraryTask(dataModel)
		t.Type = "RemoveLibrary"
	case *services.GenerateThumbnailTask:
		t.Data = SerializeGenerateThumbnailsTask(dataModel)
		t.Type = "GenerateThumbnail"
	}
	return nil
}

type ScanLibrarySerialize struct {
	TargetDir  string               `json:"targetDir"`
	LibraryId  uint                 `json:"libraryId"`
	Name       string               `json:"name"`
	Total      int64                `json:"total"`
	Current    int64                `json:"current"`
	CurrentDir string               `json:"currentDir"`
	Err        string               `json:"err"`
	ErrorFile  []services.SyncError `json:"errorFiles"`
}

func SerializeScanTask(dataModel interface{}) ScanLibrarySerialize {
	model := dataModel.(*services.ScanTask)
	t := ScanLibrarySerialize{}
	t.TargetDir = model.TargetDir
	t.LibraryId = model.LibraryId
	t.Name = model.Name
	t.Total = model.Total
	t.Current = model.Current
	t.CurrentDir = model.CurrentDir
	t.ErrorFile = model.SyncError
	if model.Err != nil {
		t.Err = model.Err.Error()
	}
	return t
}

type MatchLibrarySerialize struct {
	TargetDir  string `json:"targetDir"`
	LibraryId  uint   `json:"libraryId"`
	Name       string `json:"name"`
	Total      int64  `json:"total"`
	Current    int64  `json:"current"`
	CurrentDir string `json:"currentDir"`
}

func SerializeMatchTask(dataModel interface{}) MatchLibrarySerialize {
	model := dataModel.(*services.MatchLibraryTagTask)
	t := MatchLibrarySerialize{}
	t.TargetDir = model.TargetDir
	t.LibraryId = model.LibraryId
	t.Name = model.Name
	t.Total = model.Total
	t.Current = model.Current
	t.CurrentDir = model.CurrentDir
	return t
}

type RenameLibraryBookDirectorySerialize struct {
	TargetDir  string `json:"targetDir"`
	LibraryId  uint   `json:"libraryId"`
	Name       string `json:"name"`
	Total      int64  `json:"total"`
	Current    int64  `json:"current"`
	CurrentDir string `json:"currentDir"`
}

func SerializeRenameTask(dataModel interface{}) RenameLibraryBookDirectorySerialize {
	model := dataModel.(*services.RenameBookDirectoryTask)
	t := RenameLibraryBookDirectorySerialize{}
	t.TargetDir = model.TargetDir
	t.LibraryId = model.LibraryId
	t.Name = model.Name
	t.Total = model.Total
	t.Current = model.Current
	t.CurrentDir = model.CurrentDir
	return t
}

type MoveBookSerializer struct {
	CurrentDir string `json:"currentDir"`
	Total      int    `json:"total"`
	Current    int    `json:"current"`
}

func SerializeMoveBookTask(dataModel interface{}) MoveBookSerializer {
	model := dataModel.(*services.MoveBookTask)
	t := MoveBookSerializer{}
	t.Total = model.Total
	t.Current = model.Current
	t.CurrentDir = model.CurrentDir
	return t
}

type RemoveEmptyTagSerializer struct {
	CurrentTagName string `json:"currentTagName"`
	Total          int    `json:"total"`
	Current        int    `json:"current"`
}

func SerializeRemoveEmptyTagTask(dataModel interface{}) RemoveEmptyTagSerializer {
	model := dataModel.(*services.RemoveEmptyTagTask)
	t := RemoveEmptyTagSerializer{}
	t.Total = model.Total
	t.Current = model.Current
	if model.CurrentTag != nil {
		t.CurrentTagName = model.CurrentTag.Name
	}
	return t
}

type WriteBookMetaSerializer struct {
	Current     int    `json:"current"`
	Total       int    `json:"total"`
	CurrentBook string `json:"currentBook"`
}

func SerializeWriteBookMetaTask(dataModel interface{}) WriteBookMetaSerializer {
	model := dataModel.(*services.WriteBookMetaTask)
	t := WriteBookMetaSerializer{
		Total:       model.Total,
		Current:     model.Total,
		CurrentBook: model.CurrentBook,
	}
	return t
}

type RemoveLibrarySerializer struct {
	LibraryId int    `json:"libraryId"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Err       string `json:"err"`
}

func SerializeRemoveLibraryTask(dataModel interface{}) RemoveLibrarySerializer {
	model := dataModel.(*services.RemoveLibraryTask)
	t := RemoveLibrarySerializer{
		LibraryId: model.LibraryId,
	}
	if model.Err != nil {
		t.Err = model.Err.Error()
	}
	if model.Library != nil {
		t.Name = model.Library.Name
		t.Path = model.Library.Path
	}
	return t
}

type GenerateThumbnailsSerializer struct {
	LibraryId   int                      `json:"libraryId"`
	Total       int64                    `json:"total"`
	Current     int64                    `json:"current"`
	Skip        int64                    `json:"skip"`
	Err         error                    `json:"err"`
	FileErrors  []services.GenerateError `json:"fileErrors"`
	LibraryName string                   `json:"libraryName"`
}

func SerializeGenerateThumbnailsTask(dataModel interface{}) GenerateThumbnailsSerializer {
	model := dataModel.(*services.GenerateThumbnailTask)
	t := GenerateThumbnailsSerializer{
		LibraryId:  model.LibraryId,
		Total:      model.Total,
		Current:    model.Current,
		Skip:       model.Skip,
		Err:        model.Err,
		FileErrors: model.FileErrors,
	}
	if model.Library != nil {
		t.LibraryName = model.Library.Name
	}
	return t
}
