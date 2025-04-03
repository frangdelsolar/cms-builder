package database

import (
	"context"
	"errors"
	"reflect"

	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	dbPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	"gorm.io/gorm"
)

func HardDelete(ctx context.Context, log *loggerTypes.Logger, db *dbTypes.DatabaseConnection, entity interface{}, user *authModels.User, requestId string) error {

	log.Debug().Interface("entity", entity).Msg("Hard deleting entity")
	// Use reflection to determine if the entity is a slice or array
	val := reflect.ValueOf(entity)
	isSlice := val.Kind() == reflect.Slice || val.Kind() == reflect.Array

	// First get the current state of the entity for audit logging
	var currentStates []interface{}
	if isSlice {
		ids := make([]uint, val.Len())
		for i := 0; i < val.Len(); i++ {
			id := reflect.Indirect(val.Index(i)).FieldByName("ID").Uint()
			ids[i] = uint(id)
		}

		// Fetch current states of all entities
		modelType := reflect.TypeOf(entity).Elem()
		models := reflect.New(reflect.SliceOf(modelType)).Interface()
		if err := db.DB.WithContext(ctx).Find(models, ids).Error; err != nil {
			log.Error().Err(err).Msg("Failed to fetch current states for hard delete")
			return err
		}

		modelsVal := reflect.ValueOf(models).Elem()
		for i := 0; i < modelsVal.Len(); i++ {
			currentStates = append(currentStates, modelsVal.Index(i).Interface())
		}
	} else {
		// Fetch current state of single entity
		id := reflect.Indirect(val).FieldByName("ID").Uint()
		currentState := reflect.New(reflect.TypeOf(entity).Elem()).Interface()
		if err := db.DB.WithContext(ctx).First(currentState, id).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error().Err(err).Msg("Failed to fetch current state for hard delete")
				return err
			}
			// If record not found, use the passed entity as current state
			currentState = entity
		}
		currentStates = append(currentStates, currentState)
	}

	// Perform the hard delete using Unscoped()
	result := db.DB.WithContext(ctx).Unscoped().Delete(entity)
	if result.Error != nil {
		log.Error().
			Err(result.Error).
			Str("requestId", requestId).
			Msg("Failed to hard delete entity")
		return result.Error
	}

	// Log the hard deletion action(s)
	for _, currentState := range currentStates {
		historyEntry, err := dbPkg.NewDatabaseLogEntry(
			dbTypes.HardDeleteCRUDAction, // You may want to add this action type
			user,
			currentState,
			"",
			requestId,
		)
		if err != nil {
			log.Error().
				Err(err).
				Str("requestId", requestId).
				Msg("Failed to create database log entry for hard delete")
			return err
		}

		if err := db.DB.WithContext(ctx).Create(historyEntry).Error; err != nil {
			log.Error().
				Err(err).
				Str("requestId", requestId).
				Msg("Failed to save database log entry for hard delete")
			return err
		}
	}

	log.Info().
		Str("requestId", requestId).
		Int64("rowsAffected", result.RowsAffected).
		Msg("Successfully performed hard delete")

	return nil
}
