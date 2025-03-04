package queries

import (
	"context"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
)

func FindOne(ctx context.Context, log *logger.Logger, db *database.Database, entity interface{}, filters map[string]interface{}) error {

	// Log the filters for debugging
	log.Info().
		Interface("filters", filters).
		Msg("Executing query with filters")

	// Build the query
	query := db.DB.Model(entity).WithContext(ctx)
	for key, value := range filters {
		query = query.Where(key, value)
	}

	// Execute the query and populate the entity
	if err := query.Debug().First(entity).Error; err != nil {
		log.Error().
			Err(err).
			Interface("filters", filters).
			Msg("Failed to execute query")
		return err
	}

	return nil
}
