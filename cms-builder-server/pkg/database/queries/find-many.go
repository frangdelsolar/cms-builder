package database

import (
	"context"

	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
)

func FindMany(ctx context.Context, log *loggerTypes.Logger, db *dbTypes.DatabaseConnection, entitySlice interface{}, pagination *dbTypes.Pagination, order string, filters map[string]interface{}, preload []string) error {
	if order == "" {
		order = "id desc"
	}

	// Build the query
	query := db.DB.Model(entitySlice).WithContext(ctx)

	// Preload associations
	for _, association := range preload {
		query = query.Preload(association)
	}

	// Apply filters
	for key, value := range filters {
		query = query.Where(key, value)
	}

	// Retrieve total number of records
	if pagination != nil {
		if err := query.Count(&pagination.Total).Error; err != nil {
			log.Error().
				Err(err).
				Interface("filters", filters).
				Msg("Failed to count records")
			return err
		}
	}

	// Apply pagination and ordering
	if pagination != nil {
		limit := pagination.Limit
		offset := (pagination.Page - 1) * pagination.Limit
		query = query.Limit(limit).Offset(offset)
	}
	query = query.Order(order)

	// Execute the query
	if err := query.Find(entitySlice).Error; err != nil {
		log.Error().
			Err(err).
			Interface("filters", filters).
			Msg("Failed to find records")
		return err
	}

	return nil
}
