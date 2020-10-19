package services

import (
	"fmt"
	"github.com/allentom/youcomic-api/config"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/utils"
	"github.com/sirupsen/logrus"
	"path"
	"path/filepath"
	"sync"
	"sync/atomic"
)
var DefaultScanTaskPool = ScanTaskPool{}
type ScanTaskPool struct {

}

func (p *ScanTaskPool) StartTask(targetDir string) {
	go func() {
		resultChan :=  ScannerDir(targetDir)
		result := <- resultChan
		if err,isError := result.(error);isError {
			logrus.Info(err)
		}
	}()
}
func ScannerDir(targetDir string) chan interface{} {
	resultChan := make(chan interface{})
	go func(resultChan chan interface{}) {
		scanner := utils.Scanner{
			TargetPath:   targetDir,
			PageExt:      utils.DefaultScanPageExt,
			MinPageCount: 4,
		}
		err := scanner.Scan()
		if err != nil {
			resultChan <- err
			return
		}

		// create library
		library, err := CreateLibrary(targetDir, targetDir)
		if err != nil {
			resultChan <- err
			return
		}

		taskChan := make(chan struct{}, 20)
		for idx := 0; idx < 20; idx++ {
			taskChan <- struct{}{}
		}

		createdBooks := make([]*model.Book,0)
		for _, item := range scanner.Result {
			relativePath, _ := filepath.Rel(targetDir, item.DirPath)
			book := &model.Book{
				Name: filepath.Base(item.DirPath),
				LibraryId: library.ID,
				Path: relativePath,
				Cover: item.CoverName,
			}
			database.DB.Save(book)

			// for pages
			for idx, pageName := range item.Pages {
				page := &model.Page{
					Path: pageName,
					BookId: int(book.Model.ID),
					Order: idx + 1,
				}
				database.DB.Save(page)
			}
			createdBooks = append(createdBooks, book)
		}
		var wg sync.WaitGroup
		wg.Add(1)
		var complete int64 = 0
		for idx := 0; idx < len(createdBooks); idx++ {
			<-taskChan
			go func(doneWait *sync.WaitGroup, idx int) {
				item := createdBooks[idx]
				coverThumbnailStorePath := path.Join(config.Config.Store.Root, "generate", fmt.Sprintf("%d", item.ID))
				GenerateCoverThumbnail(
					filepath.Join(targetDir,item.Path, item.Cover),
					coverThumbnailStorePath,
				)
				taskChan <- struct{}{}
				atomic.AddInt64(&complete, 1)
				fmt.Println(complete)
				if complete == int64(len(scanner.Result)) {
					wg.Done()
				}
			}(&wg, idx)
		}
		wg.Wait()
		resultChan <- struct {}{}
	}(resultChan)
	return resultChan
}
