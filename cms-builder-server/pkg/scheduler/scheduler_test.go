package scheduler_test

import (
	"testing"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler"
	tu "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

func TestNewScheduler(t *testing.T) {
	mockDb := tu.NewTestDB()
	mockUser := tu.GetTestUser()
	mockLogger := tu.NewTestLogger()

	scheduler, err := NewScheduler(mockDb, mockUser, mockLogger)
	assert.NoError(t, err)
	assert.NotNil(t, scheduler)
	assert.Equal(t, mockUser, scheduler.User)
	assert.Equal(t, mockDb, scheduler.DB)
	assert.Equal(t, mockLogger, scheduler.Logger)
}

func TestShutdown(t *testing.T) {

	mockDb := tu.NewTestDB()
	mockUser := tu.GetTestUser()
	mockLogger := tu.NewTestLogger()

	scheduler, err := NewScheduler(mockDb, mockUser, mockLogger)
	assert.NoError(t, err)

	err = scheduler.Shutdown()
	assert.NoError(t, err)
}
