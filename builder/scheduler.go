package builder

import (
	"fmt"
	"time"

	gocron "github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

var schedulerLogger *Logger

// init initializes the scheduler logger.
//
// It creates a new logger with the specified configuration and assigns it
// to the schedulerLogger variable. If the logger initialization fails, it
// prints an error message and panics with a LoggerNotInitialized error.
func init() {
	logger, err := NewLogger(&LoggerConfig{
		LogLevel:    "debug",
		WriteToFile: true,
		LogFilePath: "logs/scheduler.log",
	})
	if err != nil {
		fmt.Println("Error initializing logger:", err)
		panic(builderErr.LoggerNotInitialized)
	}
	schedulerLogger = logger
}

type SchedulerTask struct {
	*SystemData
	JobDefinitionId string                  `gorm:"not null" json:"jobDefinitionId"`
	JobDefinition   *SchedulerJobDefinition `gorm:"foreignKey:JobDefinitionId" json:"jobDefinition"`
	Status          TaskStatus              `json:"status"`
	CronJobId       string                  `json:"cronJobId"`
	Error           string                  `json:"error"`
}

type TaskStatus string

const (
	TaskStatusRunning TaskStatus = "running"
	TaskStatusFailed  TaskStatus = "failed"
	TaskStatusDone    TaskStatus = "done"
)

type JobFrequencyType string

const (
	JobFrequencyTypeImmediate JobFrequencyType = "immediate"
	JobFrequencyTypeScheduled JobFrequencyType = "scheduled"
	JobFrequencyTypeCron      JobFrequencyType = "cron"
)

type JobFrequency struct {
	*SystemData
	FrequencyType JobFrequencyType `json:"frequencyType"`
	AtTime        time.Time
	CronExpr      string `json:"cronExpr"`
	WithSeconds   bool   `json:"withSeconds"` // Cron expression with seconds
}

type SchedulerJobDefinition struct {
	*SystemData
	Name        string        `json:"name"`
	Frequency   *JobFrequency `gorm:"foreignKey:FrequencyId" json:"frequency"`
	FrequencyId string        `gorm:"not null" json:"frequencyId"`
}

type Scheduler struct {
	Cron    gocron.Scheduler
	Builder *Builder
	User    *User
}

func (s *Scheduler) RegisterJob(name string, frequency JobFrequency, function any, parameters ...any) error {

	log.Debug().Interface("Frequency", frequency).Str("Name", name).Msg("Registering job")

	// Update the system data of the frequency
	frequency.SystemData = &SystemData{
		CreatedByID: s.User.ID,
		CreatedBy:   s.User,
	}
	s.Builder.DB.Save(&frequency)

	jobDefinition, err := s.CreateJobDefinition(name, frequency)
	if err != nil {
		log.Error().Err(err).Msg("Error creating job")
		return err
	}

	frequencyDefinition, err := getFrequencyDefinition(frequency)
	if err != nil {
		log.Error().Err(err).Msg("Error creating frequency definition")
		return err
	}

	// add a job to the scheduler
	_, err = s.Cron.NewJob(
		frequencyDefinition,
		gocron.NewTask(
			function,
			parameters...,
		),
		gocron.WithEventListeners(
			gocron.BeforeJobRuns(
				func(jobID uuid.UUID, jobName string) {
					task := SchedulerTask{
						SystemData: &SystemData{
							CreatedByID: s.User.ID,
						},
						JobDefinitionId: jobDefinition.GetIDString(),
						JobDefinition:   jobDefinition,
						Status:          TaskStatusRunning,
						CronJobId:       jobID.String(),
					}

					s.Builder.DB.Save(&task)

					schedulerLogger.Info().
						Interface("Task", task).
						Msgf("Running task")
				},
			),

			gocron.AfterJobRunsWithError(
				func(jobID uuid.UUID, jobName string, jobError error) {
					err = s.UpdateTaskStatus(jobID.String(), TaskStatusFailed, jobError.Error())
					if err != nil {
						log.Error().Err(jobError).Msg("Error updating task status")
					}

					task := s.GetSchedulerTask(jobID.String())
					task.JobDefinition = jobDefinition

					schedulerLogger.Error().
						Err(jobError).
						Interface("Task", task).
						Msgf("Task failed")

				},
			),
			gocron.AfterJobRuns(
				func(jobID uuid.UUID, jobName string) {
					err = s.UpdateTaskStatus(jobID.String(), TaskStatusDone, "")
					if err != nil {
						log.Error().Err(err).Msg("Error updating task status")
					}
					task := s.GetSchedulerTask(jobID.String())
					task.JobDefinition = jobDefinition

					schedulerLogger.Info().
						Interface("Task", task).
						Msgf("Task completed")
				},
			),
		),
	)
	if err != nil {
		log.Error().Err(err).Msg("Error creating job")
		return err
	}

	return nil
}

func (s *Scheduler) UpdateTaskStatus(id string, status TaskStatus, errMsg string) error {
	task := s.GetSchedulerTask(id)
	task.Status = status
	if errMsg != "" {
		task.Error = errMsg
	}
	return s.Builder.DB.Save(&task).Error
}

func (s *Scheduler) GetSchedulerTask(id string) *SchedulerTask {
	var task SchedulerTask
	q := "cron_job_id = '" + id + "'"
	s.Builder.DB.Find(&task, q)
	return &task
}

func (s *Scheduler) CreateJobDefinition(name string, frequency JobFrequency) (*SchedulerJobDefinition, error) {
	db := s.Builder.DB
	localJob := &SchedulerJobDefinition{
		SystemData: &SystemData{
			CreatedByID: s.User.ID,
		},
		Name:        name,
		Frequency:   &frequency,
		FrequencyId: frequency.SystemData.GetIDString(),
	}
	if err := db.Create(&localJob).Error; err != nil {
		return nil, err
	}
	return localJob, nil
}

func getFrequencyDefinition(frequency JobFrequency) (gocron.JobDefinition, error) {

	switch frequency.FrequencyType {
	case JobFrequencyTypeImmediate:
		return gocron.OneTimeJob(
			gocron.OneTimeJobStartImmediately(),
		), nil

	case JobFrequencyTypeScheduled:
		return gocron.OneTimeJob(
			gocron.OneTimeJobStartDateTimes(frequency.AtTime),
		), nil

	case JobFrequencyTypeCron:
		if frequency.CronExpr == "" {
			return nil, fmt.Errorf("cron expression is required")
		}

		return gocron.CronJob(
			frequency.CronExpr,
			frequency.WithSeconds,
		), nil

	}
	return nil, fmt.Errorf("unknown frequency type: %s", frequency.FrequencyType)
}

func NewScheduler(b *Builder) (*Scheduler, error) {

	var schedulerUser User
	email := "scheduler@" + config.GetString(EnvKeys.BaseUrl)
	q := "email = '" + email + "'"
	b.DB.Find(&schedulerUser, q)

	if schedulerUser.ID == 0 {
		schedulerUser = User{
			Email: email,
			Name:  "Scheduler",
		}
		b.DB.Create(&schedulerUser)
	}

	log.Info().Interface("Scheduler user", schedulerUser).Msg("Scheduler user created")

	s, err := gocron.NewScheduler(
		gocron.WithLogger(gocron.NewLogger(gocron.LogLevelDebug)),
	)
	if err != nil {
		log.Error().Err(err).Msg("Error creating scheduler")
		return nil, err
	}
	s.Start()
	return &Scheduler{Cron: s, Builder: b, User: &schedulerUser}, nil
}

func (s *Scheduler) Shutdown() error {
	return s.Cron.Shutdown()
}
