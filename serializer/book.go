package serializer

import (
	"fmt"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/services"
	"github.com/jinzhu/copier"
	"time"
)

type BaseBookTemplate struct {
	ID        uint        `json:"id"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	Name      string      `json:"name"`
	Cover     string      `json:"cover"`
	Tags      interface{} `json:"tags"`
}

func (b *BaseBookTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	serializerModel := dataModel.(model.Book)
	err := copier.Copy(b, serializerModel)
	if err != nil {
		return err
	}
	b.Cover = fmt.Sprintf("/assets/books/%d/%s?t=%d", serializerModel.ID, serializerModel.Cover, time.Now().Unix())
	tags, err := services.GetBookTagsByTypes(serializerModel.ID, "artist", "translator", "series", "theme")
	if err != nil {
		return err
	}
	serializedTags := SerializeMultipleTemplate(tags, &BaseTagTemplate{}, nil)
	b.Tags = serializedTags
	return nil
}
