package main

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *Database

type Database struct {
	*gorm.DB
}

func (d *Database) Register(model interface{}) {
	d.DB.AutoMigrate(model)
}

// LoadDB initializes a new SQLite database connection and returns a pointer to the Database struct and an error.
//
// It takes a filepath string as a parameter, which is the path to the SQLite database file.
// If the filepath is empty, it defaults to "./data.db".
//
// The function returns a pointer to the Database struct and an error.
func LoadDB() (*Database, error) {

	cfg, err := GetConfig()
	if err != nil {
		log.Err(err).Msg("error loading config")
	}

	if cfg.AppEnv == "prod" {
		// Connect to PostgreSQL
		dsn := cfg.DbUrl
		gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to connect to PostgreSQL database")
			return nil, err
		}
		db = &Database{
			gormDB,
		}

		log.Debug().Msgf("Connecting to database: %s", dsn)
		return db, nil
	}

	log.Debug().Msgf("Connecting to database: %s", "dbFile")

	dbFile := fmt.Sprintf("%s.db", cfg.AppEnv)
	gormDB, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
		return nil, err
	}

	db = &Database{
		gormDB,
	}

	return db, nil
}

// GetDB returns the database connection for the current application.
//
// It checks if the database connection has been initialized and if not, it
// calls the InitDB function to establish a connection to the SQLite database.
// If the connection is successful, it returns the database connection.
// If the connection fails, it logs an error message and returns the error.
// If the database connection has already been initialized, it returns the
// existing connection.
//
// Returns:
// - *Database: the database connection.
// - error: an error if the database connection fails.
func GetDB() (*Database, error) {

	if db == nil {
		log.Warn().Msg("Database was not initialized. Will initialize now...")
		return LoadDB()
	}

	return db, nil
}
