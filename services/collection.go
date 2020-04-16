package services

import (
	"database/sql"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"github.com/jinzhu/gorm"
)

type CollectionQueryBuilder struct {
	IdQueryFilter
	NameQueryFilter
	OrderQueryFilter
	OwnerQueryFilter
	DefaultPageFilter
	UsersQueryFilter
	UsersAndOwnerQueryFilter
	NameSearchQueryFilter
}

func (b *CollectionQueryBuilder) ReadModels() (int, interface{}, error) {
	query := database.DB
	query = ApplyFilters(b, query)
	var collections []model.Collection
	var count = 0
	err := query.Limit(b.getLimit()).Offset(b.getOffset()).Find(&collections).Offset(-1).Count(&count).Error
	if err == sql.ErrNoRows {
		return 0, query, nil
	}
	return count, collections, err
}

type UsersAndOwnerQueryFilter struct {
	UsersAndOwnerIds []interface{}
}

func (b *UsersAndOwnerQueryFilter) SetUsersAndOwnerQueryFilter(ids ...interface{}) {
	for _, id := range ids {
		if len((id).(string)) > 0 {
			b.UsersAndOwnerIds = append(b.UsersAndOwnerIds, id)
		}
	}
}
func (b UsersAndOwnerQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if b.UsersAndOwnerIds != nil && len(b.UsersAndOwnerIds) > 0 {
		return db.Joins("inner join collection_users on collection_users.collection_id = collections.id").Where(
			"collection_users.user_id in (?)", b.UsersAndOwnerIds).Or("collections.owner in (?)", b.UsersAndOwnerIds)
	}
	return db
}

type UsersQueryFilter struct {
	Users []interface{}
}

func (b *UsersQueryFilter) SetUsersQueryFilter(users ...interface{}) {
	for _, user := range users {
		if len((user).(string)) > 0 {
			b.Users = append(b.Users, user)
		}
	}
}
func (b UsersQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if b.Users != nil && len(b.Users) > 0 {
		return db.Joins("inner join collection_users on collection_users.collection_id = collections.id").Where("collection_users.user_id in (?)", b.Users)
	}
	return db
}

type OwnerQueryFilter struct {
	Owners []interface{}
}

func (b *OwnerQueryFilter) SetOwnerQueryFilter(owners ...interface{}) {
	for _, owner := range owners {
		if len((owner).(string)) > 0 {
			b.Owners = append(b.Owners, owner)
		}
	}
}

func (b OwnerQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if b.Owners != nil && len(b.Owners) > 0 {
		return db.Where("owner in (?)", b.Owners)
	}
	return db
}

func AddBooksToCollection(collectionId uint, bookIds ...int) error {
	books := make([]model.Book, 0)
	for _, bookId := range bookIds {
		books = append(books, model.Book{Model: gorm.Model{ID: uint(bookId)}})
	}
	err := database.DB.Model(&model.Collection{Model: gorm.Model{ID: collectionId}}).Association("Books").Append(books).Error
	return err
}

func RemoveBooksFromCollection(collectionId uint, bookIds ...int) error {
	books := make([]model.Book, 0)
	for _, bookId := range bookIds {
		books = append(books, model.Book{Model: gorm.Model{ID: uint(bookId)}})
	}
	err := database.DB.Model(&model.Collection{Model: gorm.Model{ID: collectionId}}).Association("Books").Delete(books).Error
	return err
}
func AddUsersToCollection(collectionId uint, userIds ...int) error {
	users := make([]model.User, 0)
	for _, bookId := range userIds {
		users = append(users, model.User{Model: gorm.Model{ID: uint(bookId)}})
	}
	err := database.DB.Model(&model.Collection{Model: gorm.Model{ID: collectionId}}).Association("Users").Append(users).Error
	return err
}

func RemoveUsersFromCollection(collectionId uint, userIds ...int) error {
	users := make([]model.User, 0)
	for _, bookId := range userIds {
		users = append(users, model.User{Model: gorm.Model{ID: uint(bookId)}})
	}
	err := database.DB.Model(&model.Collection{Model: gorm.Model{ID: collectionId}}).Association("Users").Delete(users).Error
	return err
}

func GetCollectionById(collectionId uint) (error, *model.Collection) {
	collection := model.Collection{Model: gorm.Model{ID: collectionId}}
	err := database.DB.First(&collection).Error
	return err, &collection
}


