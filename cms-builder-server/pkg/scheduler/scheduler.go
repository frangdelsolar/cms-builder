package scheduler

import (
	"context"
	"fmt"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"

	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	schConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/constants"
	schInterfaces "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/interfaces"
	schModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/models"
	schTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/types"
	schUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/utils"
)

// Scheduler is the main struct for managing scheduled jobs.
type Scheduler struct {
	Cron         schInterfaces.GoCronScheduler // Instance of the gocron scheduler.
	User         *authModels.User              // User associated with the scheduler.
	DB           *dbTypes.DatabaseConnection   // Database connection for persisting job data.
	Logger       *loggerTypes.Logger           // Logger for logging scheduler events.
	TaskManager  schTypes.TaskManager          // Thread-safe map for storing job results.
	JobRegistry  schTypes.JobRegistry          // Registry of task functions and their parameters.
	RunScheduler bool                          // Flag to determine if the scheduler should run.
}

// RegisterTask registers a task function with the scheduler.
// Parameters:
//   - taskName: Name of the task.
//   - taskFunction: Function to execute for the task.
//   - parameters: Parameters for the task function.
func (s *Scheduler) RegisterTask(taskName string, taskFunction schTypes.SchedulerTaskFunc, parameters ...any) {
	s.JobRegistry.Jobs[taskName] = schTypes.JobRegistryTaskDefinition{
		Function:   taskFunction,
		Parameters: parameters,
	}
	s.Logger.Info().Str("TaskName", taskName).Msg("Task registered with parameters")
}

// Shutdown stops the scheduler.
// Returns:
//   - error: Error if shutdown fails.
func (s *Scheduler) Shutdown() error {
	return s.Cron.Shutdown()
}

// RegisterJob registers a new job with the scheduler.
// Parameters:
//   - jdInput: Job definition.
//   - jobFunction: Function to execute for the job.
//   - jobParameters: Optional parameters for the job.
//
// Returns:
//   - error: Error if job registration fails.
func (s *Scheduler) RegisterSchedulerJob(jdInput schModels.SchedulerJobDefinition, jobFunction schTypes.SchedulerTaskFunc, jobParameters ...any) error {

	s.Logger.Info().Interface("JobDefinition", jdInput).Msg("Registering job")

	jobDefinition, err := schUtils.GetOrCreateJobDefinition(s.DB, s.Logger, s.User, jdInput)
	if err != nil {
		s.Logger.Error().Err(err).Msg("Error creating job")
		return err
	}

	frequencyDefinition, err := schUtils.GetFrequencyDefinition(jobDefinition)
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
func (s *Scheduler) createCronJobInstance(frequency gocron.JobDefinition, jobDefinition *schModels.SchedulerJobDefinition, taskFunction schTypes.SchedulerTaskFunc, taskParameters ...any) (gocron.Job, error) {

	// wrappedTaskFunction wraps the original taskFunction to include logging and error handling.
	wrappedTaskFunction := func() {
		if !s.RunScheduler {

			s.Logger.Info().
				Str("JobName", jobDefinition.Name).
				Msg("Job was triggered but not executed because RunScheduler flag is false")
			return
		}
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
func (s *Scheduler) Before(jobDefinition *schModels.SchedulerJobDefinition) func(jobID uuid.UUID, jobName string) {
	return func(jobID uuid.UUID, jobName string) {

		s.Logger.Info().Interface("JobDefinition", jobDefinition).Msg("Starting task job")

		task := schModels.SchedulerTask{
			SystemData: &authModels.SystemData{
				CreatedByID: s.User.ID,
				UpdatedByID: s.User.ID,
			},
			JobDefinitionName: jobDefinition.Name,
			Status:            schConstants.TaskStatusRunning,
			CronJobId:         jobID.String(),
		}

		requestId := getRequestIdForCronJob(jobID)
		err := dbQueries.Create(context.Background(), s.Logger, s.DB, &task, s.User, requestId)
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
func (s *Scheduler) WithErrors(jobDefinition *schModels.SchedulerJobDefinition) func(jobID uuid.UUID, jobName string, jobError error) {
	return func(jobID uuid.UUID, jobName string, jobError error) {

		s.Logger.Error().Err(jobError).Msgf("Task Job %s failed", jobDefinition.Name)

		// Collect the results
		results, ok := s.TaskManager.Get(jobDefinition.Name)
		if !ok {
			s.Logger.Error().Msgf("Task %s does not exist", jobDefinition.Name)
			s.Logger.Error().Msgf("There may be a bug in the scheduler. This should not happen. Data loss may occur.")
		}

		requestId := getRequestIdForCronJob(jobID)

		err := schUtils.UpdateTaskStatus(s.Logger, s.DB, s.User, jobID.String(), schConstants.TaskStatusFailed, jobError.Error(), requestId, results)
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
func (s *Scheduler) After(jobDefinition *schModels.SchedulerJobDefinition) func(jobID uuid.UUID, jobName string) {
	return func(jobID uuid.UUID, jobName string) {

		s.Logger.Info().Interface("JobDefinition", jobDefinition).Msg("Task Job Succeded")

		// Collect the results
		results, ok := s.TaskManager.Get(jobDefinition.Name)
		if !ok {
			s.Logger.Error().Msgf("Task %s does not exist", jobDefinition.Name)
			s.Logger.Error().Msgf("There may be a bug in the scheduler. This should not happen. Data loss may occur.")
		}

		requestId := getRequestIdForCronJob(jobID)

		err := schUtils.UpdateTaskStatus(s.Logger, s.DB, s.User, jobID.String(), schConstants.TaskStatusDone, "", requestId, results)
		if err != nil {
			s.Logger.Error().Err(err).Msg("Error updating task status")
		}

		// Reset the task
		s.TaskManager.Delete(jobDefinition.Name)
	}
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
