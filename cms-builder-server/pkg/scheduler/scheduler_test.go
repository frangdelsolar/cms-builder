package scheduler_test

import (
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockDatabase is a mock implementation of the database.Database interface
type MockDatabase struct {
	*database.Database
	mock.Mock
	DB *gorm.DB
}

func (m *MockDatabase) Create(value interface{}, user *models.User, requestId string) *gorm.DB {
	args := m.Called(value, user, requestId)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDatabase) FindOne(out interface{}, query string, args ...interface{}) *gorm.DB {
	mockArgs := m.Called(out, query, args)
	return mockArgs.Get(0).(*gorm.DB)
}

func (m *MockDatabase) Update(value interface{}, user *models.User, differences interface{}, requestId string) *gorm.DB {
	args := m.Called(value, user, differences, requestId)
	return args.Get(0).(*gorm.DB)
}

func GetTestResources() (
	db *database.Database,
	mockUser *models.User,
	mockLogger *logger.Logger,
) {
	testConfig := &database.DBConfig{
		Driver: "sqlite",
		Path:   ":memory:",
	}

	db, err := database.LoadDB(testConfig, nil)
	if err != nil {
		panic(err)
	}

	mockUser = &models.User{ID: uint(999)}
	mockLogger = logger.Default

	return db, mockUser, mockLogger
}

func TestNewScheduler(t *testing.T) {
	db, mockUser, mockLogger := GetTestResources()
	scheduler, err := NewScheduler(db, mockUser, mockLogger)
	assert.NoError(t, err)
	assert.NotNil(t, scheduler)
	assert.Equal(t, mockUser, scheduler.User)
	assert.Equal(t, db, scheduler.DB)
	assert.Equal(t, mockLogger, scheduler.Logger)
}

func TestShutdown(t *testing.T) {
	mockDB, mockUser, mockLogger := GetTestResources()
	scheduler, err := NewScheduler(mockDB, mockUser, mockLogger)
	assert.NoError(t, err)

	err = scheduler.Shutdown()
	assert.NoError(t, err)
}
