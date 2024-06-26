package services

import (
	"database/sql"
	"fmt"
	"github.com/projectxpolaris/youcomic/database"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/utils"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"regexp"
	"strings"
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
}

func MatchTag(raw string, pattern string) []*RawTag {
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
	return tags
}

type BatchMatchResult struct {
	Result []*RawTag `json:"result"`
	Text   string    `json:"text"`
}

func BatchMatchTag(raws []string, pattern string) []*BatchMatchResult {
	var results = make([]*BatchMatchResult, 0)
	if AiTaggerInstance != nil {
		for _, raw := range raws {
			resultItem := &BatchMatchResult{
				Text:   raw,
				Result: []*RawTag{},
			}
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
					resultItem.Result = append(resultItem.Result, &RawTag{Name: text, Type: AiTagMapping[label], Source: "ai", ID: xid.New().String()})
				}
			}
			results = append(results, resultItem)
		}
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
