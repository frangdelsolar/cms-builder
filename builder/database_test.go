package builder_test

import (
	"testing"

	"github.com/frangdelsolar/cms/builder"
	"github.com/stretchr/testify/assert"
)

// TestLoadDB_Success_SQLite tests that LoadDB successfully connects to a SQLite database.
//
// To run this test, replace "test.db" with a valid path to a SQLite database file on your system.
//
// This test will check that:
// - The error returned by LoadDB is nil.
// - The returned Database instance is not nil.
//
// Optionally, you can add additional checks in this test if needed.
func TestLoadDB_Success_SQLite(t *testing.T) {
	// Replace with a valid path to your SQLite database file
	testConfig := &builder.DBConfig{Path: "test.db"}

	// Call LoadDB with the test config
	db, err := builder.LoadDB(testConfig)

	// Assert that no error occurred
	assert.NoError(t, err)

	// Assert that the db instance is not nil
	assert.NotNil(t, db)

	// **Optional:** You can add additional checks here if needed.
}

// TestLoadDB_MissingConfig tests that LoadDB returns an error when no configuration is provided.
//
// This test will check that:
// - The error returned by LoadDB is ErrDBConfigNotProvided.
// - The returned Database instance is nil.
func TestLoadDB_MissingConfig(t *testing.T) {
	// Call LoadDB with no config
	db, err := builder.LoadDB(nil)

	// Assert that the expected error is returned
	assert.EqualError(t, err, builder.ErrDBConfigNotProvided.Error())
	assert.Nil(t, db)
}

// TestLoadDB_EmptyURLAndPath tests that LoadDB returns an error when both URL and Path are empty in the configuration.
//
// This test will check that:
// - The error returned by LoadDB is ErrDBConfigNotProvided.
// - The returned Database instance is nil.
func TestLoadDB_EmptyURLAndPath(t *testing.T) {
	// Create a config with empty URL and Path
	testConfig := &builder.DBConfig{}

	// Call LoadDB with the test config
	db, err := builder.LoadDB(testConfig)

	// Assert that the expected error is returned
	assert.EqualError(t, err, builder.ErrDBConfigNotProvided.Error())
	assert.Nil(t, db)
}
