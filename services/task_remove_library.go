package services

import (
	context2 "context"
	"fmt"
	"github.com/allentom/harukap/module/task"
	appconfig "github.com/projectxpolaris/youcomic/config"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/module"
	"github.com/projectxpolaris/youcomic/plugin"
	"github.com/sirupsen/logrus"
	"path"
)

type RemoveLibraryTaskOption struct {
	LibraryId int
	OnError   func(task *RemoveLibraryTask, err error)
	OnDone    func(task *RemoveLibraryTask)
}
type RemoveLibraryTask struct {
	*task.BaseTask
	stopFlag   bool
	LibraryId  int
	Library    *model.Library
	Err        error
	Option     RemoveLibraryTaskOption
	TaskOutput *RemoveLibraryTaskOutput
}
type RemoveLibraryTaskOutput struct {
	Library *model.Library
}

func (t *RemoveLibraryTask) Output() (interface{}, error) {
	return t.TaskOutput, nil
}

func (t *RemoveLibraryTask) Stop() error {
	t.stopFlag = true
	return nil
}
func (t *RemoveLibraryTask) AbortError(err error) {
	t.Err = err
	t.Status = StatusError
	if t.Option.OnError != nil {
		t.Option.OnError(t, err)
	}
	logrus.Error(err)
}
func (t *RemoveLibraryTask) Start() error {
	defer DefaultLibraryLockPool.TryToUnlock(uint(t.LibraryId))
	books := make([]model.Book, 0)
	err := database.Instance.Model(t.Library).Association("Books").Find(&books)
	if err != nil {
		t.AbortError(err)
		return err
	}
	for _, book := range books {
		err = database.Instance.Model(&book).Association("Tags").Clear()
		if err != nil {
			t.AbortError(err)
			return err
		}
		err = database.Instance.Model(&book).Association("Collections").Clear()
		if err != nil {
			t.AbortError(err)
			return err
		}
		err = database.Instance.Unscoped().Delete(model.Page{}, "book_id = ?", book.ID).Error
		if err != nil {
			t.AbortError(err)
			return err
		}
		err = database.Instance.Unscoped().Delete(model.History{}, "book_id = ?", book.ID).Error
		if err != nil {
			t.AbortError(err)
			return err
		}
	}

	err = database.Instance.Unscoped().Delete(model.Book{}, "library_id = ?", t.Library.ID).Error
	if err != nil {
		t.AbortError(err)
		return err
	}
	err = database.Instance.Unscoped().Delete(t.Library).Error
	if err != nil {
		t.AbortError(err)
		return err
	}
	for _, book := range books {
		storage := plugin.GetDefaultStorage()
		thumbnailExt := path.Ext(book.Cover)
		thumbnail := path.Join(appconfig.Instance.Store.Root, "generate", fmt.Sprintf("%d", book.ID), fmt.Sprintf("cover_thumbnail%s", thumbnailExt))
		err := storage.Delete(context2.Background(), plugin.GetDefaultBucket(), thumbnail)
		if err != nil {
			logrus.Error(err)
		}
	}
	t.Status = StatusComplete
	if t.Option.OnDone != nil {
		t.Option.OnDone(t)
	}
	return nil
}
func (p *TaskPool) NewRemoveLibraryTask(option RemoveLibraryTaskOption) (*RemoveLibraryTask, error) {
	if !DefaultLibraryLockPool.TryToLock(uint(option.LibraryId)) {
		return nil, LibraryLockError
	}
	info := task.NewBaseTask("RemoveLibrary", "0", StatusRunning)
	library, err := GetLibraryById(uint(option.LibraryId))
	if err != nil {
		return nil, err
	}
	task := &RemoveLibraryTask{
		BaseTask:  info,
		LibraryId: option.LibraryId,
		Option:    option,
		TaskOutput: &RemoveLibraryTaskOutput{
			Library: &library,
		},
		Library: &library,
	}
	module.Task.Pool.AddTask(task)
	return task, nil
}
