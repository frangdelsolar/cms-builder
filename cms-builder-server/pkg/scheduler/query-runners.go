package scheduler

import (
	"context"
	"fmt"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
	"github.com/google/uuid"
)

func getOrCreateJobDefinition(db *database.Database, log *loggerTypes.Logger, schedulerUser *models.User, jdInput SchedulerJobDefinition) (*SchedulerJobDefinition, error) {

	// If there is a job definition with the same name, return it
	// Name must be unique
	var instance SchedulerJobDefinition

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
	instance = SchedulerJobDefinition{
		SystemData: &models.SystemData{
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

func updateTaskStatus(log *loggerTypes.Logger, db *database.Database, schedulerUser *models.User, cronJobId string, status TaskStatus, errMsg string, requestId string, results string) error {
	task := GetSchedulerTask(log, db, cronJobId)
	task.SystemData.UpdatedByID = schedulerUser.ID
	task.Status = status
	task.Error = errMsg
	task.Results = results

	previousState := GetSchedulerTask(log, db, cronJobId)
	differences := utils.CompareInterfaces(previousState, task)
	return dbQueries.Update(context.Background(), log, db, task, schedulerUser, differences, requestId)
}

func GetSchedulerTask(log *loggerTypes.Logger, db *database.Database, cronJobId string) *SchedulerTask {
	var task SchedulerTask

	filters := map[string]interface{}{
		"cron_job_id": cronJobId,
	}

	err := dbQueries.FindOne(context.Background(), log, db, &task, filters)
	if err != nil {
		return nil
	}

	return &task
}
