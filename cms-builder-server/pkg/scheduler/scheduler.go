package scheduler

import (
	"fmt"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/queries"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

const (
	TaskStatusRunning models.TaskStatus = "running"
	TaskStatusFailed  models.TaskStatus = "failed"
	TaskStatusDone    models.TaskStatus = "done"
)

const (
	JobFrequencyTypeImmediate models.JobFrequencyType = "immediate"
	JobFrequencyTypeScheduled models.JobFrequencyType = "scheduled"
	JobFrequencyTypeCron      models.JobFrequencyType = "cron"
)

type Scheduler struct {
	Cron   gocron.Scheduler
	User   *models.User
	DB     *database.Database
	Logger *logger.Logger
}

func NewScheduler(db *database.Database, schedulerUser *models.User, log *logger.Logger) (*Scheduler, error) {

	log.Info().Msg("Initializing scheduler")

	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}
	s.Start()
	return &Scheduler{
		Cron:   s,
		User:   schedulerUser,
		DB:     db,
		Logger: log,
	}, nil
}

func (s *Scheduler) GetOrCreateJobDefinition(jdInput models.SchedulerJobDefinition) (*models.SchedulerJobDefinition, error) {

	// If there is a job definition with the same name, return it
	// Name must be unique
	var instance models.SchedulerJobDefinition
	q := "name = ?"
	res := queries.FindOne(s.DB, &instance, q, jdInput.Name)
	if res.Error != nil {
		return nil, res.Error
	}

	if instance.Name != "" {
		return &instance, nil
	}

	// If there is no job definition with the same name, create it
	instance = models.SchedulerJobDefinition{
		SystemData: &models.SystemData{
			CreatedByID: s.User.ID,
			UpdatedByID: s.User.ID,
		},
		Name:          jdInput.Name,
		FrequencyType: jdInput.FrequencyType,
		AtTime:        jdInput.AtTime,
		CronExpr:      jdInput.CronExpr,
		WithSeconds:   jdInput.WithSeconds,
	}

	id := uuid.New()
	requestId := fmt.Sprintf("scheduler-worker::%s", id.String())

	res = queries.Create(s.DB, &instance, s.User, requestId)
	if res.Error != nil {
		return nil, res.Error
	}

	return &instance, nil
}

func getFrequencyDefinition(jobDefinition *models.SchedulerJobDefinition) (gocron.JobDefinition, error) {

	switch jobDefinition.FrequencyType {
	case JobFrequencyTypeImmediate:
		return gocron.OneTimeJob(
			gocron.OneTimeJobStartImmediately(),
		), nil

	case JobFrequencyTypeScheduled:
		return gocron.OneTimeJob(
			gocron.OneTimeJobStartDateTimes(jobDefinition.AtTime),
		), nil

	case JobFrequencyTypeCron:
		if jobDefinition.CronExpr == "" {
			return nil, fmt.Errorf("cron expression is required")
		}

		return gocron.CronJob(
			jobDefinition.CronExpr,
			jobDefinition.WithSeconds,
		), nil

	}
	return nil, fmt.Errorf("unknown frequency type: %s", jobDefinition.FrequencyType)
}

func GetRequestIdForCronJob(jobID uuid.UUID) string {
	return fmt.Sprintf("scheduler-worker::%s", jobID.String())
}

func (s *Scheduler) Before(jobDefinition *models.SchedulerJobDefinition) func(jobID uuid.UUID, jobName string) {
	return func(jobID uuid.UUID, jobName string) {
		task := models.SchedulerTask{
			SystemData: &models.SystemData{
				CreatedByID: s.User.ID,
				UpdatedByID: s.User.ID,
			},
			JobDefinitionName: jobDefinition.Name,
			Status:            TaskStatusRunning,
			CronJobId:         jobID.String(),
		}

		requestId := GetRequestIdForCronJob(jobID)
		err := queries.Update(s.DB, &task, s.User, nil, requestId).Error
		if err != nil {
			s.Logger.Error().Err(err).Msg("Error saving task")
		}
	}
}

func (s *Scheduler) WithErrors(jobDefinition *models.SchedulerJobDefinition) func(jobID uuid.UUID, jobName string, jobError error) {
	return func(jobID uuid.UUID, jobName string, jobError error) {
		requestId := GetRequestIdForCronJob(jobID)
		err := s.UpdateTaskStatus(jobID.String(), TaskStatusFailed, jobError.Error(), requestId)
		if err != nil {
			s.Logger.Error().Err(err).Msg("Error updating task status")
		}
	}
}

func (s *Scheduler) After(jobDefinition *models.SchedulerJobDefinition) func(jobID uuid.UUID, jobName string) {
	return func(jobID uuid.UUID, jobName string) {
		requestId := GetRequestIdForCronJob(jobID)
		err := s.UpdateTaskStatus(jobID.String(), TaskStatusDone, "", requestId)
		if err != nil {
			s.Logger.Error().Err(err).Msg("Error updating task status")
		}
	}
}

func (s *Scheduler) RegisterJob(jdInput models.SchedulerJobDefinition, function any, parameters ...any) error {

	jobDefinition, err := s.GetOrCreateJobDefinition(jdInput)
	if err != nil {
		s.Logger.Error().Err(err).Msg("Error creating job")
		return err
	}

	frequencyDefinition, err := getFrequencyDefinition(jobDefinition)
	if err != nil {
		s.Logger.Error().Err(err).Msg("Error creating frequency definition")
		return err
	}

	// Create the job instance
	_, err = s.Cron.NewJob(
		frequencyDefinition,
		gocron.NewTask(
			function,
			parameters...,
		),
		gocron.WithEventListeners(
			gocron.BeforeJobRuns(
				s.Before(jobDefinition),
			),
			gocron.AfterJobRunsWithError(
				s.WithErrors(jobDefinition),
			),
			gocron.AfterJobRuns(
				s.After(jobDefinition),
			),
		),
	)
	if err != nil {
		s.Logger.Error().Err(err).Msg("Error creating job")
		return err
	}

	return nil
}

func (s *Scheduler) UpdateTaskStatus(cronJobId string, status models.TaskStatus, errMsg string, requestId string) error {
	task := s.GetSchedulerTask(cronJobId)
	task.Status = status
	if errMsg != "" {
		task.Error = errMsg
	}

	previousState := s.GetSchedulerTask(cronJobId)
	differences := utils.CompareInterfaces(previousState, task)

	return queries.Update(s.DB, &task, s.User, differences, requestId).Error
}

func (s *Scheduler) GetSchedulerTask(cronJobId string) *models.SchedulerTask {
	var task models.SchedulerTask

	q := "cron_job_id = ?"

	err := queries.FindOne(s.DB, &task, q, cronJobId).Error
	if err != nil {
		s.Logger.Error().Err(err).Msg("Error finding task")
	}
	return &task
}

func (s *Scheduler) Shutdown() error {
	return s.Cron.Shutdown()
}
