package services

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/allentom/harukap/module/task"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/module"
	"github.com/projectxpolaris/youcomic/utils"
	"path/filepath"
)

var StrategyMapping = map[string]TagStrategy{
	"overwrite":       Overwrite,
	"append":          Append,
	"fillEmpty":       FillEmpty,
	"replaceSameType": ReplaceSameType,
}

type MatchLibraryTagTask struct {
	*task.BaseTask
	Strategy   TagStrategy
	stopFlag   bool
	Library    *model.Library
	TaskOutput *MatchLibraryTagTaskOutput
}
type MatchLibraryTagTaskOutput struct {
	TargetDir  string
	LibraryId  uint
	Name       string
	Total      int64
	Current    int64
	CurrentDir string
}

func (t *MatchLibraryTagTask) Output() (interface{}, error) {
	return t.TaskOutput, nil
}

func (t *MatchLibraryTagTask) Stop() error {
	t.stopFlag = true
	return nil
}
func (t *MatchLibraryTagTask) Start() error {
	books := make([]model.Book, 0)
	err := database.Instance.Find(&books, "library_id = ?", t.Library.ID).Error
	if err != nil {
		return err
	}
	t.TaskOutput.Total = int64(len(books))
	defer func() {
		DefaultLibraryLockPool.TryToUnlock(t.TaskOutput.LibraryId)
	}()
	for _, book := range books {
		if t.stopFlag {
			t.Status = StatusStop
			return nil
		}
		t.TaskOutput.Current += 1
		t.TaskOutput.CurrentDir = filepath.Base(book.Path)
		result := utils.MatchName(filepath.Base(book.Path))
		if result == nil {
			continue
		}
		if len(result.Name) > 0 {
			book.Name = result.Name
			database.Instance.Save(&book)
		}
		tags := make([]*model.Tag, 0)
		if len(result.Artist) > 0 {
			tags = append(tags, &model.Tag{Name: result.Artist, Type: "artist"})
		}
		if len(result.Series) > 0 {
			tags = append(tags, &model.Tag{Name: result.Series, Type: "series"})
		}
		if len(result.Theme) > 0 {
			tags = append(tags, &model.Tag{Name: result.Theme, Type: "theme"})
		}
		if len(result.Translator) > 0 {
			tags = append(tags, &model.Tag{Name: result.Translator, Type: "translator"})
		}
		if len(tags) > 0 {
			AddOrCreateTagToBook(&book, tags, t.Strategy)
		}
	}
	t.Status = StatusComplete
	return nil
}

func (p *TaskPool) NewMatchLibraryTagTask(libraryId uint, strategy string) (*MatchLibraryTagTask, error) {
	lockSuccess := DefaultLibraryLockPool.TryToLock(libraryId)
	if !lockSuccess {
		return nil, LibraryLockError
	}
	exist := linq.From(p.Tasks).FirstWith(func(i interface{}) bool {
		if task, ok := i.(*MatchLibraryTagTask); ok {
			return task.TaskOutput.LibraryId == libraryId && task.Status == StatusRunning
		}
		return false
	})
	if exist != nil {
		return exist.(*MatchLibraryTagTask), nil
	}
	var library model.Library
	err := database.Instance.First(&library, libraryId).Error
	if err != nil {
		return nil, err
	}

	task := &MatchLibraryTagTask{
		BaseTask: task.NewBaseTask("MatchTag", "0", StatusRunning),
		Library:  &library,
		Strategy: StrategyMapping[strategy],
		TaskOutput: &MatchLibraryTagTaskOutput{
			TargetDir: library.Path,
			LibraryId: libraryId,
			Name:      library.Name,
		},
	}
	task.Status = StatusRunning
	module.Task.Pool.AddTask(task)
	return task, nil
}
