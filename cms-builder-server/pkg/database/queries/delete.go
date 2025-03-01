package queries

import (
	"reflect"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"gorm.io/gorm"
)

func Delete(db *Database, entity interface{}, user *models.User, requestId string) *gorm.DB {
	// Use reflection to determine if the entity is a slice or array
	val := reflect.ValueOf(entity)
	isSlice := val.Kind() == reflect.Slice || val.Kind() == reflect.Array

	// Delete the entity or slice of entities
	result := db.DB.Delete(entity)
	if result.Error != nil {
		return result
	}

	if isSlice {
		// Handle slice of entities
		for i := 0; i < val.Len(); i++ {
			element := val.Index(i).Interface()

			// Log the deletion action for each element
			historyEntry, err := NewDatabaseLogEntry(DeleteCRUDAction, user, element, "", requestId)
			if err != nil {
				return nil
			}
			_ = db.DB.Create(historyEntry)
		}
	} else {
		// Handle single entity
		historyEntry, err := NewDatabaseLogEntry(DeleteCRUDAction, user, entity, "", requestId)
		if err != nil {
			return nil
		}
		_ = db.DB.Create(historyEntry)
	}

	return result
}
