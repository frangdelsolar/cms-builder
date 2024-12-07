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

func (db *Database) FindById(id string, entity interface{}, permissions RolePermissionMap, permissionParams PermissionParams) *gorm.DB {
	requestedBy := permissionParams[requestedByParamKey]
	userRoles, err := db.GetUserRoles(requestedBy)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error getting user roles")
		return nil
	}

	fullAccess, query, err := permissions.HasPermission(userRoles, PermissionRead, permissionParams)

	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", requestedBy).
			Interface("permission_params", permissionParams).
			Interface("user_roles", userRoles).
			Str("action", string(PermissionRead)).
			Msg("Error getting query for user")
		return nil
	}

	// Queries
	filterByInstanceIdQuery := "id = '" + id + "'"

	if !fullAccess {
		if query == "" {
			log.Error().Msg("User has no full access, yet no query was provided")
			return nil
		}
		q := query + " AND " + filterByInstanceIdQuery
		return db.DB.Where(q).First(entity)
	}

	return db.DB.Where(filterByInstanceIdQuery).First(entity)
}

func (db *Database) GetUserRoles(userId string) ([]Role, error) {
	return []Role{VisitorRole}, nil
}

func (db *Database) Find(entity interface{}, query string, pagination *Pagination, permissions RolePermissionMap, permissionParams PermissionParams) *gorm.DB {

	requestedBy := permissionParams[requestedByParamKey]
	userRoles, err := db.GetUserRoles(requestedBy)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error getting user roles")
		return nil
	}

	fullAccess, userFilter, err := permissions.HasPermission(userRoles, PermissionRead, permissionParams)

	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", requestedBy).
			Interface("permission_params", permissionParams).
			Interface("user_roles", userRoles).
			Str("action", string(PermissionRead)).
			Msg("Error getting query for user")
		return nil
	}

	if !fullAccess {
		if userFilter == "" {
			log.Error().Msg("User has no full access, yet no query was provided")
			return nil
		}

		q := userFilter
		if query != "" {
			q += " AND " + query
		}

		if pagination == nil {
			return db.DB.Where(q).Find(entity)
		} else {

			// Retrieve total number of records
			db.DB.Model(entity).Where(q).Count(&pagination.Total)

			// Apply pagination
			filtered := db.DB.Where(q)
			limit := pagination.Limit
			offset := (pagination.Page - 1) * pagination.Limit

			return filtered.Limit(limit).Offset(offset).Find(entity)
		}
	}

	if pagination == nil {
		return db.DB.Where(query).Find(entity)
	}

	// Retrieve total number of records
	db.DB.Model(entity).Where(query).Count(&pagination.Total)

	// Apply pagination
	filtered := db.DB.Where(query)
	limit := pagination.Limit
	offset := (pagination.Page - 1) * pagination.Limit

	return filtered.Limit(limit).Offset(offset).Find(entity)
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
	return nil
}
