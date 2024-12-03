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

type Task struct {
	*SystemData
	CronID string `json:"cron_id"`
	Name   string `json:"name"`
}

type JobStatus string

const (
	JobStatusPending JobStatus = "pending"
	JobStatusRunning JobStatus = "running"
	JobStatusFailed  JobStatus = "failed"
	JobStatusDone    JobStatus = "done"
)

type Job struct {
	*SystemData
	CronID string    `json:"cron_id"`
	Name   string    `json:"name"`
	Status JobStatus `json:"status"`
}

type Scheduler struct {
	Cron    gocron.Scheduler
	Builder *Builder
}

func (s *Scheduler) RegisterJob(durationInSecs int, function any, parameters ...any) error {

	log.Debug().Interface("Scheduler", s).Msg("Registering job")

	localJob := &Job{
		Status: JobStatusPending,
		SystemData: &SystemData{
			CreatedByID: 1,
		},
	}

	db := s.Builder.DB

	err := db.Create(localJob).Error
	if err != nil {
		log.Error().Err(err).Msg("Error creating job")
		return err
	}

	// add a job to the scheduler
	j, err := s.Cron.NewJob(
		gocron.DurationJob(
			time.Duration(durationInSecs)*time.Second,
		),

		gocron.NewTask(
			function,
			parameters...,
		),

		gocron.WithEventListeners(
			gocron.AfterJobRuns(
				func(jobID uuid.UUID, jobName string) {
					// do something after the job completes
					schedulerLogger.Info().Str("Id", jobID.String()).Msgf("Job completed")

					localJob.Status = JobStatusDone
					db.Save(localJob)

					log.Debug().Interface("Job", localJob).Msg("Job completed")
				},
			),
			gocron.AfterJobRunsWithError(
				func(jobID uuid.UUID, jobName string, err error) {
					// do something when the job returns an error
					schedulerLogger.Error().Err(err).Str("Id", jobID.String()).Msgf("Job failed")

					localJob.Status = JobStatusFailed
					db.Save(localJob)

					log.Debug().Interface("Job", localJob).Msg("Job failed")
				},
			),
			gocron.BeforeJobRuns(
				func(jobID uuid.UUID, jobName string) {
					// do something immediately before the job is run
					schedulerLogger.Info().Str("Id", jobID.String()).Msgf("Job started")

					localJob.Status = JobStatusRunning
					db.Save(localJob)

					log.Debug().Interface("Job", localJob).Msg("Job started")
				},
			),
		),
	)
	if err != nil {
		log.Error().Err(err).Msg("Error creating job")
		return err
	}

	fmt.Println(j.ID())
	// 	// start the scheduler
	// 	s.Start()

	// 	// block until you are ready to shut down
	// 	select {
	// 	case <-time.After(time.Minute):
	// 	}

	// 	// when you're done, shut it down
	// 	err = s.Shutdown()
	// 	if err != nil {
	// 		// handle error
	// 	}

	return nil
}

func NewScheduler(builder *Builder) (*Scheduler, error) {
	s, err := gocron.NewScheduler(
		gocron.WithLogger(gocron.NewLogger(gocron.LogLevelDebug)),
	)
	if err != nil {
		log.Error().Err(err).Msg("Error creating scheduler")
		return nil, err
	}
	s.Start()
	return &Scheduler{Cron: s, Builder: builder}, nil
}

func (s *Scheduler) Shutdown() error {
	return s.Cron.Shutdown()
}
