package services

import (
	"fmt"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/utils"
	"github.com/sirupsen/logrus"
	"path/filepath"
)

type GenerateThumbnailTaskOption struct {
	LibraryId   int
	OnError     func(task *GenerateThumbnailTask, err error)
	OnBookError func(task *GenerateThumbnailTask, err GenerateError)
	OnDone      func(task *GenerateThumbnailTask)
}
type GenerateError struct {
	Id       uint   `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	FilePath string `json:"filePath"`
	Error    error  `json:"error"`
}
type GenerateThumbnailTask struct {
	BaseTask
	stopFlag   bool
	LibraryId  int
	Total      int64
	Current    int64
	Skip       int64
	Library    *model.Library
	Err        error
	FileErrors []GenerateError
	Option     GenerateThumbnailTaskOption
}

func (t *GenerateThumbnailTask) Stop() error {
	t.stopFlag = true
	return nil
}
func (t *GenerateThumbnailTask) AbortError(err error) {
	t.Err = err
	t.Status = StatusError
	if t.Option.OnError != nil {
		t.Option.OnError(t, err)
	}
	logrus.Error(err)
}
func (t *GenerateThumbnailTask) AbortGenerateError(book model.Book, path string, coverPath string, err error) {
	generateErr := GenerateError{
		Id:       book.ID,
		Name:     book.Name,
		Path:     path,
		FilePath: coverPath,
		Error:    err,
	}
	t.FileErrors = append(t.FileErrors, generateErr)
	if t.Option.OnBookError != nil {
		t.Option.OnBookError(t, generateErr)
	}
}
func (t *GenerateThumbnailTask) Start() error {
	go func() {
		defer DefaultLibraryLockPool.TryToUnlock(uint(t.LibraryId))
		library, err := GetLibraryById(uint(t.LibraryId))
		if err != nil {
			t.AbortError(err)
			return
		}
		t.Library = &library
		books := make([]model.Book, 0)
		err = database.DB.Model(&library).Association("Books").Find(&books)
		if err != nil {
			t.AbortError(err)
			return
		}
		t.Total = int64(len(books))
		for _, book := range books {
			t.Current += 1
			thumbnailExt := filepath.Ext(book.Cover)
			thumbnailPath := filepath.Join(utils.GetThumbnailStorePath(book.ID), fmt.Sprintf("%s%s", "cover_thumbnail", thumbnailExt))
			if !utils.CheckFileExist(thumbnailPath) {
				bookCoverPath := filepath.Join(library.Path, book.Path, book.Cover)
				option := ThumbnailTaskOption{
					Input:   bookCoverPath,
					Output:  thumbnailPath,
					ErrChan: make(chan error),
				}
				DefaultThumbnailService.Resource <- option
				err = <-option.ErrChan
				if err != nil {

				}
			} else {
				t.Skip += 1
			}
		}
		t.Status = StatusComplete
		if t.Option.OnDone != nil {
			t.Option.OnDone(t)
		}
	}()
	return nil
}
func (p *TaskPool) NewGenerateThumbnailTask(option GenerateThumbnailTaskOption) (*GenerateThumbnailTask, error) {
	if !DefaultLibraryLockPool.TryToLock(uint(option.LibraryId)) {
		return nil, LibraryLockError
	}
	task := &GenerateThumbnailTask{
		BaseTask:   NewBaseTask(),
		LibraryId:  option.LibraryId,
		Option:     option,
		FileErrors: []GenerateError{},
	}
	task.Status = StatusRunning
	p.AddTask(task)
	task.Start()
	return task, nil
}
