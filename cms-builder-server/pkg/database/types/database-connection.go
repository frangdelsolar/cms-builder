package types

import (
	"fmt"

	"gorm.io/gorm"
)

// DatabaseConnection represents a database connection managed by GORM.
type DatabaseConnection struct {
	DB     *gorm.DB // Embedded GORM DB instance for database access
	Config *DatabaseConfig
}

func (d *DatabaseConnection) Close() error {
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
