package services

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/allentom/youcomic-api/application"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"github.com/jinzhu/gorm"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
)

var (
	DefaultLibraryNotFound = errors.New("default library not found")
)

func CreateBook(name string, libraryId uint) (error, *model.Book) {
	// use default library if it not specific
	if libraryId == 0 {
		var library model.Library
		err := database.DB.Where("name = ", application.DEFAULT_LIBRARY_NAME).Find(&library).Error
		if err != nil {
			return DefaultLibraryNotFound, nil
		}
		libraryId = library.ID
	}

	book := model.Book{
		Name:      name,
		LibraryId: libraryId,
	}

	result := database.DB.Create(&book)
	err := result.Error
	if err != nil {
		return err, nil
	}
	err = database.DB.Model(&model.Book{}).Where("id = ?", book.ID).Update("path", fmt.Sprintf("/%d", book.ID)).Error
	return err, &book
}

func GetBook(book *model.Book) error {
	err := database.DB.First(book, book.ID).Error
	if err == gorm.ErrRecordNotFound {
		return RecordNotFoundError
	}
	if err != nil {
		return err
	}
	return nil
}

func UpdateBook(book *model.Book, allowFields ...string) error {
	updateMap := make(map[string]interface{})
	r := reflect.ValueOf(book)
	for _, propertyName := range allowFields {
		value := reflect.Indirect(r).FieldByName(propertyName)

		updateMap[propertyName] = value.Interface()
	}
	err := database.DB.Model(book).Updates(updateMap).Error
	return err
}

type BooksQueryBuilder struct {
	DefaultPageFilter
	IdQueryFilter
	OrderQueryFilter
	NameQueryFilter
	BookCollectionQueryFilter
	TagQueryFilter
	StartTimeQueryFilter
	EndTimeQueryFilter
	NameSearchQueryFilter
}

type EndTimeQueryFilter struct {
	endTime interface{}
}

func (f EndTimeQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.endTime != nil && len(f.endTime.(string)) != 0 {
		return db.Where("created_at <= ?", f.endTime)
	}
	return db
}

func (f *EndTimeQueryFilter) SetEndTimeQueryFilter(endTime interface{}) {

	if len(endTime.(string)) > 0 {
		f.endTime = endTime
	}

}

type StartTimeQueryFilter struct {
	startTime interface{}
}

func (f StartTimeQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.startTime != nil && len(f.startTime.(string)) != 0 {
		return db.Where("created_at >= ?", f.startTime)
	}
	return db
}

func (f *StartTimeQueryFilter) SetStartTimeQueryFilter(startTime interface{}) {

	if len(startTime.(string)) > 0 {
		f.startTime = startTime
	}

}

type TagQueryFilter struct {
	tags []interface{}
}

func (f TagQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.tags != nil && len(f.tags) != 0 {
		return db.Joins("inner join book_tags on book_tags.book_id = books.id").Where("book_tags.tag_id in (?)", f.tags)
	}
	return db
}

func (f *TagQueryFilter) SetTagQueryFilter(tagIds ...interface{}) {
	for _, tagId := range tagIds {
		if len(tagId.(string)) > 0 {
			f.tags = append(f.tags, tagId)
		}
	}
}

type BookCollectionQueryFilter struct {
	collections []interface{}
}

func (f BookCollectionQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.collections != nil && len(f.collections) != 0 {
		return db.Joins("inner join collection_books on collection_books.book_id = books.id").Where("collection_books.collection_id in (?)", f.collections)
	}
	return db
}

func (f *BookCollectionQueryFilter) SetCollectionQueryFilter(collectionIds ...interface{}) {
	for _, collectionId := range collectionIds {
		if len(collectionId.(string)) > 0 {
			f.collections = append(f.collections, collectionId)
		}
	}
}
func (b *BooksQueryBuilder) ReadModels(models interface{}) (int, error) {
	query := database.DB
	query = ApplyFilters(b, query)
	var count = 0
	err := query.Limit(b.getLimit()).Offset(b.getOffset()).Find(models).Offset(-1).Count(&count).Error

	if err == sql.ErrNoRows {
		return 0, nil
	}
	return count, err
}

func CreateBooks(books []model.Book) error {
	var err error
	tx := database.DB.Begin()
	for _, book := range books {
		err = tx.Create(&book).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

func UpdateBooks(Books []model.Book, allowFields ...string) error {
	var err error
	tx := database.DB.Begin()
	for _, book := range Books {
		updateMap := make(map[string]interface{})
		r := reflect.ValueOf(book)
		for _, propertyName := range allowFields {
			value := r.FieldByName(propertyName)
			updateMap[propertyName] = value.Interface()
		}
		err := database.DB.Model(&book).Updates(updateMap).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return err
}

func DeleteBooks(ids ...int) error {
	var err error
	tx := database.DB.Begin()
	for _, id := range ids {
		book := model.Book{}
		book.ID = uint(id)
		err = tx.Delete(&book).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}
func DeleteBooksPermanently(tx *gorm.DB, ids ...int) error {
	var err error
	deleteTx := tx
	if tx == nil {
		tx = database.DB.Begin()
	}
	for _, id := range ids {
		book := model.Book{}
		book.ID = uint(id)
		err = deleteTx.Unscoped().Delete(&book).Error
		if err != nil {
			return err
		}
	}
	deleteTx.Commit()
	return nil
}

func AddTagToBook(bookId int, tagIds ...int) error {
	tagsToAdd := make([]interface{}, 0)
	for _, tagId := range tagIds {
		tagsToAdd = append(tagsToAdd, &model.Tag{Model: gorm.Model{ID: uint(tagId)}})
	}
	return database.DB.Model(&model.Book{Model: gorm.Model{ID: uint(bookId)}}).Association("Tags").Append(tagsToAdd...).Error
}

func GetBookPath(bookPath string, libraryId uint) (error, string) {
	var err error
	library, err := GetLibraryById(libraryId)
	if err != nil {
		return err, ""
	}
	storePath := path.Join(library.Path, bookPath)
	err = os.MkdirAll(storePath, os.ModePerm)
	return err, storePath
}

func GetBookTagsByType(bookId uint, tagType string) ([]model.Tag, error) {
	tags := make([]model.Tag, 0)
	err := database.DB.Table(
		"tags",
	).Select(
		"tags.*",
	).Joins(
		"inner join book_tags as b2t on tags.id = b2t.tag_id",
	).Where(
		"tags.type = ? and b2t.book_id = ?", tagType, bookId,
	).Scan(&tags).Error
	return tags, err
}

func GetBookTagsByTypes(bookId uint, tagTypes ...string) ([]model.Tag, error) {
	tags := make([]model.Tag, 0)
	err := database.DB.Table(
		"tags",
	).Select(
		"tags.*",
	).Joins(
		"inner join book_tags as b2t on tags.id = b2t.tag_id",
	).Where(
		"tags.type in (?) and b2t.book_id = ?", tagTypes, bookId,
	).Scan(&tags).Error
	return tags, err
}

func GetBookTag(bookId uint) ([]model.Tag, error) {
	tags := make([]model.Tag, 0)
	err := database.DB.Model(&model.Book{Model: gorm.Model{ID: bookId}}).Association("Tags").Find(&tags).Error
	return tags, err
}

func GetBookById(bookId uint) (model.Book, error) {
	var book model.Book
	err := database.DB.Find(&book, bookId).Error
	return book, err
}

//generate thumbnail image
func GenerateCoverThumbnail(coverImageFilePath string, storePath string) (string, error) {
	// setup image decoder
	fileExt := filepath.Ext(coverImageFilePath)
	thumbnailImageFile, err := os.Open(coverImageFilePath)
	if err != nil {
		return "", err
	}
	var thumbnailImage image.Image
	if strings.ToLower(fileExt) == ".png" {
		thumbnailImage, err = png.Decode(thumbnailImageFile)
	}
	if strings.ToLower(fileExt) == ".jpg" {
		thumbnailImage, err = jpeg.Decode(thumbnailImageFile)
	}
	if err != nil {
		return "", err
	}

	// make thumbnail
	resizeImage := resize.Thumbnail(480, 480, thumbnailImage, resize.Lanczos3)

	// mkdir
	err = os.MkdirAll(storePath, os.ModePerm)
	if err != nil {
		return "", err
	}
	thumbnailImagePath := filepath.Join(storePath, fmt.Sprintf("cover_thumbnail%s", fileExt))
	output, err := os.Create(thumbnailImagePath)
	if err != nil {
		return "", err
	}

	defer thumbnailImageFile.Close()
	defer output.Close()

	// save result
	err = jpeg.Encode(output, resizeImage, nil)
	if err != nil {
		return "", err
	}
	return thumbnailImagePath, nil
}
