package scheduler

import (
	"context"
	"fmt"

	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	schConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/constants"
	schModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/models"
	schTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/types"
	schUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/utils"
)

func RunJobRegistryJob(jr *schTypes.JobRegistry, jd *schModels.SchedulerJobDefinition, requestId string, user *authModels.User, log *loggerTypes.Logger, db *dbTypes.DatabaseConnection) (string, error) {
	log.Info().Str("Job", jd.Name).Msg("Running task")

	// Look up the task definition in the registry
	taskDefinition, ok := jr.Jobs[jd.Name]
	if !ok {
		err := fmt.Errorf("task %s not found in registry", jd.Name)
		log.Error().Err(err).Msg("Task not found")
		return "", err
	}

	traceId, err := before(jd, db, user, requestId, log)
	if err != nil {
		return "", err
	}

	// Execute the task function with its parameters
	results, jobError := taskDefinition.Function(taskDefinition.Parameters...)

	if jobError == nil {
		err = success(traceId, db, user, requestId, log, results)
		if err != nil {
			log.Error().Err(err).Msg("Error updating task status")
			return "", err
		}
	} else {
		err = fail(traceId, db, user, requestId, log, jobError.Error(), results)
		if err != nil {
			log.Error().Err(err).Msg("Error updating task status")
			return "", err
		}
	}

	return results, nil
}

func before(jobDefinition *schModels.SchedulerJobDefinition, db *dbTypes.DatabaseConnection, user *authModels.User, requestId string, log *loggerTypes.Logger) (string, error) {
	log.Info().Interface("JobDefinition", jobDefinition).Msg("Starting task job")

	task := schModels.SchedulerTask{
		SystemData: &authModels.SystemData{
			CreatedByID: user.ID,
			UpdatedByID: user.ID,
		},
		JobDefinitionName: jobDefinition.Name,
		Status:            schConstants.TaskStatusRunning,
		CronJobId:         "user-triggered::" + requestId,
	}

	err := dbQueries.Create(context.Background(), log, db, &task, user, requestId)
	if err != nil {
		log.Error().Err(err).Msg("Error saving task")
		return "", err
	}

	return task.CronJobId, nil
}

func success(jobId string, db *dbTypes.DatabaseConnection, user *authModels.User, requestId string, log *loggerTypes.Logger, results string) error {
	log.Info().Interface("jobId", jobId).Msg("Task Job Succeded")
	err := schUtils.UpdateTaskStatus(log, db, user, jobId, schConstants.TaskStatusDone, "", requestId, results)
	if err != nil {
		log.Error().Err(err).Msg("Error updating task status")
		return err
	}
	return nil
}

func fail(jobId string, db *dbTypes.DatabaseConnection, user *authModels.User, requestId string, log *loggerTypes.Logger, jobError string, results string) error {
	log.Error().Interface("jobId", jobId).Msg("Task Job Failed")
	err := schUtils.UpdateTaskStatus(log, db, user, jobId, schConstants.TaskStatusFailed, jobError, requestId, results)
	if err != nil {
		log.Error().Err(err).Msg("Error updating task status")
		return err
	}
	return nil
}
