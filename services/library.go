package services

import (
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"os"
)

func GetLibraryById(id uint) (model.Library, error) {
	var library model.Library
	err := database.DB.Find(&library, id).Error
	return library, err
}

func CreateLibrary(name string, path string) (*model.Library, error) {
	// create library with path
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return nil, err
	}

	newLibrary := &model.Library{Name: name, Path: path}
	err = database.DB.Create(newLibrary).Error
	return newLibrary, err
}
