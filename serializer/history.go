package serializer

import (
	"github.com/allentom/youcomic-api/model"
	"github.com/jinzhu/copier"
	"time"
)

type BaseHistoryTemplate struct {
	ID        uint      `json:"id"`
	UserId     int       `json:"user_id"`
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