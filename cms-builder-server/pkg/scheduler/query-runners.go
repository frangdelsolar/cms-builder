package scheduler

import (
	"errors"
	"fmt"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func getOrCreateJobDefinition(db *database.Database, schedulerUser *models.User, jdInput SchedulerJobDefinition) (*SchedulerJobDefinition, error) {

	// If there is a job definition with the same name, return it
	// Name must be unique
	var instance SchedulerJobDefinition
	q := "name = ? AND frequency_type = ? AND at_time = ? AND cron_expr = ? AND with_seconds = ? AND at_time = ?"
	res := queries.FindOne(db, &instance, q, jdInput.Name, jdInput.FrequencyType, jdInput.AtTime, jdInput.CronExpr, jdInput.WithSeconds, jdInput.AtTime)
	if res.Error != nil {
		if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, res.Error
		}
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

	res = queries.Create(db, &instance, schedulerUser, requestId)
	if res.Error != nil {
		return nil, res.Error
	}

	return &instance, nil
}

func updateTaskStatus(db *database.Database, schedulerUser *models.User, cronJobId string, status TaskStatus, errMsg string, requestId string, results string) error {
	task := GetSchedulerTask(db, cronJobId)
	task.SystemData.UpdatedByID = schedulerUser.ID
	task.Status = status
	task.Error = errMsg
	task.Results = results

	previousState := GetSchedulerTask(db, cronJobId)
	differences := utils.CompareInterfaces(previousState, task)

	return queries.Update(db, &task, schedulerUser, differences, requestId).Error
}

func GetSchedulerTask(db *database.Database, cronJobId string) *SchedulerTask {
	var task SchedulerTask

	q := "cron_job_id = ?"

	err := queries.FindOne(db, &task, q, cronJobId).Error
	if err != nil {
		// ignore
	}
	return &task
}
