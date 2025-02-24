package queries

import (
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"gorm.io/gorm"
)

func FindById(db *database.Database, id string, entity interface{}, queryExtension string) *gorm.DB {
	q := "id = '" + id + "'"

	if queryExtension != "" {
		q += " AND " + queryExtension
	}

	return db.DB.Where(q).First(entity)
}
