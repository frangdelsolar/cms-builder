package queries

import (
	"context"
	"reflect"

	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	dbPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
)

func Delete(ctx context.Context, log *loggerTypes.Logger, db *dbTypes.DatabaseConnection, entity interface{}, user *authModels.User, requestId string) error {
	// Use reflection to determine if the entity is a slice or array
	val := reflect.ValueOf(entity)
	isSlice := val.Kind() == reflect.Slice || val.Kind() == reflect.Array

	// Delete the entity or slice of entities
	result := db.DB.WithContext(ctx).Delete(entity)
	if result.Error != nil {
		log.Error().
			Err(result.Error).
			Str("requestId", requestId).
			Msg("Failed to delete entity")
		return result.Error
	}

	// Log the deletion action
	if isSlice {
		for i := 0; i < val.Len(); i++ {
			element := val.Index(i).Interface()
			historyEntry, err := dbPkg.NewDatabaseLogEntry(dbTypes.DeleteCRUDAction, user, element, "", requestId)
			if err != nil {
				log.Error().
					Err(err).
					Str("requestId", requestId).
					Msg("Failed to create database log entry")
				return err
			}
			if err := db.DB.WithContext(ctx).Create(historyEntry).Error; err != nil {
				log.Error().
					Err(err).
					Str("requestId", requestId).
					Msg("Failed to save database log entry")
				return err
			}
		}
	} else {
		historyEntry, err := dbPkg.NewDatabaseLogEntry(dbTypes.DeleteCRUDAction, user, entity, "", requestId)
		if err != nil {
			log.Error().
				Err(err).
				Str("requestId", requestId).
				Msg("Failed to create database log entry")
			return err
		}
		if err := db.DB.WithContext(ctx).Create(historyEntry).Error; err != nil {
			log.Error().
				Err(err).
				Str("requestId", requestId).
				Msg("Failed to save database log entry")
			return err
		}
	}

	return nil
}
