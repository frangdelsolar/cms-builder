package queries_test

import (
	"testing"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/queries"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestFindOne(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto migrate the User model
	err = db.AutoMigrate(&User{})
	assert.NoError(t, err)

	// Create a Database instance
	database := &Database{DB: db}

	// Create a test user
	testUser := User{
		Name:  "John Doe",
		Email: "john.doe@example.com",
	}
	db.Create(&testUser)

	// Test case: Find user by ID
	t.Run("Find user by ID", func(t *testing.T) {
		var foundUser User
		query := "id = ?"
		result := FindOne(database, &foundUser, query, testUser.ID)

		assert.NoError(t, result.Error)
		assert.Equal(t, testUser.Name, foundUser.Name)
		assert.Equal(t, testUser.Email, foundUser.Email)
	})

	// Test case: Find user by email
	t.Run("Find user by email", func(t *testing.T) {
		var foundUser User
		query := "email = ?"
		result := FindOne(database, &foundUser, query, testUser.Email)

		assert.NoError(t, result.Error)
		assert.Equal(t, testUser.Name, foundUser.Name)
		assert.Equal(t, testUser.Email, foundUser.Email)
	})

	// Test case: User does not exist
	t.Run("User does not exist", func(t *testing.T) {
		var foundUser User
		query := "email = ?"
		result := FindOne(database, &foundUser, query, "non.existent@example.com")

		assert.Error(t, result.Error)
		assert.Equal(t, gorm.ErrRecordNotFound, result.Error)
	})
}
