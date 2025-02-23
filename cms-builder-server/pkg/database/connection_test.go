package database_test

import (
	"testing"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/stretchr/testify/assert"
)

// TestLoadDB_Success_SQLite tests that LoadDB successfully connects to a SQLite database.
func TestLoadDB_Success_SQLite(t *testing.T) {
	// Use an in-memory SQLite database for testing
	testConfig := &DBConfig{
		Driver: "sqlite",
		Path:   ":memory:", // In-memory database for isolated testing
	}

	// Call LoadDB with the test config
	db, err := LoadDB(testConfig, nil)

	// Assert that no error occurred and the database instance is valid
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Verify the connection by pinging the database
	sqlDB, err := db.DB.DB()
	assert.NoError(t, err)
	assert.NotNil(t, sqlDB)

	err = sqlDB.Ping()
	assert.NoError(t, err, "Failed to ping the database")
}

// TestLoadDB_MissingConfig tests that LoadDB returns an error when no configuration is provided.
func TestLoadDB_MissingConfig(t *testing.T) {
	// Call LoadDB with no config
	db, err := LoadDB(nil, nil)

	// Assert that the expected error is returned and the database instance is nil
	assert.Equal(t, ErrDBConfigNotProvided, err)
	assert.Nil(t, db)
}

// TestLoadDB_EmptyURLAndPath tests that LoadDB returns an error when both URL and Path are empty in the configuration.
func TestLoadDB_EmptyURLAndPath(t *testing.T) {
	// Create a config with empty URL and Path
	testConfig := &DBConfig{}

	// Call LoadDB with the test config
	db, err := LoadDB(testConfig, nil)

	// Assert that the expected error is returned and the database instance is nil
	assert.Equal(t, ErrDBConfigNotProvided, err)
	assert.Nil(t, db)
}
