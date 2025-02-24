package builder

import (
	"fmt"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	gocron "github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
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
	Cron    gocron.Scheduler
	Builder *Builder
	User    *User
}

func (s *Scheduler) GetOrCreateJobDefinition(jdInput SchedulerJobDefinition) (*SchedulerJobDefinition, error) {
	db := s.Builder.DB

	// If there is a job definition with the same name, return it
	// Name must be unique
	var instance SchedulerJobDefinition
	q := fmt.Sprintf("name = '%s'", jdInput.Name)
	res := db.Find(&instance, q, nil, "")
	if res.Error != nil {
		return nil, res.Error
	}

	if instance.Name != "" {
		return &instance, nil
	}

	// If there is no job definition with the same name, create it
	instance = SchedulerJobDefinition{
		SystemData: &SystemData{
			CreatedByID: s.User.ID,
			UpdatedByID: s.User.ID,
		},
		Name:          jdInput.Name,
		FrequencyType: jdInput.FrequencyType,
		AtTime:        jdInput.AtTime,
		CronExpr:      jdInput.CronExpr,
		WithSeconds:   jdInput.WithSeconds,
	}

	requestId := fmt.Sprintf("scheduler-worker-%s", instance.Name)
	res = db.Create(&instance, s.User, requestId)
	if res.Error != nil {
		return nil, res.Error
	}

	return &instance, nil
}

func getFrequencyDefinition(jobDefinition *SchedulerJobDefinition) (gocron.JobDefinition, error) {

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

func (s *Scheduler) Before(jobDefinition *SchedulerJobDefinition) func(jobID uuid.UUID, jobName string) {
	return func(jobID uuid.UUID, jobName string) {
		task := SchedulerTask{
			SystemData: &SystemData{
				CreatedByID: s.User.ID,
				UpdatedByID: s.User.ID,
			},
			JobDefinitionName: jobDefinition.Name,
			Status:            TaskStatusRunning,
			CronJobId:         jobID.String(),
		}

		requestId := GetRequestIdForCronJob(jobID)
		err := s.Builder.DB.Save(&task, s.User, nil, requestId).Error
		if err != nil {
			log.Error().Err(err).Msg("Error saving task")
		}
	}
}

func (s *Scheduler) WithErrors(jobDefinition *SchedulerJobDefinition) func(jobID uuid.UUID, jobName string, jobError error) {
	return func(jobID uuid.UUID, jobName string, jobError error) {
		requestId := GetRequestIdForCronJob(jobID)
		err := s.UpdateTaskStatus(jobID.String(), TaskStatusFailed, jobError.Error(), requestId)
		if err != nil {
			log.Error().Err(err).Msg("Error updating task status")
		}
	}
}

func (s *Scheduler) After(jobDefinition *SchedulerJobDefinition) func(jobID uuid.UUID, jobName string) {
	return func(jobID uuid.UUID, jobName string) {
		requestId := GetRequestIdForCronJob(jobID)
		err := s.UpdateTaskStatus(jobID.String(), TaskStatusDone, "", requestId)
		if err != nil {
			log.Error().Err(err).Msg("Error updating task status")
		}
	}
}

func (s *Scheduler) RegisterJob(jdInput SchedulerJobDefinition, function any, parameters ...any) error {

	jobDefinition, err := s.GetOrCreateJobDefinition(jdInput)
	if err != nil {
		log.Error().Err(err).Msg("Error creating job")
		return err
	}

	frequencyDefinition, err := getFrequencyDefinition(jobDefinition)
	if err != nil {
		log.Error().Err(err).Msg("Error creating frequency definition")
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
		log.Error().Err(err).Msg("Error creating job")
		return err
	}

	return nil
}

func (s *Scheduler) UpdateTaskStatus(cronJobId string, status TaskStatus, errMsg string, requestId string) error {
	task := s.GetSchedulerTask(cronJobId)
	task.Status = status
	if errMsg != "" {
		task.Error = errMsg
	}

	previousState := s.GetSchedulerTask(cronJobId)
	differences := CompareInterfaces(previousState, task)

	return s.Builder.DB.Save(&task, s.User, differences, requestId).Error
}

func (s *Scheduler) GetSchedulerTask(cronJobId string) *SchedulerTask {
	var task SchedulerTask

	q := "cron_job_id = '" + cronJobId + "'"
	err := s.Builder.DB.Find(&task, q, nil, "").Error
	if err != nil {
		log.Error().Err(err).Msg("Error finding task")
	}
	return &task
}

func NewScheduler(b *Builder) (*Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		log.Error().Err(err).Msg("Error creating scheduler")
		return nil, err
	}
	s.Start()
	return &Scheduler{Cron: s, Builder: b, User: &SchedulerUser}, nil
}

func (s *Scheduler) Shutdown() error {
	return s.Cron.Shutdown()
}
