package services

import (
	"encoding/json"
	"github.com/ahmetb/go-linq/v3"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/utils"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
)

type MetaTag struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
type BookMeta struct {
	OriginalName string    `json:"originalName"`
	Title        string    `json:"title"`
	Cover        string    `json:"cover"`
	Tags         []MetaTag `json:"tags"`
}
type WriteBookMetaTask struct {
	BaseTask
	LibraryId   uint
	Total       int
	Current     int
	CurrentBook string
	stopFlag    bool
	Err         error
}

func (t *WriteBookMetaTask) Stop() error {
	t.stopFlag = true
	return nil
}
func (t *WriteBookMetaTask) RaiseError(err error) {
	t.Status = StatusStop
	t.Err = err
}
func (t *WriteBookMetaTask) Start() error {
	go func() {
		defer func() {
			DefaultLibraryLockPool.TryToUnlock(t.LibraryId)
		}()
		t.Status = StatusRunning
		var library model.Library
		err := database.DB.Where("id = ?", t.LibraryId).Find(&library).Error
		if err != nil {
			t.RaiseError(err)
			return
		}
		var books []*model.Book
		err = database.DB.Where("library_id = ?", t.LibraryId).Preload("Tags").Find(&books).Error
		if err != nil {
			t.RaiseError(err)
			return
		}
		t.Total = len(books)
		for _, book := range books {
			if t.stopFlag {
				t.Status = StatusStop
				break
			}
			t.Current += 1
			t.CurrentBook = book.Name
			bookPath := filepath.Join(library.Path, book.Path)
			metaFilePath := filepath.Join(bookPath, "youcomic_meta.json")

			if err != nil {
				logrus.Error(err)
				continue
			}
			meta := BookMeta{}
			if utils.CheckFileExist(metaFilePath) {
				jsonFile, err := os.Open(metaFilePath)
				byteValue, _ := ioutil.ReadAll(jsonFile)
				err = json.Unmarshal(byteValue, &meta)
				if err != nil {
					jsonFile.Close()
					logrus.Error(err)
					continue
				}
				jsonFile.Close()
			}
			if len(meta.OriginalName) == 0 {
				meta.OriginalName = filepath.Base(book.Path)
			}
			meta.Title = book.Name
			meta.Cover = book.Cover
			meta.Tags = []MetaTag{}
			for _, tag := range book.Tags {
				meta.Tags = append(meta.Tags, MetaTag{
					Name: tag.Name,
					Type: tag.Type,
				})
			}

			file, _ := json.MarshalIndent(meta, "", "   ")
			if err != nil {
				continue
				logrus.Error(err)
			}
			err = ioutil.WriteFile(metaFilePath, file, 0644)
			if err != nil {
				logrus.Error(err)
			}
		}
		t.Status = StatusComplete

	}()
	return nil
}

func (p *TaskPool) NewWriteBookMetaTask(library *model.Library) (*WriteBookMetaTask, error) {
	lockSuccess := DefaultLibraryLockPool.TryToLock(library.ID)
	if !lockSuccess {
		return nil, LibraryLockError
	}
	exist := linq.From(p.Tasks).FirstWith(func(i interface{}) bool {
		if task, ok := i.(*WriteBookMetaTask); ok {
			return task.LibraryId == library.ID && (task.Status == StatusRunning)
		}
		return false
	})
	if exist != nil {
		return exist.(*WriteBookMetaTask), nil
	}
	task := &WriteBookMetaTask{
		BaseTask:  NewBaseTask(),
		LibraryId: library.ID,
	}
	task.Status = ScanStatusAnalyze
	p.AddTask(task)
	task.Start()
	return task, nil
}
