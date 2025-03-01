package scheduler_test

import (
	"fmt"
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
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
func RegisterTestTask(s *pkg.Scheduler, log *logger.Logger, store store.Store, db *database.Database, schedulerUser *models.User) (pkg.SchedulerJobDefinition, pkg.SchedulerTaskFunc) {
	jobDefinition := pkg.SchedulerJobDefinition{
		Name:          "test-job",
		FrequencyType: pkg.JobFrequencyTypeImmediate,
	}

	testFunc := func(resultsCollector pkg.ResultsCollectorFunc, jobName string, fail bool) error {
		log.Info().Msg("Test func is running")

		defer func() {
			log.Info().Msg("Running result collector")
			resultsCollector(jobName, "Test func has been ran")
		}()

		if fail {
			log.Error().Msg("Test func has failed")
			return fmt.Errorf("Test func has failed")
		}

		log.Info().Msg("Test func has completed")
		return nil
	}

	testTask := func(resultsCollector pkg.ResultsCollectorFunc, jobName string, parameters ...any) {
		testFunc(resultsCollector, jobName, parameters[0].(bool))
	}

	err := s.RegisterJob(jobDefinition, testTask)
	if err != nil {
		log.Error().Err(err).Msg("Error registering job")
		return pkg.SchedulerJobDefinition{}, nil
	}

	return jobDefinition, testTask
}

// TestRunJob tests the execution of a job with different scenarios.
func TestRunJob(t *testing.T) {
	tests := []struct {
		name           string
		failTask       bool // Whether the task should fail
		expectedStatus pkg.TaskStatus
		expectedError  string
	}{
		{
			name:           "Happy Path - Task Succeeds",
			failTask:       false,
			expectedStatus: pkg.TaskStatusDone,
			expectedError:  "",
		},
		{
			name:           "Unhappy Path - Task Fails",
			failTask:       true,
			expectedStatus: pkg.TaskStatusFailed,
			expectedError:  "Test func has failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bed := SetupSchedulerTestBed()

			cron := MockCron{}
			bed.Scheduler.Cron = cron

			jd, taskFunc := RegisterTestTask(bed.Scheduler, bed.Logger, bed.Store, bed.Db, bed.SchedulerUser)
			assert.NotNil(t, jd)
			assert.NotNil(t, taskFunc)

			beforeExec := bed.Scheduler.Before(&jd)
			assert.NotNil(t, beforeExec)

			afterExec := bed.Scheduler.After(&jd)
			assert.NotNil(t, afterExec)

			afterWithErrorsExec := bed.Scheduler.WithErrors(&jd)
			assert.NotNil(t, afterWithErrorsExec)

			jobId := uuid.New()
			jobName := jd.Name

			// Execute the "Before" hook
			beforeExec(jobId, jobName)

			// Validate task was created in db and has status of running
			task := pkg.GetSchedulerTask(bed.Db, jobId.String())
			assert.NotNil(t, task)
			assert.Equal(t, pkg.TaskStatusRunning, task.Status)

			// Validate the taskManager has the task
			results, ok := bed.Scheduler.TaskManager.Get(jd.Name)
			assert.True(t, ok)
			assert.Equal(t, "", results)

			// Execute the task
			resultsCollector := func(jobName string, results string) {
				bed.Scheduler.TaskManager.Set(jobName, results)
			}
			taskFunc(resultsCollector, jd.Name, tt.failTask)

			// Validate the taskManager has the task results
			results, ok = bed.Scheduler.TaskManager.Get(jd.Name)
			assert.True(t, ok)
			assert.Equal(t, "Test func has been ran", results)

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
			task = pkg.GetSchedulerTask(bed.Db, jobId.String())
			assert.NotNil(t, task)
			assert.Equal(t, tt.expectedStatus, task.Status)
			assert.Equal(t, "Test func has been ran", task.Results)
			assert.Equal(t, tt.expectedError, task.Error)
		})
	}
}
