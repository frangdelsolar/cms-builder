package utils

import (
	"context"
	"fmt"

	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	schModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/models"
	"github.com/google/uuid"
)

func GetOrCreateJobDefinition(db *dbTypes.DatabaseConnection, log *loggerTypes.Logger, schedulerUser *authModels.User, jdInput schModels.SchedulerJobDefinition) (*schModels.SchedulerJobDefinition, error) {

	// If there is a job definition with the same name, return it
	// Name must be unique
	var instance schModels.SchedulerJobDefinition

	filters := map[string]interface{}{
		"name":           jdInput.Name,
		"frequency_type": jdInput.FrequencyType,
		"at_time":        jdInput.AtTime,
		"cron_expr":      jdInput.CronExpr,
		"with_seconds":   jdInput.WithSeconds,
	}

	err := dbQueries.FindOne(context.Background(), log, db, &instance, filters)
	if err != nil {
		log.Error().Err(err).Interface("filters", filters).Msg("Failed to find job definition")
		// return nil, err
	}

	if instance.Name != "" {
		return &instance, nil
	}

	// If there is no job definition with the same name, create it
	instance = schModels.SchedulerJobDefinition{
		SystemData: &authModels.SystemData{
			CreatedByID: schedulerUser.ID,
			UpdatedByID: schedulerUser.ID,
		},
		Name:          jdInput.Name,
		FrequencyType: jdInput.FrequencyType,
		AtTime:        jdInput.AtTime,
		CronExpr:      jdInput.CronExpr,
		WithSeconds:   jdInput.WithSeconds,
	}

	id := uuid.New()
	requestId := fmt.Sprintf("scheduler-worker::%s", id.String())

	err = dbQueries.Create(context.Background(), log, db, &instance, schedulerUser, requestId)
	if err != nil {
		return nil, err
	}

	return &instance, nil
}
