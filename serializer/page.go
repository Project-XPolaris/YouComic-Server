package serializer

import (
	"fmt"
	"github.com/allentom/youcomic-api/model"
	"github.com/jinzhu/copier"
	"time"
)

type BasePageTemplate struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Order     int       `json:"order"`
	BookId    int       `json:"book_id"`
	Path      string    `json:"path"`
}

func (t *BasePageTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	var err error
	serializerModel := dataModel.(model.Page)
	err = copier.Copy(t, serializerModel)
	t.Path = fmt.Sprintf("/assets/books/%d/%s?t=%d", serializerModel.BookId, serializerModel.Path, time.Now().Unix())
	if err != nil {
		return err
	}
	return nil
}
