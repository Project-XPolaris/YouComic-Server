package services

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/allentom/youcomic-api/application"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/utils"
	"gorm.io/gorm"
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
		err := database.DB.Where("name = ?", application.DEFAULT_LIBRARY_NAME).Find(&library).Error
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
	LibraryQueryFilter
	DirectoryNameQueryFilter
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

type DirectoryNameQueryFilter struct {
	Key string
}

func (f DirectoryNameQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if len(f.Key) != 0 {
		return db.Where("path like ?", fmt.Sprintf("%%%s%%",f.Key))
	}
	return db
}
func (f *DirectoryNameQueryFilter) SetPathSearchQueryFilter(searchKey interface{}) {
	f.Key = searchKey.(string)

}

type LibraryQueryFilter struct {
	library []interface{}
}

func (f LibraryQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.library != nil && len(f.library) != 0 {
		return db.Where("library_id in (?)", f.library)
	}
	return db
}

func (f *LibraryQueryFilter) SetLibraryQueryFilter(libraries ...interface{}) {
	for _, libraryId := range libraries {
		if len(libraryId.(string)) > 0 {
			f.library = append(f.library, libraryId)
		}
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
func (b *BooksQueryBuilder) ReadModels(models interface{}) (int64, error) {
	query := database.DB
	query = ApplyFilters(b, query)
	var count int64 = 0
	err := query.Limit(b.getLimit()).Offset(b.getOffset()).Find(models).Offset(-1).Count(&count).Error

	if err == sql.ErrNoRows {
		return 0, nil
	}
	return count, err
}

func CreateBooks(books []model.Book) error {
	var err error

	for _, book := range books {
		err = database.DB.Create(&book).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateBooks(Books []model.Book, allowFields ...string) error {
	var err error
	for _, book := range Books {
		updateMap := make(map[string]interface{})
		r := reflect.ValueOf(book)
		for _, propertyName := range allowFields {
			value := r.FieldByName(propertyName)
			updateMap[propertyName] = value.Interface()
		}
		err := database.DB.Model(&book).Updates(updateMap).Error
		if err != nil {
			return err
		}
	}
	return err
}

func DeleteBooks(ids ...int) error {
	var err error
	for _, id := range ids {
		book := model.Book{}
		book.ID = uint(id)
		err = database.DB.Delete(&book).Error
		if err != nil {
			return err
		}
	}
	return nil
}
func DeleteBooksPermanently(tx *gorm.DB, ids ...int) error {
	var err error

	for _, id := range ids {
		book := model.Book{}
		book.ID = uint(id)
		err = database.DB.Unscoped().Delete(&book).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func AddTagToBook(bookId int, tagIds ...int) error {
	tagsToAdd := make([]interface{}, 0)
	for _, tagId := range tagIds {
		tagsToAdd = append(tagsToAdd, &model.Tag{Model: gorm.Model{ID: uint(tagId)}})
	}
	return database.DB.Model(&model.Book{Model: gorm.Model{ID: uint(bookId)}}).Association("Tags").Append(tagsToAdd...)
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
	err := database.DB.Model(&model.Book{Model: gorm.Model{ID: bookId}}).Association("Tags").Find(&tags)
	return tags, err
}

func GetBookById(bookId uint) (model.Book, error) {
	var book model.Book
	err := database.DB.Find(&book, bookId).Error
	return book, err
}

type BookDailyResult struct {
	Date  string
	Total int
}

func (b *BooksQueryBuilder) GetDailyCount() ([]BookDailyResult, int64, error) {
	query := database.DB
	query = ApplyFilters(b, query)
	var count int64 = 0
	rawRows, err := query.Model(
		&model.Book{},
	).Limit(
		b.getLimit(),
	).Offset(
		b.getOffset(),
	).Order(
		b.Order,
	).Select(
		"date(created_at) as `date` ,count(*) as total",
	).Group(
		`date(created_at)`,
	).Count(&count).Rows()

	if err != nil {
		return nil, 0, err
	}
	result := make([]BookDailyResult, 0)
	for rawRows.Next() {
		var dailyResult BookDailyResult
		err = database.DB.ScanRows(rawRows, &dailyResult)
		if err != nil {
			return nil, count, err
		}
		result = append(result, dailyResult)
	}
	return result, count, err
}

func RenderBookDirectoryRenameText(book *model.Book,Pattern string,Slots []RenameSlot) string{
	name := Pattern
	for _, slot := range Slots {
		if slot.Type == "title" {
			titleText := strings.ReplaceAll(slot.Pattern, "%content%", book.Name)
			name = strings.ReplaceAll(name, "%title%", titleText)
			continue
		}
		var tags []*model.Tag
		linq.From(book.Tags).Where(func(i interface{}) bool {
			return i.(*model.Tag).Type == slot.Type
		}).ToSlice(&tags)
		slotTexts := make([]string, 0)
		for _, tag := range tags {
			slotTexts = append(slotTexts, tag.Name)
		}
		slotContent := ""
		if len(slotTexts) != 0 {
			slotContent = strings.ReplaceAll(slot.Pattern, "%content%", strings.Join(slotTexts, slot.Sep))
		}
		name = strings.ReplaceAll(
			name,
			fmt.Sprintf("%%%s%%", slot.Type),
			slotContent,
		)
	}
	name = strings.TrimSpace(name)
	return name
}
func RenameBookDirectoryById(bookId int, pattern string,slots []RenameSlot) (*model.Book, error) {
	var book model.Book
	err := database.DB.Preload("Tags").First(&book, bookId).Error
	if err != nil {
		return nil, err
	}
	var library model.Library
	err = database.DB.First(&library, book.LibraryId).Error
	if err != nil {
		return nil, err
	}
	name := RenderBookDirectoryRenameText(&book,pattern,slots)
	if name == filepath.Base(book.Path) {
		return &book,nil
	}
	newPath := utils.ReplaceLastString(book.Path, filepath.Base(book.Path), name)
	err = os.Rename(path.Join(library.Path, book.Path), path.Join(library.Path, newPath))
	if err != nil {
		return nil, err
	}
	book.Path = newPath
	err = database.DB.Save(&book).Error
	if err != nil {
		return nil, err
	}
	return &book, nil
}
func RenameBookDirectory(book *model.Book,library *model.Library, name string) (*model.Book, error) {
	if name == filepath.Base(book.Path) {
		return book,nil
	}
	newPath := utils.ReplaceLastString(book.Path, filepath.Base(book.Path), name)
	err := os.Rename(path.Join(library.Path, book.Path), path.Join(library.Path, newPath))
	if err != nil {
		return nil, err
	}
	book.Path = newPath
	err = database.DB.Save(&book).Error
	if err != nil {
		return nil, err
	}
	return book, nil
}