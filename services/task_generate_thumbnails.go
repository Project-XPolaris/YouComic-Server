package services

import (
	"context"
	"fmt"
	"github.com/allentom/harukap/module/task"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/module"
	"github.com/projectxpolaris/youcomic/plugin"
	"github.com/projectxpolaris/youcomic/utils"
	"github.com/sirupsen/logrus"
	"path/filepath"
)

type GenerateThumbnailTaskOption struct {
	LibraryId   int
	Force       bool
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
	*task.BaseTask
	stopFlag bool

	Library    *model.Library
	Err        error
	Option     GenerateThumbnailTaskOption
	TaskOutput *GenerateThumbnailTaskOutput
}
type GenerateThumbnailTaskOutput struct {
	LibraryId  int
	Total      int64
	Current    int64
	Skip       int64
	FileErrors []GenerateError
}

func (t *GenerateThumbnailTask) Output() (interface{}, error) {
	return t.TaskOutput, nil
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
	t.TaskOutput.FileErrors = append(t.TaskOutput.FileErrors, generateErr)
	if t.Option.OnBookError != nil {
		t.Option.OnBookError(t, generateErr)
	}
}
func (t *GenerateThumbnailTask) Start() error {
	go func() {
		defer DefaultLibraryLockPool.TryToUnlock(uint(t.TaskOutput.LibraryId))
		library, err := GetLibraryById(uint(t.TaskOutput.LibraryId))
		if err != nil {
			t.AbortError(err)
			return
		}
		t.Library = &library
		books := make([]model.Book, 0)
		err = database.Instance.Model(&library).Association("Books").Find(&books)
		if err != nil {
			t.AbortError(err)
			return
		}
		t.TaskOutput.Total = int64(len(books))
		for _, book := range books {
			t.TaskOutput.Current += 1
			thumbnailExt := filepath.Ext(book.Cover)
			thumbnailPath := filepath.Join(utils.GetThumbnailStorePath(book.ID), fmt.Sprintf("%s%s", "cover_thumbnail", thumbnailExt))
			storage := plugin.GetDefaultStorage()
			isExist, err := storage.IsExist(context.Background(), plugin.GetDefaultBucket(), thumbnailPath)
			if err != nil {
				t.AbortGenerateError(book, book.Cover, thumbnailPath, err)
				continue
			}
			if !isExist || t.Option.Force {
				bookCoverPath := filepath.Join(library.Path, book.Path, book.Cover)
				option := ThumbnailTaskOption{
					Input:   bookCoverPath,
					Output:  utils.GetThumbnailStorePath(book.ID),
					ErrChan: make(chan error),
				}
				DefaultThumbnailService.Resource <- option
				err = <-option.ErrChan
				if err != nil {
					t.Err = err
				}
			} else {
				t.TaskOutput.Skip += 1
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
	info := task.NewBaseTask("GenerateThumbnail", "0", StatusRunning)
	task := &GenerateThumbnailTask{
		BaseTask: info,
		Option:   option,
		TaskOutput: &GenerateThumbnailTaskOutput{
			LibraryId:  option.LibraryId,
			FileErrors: []GenerateError{},
		},
	}
	task.Status = StatusRunning
	module.Task.Pool.AddTask(task)
	return task, nil
}
