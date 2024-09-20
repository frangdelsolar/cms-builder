package builder

import (
	"errors"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	ErrDBNotInitialized    = errors.New("database not initialized")
	ErrDBConfigNotProvided = errors.New("database config not provided")
)

// Database represents a database connection managed by GORM.
type Database struct {
	DB *gorm.DB // Embedded GORM DB instance for database access
}

// FindById retrieves an entity by its ID from the database.
func (db *Database) FindById(id string, entity interface{}, userId string, skipUserBinding bool) *gorm.DB {
	if skipUserBinding {
		return db.DB.Where("id = ?", id).First(entity)
	} else {
		return db.DB.Where("id = ? AND created_by_id = ?", id, userId).First(entity)
	}
}

// FindAll retrieves all entities from the database.
func (db *Database) FindAll(entity interface{}) *gorm.DB {
	return db.DB.Find(entity)
}

// FindAllByUserId retrieves all entities associated with a specific user ID from the database.
func (db *Database) FindAllByUserId(entity interface{}, userId string) *gorm.DB {
	return db.DB.Where("created_by_id = ?", userId).Find(entity)
}

func (db *Database) Create(entity interface{}) *gorm.DB {
	return db.DB.Create(entity)
}

func (db *Database) Delete(entity interface{}) *gorm.DB {
	return db.DB.Delete(entity)
}
func (db *Database) Save(entity interface{}) *gorm.DB {
	return db.DB.Save(entity)
}

// Find runs a query on the database using the provided query string and stores the
// results in the provided entity.
//
// The query string should be a valid GORM query, such as "name = ?" or "id > ?".
// The entity should be a pointer to a struct that matches the shape of the data
// being queried.
func (db *Database) Find(entity interface{}, query string) {
	db.DB.Where(query).Find(entity)
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

	if config == nil || (config.URL == "" && config.Path == "") {
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

// Migrate calls the AutoMigrate method on the GORM DB instance.
func (db *Database) Migrate(model interface{}) error {
	if db == nil {
		return ErrDBNotInitialized
	}
	db.DB.AutoMigrate(model)
	log.Debug().Interface("Model", model).Msg("Database migration complete")
	return nil
}
