package types

// DatabaseConfig defines the configuration options for connecting to a database.
type DatabaseConfig struct {
	// URL: Used for connecting to a PostgreSQL database.
	// Provide a complete connection string (e.g., "postgres://user:password@host:port/database").
	URL string
	// Path: Used for connecting to a SQLite database.
	// Provide the path to the SQLite database file.
	Path string

	// Driver: The driver to use for connecting to the database. postgres or sqlite
	Driver string
}
