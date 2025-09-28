package services

import (
	"fmt"
	"time"

	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	// 模板类型常量
	TemplateTypeTagPrompt      = "tag_prompt"
	TemplateTypeBatchTagPrompt = "batch_tag_prompt"

	// 默认模板名称
	DefaultTagPromptTemplateName      = "default_tag_prompt"
	DefaultBatchTagPromptTemplateName = "default_batch_tag_prompt"
)

// InitDefaultTemplates 初始化默认模板到数据库
func InitDefaultTemplates() error {
	logrus.Info("开始初始化默认模板到数据库...")

	// 初始化默认标签提示模板
	err := initDefaultTagPromptTemplate()
	if err != nil {
		return fmt.Errorf("初始化默认标签提示模板失败: %v", err)
	}
	logrus.Info("默认标签提示模板初始化完成")

	// 初始化默认批量标签提示模板
	err = initDefaultBatchTagPromptTemplate()
	if err != nil {
		return fmt.Errorf("初始化默认批量标签提示模板失败: %v", err)
	}
	logrus.Info("默认批量标签提示模板初始化完成")

	logrus.Info("所有默认模板初始化完成")
	return nil
}

// initDefaultTagPromptTemplate 初始化默认标签提示模板
func initDefaultTagPromptTemplate() error {
	var template model.Template
	err := database.Instance.Where("name = ?", DefaultTagPromptTemplateName).First(&template).Error

	if err == gorm.ErrRecordNotFound {
		// 模板不存在，创建新的
		template = model.Template{
			Name:        DefaultTagPromptTemplateName,
			Type:        TemplateTypeTagPrompt,
			Content:     getDefaultTagPromptTemplateContent(),
			Description: "默认的单个文本标签分析提示模板",
			IsDefault:   true,
			Version:     "1.0",
		}
		return database.Instance.Create(&template).Error
	} else if err != nil {
		return err
	}

	// 模板已存在，更新内容（如果需要）
	template.Content = getDefaultTagPromptTemplateContent()
	template.LastUsedAt = &[]time.Time{time.Now()}[0]
	return database.Instance.Save(&template).Error
}

// initDefaultBatchTagPromptTemplate 初始化默认批量标签提示模板
func initDefaultBatchTagPromptTemplate() error {
	var template model.Template
	err := database.Instance.Where("name = ?", DefaultBatchTagPromptTemplateName).First(&template).Error

	if err == gorm.ErrRecordNotFound {
		// 模板不存在，创建新的
		template = model.Template{
			Name:        DefaultBatchTagPromptTemplateName,
			Type:        TemplateTypeBatchTagPrompt,
			Content:     getBatchTagPromptTemplateContent(),
			Description: "默认的批量文本标签分析提示模板",
			IsDefault:   true,
			Version:     "1.0",
		}
		return database.Instance.Create(&template).Error
	} else if err != nil {
		return err
	}

	// 模板已存在，更新内容（如果需要）
	template.Content = getBatchTagPromptTemplateContent()
	template.LastUsedAt = &[]time.Time{time.Now()}[0]
	return database.Instance.Save(&template).Error
}

// GetTemplateByName 根据名称获取模板
func GetTemplateByName(name string) (*model.Template, error) {
	var template model.Template
	err := database.Instance.Where("name = ?", name).First(&template).Error
	if err != nil {
		return nil, err
	}

	// 更新最后使用时间
	now := time.Now()
	template.LastUsedAt = &now
	database.Instance.Save(&template)

	return &template, nil
}

// GetTemplatesByType 根据类型获取模板列表
func GetTemplatesByType(templateType string) ([]model.Template, error) {
	var templates []model.Template
	err := database.Instance.Where("type = ?", templateType).Find(&templates).Error
	return templates, err
}

// CreateTemplate 创建新模板
func CreateTemplate(template *model.Template) error {
	return database.Instance.Create(template).Error
}

// UpdateTemplate 更新模板
func UpdateTemplate(template *model.Template) error {
	return database.Instance.Save(template).Error
}

// DeleteTemplate 删除模板
func DeleteTemplate(id uint) error {
	return database.Instance.Delete(&model.Template{}, id).Error
}

// GetTemplateByID 根据ID获取模板
func GetTemplateByID(id uint) (*model.Template, error) {
	var template model.Template
	err := database.Instance.First(&template, id).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// GetDB 获取数据库实例（用于API处理器）
func GetDB() *gorm.DB {
	return database.Instance
}

// GetDefaultTagPromptTemplate 获取默认的标签提示模板
func GetDefaultTagPromptTemplate() (string, error) {
	template, err := GetTemplateByName(DefaultTagPromptTemplateName)
	if err != nil {
		logrus.WithError(err).Warn("无法从数据库获取默认标签提示模板，使用硬编码版本")
		return getDefaultTagPromptTemplateContent(), nil
	}
	return template.Content, nil
}

// GetBatchTagPromptTemplate 获取默认的批量标签提示模板
func GetBatchTagPromptTemplate() (string, error) {
	template, err := GetTemplateByName(DefaultBatchTagPromptTemplateName)
	if err != nil {
		logrus.WithError(err).Warn("无法从数据库获取默认批量标签提示模板，使用硬编码版本")
		return getBatchTagPromptTemplateContent(), nil
	}
	return template.Content, nil
}

// getDefaultTagPromptTemplateContent 获取默认标签提示模板的内容（硬编码版本）
func getDefaultTagPromptTemplateContent() string {
	return `你是一个专业的漫画标签分析助手。请分析以下文本，提取出漫画相关的标签信息。

文本内容："%s"

请从文本中提取以下类型的标签：
- artist: 画师/作者名称
- series: 贩售的会场（如Comic Market、Comic Market Online等）
- name: 漫画标题/名称
- theme: 主题(一般是某个动画或者游戏等的名字)
- translator: 翻译者
- type: 漫画类型(如CG、同人志等)
- lang: 语言(如日语、中文、英语等),如果有翻译，则记录翻译的语言
- magazine: 杂志名称
- societies: 画师所在的社团名称
- original-lang: 原语言(如日语、中文、英语等)
- chapter: 章节名称(如"第一章"、"序章"、"最终话"等)
- chapter_number: 章节序号(如"01"、"1"、"final"等，提取数字或序号)


请以JSON格式返回结果，格式如下：
{
  "tags": [
    {"name": "标签名称", "type": "标签类型"},
    {"name": "标签名称", "type": "标签类型"}
  ]
}

要求：
1. 只提取确实存在于文本中的信息
2. 标签名称要准确，去除多余符号
3. 如果无法确定标签类型，可以留空
4. 不要添加文本中不存在的信息
5. 返回纯JSON，不要其他说明文字`
}

// getBatchTagPromptTemplateContent 获取默认批量标签提示模板的内容（硬编码版本）
func getBatchTagPromptTemplateContent() string {
	return `你是一个专业的漫画标签分析助手。请分析以下多个文本，为每个文本提取出漫画相关的标签信息。

文本列表：
%s

请从每个文本中提取以下类型的标签：
- artist: 画师/作者名称
- series: 贩售的会场（如Comic Market、Comic Market Online等）
- name: 漫画标题/名称
- theme: 主题(一般是某个动画或者游戏等的名字)
- translator: 翻译者
- type: 漫画类型(如CG、同人志等)
- lang: 语言(如日语、中文、英语等),如果有翻译，则记录翻译的语言
- magazine: 杂志名称
- societies: 画师所在的社团名称
- original-lang: 原语言(如日语、中文、英语等)
- chapter: 章节名称(如"第一章"、"序章"、"最终话"等)
- chapter_number: 章节序号(如"01"、"1"、"final"等，提取数字或序号)

请以JSON格式返回结果，格式如下：
{
  "results": [
    {
      "text_index": 1,
      "tags": [
        {"name": "标签名称", "type": "标签类型"},
        {"name": "标签名称", "type": "标签类型"}
      ]
    },
    {
      "text_index": 2,
      "tags": [
        {"name": "标签名称", "type": "标签类型"},
        {"name": "标签名称", "type": "标签类型"}
      ]
    }
  ]
}

要求：
1. 为每个文本都提供一个结果项，即使没有识别到标签
2. text_index对应文本列表中的编号（从1开始）
3. 只提取确实存在于文本中的信息
4. 标签名称要准确，去除多余符号
5. 如果无法确定标签类型，可以留空
6. 不要添加文本中不存在的信息
7. 返回纯JSON，不要其他说明文字`
}
