package services

import (
	"fmt"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/utils"
	"gorm.io/gorm"
	"os"
	"reflect"
)

//page filter
type PageFilter interface {
	SetPageFilter(page int, pageSize int)
	getOffset() int
	getLimit() int
}
type DefaultPageFilter struct {
	Page     int
	PageSize int
}

func (b *DefaultPageFilter) SetPageFilter(page int, pageSize int) {
	b.Page = page
	b.PageSize = pageSize
}

func (b *DefaultPageFilter) getOffset() int {
	return b.PageSize * (b.Page - 1)
}

func (b *DefaultPageFilter) getLimit() int {
	if b.PageSize == 0 {
		return 1
	} else {
		return b.PageSize
	}
}

//order filter
type OrderFilter interface {
	SetOrderFilter(order string)
}

//read models from database
type ModelsReader interface {
	ReadModels() (int64, interface{}, error)
}

//delete model by model's id
func DeleteById(model interface{}) error {
	return database.DB.Delete(model).Error
}

//update models with allow fields
func UpdateModel(model interface{}, allowFields ...string) error {
	updateMap := make(map[string]interface{})
	r := reflect.ValueOf(model)
	for _, propertyName := range allowFields {
		value := reflect.Indirect(r).FieldByName(propertyName)

		updateMap[propertyName] = value.Interface()
	}
	err := database.DB.Model(model).Updates(updateMap).Error
	return err
}

//get model by id
func GetModelById(model interface{}, id int) error {
	err := database.DB.First(model, id).Error
	if err != nil {
		return err
	}
	return nil
}

//combine query filter and gorm db query
type GORMFilter interface {
	ApplyQuery(db *gorm.DB) *gorm.DB
}

//apply query builder inner filters that implement GORMFilter
func ApplyFilters(queryBuilder interface{}, db *gorm.DB) *gorm.DB {

	query := db
	builderRef := reflect.ValueOf(queryBuilder).Elem()
	builderFieldCount := builderRef.NumField()
	//walk through all field of builder
	for fieldIdx := 0; fieldIdx < builderFieldCount; fieldIdx++ {
		field := builderRef.Field(fieldIdx).Interface()
		gormFilter, isImplement := field.(GORMFilter)
		if isImplement {
			query = gormFilter.ApplyQuery(query)
		}
	}
	return query
}

//Ids filter
type IdQueryFilter struct {
	Ids []interface{}
}

func (f IdQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.Ids != nil && len(f.Ids) != 0 {
		return db.Where("id in (?)", f.Ids)
	}
	return db
}

func (f *IdQueryFilter) InId(ids ...interface{}) {
	for _, id := range ids {
		if !utils.IsZeroVal(id) {
			f.Ids = append(f.Ids, id)
		}

	}

}

//order filter
type OrderQueryFilter struct {
	Order string
}

func (f OrderQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if len(f.Order) != 0 {
		return db.Order(f.Order)
	}
	return db
}

func (f *OrderQueryFilter) SetOrderFilter(order string) {
	f.Order = order
}

type NameQueryFilter struct {
	Names []interface{}
}

func (f *NameQueryFilter) SetNameFilter(names ...interface{}) {
	for _, name := range names {
		if len(name.(string)) != 0 {
			f.Names = append(f.Names, name)
		}
	}

}
func (f NameQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.Names != nil && len(f.Names) != 0 {
		return db.Where("name in (?)", f.Names)
	}
	return db
}

type NameSearchQueryFilter struct {
	nameSearch interface{}
}

func (f NameSearchQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.nameSearch != nil && len(f.nameSearch.(string)) != 0 {
		return db.Where("name like ?", fmt.Sprintf("%%%s%%", f.nameSearch))
	}
	return db
}

func (f *NameSearchQueryFilter) SetNameSearchQueryFilter(nameSearch interface{}) {
	if len(nameSearch.(string)) > 0 {
		f.nameSearch = nameSearch
	}
}
func CreateModels(models []interface{}) error {
	var err error
	for _, modelToCreate := range models {
		err = database.DB.Create(modelToCreate).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateModels(updateModel interface{}, updateModels []interface{}, allowFields ...string) error {
	var err error
	for _, updateMapInterface := range updateModels {
		rawUpdateMap := updateMapInterface.(map[string]interface{})
		updateMap := make(map[string]interface{}, 0)
		for _, key := range allowFields {
			updateMap[key] = rawUpdateMap[key]
		}
		err := database.DB.Model(updateModel).Where("id = ?", rawUpdateMap["id"]).Updates(updateMap).Error
		if err != nil {
			return err
		}
	}
	return err
}

func DeleteModels(deleteModel interface{}, ids ...int) error {
	var err error

	err = database.DB.Where("id in (?)", ids).Delete(deleteModel).Error
	if err != nil {
		return err
	}

	return nil
}

func CreateModel(modelToCreate interface{}) error {
	result := database.DB.Create(modelToCreate)
	err := result.Error
	if err != nil {
		return err
	}
	return nil
}

func RemoveTagFromBook(bookId uint, tagId uint) error {
	return database.DB.Model(&model.Book{Model: gorm.Model{ID: bookId}}).Association("Tags").Delete(model.Tag{Model: gorm.Model{ID: tagId}})
}

func DeleteBookFile(bookId uint) (err error) {
	path := utils.GetBookStorePath(bookId)
	err = os.RemoveAll(path)
	return
}

type ModelUpdater interface {
	SetId(id interface{})
	Update(valueMapping map[string]interface{}) error
}
