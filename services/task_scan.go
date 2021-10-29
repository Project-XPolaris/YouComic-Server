package services

import (
	"encoding/json"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/utils"
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

func (p *TaskPool) NewLibraryAndScan(targetPath string, name string, option ScanLibraryOption) (*ScanTask, error) {
	library, err := CreateLibrary(name, targetPath)
	if err != nil {
		return nil, err
	}
	option.Library = library
	return p.NewScanLibraryTask(option)
}
func (p *TaskPool) NewScanLibraryTask(option ScanLibraryOption) (*ScanTask, error) {
	lockSuccess := DefaultLibraryLockPool.TryToLock(option.Library.ID)
	if !lockSuccess {
		return nil, LibraryLockError
	}
	task := &ScanTask{
		BaseTask:  NewBaseTask(),
		TargetDir: option.Library.Path,
		LibraryId: option.Library.ID,
		Name:      option.Library.Name,
		Option:    &option,
		SyncError: []SyncError{},
	}
	task.Status = ScanStatusAnalyze
	p.AddTask(task)
	task.Start()
	return task, nil
}

type ScanTask struct {
	BaseTask
	TargetDir  string
	LibraryId  uint
	Name       string
	Total      int64
	Current    int64
	CurrentDir string
	SyncError  []SyncError
	stopFlag   bool
	Err        error
	Option     *ScanLibraryOption
}

func (t *ScanTask) Start() error {
	go func() {
		defer func() {
			DefaultLibraryLockPool.TryToUnlock(t.LibraryId)
		}()
		t.scannerDir()
	}()
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
	t.SyncError = append(t.SyncError, syncError)
	if t.Option.OnDirError != nil {
		t.Option.OnDirError(t, syncError)
	}
}
func (t *ScanTask) scannerDir() {
	// sync with exist book
	var library model.Library
	err := database.DB.Where("id = ?", t.LibraryId).Preload("Books").Find(&library).Error
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
		TargetPath:   t.TargetDir,
		PageExt:      utils.DefaultScanPageExt,
		MinPageCount: 4,
	}
	err = scanner.Scan()
	if err != nil {
		t.AbortTaskError(err)
		return
	}
	t.Total = scanner.Total
	t.Status = ScanStatusAdd
	// create library
	for _, item := range scanner.Result {
		if t.stopFlag {
			t.Status = StatusStop
			if t.Option.OnStop != nil {
				t.Option.OnStop(t)
			}
			return
		}
		t.Current += 1
		t.CurrentDir = filepath.Base(item.DirPath)
		relativePath, _ := filepath.Rel(t.TargetDir, item.DirPath)
		isExist := false
		for _, book := range library.Books {
			if book.Path == relativePath {
				isExist = true
				break
			}
		}
		if isExist {
			continue
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
				continue
			}
			jsonFile.Close()
		}
		book := model.Book{
			Name:      filepath.Base(item.DirPath),
			LibraryId: t.LibraryId,
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
		err = database.DB.Save(&book).Error
		if err != nil {
			t.AbortFileError(item.DirPath, err)
			continue
		}

		// for pages
		savePages := make([]model.Page, 0)
		for idx, pageName := range item.Pages {
			page := model.Page{
				Path:      pageName,
				BookId:    int(book.Model.ID),
				PageOrder: idx + 1,
			}
			savePages = append(savePages, page)

		}
		err = database.DB.Save(savePages).Error
		if err != nil {
			t.AbortFileError(item.DirPath, err)
			continue
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
			}
		}
		coverThumbnailStorePath := utils.GetThumbnailStorePath(book.ID)
		option := ThumbnailTaskOption{
			Input:   filepath.Join(t.TargetDir, book.Path, book.Cover),
			Output:  coverThumbnailStorePath,
			ErrChan: make(chan error),
		}
		go func(item utils.ScannerResult) {
			DefaultThumbnailService.Resource <- option
			err = <-option.ErrChan
			if err != nil {
				// use page as cover
				for _, page := range item.Pages {
					option.Input = filepath.Join(t.TargetDir, book.Path, page)
					DefaultThumbnailService.Resource <- option
					err = <-option.ErrChan
					if err == nil {
						break
					}
				}
			}
			if err != nil {
				t.AbortFileError(item.DirPath, err)
			}
		}(item)

	}
	t.Status = StatusComplete
	if t.Option.OnDone != nil {
		t.Option.OnDone(t)
	}
}
