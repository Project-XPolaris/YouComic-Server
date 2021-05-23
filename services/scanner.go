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
)

var DefaultScanTaskPool = ScanTaskPool{}

type ScanTaskPool struct {
}

func (p *ScanTaskPool) StartTask(targetDir string) {
	go func() {
		resultChan := ScannerDir(targetDir)
		result := <-resultChan
		if err, isError := result.(error); isError {
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
		for _, item := range scanner.Result {
			relativePath, _ := filepath.Rel(targetDir, item.DirPath)
			book := &model.Book{
				Name:      filepath.Base(item.DirPath),
				LibraryId: library.ID,
				Path:      relativePath,
				Cover:     item.CoverName,
			}
			database.DB.Save(book)

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
				filepath.Join(targetDir, book.Path, book.Cover),
				coverThumbnailStorePath,
			)
			if err != nil {
				logrus.Error(err)
			}
		}
		resultChan <- struct{}{}
	}(resultChan)
	return resultChan
}
