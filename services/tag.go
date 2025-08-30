package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/plugin"
	"github.com/projectxpolaris/youcomic/utils"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TagQueryBuilder struct {
	IdQueryFilter
	OrderQueryFilter
	NameQueryFilter
	NameSearchQueryFilter
	DefaultPageFilter
	TagTypeQueryFilter
	TagSubscriptionQueryFilter
	RandomQueryFilter
}

type TagTypeQueryFilter struct {
	types []interface{}
}

func (f TagTypeQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.types != nil && len(f.types) != 0 {
		return db.Where("`type` in (?)", f.types)
	}
	return db
}

func (f *TagTypeQueryFilter) SetTagTypeQueryFilter(types ...interface{}) {
	for _, typeName := range types {
		if len(typeName.(string)) > 0 {
			f.types = append(f.types, typeName)
		}
	}
}

type TagSubscriptionQueryFilter struct {
	subscriptions []interface{}
}

func (f TagSubscriptionQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.subscriptions != nil && len(f.subscriptions) != 0 {
		return db.Joins("inner join user_subscriptions on user_subscriptions.tag_id = tags.id").Where("user_subscriptions.user_id in (?)", f.subscriptions)
	}
	return db
}

func (f *TagSubscriptionQueryFilter) SetTagSubscriptionQueryFilter(subscriptions ...interface{}) {
	for _, subscriptionId := range subscriptions {
		if len(subscriptionId.(string)) > 0 {
			f.subscriptions = append(f.subscriptions, subscriptionId)
		}
	}
}

func (b *TagQueryBuilder) ReadModels() (int64, interface{}, error) {
	query := database.Instance
	query = ApplyFilters(b, query)
	if b.random {
		if _, ok := query.Config.Dialector.(*mysql.Dialector); ok {
			query = query.Order("rand()")
		}
		if _, ok := query.Config.Dialector.(*sqlite.Dialector); ok {
			query = query.Order("random()")
		}
	}
	var count int64 = 0
	md := make([]model.Tag, 0)
	err := query.Limit(b.getLimit()).Offset(b.getOffset()).Find(&md).Offset(-1).Count(&count).Error
	if err == sql.ErrNoRows {
		return 0, query, nil
	}
	return count, md, err
}

func GetTagBooks(tagId uint, page int, pageSize int) (int64, []model.Book, error) {
	var books []model.Book
	var count int64 = 0
	err := database.Instance.Model(
		&model.Tag{Model: gorm.Model{ID: tagId}},
	).Limit(pageSize).Offset((page - 1) * pageSize).Preload("Books").Error
	count = database.Instance.Model(
		&model.Tag{Model: gorm.Model{ID: tagId}},
	).Association("Books").Count()
	return count, books, err
}

func AddBooksToTag(tagId int, bookIds []int) error {
	var err error
	books := make([]model.Book, 0)
	for _, bookId := range bookIds {
		books = append(books, model.Book{Model: gorm.Model{ID: uint(bookId)}})
	}
	err = database.Instance.Model(&model.Tag{Model: gorm.Model{ID: uint(tagId)}}).Association("Books").Append(books)
	return err
}

func RemoveBooksFromTag(tagId int, bookIds []int) error {
	var err error
	books := make([]model.Book, 0)
	for _, bookId := range bookIds {
		books = append(books, model.Book{Model: gorm.Model{ID: uint(bookId)}})
	}
	err = database.Instance.Model(&model.Tag{Model: gorm.Model{ID: uint(tagId)}}).Association("Books").Delete(books)
	return err
}

type TagStrategy int

const (
	Overwrite TagStrategy = iota + 1
	Append
	FillEmpty
	ReplaceSameType
)

func AddOrCreateTagToBook(book *model.Book, tags []*model.Tag, strategy TagStrategy) (err error) {
	for _, tag := range tags {
		err = database.Instance.FirstOrCreate(tag, model.Tag{Name: tag.Name, Type: tag.Type}).Error
		if err != nil {
			return err
		}
	}
	ass := database.Instance.Model(book).Association("Tags")
	if strategy == Overwrite {
		err = ass.Clear()
		if err != nil {
			return err
		}
	}
	if strategy == Append {

	}
	appendTag := tags
	if strategy == FillEmpty {
		var existTags []model.Tag
		err = ass.Find(&existTags)
		if err != nil {
			return err
		}
		appendTag = []*model.Tag{}
		for _, tag := range tags {
			isExist := false
			for _, existTag := range existTags {
				if tag.Type == existTag.Type {
					isExist = true
					break
				}
			}
			if !isExist {
				appendTag = append(appendTag, tag)
			}
		}
	}
	if strategy == ReplaceSameType {
		var existTags []model.Tag
		err = ass.Find(&existTags)
		if err != nil {
			return err
		}
		appendTag = []*model.Tag{}
		for _, tag := range tags {
			for _, existTag := range existTags {
				if tag.Type == existTag.Type {
					err = ass.Delete(existTag)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	for _, tag := range appendTag {
		err = database.Instance.Model(tag).Association("Books").Append(book)
		if err != nil {
			return err
		}
	}
	return nil
}

// add users to tag
func AddTagSubscription(tagId uint, users ...interface{}) error {
	tag := &model.Tag{Model: gorm.Model{ID: tagId}}
	err := database.Instance.Model(tag).Association("Subscriptions").Append(users...)
	return err
}

// remove users from tag
func RemoveTagSubscription(tagId uint, users ...interface{}) error {
	tag := &model.Tag{Model: gorm.Model{ID: tagId}}
	err := database.Instance.Model(tag).Association("Subscriptions").Delete(users...)
	return err
}

// get tag with tag id
func GetTagById(id uint) (*model.Tag, error) {
	tag := &model.Tag{Model: gorm.Model{ID: id}}
	err := database.Instance.Find(tag).Error
	return tag, err
}

type TagCount struct {
	Name  string `json:"name"`
	Total int    `json:"total"`
}

func (b *TagQueryBuilder) GetTagCount() (int64, interface{}, error) {
	query := database.Instance
	query = ApplyFilters(b, query)
	var count int64 = 0
	rows, err := query.Model(&model.Tag{}).Select(
		`name,count(book_tags.book_id) as total`,
	).Joins(
		`inner join book_tags on tags.id = book_tags.tag_id`,
	).Group(
		`name`,
	).Limit(b.getLimit()).Offset(b.getOffset()).Count(&count).Rows()
	if err == sql.ErrNoRows {
		return 0, query, nil
	}
	result := make([]TagCount, 0)
	for rows.Next() {
		var tagCount TagCount
		err = database.Instance.ScanRows(rows, &tagCount)
		if err != nil {
			return 0, nil, err
		}
		result = append(result, tagCount)
	}
	return count, result, err
}

type TagTypeCount struct {
	Name  string `json:"name"`
	Total int    `json:"total"`
}

func (b *TagQueryBuilder) GetTagTypeCount() (int64, []TagTypeCount, error) {
	query := database.Instance
	query = ApplyFilters(b, query)
	var count int64 = 0
	rows, err := query.Model(&model.Tag{}).Select(
		`type as name,count(*) as total`,
	).Group(
		`type`,
	).Limit(b.getLimit()).Offset(b.getOffset()).Count(&count).Rows()
	if err == sql.ErrNoRows {
		return 0, nil, nil
	}
	result := make([]TagTypeCount, 0)
	for rows.Next() {
		var tagCount TagTypeCount
		err = database.Instance.ScanRows(rows, &tagCount)
		if err != nil {
			return 0, nil, err
		}
		result = append(result, tagCount)
	}
	return count, result, err
}

func TagAdd(fromTagId uint, toTargetId uint) error {
	var fromTag, toTag model.Tag
	err := database.Instance.First(&fromTag, fromTagId).Error
	if err != nil {
		return err
	}
	err = database.Instance.First(&toTag, toTargetId).Error
	if err != nil {
		return err
	}
	var deltaBooks []model.Book
	err = database.Instance.Model(&fromTag).Association("Books").Find(&deltaBooks)
	if err != nil {
		return err
	}
	err = database.Instance.Model(&toTag).Association("Books").Append(deltaBooks)
	if err != nil {
		return err
	}
	return nil
}

type MatchRule struct {
	Scope  string `json:"scope"`
	Value1 string `json:"value1"`
	Value2 string `json:"value2"`
	Value3 string `json:"value3"`
	Action string `json:"action"`
}

type RawTag struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Source string `json:"source"`
}

var AiTagMapping = map[string]string{
	"ART": "artist",
	"SOC": "societies",
	"SER": "series",
	"THE": "theme",
	"TRA": "translator",
	"TIT": "name",
	"TPE": "type",
	"LNG": "lang",
	"MAG": "magazine",
	"CHP": "chapter",
	"CHN": "chapter_number",
}

// LLM批量处理的默认最大数量限制
const DefaultLLMBatchMaxSize = 20

// getLLMBatchMaxSize 获取LLM批量处理的最大数量限制
// 可以通过配置文件或环境变量来配置，如果没有配置则使用默认值
func getLLMBatchMaxSize() int {
	// 这里可以添加从配置文件读取的逻辑
	// 目前先使用默认值
	return DefaultLLMBatchMaxSize
}

// LLMBatchProgress 批量处理进度信息
type LLMBatchProgress struct {
	CurrentBatch   int `json:"current_batch"`
	TotalBatches   int `json:"total_batches"`
	ProcessedCount int `json:"processed_count"`
	TotalCount     int `json:"total_count"`
}

func MatchTag(raw string, pattern string, useLLM bool, customPrompt ...string) []*RawTag {
	return MatchTagWithHistory(raw, pattern, useLLM, false, customPrompt...)
}

// MatchTagWithHistory 支持历史记录的标签匹配函数
func MatchTagWithHistory(raw string, pattern string, useLLM bool, forceReprocess bool, customPrompt ...string) []*RawTag {
	var result *utils.MatchTagResult
	if len(pattern) > 0 {
		rex := regexp.MustCompile(pattern)
		result = utils.MatchAllResult(rex, raw)
	} else {
		result = utils.MatchName(raw)
	}
	tags := make([]*RawTag, 0)
	if result != nil {
		if len(result.Name) > 0 {
			tags = append(tags, &RawTag{Name: result.Name, Type: "name", Source: "pattern", ID: xid.New().String()})
		}
		if len(result.Artist) > 0 {
			tags = append(tags, &RawTag{Name: result.Artist, Type: "artist", Source: "pattern", ID: xid.New().String()})
		}
		if len(result.Series) > 0 {
			tags = append(tags, &RawTag{Name: result.Series, Type: "series", Source: "pattern", ID: xid.New().String()})
		}
		if len(result.Theme) > 0 {
			tags = append(tags, &RawTag{Name: result.Theme, Type: "theme", Source: "pattern", ID: xid.New().String()})
		}
		if len(result.Translator) > 0 {
			tags = append(tags, &RawTag{Name: result.Translator, Type: "translator", Source: "pattern", ID: xid.New().String()})
		}
	}

	rawTagStrings := utils.MatchTagTextFromText(raw)
	var queryTags []model.Tag
	if len(rawTagStrings) > 0 {
		query := database.Instance.Model(&model.Tag{})
		for idx, tagString := range rawTagStrings {
			if len(tagString) == 0 {
				continue
			}
			if idx == 0 {
				query = query.Where("name like ?", fmt.Sprintf("%%%s%%", tagString))
				continue
			}
			query = query.Or("name like ?", fmt.Sprintf("%%%s%%", tagString))
		}
		query.Find(&queryTags)
		if queryTags != nil {
			for _, tagRecord := range queryTags {
				logrus.Info(tagRecord.Name)
				appendFlag := true
				for idx := range tags {
					exist := tags[idx]
					if exist.Type == tagRecord.Type && exist.Name == tagRecord.Name {
						appendFlag = false
						break
					}
				}
				if appendFlag {
					tags = append(tags, &RawTag{Name: tagRecord.Name, Type: tagRecord.Type, Source: "database", ID: xid.New().String()})
				}
			}

		}
	}
	for _, tagString := range rawTagStrings {
		tags = append(tags, &RawTag{Name: tagString, Type: "", Source: "raw", ID: xid.New().String()})
	}
	if AiTaggerInstance != nil {
		response, err := AiTaggerInstance.predict(raw)
		if err != nil {
			logrus.Info(err)
		}
		if response != nil {
			for _, tag := range response {
				text := tag.Text
				text = strings.TrimSpace(text)
				label := tag.Label
				// if label not in mapping
				if _, ok := AiTagMapping[label]; !ok {
					continue
				}
				tags = append(tags, &RawTag{Name: text, Type: AiTagMapping[label], Source: "ai", ID: xid.New().String()})
			}
		}
	}

	// 添加LLM分析（支持历史记录）
	if useLLM {
		var prompt string
		if len(customPrompt) > 0 && customPrompt[0] != "" {
			prompt = customPrompt[0]
		}
		llmTags := extractTagsWithLLMCustomPromptWithHistory(raw, prompt, forceReprocess)
		tags = append(tags, llmTags...)
	}
	return tags
}

// LLMTagResponse LLM返回的标签结构
type LLMTagResponse struct {
	Tags []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"tags"`
}

// extractTagsWithLLM 使用LLM分析文本提取标签
func extractTagsWithLLM(rawText string) []*RawTag {
	var result []*RawTag

	// 检查LLM插件是否可用
	if plugin.LLM == nil {
		return result
	}

	client, err := plugin.LLM.GetClient()
	if err != nil {
		logrus.WithError(err).Warn("无法获取LLM客户端")
		return result
	}

	// 构建LLM分析prompt
	prompt := buildTagAnalysisPrompt(rawText, "")

	ctx := context.Background()
	response, err := client.GenerateText(ctx, prompt)
	if err != nil {
		logrus.WithError(err).Warn("LLM标签分析失败")
		return result
	}

	// 解析LLM响应
	llmTags := parseLLMTagResponse(response)
	for _, tag := range llmTags {
		result = append(result, &RawTag{
			ID:     xid.New().String(),
			Name:   tag.Name,
			Type:   tag.Type,
			Source: "llm",
		})
	}

	return result
}

// extractTagsWithLLMCustomPrompt 使用LLM分析文本提取标签（支持自定义prompt）
func extractTagsWithLLMCustomPrompt(rawText string, customPrompt string) []*RawTag {
	var result []*RawTag

	// 检查LLM插件是否可用
	if plugin.LLM == nil {
		return result
	}

	client, err := plugin.LLM.GetClient()
	if err != nil {
		logrus.WithError(err).Warn("无法获取LLM客户端")
		return result
	}

	// 构建LLM分析prompt
	prompt := buildTagAnalysisPrompt(rawText, customPrompt)

	ctx := context.Background()
	response, err := client.GenerateText(ctx, prompt)
	if err != nil {
		logrus.WithError(err).Warn("LLM标签分析失败")
		return result
	}

	// 解析LLM响应
	llmTags := parseLLMTagResponse(response)
	for _, tag := range llmTags {
		result = append(result, &RawTag{
			ID:     xid.New().String(),
			Name:   tag.Name,
			Type:   tag.Type,
			Source: "llm",
		})
	}

	return result
}

// extractTagsWithLLMCustomPromptWithHistory 使用LLM分析文本提取标签（支持历史记录和自定义prompt）
func extractTagsWithLLMCustomPromptWithHistory(rawText string, customPrompt string, forceReprocess bool) []*RawTag {
	// 如果不强制重新处理，先检查历史记录
	if !forceReprocess {
		history, err := FindLLMTagHistory(rawText, customPrompt)
		if err != nil {
			logrus.WithError(err).Warn("查询LLM历史记录失败")
		} else if history != nil {
			// 找到历史记录，更新使用计数并返回结果
			err = UpdateLLMTagHistoryUsage(history.ID)
			if err != nil {
				logrus.WithError(err).Warn("更新LLM历史记录使用次数失败")
			}

			logrus.Infof("使用LLM历史记录，文本: %s", rawText)
			return convertLLMTagResultsToRawTags(history.Results)
		}
	}

	// 没有历史记录或强制重新处理，进行新的LLM分析
	startTime := time.Now()
	llmResults := extractTagsWithLLMCustomPromptOriginal(rawText, customPrompt)
	processingTime := time.Since(startTime).Milliseconds()

	// 转换结果格式用于保存
	historyResults := make([]model.LLMTagResult, 0, len(llmResults))
	for _, tag := range llmResults {
		historyResults = append(historyResults, model.LLMTagResult{
			Name: tag.Name,
			Type: tag.Type,
		})
	}

	// 保存到历史记录
	err := SaveLLMTagHistory(rawText, customPrompt, historyResults, processingTime, true, "")
	if err != nil {
		logrus.WithError(err).Warn("保存LLM历史记录失败")
	} else {
		logrus.Infof("保存LLM识别结果到历史记录，文本: %s, 处理时间: %dms", rawText, processingTime)
	}

	return llmResults
}

// batchExtractTagsWithLLMHistory 批量处理LLM标签提取，优化历史记录查询
func batchExtractTagsWithLLMHistory(rawTexts []string, customPrompt string) [][]*RawTag {
	results := make([][]*RawTag, len(rawTexts))

	// 初始化所有结果为空
	for i := range results {
		results[i] = []*RawTag{}
	}

	if len(rawTexts) == 0 {
		return results
	}

	// 准备批量查询数据
	textsWithPrompts := make([]struct{ Text, Prompt string }, len(rawTexts))
	for i, text := range rawTexts {
		textsWithPrompts[i] = struct{ Text, Prompt string }{
			Text:   text,
			Prompt: customPrompt,
		}
	}

	// 批量查询历史记录
	historyMap, err := FindLLMTagHistoryBatch(textsWithPrompts)
	if err != nil {
		logrus.WithError(err).Warn("批量查询LLM历史记录失败")
		// 如果批量查询失败，回退到逐个处理
		for i, text := range rawTexts {
			results[i] = extractTagsWithLLMCustomPromptWithHistory(text, customPrompt, false)
		}
		return results
	}

	// 分类：有历史记录的和需要新处理的
	var needProcessTexts []string
	var needProcessIndices []int
	var foundHistoryIds []uint

	for i, text := range rawTexts {
		textHash := generateHash(text)
		promptHash := generateHash(customPrompt)
		key := textHash + "_" + promptHash

		if history, found := historyMap[key]; found && history != nil {
			// 找到历史记录，直接使用
			results[i] = convertLLMTagResultsToRawTags(history.Results)
			foundHistoryIds = append(foundHistoryIds, history.ID)
			logrus.Debugf("使用LLM历史记录，文本: %s", text)
		} else {
			// 没有历史记录，需要新处理
			needProcessTexts = append(needProcessTexts, text)
			needProcessIndices = append(needProcessIndices, i)
		}
	}

	// 批量更新找到的历史记录使用计数
	if len(foundHistoryIds) > 0 {
		err = UpdateLLMTagHistoryUsageBatch(foundHistoryIds)
		if err != nil {
			logrus.WithError(err).Warn("批量更新LLM历史记录使用次数失败")
		}
	}

	// 处理需要新计算的文本
	if len(needProcessTexts) > 0 {
		startTime := time.Now()

		// 对需要处理的文本进行LLM分析
		newResults := make([]model.LLMTagResult, 0)
		for idx, text := range needProcessTexts {
			originalIndex := needProcessIndices[idx]
			llmTags := extractTagsWithLLMCustomPromptOriginal(text, customPrompt)
			results[originalIndex] = llmTags

			// 准备保存到历史记录的数据
			for _, tag := range llmTags {
				newResults = append(newResults, model.LLMTagResult{
					Name: tag.Name,
					Type: tag.Type,
				})
			}

			// 逐个保存历史记录（这里可以进一步优化为批量保存）
			historyResults := make([]model.LLMTagResult, 0, len(llmTags))
			for _, tag := range llmTags {
				historyResults = append(historyResults, model.LLMTagResult{
					Name: tag.Name,
					Type: tag.Type,
				})
			}

			processingTime := time.Since(startTime).Milliseconds()
			err = SaveLLMTagHistory(text, customPrompt, historyResults, processingTime, true, "")
			if err != nil {
				logrus.WithError(err).Warnf("保存LLM历史记录失败，文本: %s", text)
			}
		}

		logrus.Infof("新处理了 %d 个文本的LLM分析，复用了 %d 个历史记录",
			len(needProcessTexts), len(foundHistoryIds))
	} else {
		logrus.Infof("全部 %d 个文本都使用了LLM历史记录", len(rawTexts))
	}

	return results
}

// extractTagsWithLLMCustomPromptOriginal 原始的LLM标签提取函数（不使用历史记录）
func extractTagsWithLLMCustomPromptOriginal(rawText string, customPrompt string) []*RawTag {
	var result []*RawTag

	// 检查LLM插件是否可用
	if plugin.LLM == nil {
		return result
	}

	client, err := plugin.LLM.GetClient()
	if err != nil {
		logrus.WithError(err).Warn("无法获取LLM客户端")
		return result
	}

	// 构建LLM分析prompt
	prompt := buildTagAnalysisPrompt(rawText, customPrompt)

	ctx := context.Background()
	response, err := client.GenerateText(ctx, prompt)
	if err != nil {
		logrus.WithError(err).Warn("LLM标签分析失败")
		return result
	}

	// 解析LLM响应
	llmTags := parseLLMTagResponse(response)
	for _, tag := range llmTags {
		result = append(result, &RawTag{
			ID:     xid.New().String(),
			Name:   tag.Name,
			Type:   tag.Type,
			Source: "llm",
		})
	}

	return result
}

// extractTagsWithLLMBatch 使用LLM批量分析多个文本提取标签
func extractTagsWithLLMBatch(rawTexts []string, customPrompt string) [][]*RawTag {
	results := make([][]*RawTag, len(rawTexts))

	// 初始化所有结果为空
	for i := range results {
		results[i] = []*RawTag{}
	}

	// 检查LLM插件是否可用
	if plugin.LLM == nil {
		return results
	}

	client, err := plugin.LLM.GetClient()
	if err != nil {
		logrus.WithError(err).Warn("无法获取LLM客户端")
		return results
	}

	// 构建批量LLM分析prompt
	prompt := buildBatchTagAnalysisPrompt(rawTexts, customPrompt)

	ctx := context.Background()
	response, err := client.GenerateText(ctx, prompt)
	if err != nil {
		logrus.WithError(err).Warn("LLM批量标签分析失败")
		return results
	}

	// 解析LLM批量响应
	batchResults := parseLLMBatchTagResponse(response, len(rawTexts))

	// 转换为RawTag格式
	for i, textResults := range batchResults {
		if i < len(results) {
			for _, tag := range textResults {
				if len(strings.TrimSpace(tag.Name)) > 0 {
					results[i] = append(results[i], &RawTag{
						ID:     xid.New().String(),
						Name:   strings.TrimSpace(tag.Name),
						Type:   strings.TrimSpace(tag.Type),
						Source: "llm",
					})
				}
			}
		}
	}

	return results
}

// buildTagAnalysisPrompt 构建标签分析prompt
func buildTagAnalysisPrompt(rawText string, customPrompt string) string {
	// 如果提供了自定义prompt，使用自定义的
	if customPrompt != "" {
		// 对于自定义prompt，使用简单的字符串替换，保持向后兼容
		// 先尝试新格式 {{content}}，如果没有则尝试旧格式 %s
		if strings.Contains(customPrompt, "{{content}}") {
			return strings.ReplaceAll(customPrompt, "{{content}}", rawText)
		} else {
			return fmt.Sprintf(customPrompt, rawText)
		}
	}

	// 直接使用写死的prompt
	promptTemplate := getDefaultTagPromptTemplate()
	return fmt.Sprintf(promptTemplate, rawText)
}

// buildBatchTagAnalysisPrompt 构建批量标签分析prompt
func buildBatchTagAnalysisPrompt(rawTexts []string, customPrompt string) string {
	// 如果提供了自定义prompt，使用自定义的
	if customPrompt != "" {
		// 构建文本列表
		var textList strings.Builder
		for i, text := range rawTexts {
			textList.WriteString(fmt.Sprintf("文本%d: \"%s\"\n", i+1, text))
		}

		// 对于自定义prompt，使用简单的字符串替换
		if strings.Contains(customPrompt, "{{content}}") {
			return strings.ReplaceAll(customPrompt, "{{content}}", textList.String())
		} else {
			return fmt.Sprintf(customPrompt, textList.String())
		}
	}

	// 使用批量处理的默认prompt模板
	promptTemplate := getBatchTagPromptTemplate()

	// 构建文本列表
	var textList strings.Builder
	for i, text := range rawTexts {
		textList.WriteString(fmt.Sprintf("文本%d: \"%s\"\n", i+1, text))
	}

	return fmt.Sprintf(promptTemplate, textList.String())
}

// getBatchTagPromptTemplate 获取批量处理的默认prompt模板
func getBatchTagPromptTemplate() string {
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

// getDefaultTagPromptTemplate 获取默认的prompt模板
func getDefaultTagPromptTemplate() string {
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

// parseLLMTagResponse 解析LLM返回的标签响应
func parseLLMTagResponse(response string) []struct{ Name, Type string } {
	var result []struct{ Name, Type string }

	// 尝试提取JSON部分
	jsonStr := extractJSONFromResponse(response)
	if jsonStr == "" {
		logrus.Warn("无法从LLM响应中提取JSON")
		return result
	}

	var llmResponse LLMTagResponse
	err := json.Unmarshal([]byte(jsonStr), &llmResponse)
	if err != nil {
		logrus.WithError(err).Warn("解析LLM标签响应失败")
		return result
	}

	for _, tag := range llmResponse.Tags {
		if len(strings.TrimSpace(tag.Name)) > 0 {
			result = append(result, struct{ Name, Type string }{
				Name: strings.TrimSpace(tag.Name),
				Type: strings.TrimSpace(tag.Type),
			})
		}
	}

	return result
}

// extractJSONFromResponse 从响应中提取JSON部分
func extractJSONFromResponse(response string) string {
	// 查找JSON开始和结束位置
	start := strings.Index(response, "{")
	if start == -1 {
		return ""
	}

	// 从后往前查找最后一个}
	end := strings.LastIndex(response, "}")
	if end == -1 || end <= start {
		return ""
	}

	return response[start : end+1]
}

// LLMBatchTagResponse LLM批量返回的标签结构
type LLMBatchTagResponse struct {
	Results []struct {
		TextIndex int `json:"text_index"`
		Tags      []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"tags"`
	} `json:"results"`
}

// parseLLMBatchTagResponse 解析LLM批量返回的标签响应
func parseLLMBatchTagResponse(response string, expectedCount int) [][]struct{ Name, Type string } {
	results := make([][]struct{ Name, Type string }, expectedCount)

	// 初始化所有结果为空
	for i := range results {
		results[i] = []struct{ Name, Type string }{}
	}

	// 尝试提取JSON部分
	jsonStr := extractJSONFromResponse(response)
	if jsonStr == "" {
		logrus.Warn("无法从LLM批量响应中提取JSON")
		return results
	}

	var batchResponse LLMBatchTagResponse
	err := json.Unmarshal([]byte(jsonStr), &batchResponse)
	if err != nil {
		logrus.WithError(err).Warn("解析LLM批量标签响应失败")
		return results
	}

	// 按text_index分配结果
	for _, result := range batchResponse.Results {
		// text_index从1开始，数组索引从0开始
		arrayIndex := result.TextIndex - 1
		if arrayIndex >= 0 && arrayIndex < len(results) {
			for _, tag := range result.Tags {
				if len(strings.TrimSpace(tag.Name)) > 0 {
					results[arrayIndex] = append(results[arrayIndex], struct{ Name, Type string }{
						Name: strings.TrimSpace(tag.Name),
						Type: strings.TrimSpace(tag.Type),
					})
				}
			}
		}
	}

	return results
}

type BatchMatchResult struct {
	Result []*RawTag `json:"result"`
	Text   string    `json:"text"`
}

func BatchMatchTag(raws []string, pattern string, useLLM bool, customPrompt ...string) []*BatchMatchResult {
	return BatchMatchTagWithHistory(raws, pattern, useLLM, false, customPrompt...)
}

// BatchMatchTagWithHistory 支持历史记录的批量标签匹配函数
func BatchMatchTagWithHistory(raws []string, pattern string, useLLM bool, forceReprocess bool, customPrompt ...string) []*BatchMatchResult {
	var results = make([]*BatchMatchResult, 0)

	// 初始化所有结果项
	for _, raw := range raws {
		resultItem := &BatchMatchResult{
			Text:   raw,
			Result: []*RawTag{},
		}
		results = append(results, resultItem)
	}

	// AI标签器分析（仍然需要逐个处理）
	for i, raw := range raws {
		if AiTaggerInstance != nil {
			response, err := AiTaggerInstance.predict(raw)
			if err != nil {
				logrus.Info(err)
			}
			if response != nil {
				for _, tag := range response {
					text := tag.Text
					text = strings.TrimSpace(text)
					label := tag.Label
					// if label not in mapping
					if _, ok := AiTagMapping[label]; !ok {
						continue
					}
					results[i].Result = append(results[i].Result, &RawTag{Name: text, Type: AiTagMapping[label], Source: "ai", ID: xid.New().String()})
				}
			}
		}
	}

	// LLM分析（支持历史记录，批量优化）
	if useLLM && len(raws) > 0 {
		var prompt string
		if len(customPrompt) > 0 && customPrompt[0] != "" {
			prompt = customPrompt[0]
		}

		if forceReprocess {
			// 强制重新处理时，逐个处理
			for i, raw := range raws {
				llmTags := extractTagsWithLLMCustomPromptWithHistory(raw, prompt, true)
				results[i].Result = append(results[i].Result, llmTags...)
			}
		} else {
			// 使用批量查询优化性能
			batchLLMTags := batchExtractTagsWithLLMHistory(raws, prompt)
			for i, llmTags := range batchLLMTags {
				if i < len(results) {
					results[i].Result = append(results[i].Result, llmTags...)
				}
			}
		}

		logrus.Infof("完成 %d 个项目的LLM处理（支持历史记录）", len(raws))
	}

	return results
}

func ClearEmptyTag() error {
	var tags []model.Tag
	database.Instance.Find(&tags)
	for _, tag := range tags {
		ass := database.Instance.Model(&tag).Association("Books")
		if ass.Count() == 0 {
			err := database.Instance.Unscoped().Delete(&tag).Error
			if err != nil {
				logrus.Error(err)
			}
		}
	}
	return nil
}
