package queries

import (
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"gorm.io/gorm"
)

func Update(db *Database, entity interface{}, user *models.User, differences interface{}, requestId string) *gorm.DB {
	result := db.DB.Save(entity)
	if result.Error == nil {
		historyEntry, err := NewDatabaseLogEntry(UpdateCRUDAction, user, entity, differences, requestId)
		if err != nil {
			return db.DB
		}
		_ = db.DB.Create(historyEntry)
	}

	return result
}
