package queries

import (
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"gorm.io/gorm"
)

func FindOne(db *database.Database, entity interface{}, query string, args ...interface{}) *gorm.DB {
	if query == "" {
		return db.DB.First(entity)
	}

	return db.DB.Where(query, args...).First(entity)
}
