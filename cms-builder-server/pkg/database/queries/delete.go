package queries

import (
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"gorm.io/gorm"
)

func Delete(db *Database, entity interface{}, user *models.User, requestId string) *gorm.DB {
	result := db.DB.Delete(entity)
	if result.Error == nil {
		historyEntry, err := NewDatabaseLogEntry(DeleteCRUDAction, user, entity, "", requestId)
		if err != nil {
			return nil
		}
		_ = db.DB.Create(historyEntry)
	}

	return result
}
