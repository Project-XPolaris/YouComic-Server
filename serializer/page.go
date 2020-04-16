package serializer

import (
	"fmt"
	"github.com/allentom/youcomic-api/config"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/utils"
	"github.com/jinzhu/copier"
	"os"
	"path"
	"time"
)

type BasePageTemplate struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Order     int       `json:"order"`
	BookId    int       `json:"book_id"`
	Path      string    `json:"path"`
}
type PageTemplateWithSize struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Order     int       `json:"order"`
	BookId    int       `json:"book_id"`
	Path      string    `json:"path"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
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

func (t *PageTemplateWithSize) Serializer(dataModel interface{}, context map[string]interface{}) error {
	var err error
	serializerModel := dataModel.(model.Page)
	err = copier.Copy(t, serializerModel)
	t.Path = fmt.Sprintf("/assets/books/%d/%s?t=%d", serializerModel.BookId, serializerModel.Path, time.Now().Unix())
	filePath := path.Join(config.Config.Store.Books, fmt.Sprintf("/%d/%s", serializerModel.BookId, serializerModel.Path))
	if _, err := os.Stat(filePath); err == nil {
		width, height, _ := utils.GetImageDimension(filePath)
		t.Width = width
		t.Height = height
	}

	if err != nil {
		return err
	}
	return nil
}
