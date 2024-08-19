package builder

import (
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Database represents a database connection managed by GORM.
type Database struct {
	*gorm.DB // Embedded GORM DB instance for database access
}

// DBConfig defines the configuration options for connecting to a database.
type DBConfig struct {
	// URL: Used for connecting to a PostgreSQL database.
	// Provide a complete connection string (e.g., "postgres://user:password@host:port/database").
	URL string
	// Path: Used for connecting to a SQLite database.
	// Provide the path to the SQLite database file.
	Path string
}

// LoadDB establishes a connection to the database based on the provided configuration.
//
// It takes a pointer to a DBConfig struct as input, which specifies the connection details.
// On successful connection, it returns a pointer to a Database instance encapsulating the GORM DB object.
// Otherwise, it returns an error indicating the connection failure.
func LoadDB(config *DBConfig) (*Database, error) {

	if config.URL == "" && config.Path == "" {
		return nil, ErrDBConfigNotProvided
	}

	var db *Database

	if config.URL != "" {
		// Connect to PostgreSQL
		gormDB, err := gorm.Open(postgres.Open(config.URL), &gorm.Config{})
		if err != nil {
			return db, err
		}
		return &Database{
			gormDB,
		}, nil
	}

	if config.Path != "" {
		// Connect to SQLite
		gormDB, err := gorm.Open(sqlite.Open(config.Path), &gorm.Config{})
		if err != nil {
			return db, err
		}
		return &Database{
			gormDB,
		}, nil
	}

	return db, ErrDBConfigNotProvided // Should never be reached, but added for completeness
}
