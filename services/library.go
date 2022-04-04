package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	appconfig "github.com/projectxpolaris/youcomic/config"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

func GetLibraryById(id uint) (model.Library, error) {
	var library model.Library
	err := database.Instance.Find(&library, id).Error
	return library, err
}

func CreateLibrary(name string, path string) (*model.Library, error) {
	// create library with path
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return nil, err
	}

	newLibrary := &model.Library{Name: name, Path: path}
	err = database.Instance.Create(newLibrary).Error
	return newLibrary, err
}

type LibraryExportConfig struct {
	Name  string `json:"name"`
	Books []struct {
		Name string `json:"name,omitempty"`
		Path string `json:"path"`
		Tags []struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"tags"`
		Pages []struct {
			Path  string `json:"path"`
			Order int    `json:"order"`
		} `json:"pages"`
		Cover string `json:"cover"`
	} `json:"books"`
}

func ImportLibrary(libraryPath string) error {
	file, err := ioutil.ReadFile(path.Join(libraryPath, "library_export.json"))
	if err != nil {
		return err
	}
	config := LibraryExportConfig{}

	err = json.Unmarshal(file, &config)
	if err != nil {
		return err
	}

	// register new library
	library, err := CreateLibrary(config.Name, libraryPath)
	if err != nil {
		return err
	}
	// add library book
	for _, bookConfig := range config.Books {
		book := model.Book{Name: bookConfig.Name, Path: bookConfig.Path, LibraryId: library.ID, Cover: bookConfig.Cover}
		err = database.Instance.Create(&book).Error
		if err != nil {
			return err
		}
		//generate cover thumbnail
		coverAbsolutePath := path.Join(libraryPath, bookConfig.Path, bookConfig.Cover)
		coverThumbnailStorePath := path.Join(appconfig.Instance.Store.Root, "generate", fmt.Sprintf("%d", book.ID))
		_, err = GenerateCoverThumbnail(coverAbsolutePath, coverThumbnailStorePath)
		if err != nil {
			return err
		}
		for _, tagConfig := range bookConfig.Tags {
			tag := model.Tag{}
			err = database.Instance.FirstOrCreate(&tag, model.Tag{Name: tagConfig.Name, Type: tagConfig.Type}).Error
			if err != nil {
				return err
			}
			database.Instance.Model(&book).Association("Tags").Append(&tag)
		}
		for _, pageConfig := range bookConfig.Pages {
			page := model.Page{PageOrder: pageConfig.Order, Path: pageConfig.Path, BookId: int(book.ID)}
			err = database.Instance.Create(&page).Error
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func DeleteLibrary(libraryId uint) error {
	library, err := GetLibraryById(libraryId)
	if err != nil {
		return err
	}

	books := make([]model.Book, 0)
	err = database.Instance.Model(&library).Association("Books").Find(&books)
	if err != nil {
		return err
	}
	for _, book := range books {
		err = database.Instance.Unscoped().Delete(model.Page{}, "book_id = ?", book.ID).Error
		if err != nil {
			return err
		}
	}
	err = database.Instance.Unscoped().Delete(model.Book{}, "library_id = ?", library.ID).Error
	if err != nil {
		return err
	}

	err = database.Instance.Unscoped().Delete(&library).Error
	if err != nil {
		return err
	}
	for _, book := range books {
		os.RemoveAll(filepath.Join(appconfig.Instance.Store.Root, "generate", fmt.Sprintf("%d", book.ID)))
	}

	return err
}

type LibraryQueryBuilder struct {
	DefaultPageFilter
	IdQueryFilter
	OrderQueryFilter
	NameQueryFilter
}

func (b *LibraryQueryBuilder) SetId(id interface{}) {
	b.InId(id)
}

func (b *LibraryQueryBuilder) Update(valueMapping map[string]interface{}) error {
	query := database.Instance
	query = ApplyFilters(b, query)
	err := query.Table("libraries").Updates(valueMapping).Error
	return err
}

func (b *LibraryQueryBuilder) ReadModels() (int64, interface{}, error) {
	query := database.Instance
	query = ApplyFilters(b, query)
	var count int64 = 0
	md := make([]model.Library, 0)
	err := query.Limit(b.getLimit()).Offset(b.getOffset()).Find(&md).Offset(-1).Count(&count).Error
	if err == sql.ErrNoRows {
		return 0, query, nil
	}
	return count, md, err
}

func ScanLibrary(id uint, option ScanLibraryOption) (*ScanTask, error) {
	library, err := GetLibraryById(id)
	if err != nil {
		return nil, err
	}
	option.Library = &library
	return DefaultTaskPool.NewScanLibraryTask(option)
}
