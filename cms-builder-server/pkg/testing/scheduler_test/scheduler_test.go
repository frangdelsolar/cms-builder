package scheduler_test

import (
	"testing"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing/scheduler_test"
	"github.com/stretchr/testify/assert"
)

func TestNewScheduler(t *testing.T) {
	bed := SetupSchedulerTestBed()

	assert.NotNil(t, bed.Scheduler)
	assert.NotNil(t, bed.SchedulerUser)
	assert.NotNil(t, bed.Db)
	assert.NotNil(t, bed.Logger)
}

func TestShutdown(t *testing.T) {
	bed := SetupSchedulerTestBed()
	err := bed.Scheduler.Shutdown()
	assert.NoError(t, err)
}
