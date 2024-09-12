package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
)

const (
	defaultLogFilePath = "logs/default.log"
	defaultLogLevel    = zerolog.DebugLevel
)

type Logger struct {
	*zerolog.Logger
}

// LoggerConfig defines the configuration options for the logger.
type LoggerConfig struct {
	LogLevel    string `json:"logLevel"`
	WriteToFile bool   `json:"writeToFile"`
	LogFilePath string `json:"logFilePath"`
}

// NewLogger creates a new zerolog.Logger instance based on the provided configuration.
func NewLogger(config *LoggerConfig) (*Logger, error) {

	// Handle nil config gracefully
	if config == nil {
		config = &LoggerConfig{
			LogLevel:    defaultLogLevel.String(),
			WriteToFile: true,
			LogFilePath: defaultLogFilePath,
		}
	}

	// Validate log level
	level, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		fmt.Printf("Invalid log level: %s\n", config.LogLevel)
		level = defaultLogLevel
	}

	// Set global level
	zerolog.SetGlobalLevel(level)

	// Configure caller
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		path := filepath.Dir(file)
		file = filepath.Base(path) + "/" + filepath.Base(file)
		return file + ":" + strconv.Itoa(line)
	}

	var logger zerolog.Logger

	// CONSOLE MODE
	if !config.WriteToFile {
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
		}).
			With().
			Caller().
			Timestamp().
			Logger()

		return &Logger{&logger}, nil
	}

	// FILE MODE

	// Create log directory if it doesn't exist
	err = os.MkdirAll(filepath.Dir(config.LogFilePath), os.ModePerm)
	if err != nil {
		return nil, err
	}

	// Open log file
	logFile, err := os.OpenFile(
		config.LogFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)

	if err != nil {
		return nil, err
	}

	// Create writer
	writer := zerolog.MultiLevelWriter(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
	}, logFile)

	logger = zerolog.New(writer).
		With().
		Caller().
		Timestamp().
		Logger()

	return &Logger{&logger}, nil
}
