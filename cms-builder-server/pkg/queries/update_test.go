package queries_test

import (
	"testing"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/queries"
	tu "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

func TestSave(t *testing.T) {
	// Setup in-memory SQLite database
	database := tu.NewTestDB()

	// Auto migrate the User and HistoryEntry models
	err := database.DB.AutoMigrate(&User{}, &DatabaseLog{})
	assert.NoError(t, err)

	// Create a test user
	testUser := User{
		Name:  "John Doe",
		Email: "john.doe@example.com",
	}
	database.DB.Create(&testUser)

	// Test case: Successful save (update)
	t.Run("Successful save (update)", func(t *testing.T) {
		// Update the user's email
		testUser.Email = "john.doe.updated@example.com"
		differences := map[string]interface{}{"Email": "john.doe.updated@example.com"}

		// Save the user
		result := Update(database, &testUser, &testUser, differences, "request-123")
		assert.NoError(t, result.Error)

		// Verify the user was updated
		var updatedUser User
		err := database.DB.First(&updatedUser, testUser.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "john.doe.updated@example.com", updatedUser.Email)

		// Verify the history entry was created
		var historyEntry DatabaseLog
		err = database.DB.Where("trace_id = ?", "request-123").First(&historyEntry).Error
		assert.NoError(t, err)
		assert.Equal(t, UpdateCRUDAction, historyEntry.Action)
		assert.Equal(t, testUser.StringID(), historyEntry.UserId)
		assert.Equal(t, testUser.StringID(), historyEntry.ResourceId)
		assert.Equal(t, "User", historyEntry.ResourceName)
		assert.Equal(t, "request-123", historyEntry.TraceId)
	})

	// Test case: Successful save (create)
	t.Run("Successful save (create)", func(t *testing.T) {
		// Create a new user
		newUser := User{
			Name:  "Jane Doe",
			Email: "jane.doe@example.com",
		}
		differences := map[string]interface{}{"Name": "Jane Doe", "Email": "jane.doe@example.com"}

		// Save the new user
		result := Update(database, &newUser, &testUser, differences, "request-456")
		assert.NoError(t, result.Error)

		// Verify the user was created
		var createdUser User
		err := database.DB.First(&createdUser, newUser.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "Jane Doe", createdUser.Name)
		assert.Equal(t, "jane.doe@example.com", createdUser.Email)

		// Verify the history entry was created
		var historyEntry DatabaseLog
		err = database.DB.Where("trace_id = ?", "request-456").First(&historyEntry).Error
		assert.NoError(t, err)
		assert.Equal(t, UpdateCRUDAction, historyEntry.Action)
		assert.Equal(t, testUser.StringID(), historyEntry.UserId)
		assert.Equal(t, newUser.StringID(), historyEntry.ResourceId)
		assert.Equal(t, "User", historyEntry.ResourceName)
		assert.Equal(t, "request-456", historyEntry.TraceId)
	})

	// TODO: Test case: Error creating history entry
}
