package services

import (
	"database/sql"
	"fmt"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"gorm.io/gorm"
	"time"
)

func AddBookHistory(userId uint, bookId uint) error {
	var count int64 = 0
	database.Instance.Model(&model.History{}).Where(&model.History{UserId: userId, BookId: bookId}).Count(&count)
	fmt.Println(count)
	if count > 0 {
		return database.Instance.Model(&model.History{}).Where(&model.History{UserId: userId, BookId: bookId}).Update("UpdatedAt", time.Now()).Error
	} else {
		return database.Instance.Create(&model.History{UserId: userId, BookId: bookId}).Error
	}
}

type HistoryQueryBuilder struct {
	DefaultPageFilter
	IdQueryFilter
	OrderQueryFilter
	UserIdFilter
}

func (b *HistoryQueryBuilder) ReadModels() (int64, interface{}, error) {
	query := database.Instance
	query = ApplyFilters(b, query)
	var count int64 = 0
	md := make([]model.History, 0)
	err := query.Limit(b.getLimit()).Offset(b.getOffset()).Find(&md).Offset(-1).Count(&count).Error
	if err == sql.ErrNoRows {
		return 0, query, nil
	}
	return count, md, err
}
func (b *HistoryQueryBuilder) DeleteModels(permanently bool) error {
	query := database.Instance
	query = ApplyFilters(b, query)
	if permanently {
		query = query.Unscoped()
	}
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

func CreateHistoryByBook(bookId uint, pagePos uint, userId uint) (*model.History, error) {
	book := model.Book{}
	database.Instance.First(&book, bookId)
	if book.ID == 0 {
		return nil, fmt.Errorf("book not found")
	}
	var history model.History
	database.Instance.Where(&model.History{BookId: bookId, UserId: userId}).First(&history)
	if history.ID != 0 {
		database.Instance.Model(&history).Update("page_pos", pagePos)
		return &history, nil
	}
	history = model.History{
		BookId:  bookId,
		PagePos: pagePos,
		UserId:  userId,
	}

	err := database.Instance.Create(&history).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}
