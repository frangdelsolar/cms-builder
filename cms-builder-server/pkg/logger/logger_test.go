package logger_test

import (
	"io/ioutil"
	"os"
	"testing"

	loggerPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewLogger_NoConfig(t *testing.T) {
	// Call loggerPkg.NewLogger with nil config (default will be used)
	logger, err := loggerPkg.NewLogger(nil)

	// Assert that no error occurred
	assert.Error(t, err, "expected error but got nil")
	assert.Nil(t, logger)

}

func TestNewLogger_ValidConfig(t *testing.T) {
	// Define a valid configuration
	testConfig := &loggerTypes.LoggerConfig{
		LogLevel:    "info",
		WriteToFile: true,
		LogFilePath: "test.log",
	}

	// Call loggerPkg.NewLogger with the test config
	logger, err := loggerPkg.NewLogger(testConfig)

	// Assert that no error occurred
	assert.NoError(t, err)

	// Assert that the logger is not nil
	assert.NotNil(t, logger)

	// Check the actual log level (may require additional verification depending on your setup)
	actualLevel := zerolog.GlobalLevel()
	assert.Equal(t, zerolog.InfoLevel, actualLevel)
}

func TestNewLogger_InvalidLogLevel(t *testing.T) {
	// Define a config with invalid log level
	testConfig := &loggerTypes.LoggerConfig{
		LogLevel:    "invalid_level",
		WriteToFile: true,
		LogFilePath: "test.log",
	}

	// Call loggerPkg.NewLogger with the test config
	logger, err := loggerPkg.NewLogger(testConfig)

	// Assert that an error occurred
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	// Check the actual log level (may require additional verification depending on your setup)
	actualLevel := zerolog.GlobalLevel()
	assert.Equal(t, zerolog.DebugLevel, actualLevel)
}

func TestNewLogger_WriteToFile_Success(t *testing.T) {
	defer os.Remove("test.log") // Clean up after the test

	// Define a config for writing to a file
	testConfig := &loggerTypes.LoggerConfig{
		LogLevel:    "debug",
		WriteToFile: true,
		LogFilePath: "test.log",
	}

	// Call loggerPkg.NewLogger with the test config
	logger, err := loggerPkg.NewLogger(testConfig)

	// Assert that no error occurred
	assert.NoError(t, err)

	// Assert that the logger is not nil
	assert.NotNil(t, logger)

	// Log a debug message
	logger.Debug().Msg("Test debug message")

	// Read the contents of the log file
	logFileContents, err := ioutil.ReadFile("test.log")
	assert.NoError(t, err)

	// Assert that the log file contains the message
	assert.Contains(t, string(logFileContents), "Test debug message")
}

func TestNewLogger_WriteToFile_Error(t *testing.T) {
	// Define a config with invalid file path
	testConfig := &loggerTypes.LoggerConfig{
		LogLevel:    "debug",
		WriteToFile: true,
		LogFilePath: "/invalid/path/test.log",
	}

	// Call loggerPkg.NewLogger with the test config
	logger, err := loggerPkg.NewLogger(testConfig)

	// Assert that an error occurred
	assert.Error(t, err)
	assert.Nil(t, logger)
}
