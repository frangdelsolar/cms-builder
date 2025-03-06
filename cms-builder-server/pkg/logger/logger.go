package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"

	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
)

var Default *loggerTypes.Logger

func init() {
	log, _ := NewLogger(&loggerTypes.LoggerConfig{
		LogLevel: "debug",
	})

	Default = log
}

// NewLogger creates a new zerolog.Logger instance based on the provided configuration.
func NewLogger(config *loggerTypes.LoggerConfig) (*loggerTypes.Logger, error) {

	// Handle nil config by providing a default configuration
	if config == nil {
		return nil, fmt.Errorf("nil config provided")
	}

	// Validate log level
	level, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		fmt.Printf("Invalid log level: %s. Defaulting to debug level.\n", config.LogLevel)
		level = zerolog.DebugLevel // Use default level if invalid
	}

	// Set global log level for zerolog
	zerolog.SetGlobalLevel(level)

	// Configure caller information format
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		path := filepath.Dir(file)
		file = filepath.Base(path) + "/" + filepath.Base(file)
		return file + ":" + strconv.Itoa(line)
	}

	var logger zerolog.Logger

	// CONSOLE MODE (if WriteToFile is false)
	if !config.WriteToFile {
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
		}).
			With().
			Caller().
			Timestamp().
			Logger()

		return &loggerTypes.Logger{&logger}, nil
	}

	// FILE MODE (if WriteToFile is true)

	// Create log directory if it doesn't exist
	err = os.MkdirAll(filepath.Dir(config.LogFilePath), os.ModePerm)
	if err != nil {
		return nil, err
	}

	// Open log file for appending, creating if necessary
	logFile, err := os.OpenFile(
		config.LogFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)

	if err != nil {
		return nil, err
	}

	// Create a writer that logs to both console and file
	writer := zerolog.MultiLevelWriter(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
	}, logFile)

	logger = zerolog.New(writer).
		With().
		Caller().
		Timestamp().
		Logger()

	return &loggerTypes.Logger{&logger}, nil
}
