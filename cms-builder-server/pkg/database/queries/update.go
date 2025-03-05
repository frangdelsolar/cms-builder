package queries

import (
	"context"

	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
)

func Update(ctx context.Context, log *loggerTypes.Logger, db *dbTypes.DatabaseConnection, entity interface{}, user *models.User, differences interface{}, requestId string) error {
	// Update the entity
	result := db.DB.WithContext(ctx).Save(entity)
	if result.Error != nil {
		log.Error().
			Err(result.Error).
			Str("requestId", requestId).
			Msg("Failed to update entity")
		return result.Error
	}

	// Log the update action
	historyEntry, err := NewDatabaseLogEntry(dbTypes.UpdateCRUDAction, user, entity, differences, requestId)
	if err != nil {
		log.Error().
			Interface("differences", differences).
			Interface("entity", entity).
			Interface("user", user).
			Err(err).
			Str("requestId", requestId).
			Msg("Failed to create database log entry")
		return err
	}
	if err := db.DB.WithContext(ctx).Create(historyEntry).Error; err != nil {
		log.Error().
			Interface("differences", differences).
			Interface("entity", entity).
			Interface("user", user).
			Err(err).
			Str("requestId", requestId).
			Msg("Failed to save database log entry")
		return err
	}

	return nil
}
