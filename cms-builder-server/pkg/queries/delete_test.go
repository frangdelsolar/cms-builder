package queries_test

import (
	"testing"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/queries"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDelete(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto migrate the User and LogHistoryEntry models
	err = db.AutoMigrate(&User{}, &DatabaseLog{})
	assert.NoError(t, err)

	// Create a Database instance
	database := &Database{DB: db}

	// Create a test user
	testUser := User{
		Name:  "John Doe",
		Email: "john.doe@example.com",
	}
	db.Create(&testUser)

	// Test case: Successful deletion
	t.Run("Successful deletion", func(t *testing.T) {
		// Delete the user
		result := Delete(database, &testUser, &testUser, "request-123")
		assert.NoError(t, result.Error)

		// Verify the user was deleted
		var deletedUser User
		err := db.First(&deletedUser, testUser.ID).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)

		// Verify the history entry was created
		var historyEntry DatabaseLog
		err = db.Where("trace_id = ?", "request-123").First(&historyEntry).Error
		assert.NoError(t, err)
		assert.Equal(t, models.DeleteCRUDAction, historyEntry.Action)
		assert.Equal(t, testUser.StringID(), historyEntry.UserId)
		assert.Equal(t, testUser.StringID(), historyEntry.ResourceId)
		assert.Equal(t, "User", historyEntry.ResourceName)
		assert.Equal(t, "request-123", historyEntry.TraceId)
	})

	// TODO: Test case: Entity does not exist

	// TODO: Test case: Error creating history entry

}
