package serializer

import (
	"github.com/allentom/youcomic-api/model"
	"github.com/jinzhu/copier"
	"time"
)

type BaseLibraryTemplate struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
}

func (b *BaseLibraryTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	serializerModel := dataModel.(model.Library)
	err := copier.Copy(b, serializerModel)
	return err
}
