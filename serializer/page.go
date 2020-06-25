package serializer

import (
	"fmt"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/services"
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
	t.Path = fmt.Sprintf("/content/book/%d/%s?t=%d", serializerModel.BookId, serializerModel.Path, time.Now().Unix())
	if err != nil {
		return err
	}
	return nil
}

func (t *PageTemplateWithSize) Serializer(dataModel interface{}, context map[string]interface{}) error {
	var err error
	serializerModel := dataModel.(model.Page)
	err = copier.Copy(t, serializerModel)
	t.Path = fmt.Sprintf("%s/?t=%d", path.Join("/content/book",fmt.Sprintf("%d",serializerModel.BookId),serializerModel.Path), time.Now().Unix())

	book, err := services.GetBookById(uint(serializerModel.BookId))
	if err != nil {
		return err
	}

	library, err := services.GetLibraryById(book.LibraryId)
	if err != nil {
		return err
	}

	filePath := path.Join(library.Path, book.Path, serializerModel.Path)
	if _, err := os.Stat(filePath); err == nil {
		width, height, _ := utils.GetImageDimension(filePath)
		t.Width = width
		t.Height = height
	} else {
		return err
	}

	return nil
}
