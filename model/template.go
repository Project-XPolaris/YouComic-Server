package model

import (
	"time"

	"gorm.io/gorm"
)

// Template 模板表，用于存储各种模板内容
type Template struct {
	gorm.Model
	// 模板名称，唯一标识
	Name string `gorm:"type:varchar(100);uniqueIndex" json:"name"`

	// 模板类型，如 "tag_prompt", "batch_tag_prompt" 等
	Type string `gorm:"type:varchar(50);index" json:"type"`

	// 模板内容
	Content string `gorm:"type:text" json:"content"`

	// 模板描述
	Description string `gorm:"type:text" json:"description"`

	// 是否为系统默认模板
	IsDefault bool `gorm:"default:false;index" json:"is_default"`

	// 版本号
	Version string `gorm:"type:varchar(20)" json:"version"`

	// 最后更新时间
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// TableName 指定表名
func (Template) TableName() string {
	return "templates"
}
