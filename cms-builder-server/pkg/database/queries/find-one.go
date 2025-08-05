package database

import (
	"context"

	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
)

func FindOne(ctx context.Context, log *loggerTypes.Logger, db *dbTypes.DatabaseConnection, entity interface{}, filters map[string]interface{}, preload []string) error {

	// Log the filters for debugging
	log.Info().
		Interface("filters", filters).
		Msg("Executing query with filters")

	// Build the query
	query := db.DB.Model(entity).WithContext(ctx)

	// Preload associations
	for _, association := range preload {
		query = query.Preload(association)
	}

	// Apply filters
	for key, value := range filters {
		query = query.Where(key, value)
	}

	// Execute the query and populate the entity
	if err := query.First(entity).Error; err != nil {
		log.Error().
			Err(err).
			Interface("filters", filters).
			Msg("Failed to execute query")
		return err
	}

	return nil
}
