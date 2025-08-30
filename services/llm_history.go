package services

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/plugin"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// generateHash 生成文本的哈希值
func generateHash(text string) string {
	hash := sha256.Sum256([]byte(text))
	return fmt.Sprintf("%x", hash)
}

// getLLMModelInfo 获取当前LLM模型信息
func getLLMModelInfo() (modelName, modelVersion string) {
	if plugin.LLM == nil {
		return "unknown", "unknown"
	}

	currentConfig := plugin.LLM.GetCurrentConfig()
	if currentConfig == nil {
		return "unknown", "unknown"
	}

	// 获取当前使用的默认提供商
	defaultProvider := currentConfig.Default
	if defaultProvider == "" {
		return "unknown", "unknown"
	}

	// 根据提供商获取对应的模型信息
	switch defaultProvider {
	case "openai":
		if currentConfig.OpenAI != nil && currentConfig.OpenAI.Enable {
			modelName = currentConfig.OpenAI.Model
			if modelName == "" {
				modelName = "gpt-3.5-turbo" // OpenAI默认模型
			}
			modelVersion = "openai-api"
		}
	case "gemini":
		if currentConfig.Gemini != nil && currentConfig.Gemini.Enable {
			modelName = currentConfig.Gemini.Model
			if modelName == "" {
				modelName = "gemini-pro" // Gemini默认模型
			}
			modelVersion = "gemini-api"
		}
	case "ollama":
		if currentConfig.Ollama != nil && currentConfig.Ollama.Enable {
			modelName = currentConfig.Ollama.Model
			if modelName == "" {
				modelName = "llama2" // Ollama默认模型
			}
			modelVersion = "ollama-api"
		}
	default:
		modelName = defaultProvider
		modelVersion = "unknown"
	}

	// 如果模型名为空，返回提供商名作为模型名
	if modelName == "" {
		modelName = defaultProvider
		modelVersion = "unknown"
	}

	return modelName, modelVersion
}

// FindLLMTagHistory 查找LLM标签识别历史记录
func FindLLMTagHistory(originalText, customPrompt string) (*model.LLMTagHistory, error) {
	textHash := generateHash(originalText)
	promptHash := generateHash(customPrompt)
	modelName, modelVersion := getLLMModelInfo()

	var history model.LLMTagHistory
	err := database.Instance.Where(
		"text_hash = ? AND model_name = ? AND model_version = ? AND prompt_hash = ? AND success = ?",
		textHash, modelName, modelVersion, promptHash, true,
	).First(&history).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 没有找到历史记录，返回nil而不是错误
		}
		return nil, err
	}

	return &history, nil
}

// SaveLLMTagHistory 保存LLM标签识别历史记录
func SaveLLMTagHistory(originalText, customPrompt string, results []model.LLMTagResult, processingTimeMs int64, success bool, errorMessage string) error {
	textHash := generateHash(originalText)
	promptHash := generateHash(customPrompt)
	modelName, modelVersion := getLLMModelInfo()

	history := model.LLMTagHistory{
		OriginalText:     originalText,
		TextHash:         textHash,
		ModelName:        modelName,
		ModelVersion:     modelVersion,
		CustomPrompt:     customPrompt,
		PromptHash:       promptHash,
		Results:          model.LLMTagResults(results),
		ProcessingTimeMs: processingTimeMs,
		Success:          success,
		ErrorMessage:     errorMessage,
		UsageCount:       1,
	}

	return database.Instance.Create(&history).Error
}

// UpdateLLMTagHistoryUsage 更新LLM标签历史记录的使用信息
func UpdateLLMTagHistoryUsage(historyId uint) error {
	now := time.Now()
	return database.Instance.Model(&model.LLMTagHistory{}).
		Where("id = ?", historyId).
		Updates(map[string]interface{}{
			"usage_count":  gorm.Expr("usage_count + 1"),
			"last_used_at": &now,
		}).Error
}

// CleanOldLLMTagHistory 清理旧的LLM标签历史记录
// 保留最近30天的记录，或者使用次数大于1的记录
func CleanOldLLMTagHistory() error {
	cutoffDate := time.Now().AddDate(0, 0, -30)

	result := database.Instance.Where(
		"created_at < ? AND usage_count <= 1",
		cutoffDate,
	).Delete(&model.LLMTagHistory{})

	if result.Error != nil {
		return result.Error
	}

	logrus.Infof("清理了 %d 条旧的LLM标签历史记录", result.RowsAffected)
	return nil
}

// FindLLMTagHistoryBatch 批量查找LLM标签识别历史记录
func FindLLMTagHistoryBatch(textsWithPrompts []struct{ Text, Prompt string }) (map[string]*model.LLMTagHistory, error) {
	if len(textsWithPrompts) == 0 {
		return map[string]*model.LLMTagHistory{}, nil
	}

	modelName, modelVersion := getLLMModelInfo()
	result := make(map[string]*model.LLMTagHistory)

	// 构建查询条件
	var conditions []string
	var args []interface{}

	for _, item := range textsWithPrompts {
		textHash := generateHash(item.Text)
		promptHash := generateHash(item.Prompt)

		conditions = append(conditions, "(text_hash = ? AND model_name = ? AND model_version = ? AND prompt_hash = ? AND success = ?)")
		args = append(args, textHash, modelName, modelVersion, promptHash, true)

		// 建立查找键
		key := textHash + "_" + promptHash
		result[key] = nil
	}

	// 批量查询
	var histories []model.LLMTagHistory
	query := "(" + strings.Join(conditions, " OR ") + ")"
	err := database.Instance.Where(query, args...).Find(&histories).Error
	if err != nil {
		return nil, err
	}

	// 将结果按键值对应起来
	for i, history := range histories {
		key := history.TextHash + "_" + history.PromptHash
		result[key] = &histories[i]
	}

	return result, nil
}

// UpdateLLMTagHistoryUsageBatch 批量更新LLM标签历史记录的使用信息
func UpdateLLMTagHistoryUsageBatch(historyIds []uint) error {
	if len(historyIds) == 0 {
		return nil
	}

	now := time.Now()
	return database.Instance.Model(&model.LLMTagHistory{}).
		Where("id IN (?)", historyIds).
		Updates(map[string]interface{}{
			"usage_count":  gorm.Expr("usage_count + 1"),
			"last_used_at": &now,
		}).Error
}

// GetLLMTagHistoryList 获取LLM标签历史记录列表
func GetLLMTagHistoryList(page, pageSize int, search string) (int64, []model.LLMTagHistory, error) {
	var histories []model.LLMTagHistory
	var count int64

	query := database.Instance.Model(&model.LLMTagHistory{}).Where("success = ?", true)

	// 如果有搜索条件，使用更智能的相关性搜索
	if search != "" {
		searchWords := strings.Fields(search)
		if len(searchWords) > 0 {
			// 构建多个搜索条件，提高相关性匹配
			conditions := make([]string, 0)
			args := make([]interface{}, 0)

			// 完全匹配（最高优先级）
			conditions = append(conditions, "original_text = ?")
			args = append(args, search)

			// 包含完整搜索文本
			conditions = append(conditions, "original_text LIKE ?")
			args = append(args, "%"+search+"%")

			// 包含搜索文本中的任一个词
			for _, word := range searchWords {
				if len(word) > 1 { // 忽略太短的词
					conditions = append(conditions, "original_text LIKE ?")
					args = append(args, "%"+word+"%")
				}
			}

			// 使用 OR 连接所有条件
			if len(conditions) > 0 {
				query = query.Where("("+strings.Join(conditions, " OR ")+")", args...)
			}
		}
	}

	// 获取总数
	err := query.Count(&count).Error
	if err != nil {
		return 0, nil, err
	}

	// 分页查询，按使用次数和创建时间排序（更常用的历史记录优先）
	offset := (page - 1) * pageSize
	err = query.Order("usage_count DESC, created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&histories).Error

	if err != nil {
		return 0, nil, err
	}

	return count, histories, nil
}

// convertLLMTagResultsToRawTags 将LLM历史记录结果转换为RawTag格式
func convertLLMTagResultsToRawTags(results model.LLMTagResults) []*RawTag {
	rawTags := make([]*RawTag, 0, len(results))

	for _, result := range results {
		if len(result.Name) > 0 {
			rawTags = append(rawTags, &RawTag{
				ID:     xid.New().String(),
				Name:   result.Name,
				Type:   result.Type,
				Source: "llm",
			})
		}
	}

	return rawTags
}
