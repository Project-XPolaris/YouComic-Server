package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// LLMTagResult 存储LLM识别的单个标签结果
type LLMTagResult struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// LLMTagResults LLM标签结果的切片，实现gorm的Scanner和Valuer接口
type LLMTagResults []LLMTagResult

// Scan 实现sql.Scanner接口，从数据库读取JSON数据
func (l *LLMTagResults) Scan(value interface{}) error {
	if value == nil {
		*l = LLMTagResults{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("cannot scan non-[]byte value into LLMTagResults")
	}

	return json.Unmarshal(bytes, l)
}

// Value 实现driver.Valuer接口，将数据写入数据库
func (l LLMTagResults) Value() (driver.Value, error) {
	if len(l) == 0 {
		return "[]", nil
	}
	return json.Marshal(l)
}

// LLMTagHistory LLM标签识别历史记录
type LLMTagHistory struct {
	gorm.Model
	// 原始文本内容，用作查找的键
	OriginalText string `gorm:"type:text" json:"original_text"`

	// 文本内容的哈希值，用于快速比较（主要查询字段）
	TextHash string `gorm:"type:varchar(64);index:idx_hash_model_prompt" json:"text_hash"`

	// LLM模型信息
	ModelName    string `gorm:"type:varchar(100);index:idx_hash_model_prompt" json:"model_name"`
	ModelVersion string `gorm:"type:varchar(50);index:idx_hash_model_prompt" json:"model_version"`

	// 使用的Prompt（可能为空，表示使用默认prompt）
	CustomPrompt string `gorm:"type:text" json:"custom_prompt"`

	// Prompt的哈希值，用于快速比较（主要查询字段）
	PromptHash string `gorm:"type:varchar(64);index:idx_hash_model_prompt" json:"prompt_hash"`

	// LLM识别的结果，以JSON格式存储
	Results LLMTagResults `gorm:"type:json" json:"results"`

	// 处理耗时（毫秒）
	ProcessingTimeMs int64 `json:"processing_time_ms"`

	// 是否成功
	Success bool `gorm:"default:true" json:"success"`

	// 错误信息（如果有）
	ErrorMessage string `gorm:"type:text" json:"error_message,omitempty"`

	// 使用计数，每次复用时递增
	UsageCount int `gorm:"default:1" json:"usage_count"`

	// 最后使用时间
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// TableName 指定表名
func (LLMTagHistory) TableName() string {
	return "llm_tag_histories"
}
