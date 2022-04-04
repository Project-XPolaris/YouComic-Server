package serializer

import (
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/projectxpolaris/youcomic/model"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

type BaseBookTemplate struct {
	ID           uint        `json:"id"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	Name         string      `json:"name"`
	Cover        string      `json:"cover"`
	LibraryId    uint        `json:"library_id"`
	Tags         interface{} `json:"tags"`
	DirName      string      `json:"dirName"`
	OriginalName string      `json:"originalName"`
}

func (b *BaseBookTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	serializerModel := dataModel.(model.Book)
	err := copier.Copy(b, serializerModel)
	if err != nil {
		return err
	}
	if len(b.Cover) != 0 {
		b.Cover = fmt.Sprintf("%s?t=%d",
			path.Join("/", "content", "book", strconv.Itoa(int(serializerModel.ID)), serializerModel.Cover),
			time.Now().Unix(),
		)
	}
	b.DirName = filepath.Base(serializerModel.Path)
	if len(b.OriginalName) == 0 {
		b.OriginalName = b.DirName
	}
	//tags, err := services.GetBookTagsByTypes(serializerModel.ID, "artist", "translator", "series", "theme")
	//if err != nil {
	//	return err
	//}
	serializedTags := SerializeMultipleTemplate(serializerModel.Tags, &BaseTagTemplate{}, nil)
	b.Tags = serializedTags
	return nil
}

type BookDailySummaryTemplate struct {
	Date  string `json:"date"`
	Total int    `json:"total"`
}

func (b *BookDailySummaryTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	err := copier.Copy(b, dataModel)
	return err
}
