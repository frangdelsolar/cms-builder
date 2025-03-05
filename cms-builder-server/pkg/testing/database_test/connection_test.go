package database_test

import (
	"testing"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	"github.com/stretchr/testify/assert"
)

func TestLoadDB_SQLite_Success(t *testing.T) {
	// Test loading a SQLite database with a valid configuration
	config := &dbTypes.DatabaseConfig{
		Driver: "sqlite",
		Path:   ":memory:", // Use an in-memory SQLite database for testing
	}

	log := loggerPkg.Default

	db, err := NewDatabaseConnection(config, log)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	assert.NotNil(t, db.DB)
	assert.Equal(t, "sqlite", db.Config.Driver)
}

// func TestLoadDB_Postgres_Success(t *testing.T) {
// 	// Test loading a PostgreSQL database with a valid configuration
// 	config := &dbTypes.DatabaseConfig{
// 		Driver: "postgres",
// 		URL:    "postgres://user:password@localhost:5432/dbname", // Replace with a valid test URL
// 	}

// 	log := loggerPkg.Default

// 	db, err := LoadDB(config, log)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, db)
// 	assert.NotNil(t, db.DB)
// 	assert.Equal(t, "postgres", db.Config.Driver)
// }

func TestLoadDB_InvalidDriver(t *testing.T) {
	// Test loading a database with an invalid driver
	config := &dbTypes.DatabaseConfig{
		Driver: "invalid",
	}

	log := loggerPkg.Default

	db, err := NewDatabaseConnection(config, log)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidDriver, err)
	assert.Nil(t, db)
}

func TestLoadDB_NilConfig(t *testing.T) {
	// Test loading a database with a nil config
	log := loggerPkg.Default

	db, err := NewDatabaseConnection(nil, log)
	assert.Error(t, err)
	assert.Equal(t, ErrDBConfigNotProvided, err)
	assert.Nil(t, db)
}

func TestLoadDB_EmptySQLitePath(t *testing.T) {
	// Test loading a SQLite database with an empty path
	config := &dbTypes.DatabaseConfig{
		Driver: "sqlite",
		Path:   "", // Empty path
	}

	log := loggerPkg.Default

	db, err := NewDatabaseConnection(config, log)
	assert.Error(t, err)
	assert.Equal(t, ErrEmptySQLitePath, err)
	assert.Nil(t, db)
}

func TestLoadDB_EmptyPostgresURL(t *testing.T) {
	// Test loading a PostgreSQL database with an empty URL
	config := &dbTypes.DatabaseConfig{
		Driver: "postgres",
		URL:    "", // Empty URL
	}

	log := loggerPkg.Default

	db, err := NewDatabaseConnection(config, log)
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyPostgresURL, err)
	assert.Nil(t, db)
}

func TestLoadDB_SQLite_ConnectionError(t *testing.T) {
	// Test loading a SQLite database with an invalid path
	config := &dbTypes.DatabaseConfig{
		Driver: "sqlite",
		Path:   "/invalid/path/to/database.db", // Invalid path
	}

	log := loggerPkg.Default

	db, err := NewDatabaseConnection(config, log)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to database")
	assert.Nil(t, db)
}

func TestLoadDB_Postgres_ConnectionError(t *testing.T) {
	// Test loading a PostgreSQL database with an invalid URL
	config := &dbTypes.DatabaseConfig{
		Driver: "postgres",
		URL:    "postgres://invalid:invalid@localhost:5432/invalid", // Invalid URL
	}

	log := loggerPkg.Default

	db, err := NewDatabaseConnection(config, log)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to database")
	assert.Nil(t, db)
}

func TestDatabase_Close(t *testing.T) {
	// Test closing a database connection
	config := &dbTypes.DatabaseConfig{
		Driver: "sqlite",
		Path:   ":memory:", // Use an in-memory SQLite database for testing
	}

	log := loggerPkg.Default

	db, err := NewDatabaseConnection(config, log)
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
	db := &dbTypes.DatabaseConnection{}

	err := db.Close()
	assert.Error(t, err)
	assert.Equal(t, ErrDBNotInitialized, err)
}
