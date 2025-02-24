package queries

import (
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"gorm.io/gorm"
)

func FindMany(db *database.Database, entitySlice interface{}, query string, pagination *Pagination, order string) *gorm.DB {
	if order == "" {
		order = "id desc"
	}

	if pagination == nil {
		return db.DB.Order(order).Where(query).Find(entitySlice)
	}

	// Retrieve total number of records
	res := db.DB.Model(entitySlice).Where(query).Count(&pagination.Total)
	if res.Error != nil {
		return res
	}

	// Apply pagination
	filtered := db.DB.Where(query).Order(order)
	limit := pagination.Limit
	offset := (pagination.Page - 1) * pagination.Limit

	return filtered.Limit(limit).Offset(offset).Find(entitySlice)
}
