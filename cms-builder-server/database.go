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
	DB      *gorm.DB // Embedded GORM DB instance for database access
	Builder *Builder
	Config  *DBConfig
}

// FindById retrieves a single record from the database that matches the provided ID.
// It allows for an optional query extension to refine the search criteria.
//
// Parameters:
//   - id: the unique identifier of the record to be retrieved.
//   - entity: the destination where the result will be stored.
//   - queryExtension: an optional additional query condition.
//
// Returns:
//   - *gorm.DB: the result of the database query, which can be used to check for errors.
func (db *Database) FindById(id string, entity interface{}, queryExtension string) *gorm.DB {
	q := "id = '" + id + "'"

	if queryExtension != "" {
		q += " AND " + queryExtension
	}

	return db.DB.Where(q).First(entity)
}

// FindUserByFirebaseId retrieves a user from the database by its Firebase ID.
//
// Parameters:
//   - firebaseId: the Firebase ID of the user to be retrieved.
//   - entity: the destination where the result will be stored.
//
// Returns:
//   - *gorm.DB: the result of the database query, which can be used to check for errors.
func (db *Database) FindUserByFirebaseId(firebaseId string, user *User) *gorm.DB {
	return db.DB.Where("firebase_id = ?", firebaseId).First(user)
}

// Find retrieves records from the database based on the provided query.
// If pagination is provided, the query will be limited to the specified number of records
// and offset to the correct page.
//
// Parameters:
//   - entity: the destination where the result will be stored.
//   - query: the query to be executed, it can be a raw SQL query or a GORM query.
//   - pagination: optional pagination information.
//
// Returns:
//   - *gorm.DB: the result of the database query, which can be used to check for errors.
func (db *Database) Find(entity interface{}, query string, pagination *Pagination, order string) *gorm.DB {
	if order == "" {
		order = "id desc"
	}

	if pagination == nil {
		return db.DB.Order(order).Where(query).Find(entity)
	}

	// Retrieve total number of records
	res := db.DB.Model(entity).Where(query).Count(&pagination.Total)
	if res.Error != nil {
		log.Error().Err(res.Error).Msg("Error retrieving total number of records")
		return res
	}

	// Apply pagination
	filtered := db.DB.Where(query).Order(order)
	limit := pagination.Limit
	offset := (pagination.Page - 1) * pagination.Limit

	return filtered.Limit(limit).Offset(offset).Find(entity)
}

// FindOne retrieves a single record from the database that matches the provided query.
//
// Parameters:
//   - entity: the destination where the result will be stored.
//   - query: the query to be executed, it can be a raw SQL query or a GORM query.
//
// Returns:
//   - *gorm.DB: the result of the database query, which can be used to check for errors.
func (db *Database) FindOne(entity interface{}, query string) *gorm.DB {
	return db.DB.Where(query).First(entity)
}

// Create creates a new record in the database.
//
// Parameters:
//   - entity: the model instance to be created.
//
// Returns:
//   - *gorm.DB: the result of the database query, which can be used to check for errors.
func (db *Database) Create(entity interface{}, user *User) *gorm.DB {

	result := db.DB.Create(entity)
	if result.Error == nil {
		historyEntry, err := NewLogHistoryEntry(CreateCRUDAction, user, entity, "")
		if err != nil {
			return nil
		}
		_ = db.DB.Create(historyEntry)
	}

	return result
}

// Delete deletes the record in the database.
//
// Parameters:
//   - entity: the model instance to be deleted.
//
// Returns:
//   - *gorm.DB: the result of the database query, which can be used to check for errors.
func (db *Database) Delete(entity interface{}, user *User) *gorm.DB {

	result := db.DB.Delete(entity)
	if result.Error == nil {
		historyEntry, err := NewLogHistoryEntry(DeleteCRUDAction, user, entity, "")
		if err != nil {
			return nil
		}
		_ = db.DB.Create(historyEntry)
	}

	return result
}

// Save updates a record in the database if it already exists, or creates a new one if it does not.
//
// Parameters:
//   - entity: the model instance to be saved.
//
// Returns:
//   - *gorm.DB: the result of the database query, which can be used to check for errors.
func (db *Database) Save(entity interface{}, user *User, differences interface{}) *gorm.DB {

	result := db.DB.Save(entity)
	if result.Error == nil {
		historyEntry, err := NewLogHistoryEntry(UpdateCRUDAction, user, entity, differences)
		if err != nil {
			return db.DB
		}
		_ = db.DB.Create(historyEntry)
	}

	return result
}

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

	Builder *Builder
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

	if config.Driver == "" || (config.Driver != "postgres" && config.Driver != "sqlite") {
		log.Warn().Msg("Driver not provided or invalid. Defaulting to SQLite")
		config.Driver = "sqlite"
	}

	db := &Database{}

	switch config.Driver {
	case "postgres":
		connection, err := gorm.Open(postgres.Open(config.URL), &gorm.Config{
			// Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			return db, err
		}
		db.DB = connection

	case "sqlite":
		connection, err := gorm.Open(sqlite.Open(config.Path), &gorm.Config{
			// Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			return db, err
		}
		db.DB = connection
	}

	db.Config = config
	db.Builder = config.Builder

	return db, nil
}

// Migrate calls the AutoMigrate method on the GORM DB instance.
func (db *Database) Migrate(model interface{}) error {
	if db == nil {
		return ErrDBNotInitialized
	}
	db.DB.AutoMigrate(model)
	return nil
}
