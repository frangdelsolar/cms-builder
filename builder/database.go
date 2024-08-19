package builder

import (
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Database struct {
	*gorm.DB
}

type DBConfig struct {
	URL  string // Used for postgres
	Path string // Used for sqlite
}

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

	return db, ErrDBConfigNotProvided
}
