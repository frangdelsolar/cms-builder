package database

import (
	"context"

	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	dbPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
)

func Create(ctx context.Context, log *loggerTypes.Logger, db *dbTypes.DatabaseConnection, instance interface{}, user *authModels.User, requestId string) error {
	// Create the instance
	result := db.DB.WithContext(ctx).Create(instance)
	if result.Error != nil {
		log.Error().
			Err(result.Error).
			Str("requestId", requestId).
			Msg("Failed to create instance")
		return result.Error
	}

	// Log the create action
	historyEntry, err := dbPkg.NewDatabaseLogEntry(dbTypes.CreateCRUDAction, user, instance, "", requestId)
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

	return nil
}
