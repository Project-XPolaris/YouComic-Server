package services

import (
	"errors"
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	appconfig "github.com/allentom/youcomic-api/config"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

type RemoveLibraryTask struct {
	BaseTask
	stopFlag  bool
	LibraryId int
	Err       error
}

func (t *RemoveLibraryTask) Stop() error {
	t.stopFlag = true
	return nil
}
func (t *RemoveLibraryTask) AbortError(err error) {
	t.Err = err
	t.Status = StatusError
	logrus.Error(err)
}
func (t *RemoveLibraryTask) Start() error {
	go func() {
		library, err := GetLibraryById(uint(t.LibraryId))
		if err != nil {
			t.AbortError(err)
			return
		}
		books := make([]model.Book, 0)
		err = database.DB.Model(&library).Association("Books").Find(&books)
		if err != nil {
			t.AbortError(err)
			return
		}
		for _, book := range books {
			err = database.DB.Unscoped().Delete(model.Page{}, "book_id = ?", book.ID).Error
			if err != nil {
				t.AbortError(err)
				return
			}
		}
		err = database.DB.Unscoped().Delete(model.Book{}, "library_id = ?", library.ID).Error
		if err != nil {
			t.AbortError(err)
			return
		}

		err = database.DB.Unscoped().Delete(&library).Error
		if err != nil {
			t.AbortError(err)
			return
		}
		for _, book := range books {
			os.RemoveAll(filepath.Join(appconfig.Instance.Store.Root, "generate", fmt.Sprintf("%d", book.ID)))
		}
		t.Status = StatusComplete
	}()
	return nil
}
func (p *TaskPool) NewRemoveLibraryTask(libraryId int) (*RemoveLibraryTask, error) {
	exist := linq.From(p.Tasks).FirstWith(func(i interface{}) bool {
		if task, ok := i.(*RemoveLibraryTask); ok {
			if task.Status == StatusRunning && task.LibraryId == libraryId {
				return true
			}
		}
		if task, ok := i.(*MatchLibraryTagTask); ok {
			if task.Status == StatusRunning && task.LibraryId == uint(libraryId) {
				return true
			}
		}
		if task, ok := i.(*ScanTask); ok {
			if task.Status == StatusRunning && task.LibraryId == uint(libraryId) {
				return true
			}
		}
		if task, ok := i.(*RenameBookDirectoryTask); ok {
			if task.Status == StatusRunning && task.LibraryId == uint(libraryId) {
				return true
			}
		}
		if task, ok := i.(*WriteBookMetaTask); ok {
			if task.Status == StatusRunning && task.LibraryId == uint(libraryId) {
				return true
			}
		}
		return false
	})
	if exist != nil {
		return nil, errors.New("library lock by other task")
	}
	task := &RemoveLibraryTask{
		BaseTask:  NewBaseTask(),
		LibraryId: libraryId,
	}
	task.Status = StatusRunning
	p.AddTask(task)
	task.Start()
	return task, nil
}
