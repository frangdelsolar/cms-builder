package queries

import (
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"

	"gorm.io/gorm"
)

func Create(db *Database, instance interface{}, user *models.User, requestId string) *gorm.DB {
	result := db.DB.Create(instance)
	if result.Error == nil {
		historyEntry, err := NewDatabaseLogEntry(CreateCRUDAction, user, instance, "", requestId)
		if err != nil {
			return nil
		}
		_ = db.DB.Create(historyEntry)
	}

	return result
}
