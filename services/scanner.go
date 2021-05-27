package services

import (
	"fmt"
	linq "github.com/ahmetb/go-linq/v3"
	"github.com/allentom/youcomic-api/config"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/utils"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"path"
	"path/filepath"
	"sync"
	"time"
)

var DefaultScanTaskPool = ScanTaskPool{
	Tasks: []*ScanTask{},
}

const (
	ScanStatusAnalyze  = "Analyze"
	ScanStatusAdd      = "Add"
	ScanStatusComplete = "Complete"
	ScanStatusStop     = "Stop"
)

type ScanTaskPool struct {
	Tasks []*ScanTask
	sync.Mutex
}

type ScanTask struct {
	ID         string
	TargetDir  string
	LibraryId  uint
	Name       string
	Total      int64
	Current    int64
	Status     string
	Created    time.Time
	CurrentDir string
	stopFlag   bool
}

func (p *ScanTaskPool) AddTask(task *ScanTask) {
	p.Lock()
	defer p.Unlock()
	p.Tasks = append(p.Tasks, task)
}
func (p *ScanTaskPool) StopTask(id string) {
	task := linq.From(p.Tasks).FirstWith(func(i interface{}) bool {
		return i.(*ScanTask).ID == id
	}).(*ScanTask)
	task.StopTask()
}
func (p *ScanTaskPool) NewLibraryAndScan(targetPath string, name string) (*ScanTask, error) {
	library, err := CreateLibrary(name, targetPath)
	if err != nil {
		return nil, err
	}
	task := &ScanTask{
		ID:        xid.New().String(),
		TargetDir: library.Path,
		LibraryId: library.ID,
		Name:      library.Name,
		Status:    ScanStatusAnalyze,
		Created:   time.Now(),
	}
	p.AddTask(task)
	task.StartTask()
	return task, err
}

func (p *ScanTaskPool) NewScanLibraryTask(library *model.Library) (*ScanTask, error) {
	exist := linq.From(p.Tasks).FirstWith(func(i interface{}) bool {
		task := i.(*ScanTask)
		return task.LibraryId == library.ID && ( task.Status == ScanStatusAdd || task.Status == ScanStatusAnalyze)
	})
	if exist != nil {
		return exist.(*ScanTask),nil
	}
	task := &ScanTask{
		ID:        xid.New().String(),
		TargetDir: library.Path,
		LibraryId: library.ID,
		Name:      library.Name,
		Status:    ScanStatusAnalyze,
		Created:   time.Now(),
	}
	p.AddTask(task)
	task.StartTask()
	return task, nil
}
func (t *ScanTask) StartTask() {
	resultChan := t.ScannerDir()
	go func() {
		result := <-resultChan
		if err, isError := result.(error); isError {
			logrus.Info(err)
		}
	}()
}
func (t *ScanTask) StopTask() {
	t.stopFlag = true
}
func (t *ScanTask) ScannerDir() chan interface{} {
	resultChan := make(chan interface{})
	go func(resultChan chan interface{}) {
		scanner := utils.Scanner{
			TargetPath:   t.TargetDir,
			PageExt:      utils.DefaultScanPageExt,
			MinPageCount: 4,
		}
		err := scanner.Scan()
		if err != nil {
			resultChan <- err
			return
		}
		t.Total = scanner.Total
		t.Status = ScanStatusAdd
		// create library
		for _, item := range scanner.Result {
			if t.stopFlag {
				t.Status = ScanStatusStop
				return
			}
			t.Current += 1
			t.CurrentDir = filepath.Base(item.DirPath)
			relativePath, _ := filepath.Rel(t.TargetDir, item.DirPath)
			var book model.Book
			database.DB.Model(&model.Book{}).Where("path = ?", relativePath).First(&book)
			if len(book.Path) != 0 {
				continue
			}
			book = model.Book{
				Name:      filepath.Base(item.DirPath),
				LibraryId: t.LibraryId,
				Path:      relativePath,
				Cover:     item.CoverName,
			}
			database.DB.Save(&book)

			// for pages
			for idx, pageName := range item.Pages {
				page := &model.Page{
					Path:   pageName,
					BookId: int(book.Model.ID),
					Order:  idx + 1,
				}
				database.DB.Save(page)
			}
			coverThumbnailStorePath := path.Join(config.Config.Store.Root, "generate", fmt.Sprintf("%d", book.ID))
			option := ThumbnailTaskOption{
				Input:   filepath.Join(t.TargetDir, book.Path, book.Cover),
				Output:  coverThumbnailStorePath,
				ErrChan: make(chan error),
			}
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
		}
		t.Status = ScanStatusComplete
		resultChan <- struct{}{}
	}(resultChan)
	return resultChan
}
