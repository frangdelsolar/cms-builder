package queries

import (
	"context"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
)

func Create(ctx context.Context, log *logger.Logger, db *database.Database, instance interface{}, user *models.User, requestId string) error {
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
	historyEntry, err := NewDatabaseLogEntry(CreateCRUDAction, user, instance, "", requestId)
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
