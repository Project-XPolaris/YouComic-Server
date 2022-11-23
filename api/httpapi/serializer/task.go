package serializer

import (
	"github.com/projectxpolaris/youcomic/services"
)

type TaskSerializer struct {
	ID      string      `json:"id"`
	Status  string      `json:"status"`
	Created string      `json:"created"`
	Type    string      `json:"type"`
	Output  interface{} `json:"output"`
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
	//case *services.ScanTask:
	//	t.Output = SerializeScanTask(dataModel)
	//	t.Type = "ScanLibrary"
	//case *services.MatchLibraryTagTask:
	//	t.Output = SerializeMatchTask(dataModel)
	//	t.Type = "MatchLibrary"
	case *services.RenameBookDirectoryTask:
		t.Output = SerializeRenameTask(dataModel)
		t.Type = "RenameLibraryBookDirectory"
		//case *services.MoveBookTask:
		//	t.Output = SerializeMoveBookTask(dataModel)
		//	t.Type = "MoveBook"
		//case *services.RemoveEmptyTagTask:
		//	t.Output = SerializeRemoveEmptyTagTask(dataModel)
		//	t.Type = "RemoveEmptyTag"
		//case *services.WriteBookMetaTask:
		//	t.Output = SerializeWriteBookMetaTask(dataModel)
		//	t.Type = "WriteBookMeta"
		//case *services.RemoveLibraryTask:
		//	t.Output = SerializeRemoveLibraryTask(dataModel)
		//	t.Type = "RemoveLibrary"
		//case *services.GenerateThumbnailTask:
		//	t.Output = SerializeGenerateThumbnailsTask(dataModel)
		//	t.Type = "GenerateThumbnail"
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
	ErrorFile  []services.SyncError `json:"errorFiles"`
}

func NewScanLibraryDetail(output *services.ScanTaskOutput) (*ScanLibrarySerialize, error) {
	return &ScanLibrarySerialize{
		Name:       output.Name,
		Total:      output.Total,
		Current:    output.Current,
		CurrentDir: output.CurrentDir,
		ErrorFile:  output.SyncError,
		LibraryId:  output.LibraryId,
		TargetDir:  output.TargetDir,
	}, nil
}

type MatchLibrarySerialize struct {
	TargetDir  string `json:"targetDir"`
	LibraryId  uint   `json:"libraryId"`
	Name       string `json:"name"`
	Total      int64  `json:"total"`
	Current    int64  `json:"current"`
	CurrentDir string `json:"currentDir"`
}

func SerializeMatchTask(output *services.MatchLibraryTagTaskOutput) (*MatchLibrarySerialize, error) {
	t := &MatchLibrarySerialize{}
	t.TargetDir = output.TargetDir
	t.LibraryId = output.LibraryId
	t.Name = output.Name
	t.Total = output.Total
	t.Current = output.Current
	t.CurrentDir = output.CurrentDir
	return t, nil
}

type RemoveLibrarySerializer struct {
	LibraryId int    `json:"libraryId"`
	Name      string `json:"name"`
	Path      string `json:"path"`
}

func SerializeRemoveLibraryTask(output *services.RemoveLibraryTaskOutput) (*RemoveLibrarySerializer, error) {
	t := &RemoveLibrarySerializer{
		LibraryId: int(output.Library.ID),
		Name:      output.Library.Name,
		Path:      output.Library.Path,
	}
	return t, nil
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

func SerializeMoveBookTask(output *services.MoveBookTaskOutput) (*MoveBookSerializer, error) {
	t := MoveBookSerializer{}
	t.Total = output.Total
	t.Current = output.Current
	t.CurrentDir = output.CurrentDir
	return &t, nil
}

type RemoveEmptyTagSerializer struct {
	CurrentTagName string `json:"currentTagName"`
	Total          int    `json:"total"`
	Current        int    `json:"current"`
}

func SerializeRemoveEmptyTagTask(output *services.RemoveEmptyTagTaskOutput) (*RemoveEmptyTagSerializer, error) {
	t := &RemoveEmptyTagSerializer{
		Total:   output.Total,
		Current: output.Current,
	}
	if output.CurrentTag != nil {
		t.CurrentTagName = output.CurrentTag.Name
	}
	return t, nil
}

type WriteBookMetaSerializer struct {
	Current     int    `json:"current"`
	Total       int    `json:"total"`
	CurrentBook string `json:"currentBook"`
}

func SerializeWriteBookMetaTask(output *services.WriteBookMetaTaskOutput) (*WriteBookMetaSerializer, error) {
	t := &WriteBookMetaSerializer{
		Total:       output.Total,
		Current:     output.Total,
		CurrentBook: output.CurrentBook,
	}
	return t, nil
}

type GenerateThumbnailsSerializer struct {
	LibraryId  int                      `json:"libraryId"`
	Total      int64                    `json:"total"`
	Current    int64                    `json:"current"`
	Skip       int64                    `json:"skip"`
	Err        error                    `json:"err"`
	FileErrors []services.GenerateError `json:"fileErrors"`
}

func SerializeGenerateThumbnailsTask(output *services.GenerateThumbnailTaskOutput) (*GenerateThumbnailsSerializer, error) {
	t := &GenerateThumbnailsSerializer{
		LibraryId:  output.LibraryId,
		Total:      output.Total,
		Current:    output.Current,
		Skip:       output.Skip,
		FileErrors: output.FileErrors,
	}
	return t, nil
}
