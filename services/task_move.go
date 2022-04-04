package services

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/utils"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

type MoveBookTask struct {
	BaseTask
	BookIds    []int
	CurrentDir string
	From       *model.Library
	To         *model.Library
	Total      int
	Current    int
	stopFlag   bool
}

func (t *MoveBookTask) Stop() error {
	t.stopFlag = true
	return nil
}

func (t *MoveBookTask) Start() error {
	go func() {
		defer func() {
			DefaultLibraryLockPool.TryToUnlock(t.To.ID)
		}()
		libraries := make([]*model.Library, 0)
		for _, id := range t.BookIds {
			if t.stopFlag {
				t.Status = StatusStop
				return
			}
			t.Current += 1
			var book model.Book
			err := database.Instance.First(&book, id).Error
			if err != nil {
				logrus.Error(err)
				continue
			}
			// in position,skip
			if book.LibraryId == t.To.ID {
				continue
			}
			t.CurrentDir = filepath.Base(book.Path)
			// get target library
			var library *model.Library
			for _, cacheLibrary := range libraries {
				if cacheLibrary.ID == book.LibraryId {
					library = cacheLibrary
				}
			}
			if library == nil {
				err = database.Instance.First(&library, book.LibraryId).Error
				libraries = append(libraries, library)
				if err != nil {
					logrus.Error(err)
					continue
				}
			}
			t.From = library
			sourcePath := filepath.Join(library.Path, book.Path)
			toPath := filepath.Join(t.To.Path, book.Path)
			if utils.CheckFileExist(toPath) {
				logrus.Warn("move target exist,skip")
				continue
			}
			// try to move
			err = os.Rename(sourcePath, toPath)
			if err != nil {
				// failed to move,try to copy
				err = os.MkdirAll(toPath, 0644)
				if err != nil {
					logrus.Error(err)
					continue
				}
				err = utils.CopyDirectory(sourcePath, toPath)
				if err != nil {
					logrus.Error(err)
					continue
				}
				err = os.RemoveAll(sourcePath)
				if err != nil {
					logrus.Error(err)
					continue
				}
			}
			book.LibraryId = t.To.ID
			err = database.Instance.Save(&book).Error
			if err != nil {
				logrus.Error(err)
				continue
			}
		}
		t.Status = StatusComplete
	}()
	return nil
}
func (p *TaskPool) NewMoveBookTask(bookIds []int, toLibraryId int) (*MoveBookTask, error) {
	lockSuccess := DefaultLibraryLockPool.TryToLock(uint(toLibraryId))
	if !lockSuccess {
		return nil, LibraryLockError
	}
	exist := linq.From(p.Tasks).FirstWith(func(i interface{}) bool {
		if task, ok := i.(*MoveBookTask); ok {
			if task.Status != StatusRunning {
				return false
			}
			for _, taskBookId := range task.BookIds {
				for _, id := range bookIds {
					if id == taskBookId {
						return false
					}
				}
			}
		}
		return false
	})
	if exist != nil {
		return exist.(*MoveBookTask), nil
	}
	var library model.Library
	err := database.Instance.First(&library, toLibraryId).Error
	if err != nil {
		return nil, err
	}
	task := &MoveBookTask{
		BaseTask: NewBaseTask(),
		BookIds:  bookIds,
		To:       &library,
		Total:    len(bookIds),
	}
	task.Status = StatusRunning
	p.AddTask(task)
	task.Start()
	return task, nil
}
