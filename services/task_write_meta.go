package services

import (
	"encoding/json"
	"github.com/allentom/harukap/module/task"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/module"
	"github.com/projectxpolaris/youcomic/utils"
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
	*task.BaseTask
	LibraryId  uint
	stopFlag   bool
	Err        error
	TaskOutput *WriteBookMetaTaskOutput
}

func (t *WriteBookMetaTask) Output() (interface{}, error) {
	return t.TaskOutput, nil
}

type WriteBookMetaTaskOutput struct {
	Total       int
	Current     int
	CurrentBook string
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

	defer func() {
		DefaultLibraryLockPool.TryToUnlock(t.LibraryId)
	}()
	t.Status = StatusRunning
	var library model.Library
	err := database.Instance.Where("id = ?", t.LibraryId).Find(&library).Error
	if err != nil {
		t.RaiseError(err)
		return err
	}
	var books []*model.Book
	err = database.Instance.Where("library_id = ?", t.LibraryId).Preload("Tags").Find(&books).Error
	if err != nil {
		t.RaiseError(err)
		return err
	}
	t.TaskOutput.Total = len(books)
	for _, book := range books {
		if t.stopFlag {
			t.Status = StatusStop
			break
		}
		t.TaskOutput.Current += 1
		t.TaskOutput.CurrentBook = book.Name
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

	return nil
}

func NewWriteBookMetaTask(library *model.Library) (*WriteBookMetaTask, error) {
	lockSuccess := DefaultLibraryLockPool.TryToLock(library.ID)
	if !lockSuccess {
		return nil, LibraryLockError
	}
	info := task.NewBaseTask("WriteBookMeta", "0", StatusRunning)
	task := &WriteBookMetaTask{
		BaseTask:  info,
		LibraryId: library.ID,
		TaskOutput: &WriteBookMetaTaskOutput{},
	}
	module.Task.Pool.AddTask(task)
	return task, nil
}
