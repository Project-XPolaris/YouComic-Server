package serializer

import (
	"fmt"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jinzhu/copier"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/services"
)

type BaseBookTemplate struct {
	ID                uint              `json:"id"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	Name              string            `json:"name"`
	Cover             string            `json:"cover"`
	CoverThumbnail    string            `json:"coverThumbnail"`
	LibraryId         uint              `json:"library_id"`
	Tags              interface{}       `json:"tags"`
	DirName           string            `json:"dirName"`
	OriginalName      string            `json:"originalName"`
	PageCount         int               `json:"pageCount"`
	TitleTranslations map[string]string `json:"titleTranslations"`
	HasThumbnail      bool              `json:"hasThumbnail"`
}

func (b *BaseBookTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	serializerModel := dataModel.(model.Book)
	err := copier.Copy(b, serializerModel)
	if err != nil {
		return err
	}
	if len(b.Cover) != 0 {
		// 原图URL
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
	if serializerModel.Tags != nil {
		serializedTags := SerializeMultipleTemplate(serializerModel.Tags, &BaseTagTemplate{}, nil)
		b.Tags = serializedTags
	}
	if serializerModel.Page != nil {
		b.PageCount = len(serializerModel.Page)
	}

	// 检查缩略图是否存在，如果存在则生成缩略图URL
	b.HasThumbnail = services.CheckBookThumbnailExists(serializerModel.ID, serializerModel.Cover)
	if b.HasThumbnail && len(serializerModel.Cover) > 0 {
		// 只有缩略图存在时才生成缩略图URL
		thumbnailExt := filepath.Ext(serializerModel.Cover)
		if thumbnailExt == "" {
			thumbnailExt = ".jpg" // 默认扩展名
		}
		b.CoverThumbnail = fmt.Sprintf("%s?t=%d",
			path.Join("/", "content", "book", strconv.Itoa(int(serializerModel.ID)), fmt.Sprintf("cover_thumbnail%s", thumbnailExt)),
			time.Now().Unix(),
		)
	}

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
