package services

import (
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
)

type PermissionQueryBuilder struct {
	NameQueryFilter
}

func (b *PermissionQueryBuilder) ReadModels() (int, interface{}, error) {
	query := database.DB
	query = ApplyFilters(b, query)
	var count = 0
	md := make([]model.Permission, 0)
	err := query.Find(&md).Offset(-1).Count(&count).Error
	return count, md, err
}

