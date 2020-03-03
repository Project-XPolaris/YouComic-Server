package services

import (
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"github.com/jinzhu/gorm"
)

type TagQueryBuilder struct {
	IdQueryFilter
	OrderQueryFilter
	NameQueryFilter
	NameSearchQueryFilter
	DefaultPageFilter
	TagTypeQueryFilter
}

type TagTypeQueryFilter struct {
	types []interface{}
}

func (f TagTypeQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.types != nil && len(f.types) != 0 {
		return db.Where("`type` in (?)", f.types)
	}
	return db
}

func (f *TagTypeQueryFilter) SetTagTypeQueryFilter(types ...interface{}) {
	for _, typeName := range types {
		if len(typeName.(string)) > 0 {
			f.types = append(f.types, typeName)
		}
	}
}

func (b *TagQueryBuilder) ReadModels() (int, interface{}, error) {
	query := database.DB
	query = ApplyFilters(b, query)
	var count = 0
	md := make([]model.Tag, 0)
	err := query.Limit(b.PageSize).Offset(b.getOffset()).Find(&md).Offset(-1).Count(&count).Error
	return count, md, err
}

func GetTagBooks(tagId uint, page int, pageSize int) (int, []model.Book, error) {
	var books []model.Book
	count := 0
	err := database.DB.Model(
		&model.Tag{Model: gorm.Model{ID: tagId}},
	).Limit(pageSize).Offset((page-1)*pageSize).Related(
		&books, "Books",
	).Error
	count = database.DB.Model(
		&model.Tag{Model: gorm.Model{ID: tagId}},
	).Association("Books").Count()
	return count, books, err
}

func AddBooksToTag(tagId int, bookIds []int) error {
	var err error
	books := make([]model.Book, 0)
	for _, bookId := range bookIds {
		books = append(books, model.Book{Model: gorm.Model{ID: uint(bookId)}})
	}
	err = database.DB.Model(&model.Tag{Model: gorm.Model{ID: uint(tagId)}}).Association("Books").Append(books).Error
	return err
}

func RemoveBooksFromTag(tagId int, bookIds []int) error {
	var err error
	books := make([]model.Book, 0)
	for _, bookId := range bookIds {
		books = append(books, model.Book{Model: gorm.Model{ID: uint(bookId)}})
	}
	err = database.DB.Model(&model.Tag{Model: gorm.Model{ID: uint(tagId)}}).Association("Books").Delete(books).Error
	return err
}

func AddOrCreateTagToBook(book *model.Book, tags []*model.Tag) (err error) {
	tx := database.DB.Begin()
	for _, tag := range tags {
		err = tx.Model(tag).FirstOrCreate(tag).Error
		if err != nil {
			return err
		}
		err = tx.Model(book).Association("Tags").Append(tag).Error
		if err != nil {
			return err
		}
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	return nil
}
