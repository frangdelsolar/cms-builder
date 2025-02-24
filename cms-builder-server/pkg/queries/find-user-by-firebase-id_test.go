package queries_test

import (
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/queries"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestFindUserByFirebaseId(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto migrate the User model
	err = db.AutoMigrate(&User{})
	assert.NoError(t, err)

	// Create a test user
	testUser := User{
		FirebaseId: "testFirebaseId",
		Name:       "Test User",
	}
	db.Create(&testUser)

	// Create a Database instance
	database := &database.Database{DB: db}

	// Test case: User exists
	t.Run("User exists", func(t *testing.T) {
		var foundUser User
		result := FindUserByFirebaseId(database, "testFirebaseId", &foundUser)

		assert.NoError(t, result.Error)
		assert.Equal(t, testUser.FirebaseId, foundUser.FirebaseId)
		assert.Equal(t, testUser.Name, foundUser.Name)
	})

	// Test case: User does not exist
	t.Run("User does not exist", func(t *testing.T) {
		var foundUser User
		result := FindUserByFirebaseId(database, "nonExistentFirebaseId", &foundUser)

		assert.Error(t, result.Error)
		assert.Equal(t, gorm.ErrRecordNotFound, result.Error)
	})
}
