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
	TaskStatusRunning TaskStatus = "running" // Task is currently running.
	TaskStatusFailed  TaskStatus = "failed"  // Task failed during execution.
	TaskStatusDone    TaskStatus = "done"    // Task completed successfully.
)

const (
	JobFrequencyTypeImmediate JobFrequencyType = "immediate" // Job runs immediately.
	JobFrequencyTypeScheduled JobFrequencyType = "scheduled" // Job runs at a specific time.
	JobFrequencyTypeCron      JobFrequencyType = "cron"      // Job runs based on a cron expression.
)

// GoCronScheduler is an interface for interacting with the gocron scheduler.
type GoCronScheduler interface {
	// NewJob creates a new job with the given frequency, task, and event listeners.
	NewJob(frequency gocron.JobDefinition, task gocron.Task, eventListeners ...gocron.JobOption) (gocron.Job, error)
	// Shutdown stops the scheduler.
	Shutdown() error
}

// Scheduler is the main struct for managing scheduled jobs.
type Scheduler struct {
	Cron        GoCronScheduler    // Instance of the gocron scheduler.
	User        *models.User       // User associated with the scheduler.
	DB          *database.Database // Database connection for persisting job data.
	Logger      *logger.Logger     // Logger for logging scheduler events.
	TaskManager TaskManager        // Thread-safe map for storing job results.
	JobRegistry JobRegistry        // Registry of task functions and their parameters.
}

// RegisterTask registers a task function with the scheduler.
// Parameters:
//   - taskName: Name of the task.
//   - taskFunction: Function to execute for the task.
//   - parameters: Parameters for the task function.
func (s *Scheduler) RegisterTask(taskName string, taskFunction SchedulerTaskFunc, parameters ...any) {
	s.JobRegistry.Jobs[taskName] = TaskDefinition{
		Function:   taskFunction,
		Parameters: parameters,
	}
	s.Logger.Info().Str("TaskName", taskName).Msg("Task registered with parameters")
}

// NewScheduler initializes a new scheduler instance.
// Parameters:
//   - db: Database connection.
//   - schedulerUser: User associated with the scheduler.
//   - log: Logger instance.
//
// Returns:
//   - *Scheduler: Initialized scheduler instance.
//   - error: Error if initialization fails.
func NewScheduler(db *database.Database, schedulerUser *models.User, log *logger.Logger) (*Scheduler, error) {
	log.Info().Msg("Initializing scheduler")

	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}
	s.Start()
	return &Scheduler{
		Cron:        s,
		User:        schedulerUser,
		DB:          db,
		Logger:      log,
		TaskManager: TaskManager{Tasks: map[string]string{}},
		JobRegistry: JobRegistry{Jobs: map[string]TaskDefinition{}},
	}, nil
}

// RegisterJob registers a new job with the scheduler.
// Parameters:
//   - jdInput: Job definition.
//   - jobFunction: Function to execute for the job.
//   - jobParameters: Optional parameters for the job.
//
// Returns:
//   - error: Error if job registration fails.
func (s *Scheduler) RegisterJob(jdInput SchedulerJobDefinition, jobFunction SchedulerTaskFunc, jobParameters ...any) error {

	s.Logger.Info().Interface("JobDefinition", jdInput).Msg("Registering job")

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

	s.RegisterTask(jobDefinition.Name, jobFunction, jobParameters...)

	return nil
}

// ResultsCollectorFunc is a callback function to collect job results.
type ResultsCollectorFunc func(jobName string, results string)

// SchedulerTaskFunc is a function to execute a job.
type SchedulerTaskFunc func(jobParameters ...any) (string, error) // results, error

// createCronJobInstance creates a new cron job instance.
// Parameters:
//   - frequency: Job frequency definition.
//   - jobDefinition: Job definition.
//   - taskFunction: Function to execute for the job.
//   - taskParameters: Optional parameters for the task.
//
// Returns:
//   - gocron.Job: Created job instance.
//   - error: Error if job creation fails.
func (s *Scheduler) createCronJobInstance(frequency gocron.JobDefinition, jobDefinition *SchedulerJobDefinition, taskFunction SchedulerTaskFunc, taskParameters ...any) (gocron.Job, error) {
	s.Logger.Info().Int("parameters", len(taskParameters)).Interface("params", taskParameters).Msg("Running createCronJobInstance")

	// wrappedTaskFunction wraps the original taskFunction to include logging and error handling.
	wrappedTaskFunction := func() {
		s.Logger.Debug().Msg("Running wrappedTaskFunction")

		// Execute the original task function
		results, err := taskFunction(taskParameters...)

		// Store the results in the TaskManager
		s.TaskManager.Set(jobDefinition.Name, results)

		if err != nil {
			s.Logger.Error().Err(err).Str("JobName", jobDefinition.Name).Msg("Task execution failed")
			return
		}

		s.Logger.Info().Str("JobName", jobDefinition.Name).Str("Results", results).Msg("Task execution succeeded")
	}

	// Create and return a new cron job using the gocron scheduler.
	// The job is configured with:
	// - The specified frequency.
	// - A task that executes the wrappedTaskFunction.
	// - Event listeners for before, after, and error handling.
	return s.Cron.NewJob(
		frequency,
		gocron.NewTask(
			wrappedTaskFunction,
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

// Before returns a function that is executed before a job runs.
// Parameters:
//   - jobDefinition: Job definition.
//
// Returns:
//   - func(jobID uuid.UUID, jobName string): Function to execute before the job runs.
func (s *Scheduler) Before(jobDefinition *SchedulerJobDefinition) func(jobID uuid.UUID, jobName string) {
	return func(jobID uuid.UUID, jobName string) {

		s.Logger.Info().Interface("JobDefinition", jobDefinition).Msg("Starting task job")

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

		// Validate there is no existing task
		currentResults, ok := s.TaskManager.Get(jobDefinition.Name)
		if ok {
			s.Logger.Error().Interface("results", currentResults).Msgf("Task %s already exists", jobDefinition.Name)
			s.Logger.Error().Msg("Overwriting task. This should not happen.")
		}

		s.TaskManager.Set(jobDefinition.Name, "")
	}
}

// WithErrors returns a function that is executed if a job fails.
// Parameters:
//   - jobDefinition: Job definition.
//
// Returns:
//   - func(jobID uuid.UUID, jobName string, jobError error): Function to execute if the job fails.
func (s *Scheduler) WithErrors(jobDefinition *SchedulerJobDefinition) func(jobID uuid.UUID, jobName string, jobError error) {
	return func(jobID uuid.UUID, jobName string, jobError error) {

		s.Logger.Error().Err(jobError).Msgf("Task Job %s failed", jobDefinition.Name)

		// Collect the results
		results, ok := s.TaskManager.Get(jobDefinition.Name)
		if !ok {
			s.Logger.Error().Msgf("Task %s does not exist", jobDefinition.Name)
			s.Logger.Error().Msgf("There may be a bug in the scheduler. This should not happen. Data loss may occur.")
		}

		requestId := getRequestIdForCronJob(jobID)

		err := updateTaskStatus(s.DB, s.User, jobID.String(), TaskStatusFailed, jobError.Error(), requestId, results)
		if err != nil {
			s.Logger.Error().Err(err).Msg("Error updating task status")
		}

		// Reset the task
		s.TaskManager.Delete(jobDefinition.Name)
	}
}

// After returns a function that is executed after a job completes successfully.
// Parameters:
//   - jobDefinition: Job definition.
//
// Returns:
//   - func(jobID uuid.UUID, jobName string): Function to execute after the job completes.
func (s *Scheduler) After(jobDefinition *SchedulerJobDefinition) func(jobID uuid.UUID, jobName string) {
	return func(jobID uuid.UUID, jobName string) {

		s.Logger.Info().Interface("JobDefinition", jobDefinition).Msg("Task Job Succeded")

		// Collect the results
		results, ok := s.TaskManager.Get(jobDefinition.Name)
		if !ok {
			s.Logger.Error().Msgf("Task %s does not exist", jobDefinition.Name)
			s.Logger.Error().Msgf("There may be a bug in the scheduler. This should not happen. Data loss may occur.")
		}

		requestId := getRequestIdForCronJob(jobID)

		err := updateTaskStatus(s.DB, s.User, jobID.String(), TaskStatusDone, "", requestId, results)
		if err != nil {
			s.Logger.Error().Err(err).Msg("Error updating task status")
		}

		// Reset the task
		s.TaskManager.Delete(jobDefinition.Name)
	}
}

// Shutdown stops the scheduler.
// Returns:
//   - error: Error if shutdown fails.
func (s *Scheduler) Shutdown() error {
	return s.Cron.Shutdown()
}

// getRequestIdForCronJob generates a unique request ID for a cron job.
// Parameters:
//   - jobID: UUID of the job.
//
// Returns:
//   - string: Unique request ID.
func getRequestIdForCronJob(jobID uuid.UUID) string {
	return fmt.Sprintf("scheduler-worker::%s", jobID.String())
}
