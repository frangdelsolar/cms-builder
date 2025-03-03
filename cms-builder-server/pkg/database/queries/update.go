package queries

import (
	"context"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
)

func Update(ctx context.Context, log *logger.Logger, db *database.Database, entity interface{}, user *models.User, differences interface{}, requestId string) error {
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
	historyEntry, err := NewDatabaseLogEntry(UpdateCRUDAction, user, entity, differences, requestId)
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
