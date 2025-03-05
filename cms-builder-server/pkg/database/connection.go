package database

import (
	"fmt"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DBConfig defines the configuration options for connecting to a database.
type DBConfig struct {
	// URL: Used for connecting to a PostgreSQL database.
	// Provide a complete connection string (e.g., "postgres://user:password@host:port/database").
	URL string
	// Path: Used for connecting to a SQLite database.
	// Provide the path to the SQLite database file.
	Path string

	// Driver: The driver to use for connecting to the database. postgres or sqlite
	Driver string
}

var (
	ErrDBNotInitialized    = fmt.Errorf("database not initialized")
	ErrDBConfigNotProvided = fmt.Errorf("database config not provided")
	ErrInvalidDriver       = fmt.Errorf("invalid driver: must be 'sqlite' or 'postgres'")
	ErrEmptySQLitePath     = fmt.Errorf("empty database path for SQLite")
	ErrEmptyPostgresURL    = fmt.Errorf("empty database URL for PostgreSQL")
)

// Database represents a database connection managed by GORM.
type Database struct {
	DB     *gorm.DB // Embedded GORM DB instance for database access
	Config *DBConfig
}

func (d *Database) Close() error {
	if d.DB != nil {
		sqlDB, err := d.DB.DB() // Get the underlying *sql.DB instance
		if err != nil {
			return fmt.Errorf("failed to get underlying database connection: %v", err)
		}
		err = sqlDB.Close() // Close the database connection
		if err != nil {
			return fmt.Errorf("failed to close database connection: %v", err)
		}
		d.DB = nil // Set the DB field to nil
		return nil
	}
	return fmt.Errorf("database not initialized")
}

// LoadDB establishes a connection to the database based on the provided configuration.
//
// It takes a pointer to a DBConfig struct as input, which specifies the connection details.
// On successful connection, it returns a pointer to a Database instance encapsulating the GORM DB object.
// Otherwise, it returns an error indicating the connection failure.
func LoadDB(config *DBConfig, log *loggerTypes.Logger) (*Database, error) {
	if config == nil {
		return nil, ErrDBConfigNotProvided
	}

	if log == nil {
		log = logger.Default
	}

	// Validate the driver
	if config.Driver != "sqlite" && config.Driver != "postgres" {
		return nil, ErrInvalidDriver
	}

	// Validate required fields based on the driver
	switch config.Driver {
	case "sqlite":
		if config.Path == "" {
			return nil, ErrEmptySQLitePath
		}
	case "postgres":
		if config.URL == "" {
			return nil, ErrEmptyPostgresURL
		}
	}

	// Load the database based on the driver
	db := &Database{Config: config}
	var err error

	switch config.Driver {
	case "sqlite":
		db.DB, err = gorm.Open(sqlite.Open(config.Path), &gorm.Config{})
	case "postgres":
		db.DB, err = gorm.Open(postgres.Open(config.URL), &gorm.Config{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	log.Debug().Interface("config", config).Msg("Database loaded successfully")
	return db, nil
}
