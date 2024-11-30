package builder_test

import (
	"testing"

	"github.com/frangdelsolar/cms/builder"
	"github.com/stretchr/testify/assert"
)

// TestNewConfigReader_Success tests the NewConfigReader function with a valid configuration file.
//
// It tests the following:
//   - The returned ConfigReader is not nil.
//   - The returned error is nil.
//   - The ConfigReader can successfully retrieve values from the configuration file.
func TestNewConfigReader_Success(t *testing.T) {
	// Define a test config path with valid data (replace with your actual data)
	testConfigPath := "config.yaml"

	// Create a ReaderConfig with the test path
	config := &builder.ReaderConfig{ConfigFilePath: testConfigPath}

	// Call NewConfigReader to create a ConfigReader instance
	reader, err := builder.NewConfigReader(config)

	// Assert that no error occurred
	assert.NoError(t, err)

	// Test getting values using the reader (replace with your actual keys)
	value := reader.GetString("logLevel")
	assert.Equal(t, "info", value)
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
