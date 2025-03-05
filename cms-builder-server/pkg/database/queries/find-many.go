package queries

import (
	"context"

	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
)

type Pagination struct {
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

func FindMany(ctx context.Context, log *loggerTypes.Logger, db *dbTypes.DatabaseConnection, entitySlice interface{}, pagination *Pagination, order string, filters map[string]interface{}) error {
	if order == "" {
		order = "id desc"
	}

	// Build the query
	query := db.DB.Model(entitySlice).WithContext(ctx)
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
