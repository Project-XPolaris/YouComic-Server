package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

type JSONMap map[string]string

func (m JSONMap) Value() (driver.Value, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (m *JSONMap) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, m)
	case string:
		return json.Unmarshal([]byte(v), m)
	default:
		return fmt.Errorf("unsupported Scan type for JSONMap: %T", value)
	}
}

type Book struct {
	gorm.Model
	Name         string
	Cover        string
	History      []*History    `gorm:"foreignkey:BookId"`
	Page         []Page        `gorm:"foreignkey:BookId"`
	Tags         []*Tag        `gorm:"many2many:book_tags;"`
	Collections  []*Collection `gorm:"many2many:collection_books;"`
	Path         string
	LibraryId    uint
	OriginalName string
	// TitleTranslations stores i18n titles keyed by BCP-47 language tags, e.g. "en", "zh-CN"
	TitleTranslations JSONMap `gorm:"type:text"`
}
