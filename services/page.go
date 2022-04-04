package services

import (
	"database/sql"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"gorm.io/gorm"
	"os"
)

func CreatePage(page *model.Page) error {
	return database.Instance.Create(page).Error
}

type PageQueryBuilder struct {
	DefaultPageFilter
	IdQueryFilter
	OrderQueryFilter
	BookIdFilter
}

type BookIdFilter struct {
	bookId []interface{}
}

func (f BookIdFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	queryParams := make([]interface{}, 0)
	if f.bookId != nil {
		for _, bookId := range f.bookId {
			if len((bookId).(string)) != 0 {
				queryParams = append(queryParams, bookId)
			}
		}
		return db.Where("book_id in (?)", queryParams)
	}
	return db
}

func (f *BookIdFilter) SetBookIdFilter(bookIds ...interface{}) {
	f.bookId = append(f.bookId, bookIds...)
}

func (b *PageQueryBuilder) ReadModels(models interface{}) (int64, error) {
	query := database.Instance
	query = ApplyFilters(b, query)
	var count int64 = 0
	err := query.Unscoped().Limit(b.getLimit()).Offset(b.getOffset()).Find(models).Offset(-1).Count(&count).Error
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return count, err
}

func DeletePages(id ...int) error {
	var err error
	for _, pageId := range id {
		pageModel := model.Page{}
		err = GetModelById(&pageModel, pageId)
		if err != nil {
			return err
		}
		file, _ := os.Stat(pageModel.Path)
		if file != nil {
			err = os.Remove(pageModel.Path)
			if err != nil {
				return err
			}
		}
		err = DeleteModels(pageModel, pageId)
		if err != nil {
			return err
		}
	}
	return nil
}
