package builder_test

import (
	"testing"

	"github.com/frangdelsolar/cms/builder"
	"github.com/stretchr/testify/assert"
)

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

func TestLoadDB_MissingConfig(t *testing.T) {
	// Call LoadDB with no config
	db, err := builder.LoadDB(nil)

	// Assert that the expected error is returned
	assert.EqualError(t, err, builder.ErrDBConfigNotProvided.Error())
	assert.Nil(t, db)
}

func TestLoadDB_EmptyURLAndPath(t *testing.T) {
	// Create a config with empty URL and Path
	testConfig := &builder.DBConfig{}

	// Call LoadDB with the test config
	db, err := builder.LoadDB(testConfig)

	// Assert that the expected error is returned
	assert.EqualError(t, err, builder.ErrDBConfigNotProvided.Error())
	assert.Nil(t, db)
}
