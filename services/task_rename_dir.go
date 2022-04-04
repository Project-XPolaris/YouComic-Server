package services

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/sirupsen/logrus"
	"path/filepath"
)

type RenameSlot struct {
	Type    string `json:"type"`
	Pattern string `json:"pattern"`
	Sep     string `json:"sep"`
}
type RenameBookDirectoryTaskOption struct {
	LibraryId uint
	BookIds   uint
}
type RenameBookDirectoryTask struct {
	BaseTask
	TargetDir  string
	LibraryId  uint
	Name       string
	Total      int64
	Current    int64
	CurrentDir string
	stopFlag   bool
	Library    *model.Library
	Pattern    string
	Slots      []RenameSlot
	Option     RenameBookDirectoryTaskOption
}

func (t *RenameBookDirectoryTask) Stop() error {
	t.stopFlag = true
	return nil
}
func (t *RenameBookDirectoryTask) Start() error {
	books := make([]model.Book, 0)
	err := database.Instance.Preload("Tags").Find(&books, "library_id = ?", t.Library.ID).Error
	if err != nil {
		return err
	}
	t.Total = int64(len(books))
	go func() {
		defer func() {
			DefaultLibraryLockPool.TryToUnlock(t.LibraryId)
		}()
		for _, book := range books {
			if t.stopFlag {
				t.Status = StatusStop
				return
			}
			t.Current += 1
			t.CurrentDir = filepath.Base(book.Path)
			name := RenderBookDirectoryRenameText(&book, t.Pattern, t.Slots)
			_, err := RenameBookDirectory(&book, t.Library, name)
			if err != nil {
				logrus.Error(err)
			}
		}
		t.Status = StatusComplete
	}()
	return nil
}

func (p *TaskPool) NewRenameBookDirectoryLibraryTask(libraryId uint, pattern string, slots []RenameSlot) (*RenameBookDirectoryTask, error) {
	lockSuccess := DefaultLibraryLockPool.TryToLock(libraryId)
	if !lockSuccess {
		return nil, LibraryLockError
	}
	exist := linq.From(p.Tasks).FirstWith(func(i interface{}) bool {
		if task, ok := i.(*RenameBookDirectoryTask); ok {
			return task.LibraryId == libraryId && task.Status == StatusRunning
		}
		return false
	})
	if exist != nil {
		return exist.(*RenameBookDirectoryTask), nil
	}
	var library model.Library
	err := database.Instance.First(&library, libraryId).Error
	if err != nil {
		return nil, err
	}
	task := &RenameBookDirectoryTask{
		BaseTask:  NewBaseTask(),
		TargetDir: library.Path,
		LibraryId: library.ID,
		Name:      library.Name,
		Library:   &library,
		Pattern:   pattern,
		Slots:     slots,
	}
	task.Status = StatusRunning
	p.AddTask(task)
	task.Start()
	return task, nil
}
