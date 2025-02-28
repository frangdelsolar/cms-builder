package database

import (
	"fmt"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
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
)

// Database represents a database connection managed by GORM.
type Database struct {
	DB     *gorm.DB // Embedded GORM DB instance for database access
	Config *DBConfig
}

func (d *Database) Close() {
	if d.DB != nil {
		sqlDB, _ := d.DB.DB()
		sqlDB.Close()
	}
}

// LoadDB establishes a connection to the database based on the provided configuration.
//
// It takes a pointer to a DBConfig struct as input, which specifies the connection details.
// On successful connection, it returns a pointer to a Database instance encapsulating the GORM DB object.
// Otherwise, it returns an error indicating the connection failure.
func LoadDB(config *DBConfig, log *logger.Logger) (*Database, error) {

	log.Debug().Interface("config", config).Msg("Loading database...")

	if config == nil {
		return nil, ErrDBConfigNotProvided
	}

	if log == nil {
		log = logger.Default
	}

	if config.Driver == "" || (config.Driver != "postgres" && config.Driver != "sqlite") {
		log.Warn().Msg("Driver not provided or invalid. Defaulting to SQLite")
		config.Driver = "sqlite"
	}

	db := &Database{}

	switch config.Driver {
	case "postgres":

		if config.URL == "" {
			return db, fmt.Errorf("empty database URL")
		}

		connection, err := gorm.Open(postgres.Open(config.URL), &gorm.Config{
			// Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			return db, err
		}
		db.DB = connection

	case "sqlite":

		if config.Path == "" {
			return db, fmt.Errorf("empty database path")
		}

		connection, err := gorm.Open(sqlite.Open(config.Path), &gorm.Config{
			// Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			return db, err
		}
		db.DB = connection
	}

	db.Config = config

	return db, nil
}
