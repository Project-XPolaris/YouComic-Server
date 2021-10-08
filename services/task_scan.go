package services

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/utils"
	"github.com/sirupsen/logrus"
	"path/filepath"
)

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
