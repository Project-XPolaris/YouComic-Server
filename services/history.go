package services

import (
	"database/sql"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"github.com/jinzhu/gorm"
)

func AddBookHistory(userId uint, bookId uint) error {
	return database.DB.Create(&model.History{UserId: userId, BookId: bookId}).Error
}

type HistoryQueryBuilder struct {
	DefaultPageFilter
	IdQueryFilter
	UserIdFilter
}

func (b *HistoryQueryBuilder) ReadModels() (int, interface{}, error) {
	query := database.DB
	query = ApplyFilters(b, query)
	var count = 0
	md := make([]model.History, 0)
	err := query.Limit(b.getLimit()).Offset(b.getOffset()).Find(&md).Offset(-1).Count(&count).Error
	if err == sql.ErrNoRows {
		return 0, query, nil
	}
	return count, md, err
}
func (b *HistoryQueryBuilder) DeleteModels() error {
	query := database.DB
	query = ApplyFilters(b, query)
	err := query.Delete(model.History{}).Error
	return err
}

type UserIdFilter struct {
	userId []interface{}
}

func (f UserIdFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	queryParams := make([]interface{}, 0)
	if f.userId != nil {
		for _, userId := range f.userId {
			queryParams = append(queryParams, userId)
		}
		return db.Where("user_id in (?)", queryParams)
	}
	return db
}

func (f *UserIdFilter) SetUserIdFilter(userIds ...interface{}) {
	f.userId = append(f.userId, userIds...)
}
