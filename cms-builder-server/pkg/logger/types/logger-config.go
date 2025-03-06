package types

// LoggerConfig defines the configuration options for the logger
type LoggerConfig struct {
	// LogLevel defines the desired logging level (e.g., "debug", "info", "warn", "error")
	LogLevel string
	// WriteToFile specifies whether logs should be written to a file
	WriteToFile bool
	// LogFilePath defines the path to the log file
	LogFilePath string
}
