package main

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
)

const (
	defaultLogFile  = "default.log"
	logsDir         = "logs"
	defaultLogLevel = zerolog.DebugLevel
)

var (
	log *Logger
)

type Logger struct {
	*zerolog.Logger
}

// NewLogger creates a new Logger instance with the provided zerolog.Logger.
//
// Parameters:
// - logger: A pointer to a zerolog.Logger instance.
//
// Returns:
// - A pointer to the newly created Logger instance.

type LoggerConfig struct {
	LogLevel    string
	LogFile     string
	WriteToFile bool
}

// GetZeroLevel returns the zero log level based on the logger configuration.
//
// It takes no parameters and returns a zerolog.Level.
func (lc LoggerConfig) GetZeroLevel() zerolog.Level {
	if lc.LogLevel == "" {
		return defaultLogLevel
	}
	level, err := zerolog.ParseLevel(lc.LogLevel)
	if err != nil {
		level = defaultLogLevel
	}
	return level
}

// GetZeroLogger returns a zerolog.Logger instance with the specified logging level and output configuration.
//
// Parameters:
// - lc: A LoggerConfig struct containing the configuration for the logger.
//
// Returns:
// - A zerolog.Logger instance with the specified logging level and output configuration.
func (lc LoggerConfig) GetZeroLogger() zerolog.Logger {
	// Set Loggging Level
	zerolog.SetGlobalLevel(lc.GetZeroLevel())

	// Set Caller -> This will show the file name and line number in the log output
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		path := filepath.Dir(file)
		file = filepath.Base(path) + "/" + filepath.Base(file)
		return file + ":" + strconv.Itoa(line)
	}

	if !lc.WriteToFile {
		return zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
		}).
			With().
			Caller().
			Timestamp().
			Logger()
	}

	runLogFile, _ := os.OpenFile(
		lc.LogFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)

	// This are the writters that will be used in the logger: console and file
	multi := zerolog.MultiLevelWriter(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
	}, runLogFile)

	return zerolog.New(multi).
		With().
		Caller().
		Timestamp().
		Logger()
}

// NewLogger creates a new Logger instance based on the provided LoggerConfig.
//
// Parameters:
// - cfg: The LoggerConfig used to configure the Logger.
//
// Returns:
// - *Logger: A pointer to the newly created Logger instance.
func NewLogger() *Logger {
	// Define defaults
	logLevel := zerolog.DebugLevel.String()
	writeToFile := true

	// Set default log level if not provided either in config or environment
	loggerConfig := LoggerConfig{
		LogLevel:    logLevel,
		LogFile:     filepath.Join(logsDir, defaultLogFile),
		WriteToFile: writeToFile,
	}

	logger := loggerConfig.GetZeroLogger()
	log = &Logger{&logger}
	return log
}

// GetLogger returns a pointer to the Logger instance.
//
// Returns:
// - *Logger: A pointer to the Logger instance.
func GetLogger() *Logger {
	if log == nil {
		log = NewLogger()
	}

	return log
}
