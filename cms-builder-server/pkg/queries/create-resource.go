package queries

import (
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"gorm.io/gorm"
)

func Create(db *database.Database, instance interface{}, user *User, requestId string) *gorm.DB {
	result := db.DB.Create(instance)
	if result.Error == nil {
		historyEntry, err := NewLogHistoryEntry(CreateCRUDAction, user, instance, "", requestId)
		if err != nil {
			return nil
		}
		_ = db.DB.Create(historyEntry)
	}

	return result
}
