package serializer

import (
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/services"
	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
	"time"
)

type BaseHistoryTemplate struct {
	ID        uint      `json:"id"`
	UserId    int       `json:"user_id"`
	BookId    int       `json:"book_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (t *BaseHistoryTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	var err error
	serializerModel := dataModel.(model.History)
	err = copier.Copy(t, serializerModel)
	if err != nil {
		return err
	}
	return nil
}

type HistoryWithBookTemplate struct {
	ID        uint             `json:"id"`
	UserId    int              `json:"user_id"`
	BookId    int              `json:"book_id"`
	Book      BaseBookTemplate `json:"book"`
	CreatedAt time.Time        `json:"created_at"`
}

func (t *HistoryWithBookTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	var err error
	serializerModel := dataModel.(model.History)
	err = copier.Copy(t, serializerModel)
	if err != nil {
		return err
	}
	book := model.Book{Model: gorm.Model{ID: serializerModel.BookId}}
	err = services.GetBook(&book)
	if err != nil {
		return err
	}
	bookTemplate := BaseBookTemplate{}
	err = bookTemplate.Serializer(book, nil)
	if err != nil {
		return err
	}
	t.Book = bookTemplate
	return nil
}
