package queries

import (
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"gorm.io/gorm"
)

type Pagination struct {
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

func FindMany(db *database.Database, entitySlice interface{}, pagination *Pagination, order string, query string, args ...interface{}) *gorm.DB {
	if order == "" {
		order = "id desc"
	}

	if pagination == nil {
		return db.DB.Order(order).Where(query, args...).Find(entitySlice)
	}

	// Retrieve total number of records
	res := db.DB.Model(entitySlice).Where(query, args...).Count(&pagination.Total)
	if res.Error != nil {
		return res
	}

	// Apply pagination
	filtered := db.DB.Where(query, args...).Order(order)
	limit := pagination.Limit
	offset := (pagination.Page - 1) * pagination.Limit

	return filtered.Limit(limit).Offset(offset).Find(entitySlice)
}
