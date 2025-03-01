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

// ResultsCollectorFunc is a callback function to collect job results.
type ResultsCollectorFunc func(jobName string, results string)

// SchedulerTaskFunc is a function to execute a job.
type SchedulerTaskFunc func(resultsCollector ResultsCollectorFunc, jobName string, jobParameters ...any)

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

	// resultsCollector is a callback function that collects the results of the job
	// and stores them in the TaskManager. It is passed to the taskFunction so that
	// the task can call it to store its results after execution.
	resultsCollector := func(jobName string, results string) {
		s.TaskManager.Set(jobName, results)
	}

	// updatedParameters is a slice that combines the required parameters for the taskFunction:
	// 1. resultsCollector: To allow the task to store its results.
	// 2. jobDefinition.Name: To identify the job.
	// 3. Any additional parameters (taskParameters) provided by the caller.
	//
	// This ensures that the taskFunction receives all the parameters it needs in the correct order.
	updatedParameters := []any{resultsCollector, jobDefinition.Name}
	if len(taskParameters) > 0 {
		updatedParameters = append(updatedParameters, taskParameters...)
	}

	// Create and return a new cron job using the gocron scheduler.
	// The job is configured with:
	// - The specified frequency.
	// - A task that executes the taskFunction with the updatedParameters.
	// - Event listeners for before, after, and error handling.
	return s.Cron.NewJob(
		frequency,
		gocron.NewTask(
			taskFunction,
			updatedParameters...,
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
