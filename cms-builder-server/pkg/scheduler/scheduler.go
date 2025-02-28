package scheduler

import (
	"fmt"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

const (
	TaskStatusRunning TaskStatus = "running"
	TaskStatusFailed  TaskStatus = "failed"
	TaskStatusDone    TaskStatus = "done"
)

const (
	JobFrequencyTypeImmediate JobFrequencyType = "immediate"
	JobFrequencyTypeScheduled JobFrequencyType = "scheduled"
	JobFrequencyTypeCron      JobFrequencyType = "cron"
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

type SchedulerJobFunction any

func (s *Scheduler) RegisterJob(jdInput SchedulerJobDefinition, jobFunction SchedulerJobFunction, jobParameters ...any) error {

	jobDefinition, err := getOrCreateJobDefinition(s.DB, s.User, jdInput)
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
	_, err = s.createCronJobInstance(frequencyDefinition, jobDefinition, jobFunction, jobParameters...)
	if err != nil {
		s.Logger.Error().Err(err).Msg("Error creating job")
		return err
	}

	return nil
}

func (s *Scheduler) createCronJobInstance(frequency gocron.JobDefinition, jobDefinition *SchedulerJobDefinition, taskFunction any, taskParameters ...any) (gocron.Job, error) {
	return s.Cron.NewJob(
		frequency,
		gocron.NewTask(
			taskFunction,
			taskParameters...,
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
}

func (s *Scheduler) Before(jobDefinition *SchedulerJobDefinition) func(jobID uuid.UUID, jobName string) {
	return func(jobID uuid.UUID, jobName string) {
		task := SchedulerTask{
			SystemData: &models.SystemData{
				CreatedByID: s.User.ID,
				UpdatedByID: s.User.ID,
			},
			JobDefinitionName: jobDefinition.Name,
			Status:            TaskStatusRunning,
			CronJobId:         jobID.String(),
		}

		requestId := getRequestIdForCronJob(jobID)
		err := queries.Update(s.DB, &task, s.User, nil, requestId).Error
		if err != nil {
			s.Logger.Error().Err(err).Msg("Error saving task")
		}
	}
}

func (s *Scheduler) WithErrors(jobDefinition *SchedulerJobDefinition) func(jobID uuid.UUID, jobName string, jobError error) {
	return func(jobID uuid.UUID, jobName string, jobError error) {
		requestId := getRequestIdForCronJob(jobID)
		err := updateTaskStatus(s.DB, s.User, jobID.String(), TaskStatusFailed, jobError.Error(), requestId)
		if err != nil {
			s.Logger.Error().Err(err).Msg("Error updating task status")
		}
	}
}

func (s *Scheduler) After(jobDefinition *SchedulerJobDefinition) func(jobID uuid.UUID, jobName string) {
	return func(jobID uuid.UUID, jobName string) {
		requestId := getRequestIdForCronJob(jobID)
		err := updateTaskStatus(s.DB, s.User, jobID.String(), TaskStatusDone, "", requestId)
		if err != nil {
			s.Logger.Error().Err(err).Msg("Error updating task status")
		}
	}
}

func (s *Scheduler) Shutdown() error {
	return s.Cron.Shutdown()
}

func getRequestIdForCronJob(jobID uuid.UUID) string {
	return fmt.Sprintf("scheduler-worker::%s", jobID.String())
}
