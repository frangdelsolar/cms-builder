package scheduler_test

import (
	"fmt"
	"testing"

	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	pkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestNewScheduler tests the creation of a new scheduler instance.
func TestNewScheduler(t *testing.T) {
	bed := SetupSchedulerTestBed()

	assert.NotNil(t, bed.Scheduler)
	assert.NotNil(t, bed.SchedulerUser)
	assert.NotNil(t, bed.Db)
	assert.NotNil(t, bed.Logger)
}

// TestShutdown tests the shutdown functionality of the scheduler.
func TestShutdown(t *testing.T) {
	bed := SetupSchedulerTestBed()
	err := bed.Scheduler.Shutdown()
	assert.NoError(t, err)
}

// RegisterTestTask registers a test task with the scheduler.
func RegisterTestTask(s *pkg.Scheduler, log *loggerTypes.Logger, store store.Store, db *dbTypes.DatabaseConnection, schedulerUser *models.User) (pkg.SchedulerJobDefinition, pkg.SchedulerTaskFunc) {
	jobDefinition := pkg.SchedulerJobDefinition{
		Name:          "test-job",
		FrequencyType: pkg.JobFrequencyTypeImmediate,
	}

	testFunc := func(fail bool) (string, error) {
		log.Info().Msg("Test func is running")

		if fail {
			log.Error().Msg("Test func has failed")
			return "Failed Results", fmt.Errorf("Test func has failed")
		}

		log.Info().Msg("Test func has completed")
		return "Success results", nil
	}

	wrappedTestFunc := func(jobParameters ...any) (string, error) {
		return testFunc(jobParameters[0].(bool))
	}

	err := s.RegisterJob(jobDefinition, wrappedTestFunc)
	if err != nil {
		log.Error().Err(err).Msg("Error registering job")
		return pkg.SchedulerJobDefinition{}, nil
	}

	return jobDefinition, wrappedTestFunc
}

// TestRunJob tests the execution of a job with different scenarios.
func TestRunJob(t *testing.T) {
	tests := []struct {
		name           string
		failTask       bool // Whether the task should fail
		expectedStatus pkg.TaskStatus
		expectedError  string
		expectedResult string
	}{
		{
			name:           "Happy Path - Task Succeeds",
			failTask:       false,
			expectedStatus: pkg.TaskStatusDone,
			expectedError:  "",
			expectedResult: "Success results",
		},
		{
			name:           "Unhappy Path - Task Fails",
			failTask:       true,
			expectedStatus: pkg.TaskStatusFailed,
			expectedError:  "Test func has failed",
			expectedResult: "Failed Results",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bed := SetupSchedulerTestBed()

			cron := MockCron{}
			bed.Scheduler.Cron = cron

			// Register the test task
			jd, taskFunc := RegisterTestTask(bed.Scheduler, bed.Logger, bed.Store, bed.Db, bed.SchedulerUser)
			assert.NotNil(t, jd)
			assert.NotNil(t, taskFunc)

			// Get the event listeners
			beforeExec := bed.Scheduler.Before(&jd)
			assert.NotNil(t, beforeExec)

			afterExec := bed.Scheduler.After(&jd)
			assert.NotNil(t, afterExec)

			afterWithErrorsExec := bed.Scheduler.WithErrors(&jd)
			assert.NotNil(t, afterWithErrorsExec)

			// Generate a unique job ID and name
			jobId := uuid.New()
			jobName := jd.Name

			// Execute the "Before" hook
			beforeExec(jobId, jobName)

			// Validate task was created in db and has status of running
			task := pkg.GetSchedulerTask(bed.Logger, bed.Db, jobId.String())
			assert.NotNil(t, task)
			assert.Equal(t, pkg.TaskStatusRunning, task.Status)

			// Validate the taskManager has the task (initial state)
			results, ok := bed.Scheduler.TaskManager.Get(jd.Name)
			assert.True(t, ok)
			assert.Equal(t, "", results)

			// Execute the task
			// The scheduler will handle the resultsCollector logic internally
			results, err := taskFunc(tt.failTask)
			assert.Equal(t, tt.expectedResult, results)

			if err != nil {
				assert.Equal(t, tt.expectedError, err.Error())
			}
			t.Log(tt.name, results, err)

			// In this settup, the results collector has not been called as in the main flow
			// we will just mimic that for the sake of the testing.
			// Any changes in the logic will need to be made to the main flow as well.
			//
			// resultsCollector := func(jobName string, parameters ...any) error {
			// 	results, err := taskFunction(parameters...)

			// 	s.Logger.Info().Str("JobName", jobName).Str("Results", results).Msg("Running results collector for")
			// 	s.TaskManager.Set(jobName, results)

			// 	return err
			// }

			bed.Scheduler.TaskManager.Set(jd.Name, results)

			// Validate the taskManager has the task results
			results, ok = bed.Scheduler.TaskManager.Get(jd.Name)
			assert.True(t, ok)
			assert.Equal(t, tt.expectedResult, results)

			// Execute the appropriate "After" hook based on the test case
			if tt.failTask {
				afterWithErrorsExec(jobId, jobName, fmt.Errorf(tt.expectedError))
			} else {
				afterExec(jobId, jobName)
			}

			// Validate the taskManager no longer has the task
			results, ok = bed.Scheduler.TaskManager.Get(jd.Name)
			assert.False(t, ok)
			assert.Equal(t, "", results)

			// Validate the task status and results in the database
			task = pkg.GetSchedulerTask(bed.Logger, bed.Db, jobId.String())
			assert.NotNil(t, task)
			assert.Equal(t, tt.expectedStatus, task.Status)
			assert.Equal(t, tt.expectedResult, task.Results)
			assert.Equal(t, tt.expectedError, task.Error)
		})
	}
}
