package queries

import (
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"gorm.io/gorm"
)

func Update(db *database.Database, entity interface{}, user *User, differences interface{}, requestId string) *gorm.DB {

	result := db.DB.Save(entity)
	if result.Error == nil {
		historyEntry, err := NewLogHistoryEntry(UpdateCRUDAction, user, entity, differences, requestId)
		if err != nil {
			return db.DB
		}
		_ = db.DB.Create(historyEntry)
	}

	return result
}
