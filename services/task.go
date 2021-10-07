package services

import (
	linq "github.com/ahmetb/go-linq/v3"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/utils"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var DefaultTaskPool = TaskPool{
	Tasks: []Task{},
}

const (
	TaskStatusInit    = "Init"
	StatusRunning     = "Running"
	StatusComplete    = "Complete"
	StatusStop        = "Stop"
	ScanStatusAnalyze = "Analyze"
	ScanStatusAdd     = "Add"
)

type TaskPool struct {
	Tasks []Task
	sync.Mutex
}

func (p *TaskPool) AddTask(task Task) {
	p.Lock()
	defer p.Unlock()
	p.Tasks = append(p.Tasks, task)
}
func (p *TaskPool) StopTask(id string) error {
	task := linq.From(p.Tasks).FirstWith(func(i interface{}) bool {
		return i.(Task).GetBaseInfo().ID == id
	}).(Task)
	return task.Stop()
}
func (p *TaskPool) NewLibraryAndScan(targetPath string, name string) (*ScanTask, error) {
	library, err := CreateLibrary(name, targetPath)
	if err != nil {
		return nil, err
	}
	task := &ScanTask{
		BaseTask:  NewBaseTask(),
		TargetDir: library.Path,
		LibraryId: library.ID,
		Name:      library.Name,
	}
	task.Status = ScanStatusAnalyze
	p.AddTask(task)
	task.Start()
	return task, err
}
func (p *TaskPool) NewScanLibraryTask(library *model.Library) (*ScanTask, error) {
	lockSuccess := DefaultLibraryLockPool.TryToLock(library.ID)
	if !lockSuccess {
		return nil, LibraryLockError
	}
	exist := linq.From(p.Tasks).FirstWith(func(i interface{}) bool {
		if task, ok := i.(*ScanTask); ok {
			return task.LibraryId == library.ID && (task.Status == ScanStatusAdd || task.Status == ScanStatusAnalyze)
		}
		return false
	})
	if exist != nil {
		return exist.(*ScanTask), nil
	}
	task := &ScanTask{
		BaseTask:  NewBaseTask(),
		TargetDir: library.Path,
		LibraryId: library.ID,
		Name:      library.Name,
	}
	task.Status = ScanStatusAnalyze
	p.AddTask(task)
	task.Start()
	return task, nil
}
func (p *TaskPool) NewMatchLibraryTagTask(libraryId uint) (*MatchLibraryTagTask, error) {
	lockSuccess := DefaultLibraryLockPool.TryToLock(libraryId)
	if !lockSuccess {
		return nil, LibraryLockError
	}
	exist := linq.From(p.Tasks).FirstWith(func(i interface{}) bool {
		if task, ok := i.(*MatchLibraryTagTask); ok {
			return task.LibraryId == libraryId && task.Status == StatusRunning
		}
		return false
	})
	if exist != nil {
		return exist.(*MatchLibraryTagTask), nil
	}
	var library model.Library
	err := database.DB.First(&library, libraryId).Error
	if err != nil {
		return nil, err
	}
	task := &MatchLibraryTagTask{
		BaseTask:  NewBaseTask(),
		TargetDir: library.Path,
		LibraryId: library.ID,
		Name:      library.Name,
		Library:   &library,
	}
	task.Status = StatusRunning
	p.AddTask(task)
	task.Start()
	return task, nil
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
	err := database.DB.First(&library, libraryId).Error
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
	err := database.DB.First(&library, toLibraryId).Error
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
func (p *TaskPool) NewRemoveEmptyTagTask() (*RemoveEmptyTagTask, error) {
	exist := linq.From(p.Tasks).FirstWith(func(i interface{}) bool {
		if task, ok := i.(*RemoveEmptyTagTask); ok {
			if task.Status == StatusRunning {
				return true
			}
		}
		return false
	})
	if exist != nil {
		return exist.(*RemoveEmptyTagTask), nil
	}
	task := &RemoveEmptyTagTask{
		BaseTask: NewBaseTask(),
	}
	task.Status = StatusRunning
	p.AddTask(task)
	task.Start()
	return task, nil
}

type Task interface {
	Stop() error
	Start() error
	GetBaseInfo() *BaseTask
}
type BaseTask struct {
	ID      string
	Status  string
	Created time.Time
}

func (t *BaseTask) GetBaseInfo() *BaseTask {
	return t
}
func NewBaseTask() BaseTask {
	return BaseTask{
		ID:      xid.New().String(),
		Status:  TaskStatusInit,
		Created: time.Now(),
	}
}

type ScanTask struct {
	BaseTask
	TargetDir  string
	LibraryId  uint
	Name       string
	Total      int64
	Current    int64
	CurrentDir string
	stopFlag   bool
}

func (t *ScanTask) Start() error {
	resultChan := t.scannerDir()
	go func() {
		result := <-resultChan
		if err, isError := result.(error); isError {
			logrus.Info(err)
		}
	}()
	return nil
}
func (t *ScanTask) Stop() error {
	t.stopFlag = true
	return nil
}
func (t *ScanTask) scannerDir() chan interface{} {
	resultChan := make(chan interface{})
	go func(resultChan chan interface{}) {
		defer func() {
			DefaultLibraryLockPool.TryToUnlock(t.LibraryId)
		}()
		// sync with exist book
		var library model.Library
		err := database.DB.Where("id = ?", t.LibraryId).Preload("Books").Find(&library).Error
		if err != nil {
			resultChan <- err
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
				resultChan <- err
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
			resultChan <- err
			return
		}
		t.Total = scanner.Total
		t.Status = ScanStatusAdd
		// create library
		for _, item := range scanner.Result {
			if t.stopFlag {
				t.Status = StatusStop
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
			book := model.Book{
				Name:      filepath.Base(item.DirPath),
				LibraryId: t.LibraryId,
				Path:      relativePath,
				Cover:     item.CoverName,
			}
			database.DB.Save(&book)

			// for pages
			for idx, pageName := range item.Pages {
				page := &model.Page{
					Path:      pageName,
					BookId:    int(book.Model.ID),
					PageOrder: idx + 1,
				}
				database.DB.Save(page)
			}
			coverThumbnailStorePath := utils.GetThumbnailStorePath(book.ID)
			option := ThumbnailTaskOption{
				Input:   filepath.Join(t.TargetDir, book.Path, book.Cover),
				Output:  coverThumbnailStorePath,
				ErrChan: make(chan error),
			}
			go func() {
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
					logrus.Error(err)
				}
			}()

		}
		t.Status = StatusComplete
		resultChan <- struct{}{}
	}(resultChan)
	return resultChan
}

type MatchLibraryTagTask struct {
	BaseTask
	TargetDir  string
	LibraryId  uint
	Name       string
	Total      int64
	Current    int64
	CurrentDir string
	stopFlag   bool
	Library    *model.Library
}

func (t *MatchLibraryTagTask) Stop() error {
	t.stopFlag = true
	return nil
}
func (t *MatchLibraryTagTask) Start() error {
	books := make([]model.Book, 0)
	err := database.DB.Find(&books, "library_id = ?", t.Library.ID).Error
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
			result := utils.MatchName(filepath.Base(book.Path))
			if result == nil {
				continue
			}
			if len(result.Name) > 0 {
				book.Name = result.Name
				database.DB.Save(&book)
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
				AddOrCreateTagToBook(&book, tags, true)
			}
		}
		t.Status = StatusComplete
	}()

	return nil
}

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
	err := database.DB.Preload("Tags").Find(&books, "library_id = ?", t.Library.ID).Error
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
			err := database.DB.First(&book, id).Error
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
				err = database.DB.First(&library, book.LibraryId).Error
				libraries = append(libraries, library)
				if err != nil {
					logrus.Error(err)
					continue
				}
			}
			t.From = library
			sourcePath := filepath.Join(library.Path, book.Path)
			toPath := filepath.Join(t.To.Path, book.Path)

			// try to move
			err = os.Rename(sourcePath+"/", toPath+"/")
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
			err = database.DB.Save(&book).Error
			if err != nil {
				logrus.Error(err)
				continue
			}
		}
		t.Status = StatusComplete
	}()
	return nil
}

type RemoveEmptyTagTask struct {
	BaseTask
	CurrentTag *model.Tag
	Total      int
	Current    int
	stopFlag   bool
}

func (t *RemoveEmptyTagTask) Stop() error {
	t.stopFlag = true
	return nil
}

func (t *RemoveEmptyTagTask) Start() error {
	go func() {
		var tags []model.Tag
		database.DB.Find(&tags)
		t.Total = len(tags)
		for _, tag := range tags {
			t.Current += 1
			t.CurrentTag = &tag
			ass := database.DB.Model(&tag).Association("Books")
			if ass.Count() == 0 {
				err := database.DB.Unscoped().Delete(&tag).Error
				if err != nil {
					logrus.Error(err)
				}
			}
		}
		t.Status = StatusComplete
	}()
	return nil
}
