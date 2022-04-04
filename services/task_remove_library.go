package services

import (
	"fmt"
	appconfig "github.com/projectxpolaris/youcomic/config"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

type RemoveLibraryTaskOption struct {
	LibraryId int
	OnError   func(task *RemoveLibraryTask, err error)
	OnDone    func(task *RemoveLibraryTask)
}
type RemoveLibraryTask struct {
	BaseTask
	stopFlag  bool
	LibraryId int
	Library   *model.Library
	Err       error
	Option    RemoveLibraryTaskOption
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
	go func() {
		defer DefaultLibraryLockPool.TryToUnlock(uint(t.LibraryId))
		library, err := GetLibraryById(uint(t.LibraryId))
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
		for _, book := range books {
			err = database.Instance.Unscoped().Delete(model.Page{}, "book_id = ?", book.ID).Error
			if err != nil {
				t.AbortError(err)
				return
			}
		}
		err = database.Instance.Unscoped().Delete(model.Book{}, "library_id = ?", library.ID).Error
		if err != nil {
			t.AbortError(err)
			return
		}
		err = database.Instance.Unscoped().Delete(&library).Error
		if err != nil {
			t.AbortError(err)
			return
		}
		for _, book := range books {
			err = os.RemoveAll(filepath.Join(appconfig.Instance.Store.Root, "generate", fmt.Sprintf("%d", book.ID)))
			if err != nil {
				t.AbortError(err)
			}
		}
		t.Status = StatusComplete
		if t.Option.OnDone != nil {
			t.Option.OnDone(t)
		}
	}()
	return nil
}
func (p *TaskPool) NewRemoveLibraryTask(option RemoveLibraryTaskOption) (*RemoveLibraryTask, error) {
	if !DefaultLibraryLockPool.TryToLock(uint(option.LibraryId)) {
		return nil, LibraryLockError
	}
	task := &RemoveLibraryTask{
		BaseTask:  NewBaseTask(),
		LibraryId: option.LibraryId,
		Option:    option,
	}
	task.Status = StatusRunning
	p.AddTask(task)
	task.Start()
	return task, nil
}
