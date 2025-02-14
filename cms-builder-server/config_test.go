package builder_test

import (
	"testing"

	builder "github.com/frangdelsolar/cms/cms-builder-server"
	"github.com/stretchr/testify/assert"
)

// TestNewConfigReader_Success tests the NewConfigReader function with a valid configuration file.
//
// It tests the following:
//   - The returned ConfigReader is not nil.
//   - The returned error is nil.
//   - The ConfigReader can successfully retrieve values from the configuration file.
func TestNewConfigReader_Success(t *testing.T) {

	// DEPRECATED
	t.Skip("This functionality is being deprecated")
	// Define a test config path with valid data (replace with your actual data)
	testConfigPath := ".test.env"

	// Create a ReaderConfig with the test path
	config := &builder.ReaderConfig{
		ConfigFilePath: testConfigPath,
		ReadFile:       true,
	}

	// Call NewConfigReader to create a ConfigReader instance
	reader, err := builder.NewConfigReader(config)

	// Assert that no error occurred
	assert.NoError(t, err)

	// Test getting values using the reader (replace with your actual keys)
	value := reader.GetString(builder.EnvKeys.LogLevel)
	assert.Equal(t, "debug", value)
}

func TestNewConfigReader_EmptyConfigPath(t *testing.T) {
	// Call NewConfigReader with nil config
	reader, err := builder.NewConfigReader(nil)

	// Assert that the expected error is returned
	assert.EqualError(t, err, builder.ErrConfigFileNotProvided.Error())
	assert.Nil(t, reader)
}

// TestNewConfigReader_InvalidConfigPath tests the NewConfigReader function with a non-existent configuration file.
//
// It tests the following:
//   - The returned ConfigReader is nil.
//   - The returned error is not nil.
func TestNewConfigReader_InvalidConfigPath(t *testing.T) {
	// Define a non-existent config path
	invalidPath := "invalid/path/config.yaml"

	// Create a ReaderConfig with the invalid path
	config := &builder.ReaderConfig{
		ConfigFilePath: invalidPath,
		ReadFile:       true,
	}

	// Call NewConfigReader
	reader, err := builder.NewConfigReader(config)

	// Assert that an error is returned (specific error may vary depending on your implementation)
	assert.Error(t, err)
	assert.Nil(t, reader)
}
