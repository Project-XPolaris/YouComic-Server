package services

import (
	"fmt"
	"github.com/allentom/youcomic-api/config"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/utils"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"path"
	"path/filepath"
	"sync"
)

var DefaultScanTaskPool = ScanTaskPool{
	Tasks: []*ScanTask{},
}

const (
	ScanStatusAnalyze  = "Analyze"
	ScanStatusAdd      = "Add"
	ScanStatusComplete = "Complete"
)

type ScanTaskPool struct {
	Tasks []*ScanTask
	sync.Mutex
}

type ScanTask struct {
	ID        string `json:"id"`
	TargetDir string `json:"targetDir"`
	LibraryId uint   `json:"libraryId"`
	Name      string `json:"name"`
	Total     int64  `json:"total"`
	Current   int64  `json:"current"`
	Status    string `json:"status"`
}

func (p *ScanTaskPool) AddTask(task *ScanTask) {
	p.Lock()
	defer p.Unlock()
	p.Tasks = append(p.Tasks, task)
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
	}
	p.AddTask(task)
	task.StartTask()
	return task, err
}

func (p *ScanTaskPool) NewScanLibraryTask(library *model.Library) (*ScanTask, error) {
	task := &ScanTask{
		ID:        xid.New().String(),
		TargetDir: library.Path,
		LibraryId: library.ID,
		Name:      library.Name,
		Status:    ScanStatusAnalyze,
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
			t.Current += 1
			relativePath, _ := filepath.Rel(t.TargetDir, item.DirPath)
			var book model.Book
			database.DB.Model(&model.Book{}).Where("path = ?", relativePath).First(&book)
			if len(book.Path) != 0 {
				logrus.Info(fmt.Sprintf("ID = %d exist,skip", book.ID))
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
			_, err = GenerateCoverThumbnail(
				filepath.Join(t.TargetDir, book.Path, book.Cover),
				coverThumbnailStorePath,
			)
			if err != nil {
				logrus.Error(err)
			}
		}
		t.Status = ScanStatusComplete
		resultChan <- struct{}{}
	}(resultChan)
	return resultChan
}
