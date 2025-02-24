package queries

import (
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Define a test model
type TestModel struct {
	gorm.Model
	Name string
}

func TestCreate(t *testing.T) {
	// Initialize in-memory database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// AutoMigrate the test model
	err = db.AutoMigrate(&TestModel{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&models.HistoryEntry{})
	assert.NoError(t, err)

	// Wrap in our database struct
	testDB := &database.Database{DB: db}

	// Create a test user
	user := &models.User{ID: 1, Name: "Test User"}

	// Create test instance
	instance := &TestModel{Name: "Test Record"}

	// Call the function
	result := Create(testDB, instance, user, "test-request-id")

	// Assertions
	assert.NoError(t, result.Error, "Expected no error when creating a record")
	assert.NotZero(t, instance.ID, "Expected instance ID to be set after creation")

	// Check if record was actually inserted
	var retrieved TestModel
	err = db.First(&retrieved, instance.ID).Error
	assert.NoError(t, err, "Expected to retrieve created record")
	assert.Equal(t, "Test Record", retrieved.Name, "Expected the name to match")

	// Check if history entry was created
	var historyEntry models.HistoryEntry
	err = db.First(&historyEntry).Error
	assert.NoError(t, err, "Expected a log history entry to be created")
}
