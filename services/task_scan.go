package services

import (
	"encoding/json"
	"github.com/allentom/harukap/module/task"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/module"
	"github.com/projectxpolaris/youcomic/utils"
	"io/ioutil"
	"os"
	"path/filepath"
)

type SyncError struct {
	Path  string `json:"path"`
	Name  string `json:"name"`
	Error error  `json:"error"`
}
type ScanLibraryOption struct {
	Library    *model.Library
	OnDone     func(task *ScanTask)
	OnError    func(task *ScanTask, err error)
	OnDirError func(task *ScanTask, syncErr SyncError)
	OnStop     func(task *ScanTask)
}

func NewLibraryAndScan(targetPath string, name string, option ScanLibraryOption) (*ScanTask, error) {
	library, err := CreateLibrary(name, targetPath)
	if err != nil {
		return nil, err
	}
	option.Library = library
	return NewScanLibraryTask(option)
}
func NewScanLibraryTask(option ScanLibraryOption) (*ScanTask, error) {
	lockSuccess := DefaultLibraryLockPool.TryToLock(option.Library.ID)
	if !lockSuccess {
		return nil, LibraryLockError
	}
	info := task.NewBaseTask("ScanLibrary", "all", TaskStatusInit)
	task := &ScanTask{
		BaseTask: info,
		Option:   &option,
		TaskOutput: &ScanTaskOutput{
			Id:        info.Id,
			TargetDir: option.Library.Path,
			LibraryId: option.Library.ID,
			Name:      option.Library.Name,
			SyncError: []SyncError{},
		},
	}
	task.Status = ScanStatusAnalyze
	module.Task.Pool.AddTask(task)
	return task, nil
}

type ScanTaskOutput struct {
	Id         string
	TargetDir  string
	LibraryId  uint
	Name       string
	Total      int64
	Current    int64
	CurrentDir string
	SyncError  []SyncError
}
type ScanTask struct {
	*task.BaseTask
	stopFlag   bool
	Option     *ScanLibraryOption
	TaskOutput *ScanTaskOutput
}

func (t *ScanTask) Output() (interface{}, error) {
	return t.TaskOutput, nil
}

func (t *ScanTask) Start() error {
	defer func() {
		DefaultLibraryLockPool.TryToUnlock(t.TaskOutput.LibraryId)
	}()
	t.scannerDir()
	return nil
}
func (t *ScanTask) Stop() error {
	t.stopFlag = true
	return nil
}
func (t *ScanTask) AbortTaskError(err error) {
	t.Status = StatusError
	t.Err = err
	if t.Option.OnError != nil {
		t.Option.OnError(t, err)
	}
}
func (t *ScanTask) AbortFileError(path string, err error) {
	syncError := SyncError{
		Path:  path,
		Name:  filepath.Base(path),
		Error: err,
	}
	t.TaskOutput.SyncError = append(t.TaskOutput.SyncError, syncError)
	if t.Option.OnDirError != nil {
		t.Option.OnDirError(t, syncError)
	}
}
func (t *ScanTask) scannerDir() {
	// sync with exist book
	var library model.Library
	err := database.Instance.Where("id = ?", t.TaskOutput.LibraryId).Preload("Books").Find(&library).Error
	if err != nil {
		t.AbortTaskError(err)
		return
	}
	removeBookIds := make([]int, 0)
	for _, book := range library.Books {
		if !utils.CheckFileExist(filepath.Join(library.Path, book.Path)) {
			removeBookIds = append(removeBookIds, int(book.ID))
		}
	}
	if len(removeBookIds) > 0 {
		err = DeleteBooks(removeBookIds...)
		if err != nil {
			t.AbortTaskError(err)
			return
		}
	}
	scanner := utils.Scanner{
		TargetPath: t.TaskOutput.TargetDir,
	}
	var count int64 = 0
	err = scanner.Scan(func(result utils.ScannerResult) {
		count++
	})
	t.TaskOutput.Total = count
	if err != nil {
		t.AbortTaskError(err)
		return
	}
	t.Status = ScanStatusAdd
	err = scanner.Scan(func(item utils.ScannerResult) {
		if t.stopFlag {
			t.Status = StatusStop
			if t.Option.OnStop != nil {
				t.Option.OnStop(t)
			}
			return
		}
		t.TaskOutput.Current += 1
		t.TaskOutput.CurrentDir = filepath.Base(item.DirPath)
		relativePath, _ := filepath.Rel(t.TaskOutput.TargetDir, item.DirPath)
		isExist := false
		for _, book := range library.Books {
			if book.Path == relativePath {
				isExist = true
				break
			}
		}
		if isExist {
			return
		}
		// try to find out meta file
		metaFilePath := filepath.Join(item.DirPath, "youcomic_meta.json")
		meta := BookMeta{}
		if utils.CheckFileExist(metaFilePath) {
			jsonFile, err := os.Open(metaFilePath)
			byteValue, _ := ioutil.ReadAll(jsonFile)
			err = json.Unmarshal(byteValue, &meta)
			if err != nil {
				t.AbortFileError(item.DirPath, err)
				jsonFile.Close()
				return
			}
			jsonFile.Close()
		}
		book := model.Book{
			Name:      filepath.Base(item.DirPath),
			LibraryId: t.TaskOutput.LibraryId,
			Path:      relativePath,
			Cover:     item.CoverName,
		}
		if len(meta.OriginalName) > 0 {
			book.OriginalName = meta.OriginalName
		}
		if len(meta.Cover) > 0 && utils.CheckFileExist(filepath.Join(item.DirPath, meta.Cover)) {
			book.Cover = meta.Cover
		}
		if len(meta.Title) > 0 {
			book.Name = meta.Title
		}
		err = database.Instance.Save(&book).Error
		if err != nil {
			t.AbortFileError(item.DirPath, err)
			return
		}
		thumbnailsSource := make([]string, 0)
		thumbnailsSource = append(thumbnailsSource, filepath.Join(t.TaskOutput.TargetDir, book.Path, book.Cover))
		// for pages
		savePages := make([]model.Page, 0)

		for idx, pageName := range item.Pages {
			page := model.Page{
				Path:      pageName,
				BookId:    int(book.Model.ID),
				PageOrder: idx + 1,
			}
			thumbnailsSource = append(thumbnailsSource, filepath.Join(t.TaskOutput.TargetDir, book.Path, pageName))
			savePages = append(savePages, page)

		}
		err = database.Instance.Save(savePages).Error
		if err != nil {
			t.AbortFileError(item.DirPath, err)
			return
		}
		// for tags
		if meta.Tags != nil && len(meta.Tags) > 0 {
			tags := make([]*model.Tag, 0)
			for _, metaTag := range meta.Tags {
				tags = append(tags, &model.Tag{
					Name: metaTag.Name,
					Type: metaTag.Type,
				})
			}
			err = AddOrCreateTagToBook(&book, tags, FillEmpty)
			if err != nil {
				t.AbortFileError(item.DirPath, err)
				return
			}
		}

		coverThumbnailStorePath := utils.GetThumbnailStorePath(book.ID)
		for _, sourcePath := range thumbnailsSource {
			_, err := GenerateCoverThumbnail(sourcePath, coverThumbnailStorePath)
			if err == nil {
				break
			}
		}
	})
	t.Status = StatusComplete
	if t.Option.OnDone != nil {
		t.Option.OnDone(t)
	}
}
