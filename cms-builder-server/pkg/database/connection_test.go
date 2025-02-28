package database

import (
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestLoadDB_SQLite_Success(t *testing.T) {
	// Test loading a SQLite database with a valid configuration
	config := &DBConfig{
		Driver: "sqlite",
		Path:   ":memory:", // Use an in-memory SQLite database for testing
	}

	log := logger.Default

	db, err := LoadDB(config, log)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	assert.NotNil(t, db.DB)
	assert.Equal(t, "sqlite", db.Config.Driver)
}

// func TestLoadDB_Postgres_Success(t *testing.T) {
// 	// Test loading a PostgreSQL database with a valid configuration
// 	config := &DBConfig{
// 		Driver: "postgres",
// 		URL:    "postgres://user:password@localhost:5432/dbname", // Replace with a valid test URL
// 	}

// 	log := logger.Default

// 	db, err := LoadDB(config, log)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, db)
// 	assert.NotNil(t, db.DB)
// 	assert.Equal(t, "postgres", db.Config.Driver)
// }

func TestLoadDB_InvalidDriver(t *testing.T) {
	// Test loading a database with an invalid driver
	config := &DBConfig{
		Driver: "invalid",
	}

	log := logger.Default

	db, err := LoadDB(config, log)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidDriver, err)
	assert.Nil(t, db)
}

func TestLoadDB_NilConfig(t *testing.T) {
	// Test loading a database with a nil config
	log := logger.Default

	db, err := LoadDB(nil, log)
	assert.Error(t, err)
	assert.Equal(t, ErrDBConfigNotProvided, err)
	assert.Nil(t, db)
}

func TestLoadDB_EmptySQLitePath(t *testing.T) {
	// Test loading a SQLite database with an empty path
	config := &DBConfig{
		Driver: "sqlite",
		Path:   "", // Empty path
	}

	log := logger.Default

	db, err := LoadDB(config, log)
	assert.Error(t, err)
	assert.Equal(t, ErrEmptySQLitePath, err)
	assert.Nil(t, db)
}

func TestLoadDB_EmptyPostgresURL(t *testing.T) {
	// Test loading a PostgreSQL database with an empty URL
	config := &DBConfig{
		Driver: "postgres",
		URL:    "", // Empty URL
	}

	log := logger.Default

	db, err := LoadDB(config, log)
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyPostgresURL, err)
	assert.Nil(t, db)
}

func TestLoadDB_SQLite_ConnectionError(t *testing.T) {
	// Test loading a SQLite database with an invalid path
	config := &DBConfig{
		Driver: "sqlite",
		Path:   "/invalid/path/to/database.db", // Invalid path
	}

	log := logger.Default

	db, err := LoadDB(config, log)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to database")
	assert.Nil(t, db)
}

func TestLoadDB_Postgres_ConnectionError(t *testing.T) {
	// Test loading a PostgreSQL database with an invalid URL
	config := &DBConfig{
		Driver: "postgres",
		URL:    "postgres://invalid:invalid@localhost:5432/invalid", // Invalid URL
	}

	log := logger.Default

	db, err := LoadDB(config, log)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to database")
	assert.Nil(t, db)
}

func TestDatabase_Close(t *testing.T) {
	// Test closing a database connection
	config := &DBConfig{
		Driver: "sqlite",
		Path:   ":memory:", // Use an in-memory SQLite database for testing
	}

	log := logger.Default

	db, err := LoadDB(config, log)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Close the database
	err = db.Close()
	assert.NoError(t, err)
	assert.Nil(t, db.DB)

	// Attempt to close again (should return an error)
	err = db.Close()
	assert.Error(t, err)
	assert.Equal(t, ErrDBNotInitialized, err)
}

func TestDatabase_Close_NotInitialized(t *testing.T) {
	// Test closing a database that is not initialized
	db := &Database{}

	err := db.Close()
	assert.Error(t, err)
	assert.Equal(t, ErrDBNotInitialized, err)
}
