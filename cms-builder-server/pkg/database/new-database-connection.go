package database

import (
	"fmt"

	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	ErrDBNotInitialized    = fmt.Errorf("database not initialized")
	ErrDBConfigNotProvided = fmt.Errorf("database config not provided")
	ErrInvalidDriver       = fmt.Errorf("invalid driver: must be 'sqlite' or 'postgres'")
	ErrEmptySQLitePath     = fmt.Errorf("empty database path for SQLite")
	ErrEmptyPostgresURL    = fmt.Errorf("empty database URL for PostgreSQL")
)

// NewDatabaseConnection establishes a connection to the database based on the provided configuration.
//
// It takes a pointer to a DBConfig struct as input, which specifies the connection details.
// On successful connection, it returns a pointer to a Database instance encapsulating the GORM DB object.
// Otherwise, it returns an error indicating the connection failure.
func NewDatabaseConnection(config *dbTypes.DatabaseConfig, log *loggerTypes.Logger) (*dbTypes.DatabaseConnection, error) {
	if config == nil {
		return nil, ErrDBConfigNotProvided
	}

	if log == nil {
		log = loggerPkg.Default
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
	db := &dbTypes.DatabaseConnection{Config: config}
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
