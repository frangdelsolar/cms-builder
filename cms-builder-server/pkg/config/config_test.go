package config_test

import (
	"os"
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// TestNewConfigReader_ValidConfigFile tests the initialization of ConfigReader with a valid config file.
func TestNewConfigReader_ValidConfigFile(t *testing.T) {
	// Create a temporary config file
	configContent := `
key1: value1
key2: true
key3: 42
key4: 3.14
`
	tmpFile, err := os.CreateTemp("", "config.*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	assert.NoError(t, err)
	tmpFile.Close()

	// Initialize ConfigReader
	cfg := &config.ReaderConfig{
		ReadFile:       true,
		ConfigFilePath: tmpFile.Name(),
	}
	reader, err := config.NewConfigReader(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, reader)

	// Test reading values
	assert.Equal(t, "value1", reader.GetString("key1"))
	assert.Equal(t, true, reader.GetBool("key2"))
	assert.Equal(t, 42, reader.GetInt("key3"))
	assert.Equal(t, int64(42), reader.GetInt64("key3"))
	assert.Equal(t, 3.14, reader.GetFloat64("key4"))
}

// TestNewConfigReader_NoConfigFile tests the initialization of ConfigReader without a config file.
func TestNewConfigReader_NoConfigFile(t *testing.T) {
	cfg := &config.ReaderConfig{
		ReadFile: false,
	}
	reader, err := config.NewConfigReader(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, reader)
}

// TestNewConfigReader_ConfigFileNotFound tests the initialization of ConfigReader with a non-existent config file.
func TestNewConfigReader_ConfigFileNotFound(t *testing.T) {
	cfg := &config.ReaderConfig{
		ReadFile:       true,
		ConfigFilePath: "nonexistent.yaml",
	}
	reader, err := config.NewConfigReader(cfg)
	assert.Error(t, err)
	assert.Nil(t, reader)
	assert.Equal(t, config.ErrConfigFileNotFound, err)
}

// TestNewConfigReader_ConfigFileNotProvided tests the initialization of ConfigReader without a config.ReaderConfig.
func TestNewConfigReader_ConfigFileNotProvided(t *testing.T) {
	reader, err := config.NewConfigReader(nil)
	assert.Error(t, err)
	assert.Nil(t, reader)
	assert.Equal(t, config.ErrConfigFileNotProvided, err)
}

// TestGetString_EnvOverride tests that environment variables override config file values.
func TestGetString_EnvOverride(t *testing.T) {
	// Create a temporary config file
	configContent := `
key1: value1
`
	tmpFile, err := os.CreateTemp("", "config.*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	assert.NoError(t, err)
	tmpFile.Close()

	// Set an environment variable
	os.Setenv("KEY1", "env_value1")
	defer os.Unsetenv("KEY1")

	// Initialize ConfigReader
	cfg := &config.ReaderConfig{
		ReadFile:       true,
		ConfigFilePath: tmpFile.Name(),
	}
	reader, err := config.NewConfigReader(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, reader)

	// Test that the environment variable overrides the config file value
	assert.Equal(t, "env_value1", reader.GetString("key1"))
}

// TestGetBool_DefaultValue tests the GetBool method with a default value.
func TestGetBool_DefaultValue(t *testing.T) {
	reader := &config.ConfigReader{viper.New()}
	assert.Equal(t, false, reader.GetBool("nonexistent_key"))
}

// TestGetInt_DefaultValue tests the GetInt method with a default value.
func TestGetInt_DefaultValue(t *testing.T) {
	reader := &config.ConfigReader{viper.New()}
	assert.Equal(t, 0, reader.GetInt("nonexistent_key"))
}

// TestGetInt64_DefaultValue tests the GetInt64 method with a default value.
func TestGetInt64_DefaultValue(t *testing.T) {
	reader := &config.ConfigReader{viper.New()}
	assert.Equal(t, int64(0), reader.GetInt64("nonexistent_key"))
}

// TestGetFloat64_DefaultValue tests the GetFloat64 method with a default value.
func TestGetFloat64_DefaultValue(t *testing.T) {
	reader := &config.ConfigReader{viper.New()}
	assert.Equal(t, 0.0, reader.GetFloat64("nonexistent_key"))
}

// TestGet_DefaultValue tests the Get method with a default value.
func TestGet_DefaultValue(t *testing.T) {
	reader := &config.ConfigReader{viper.New()}
	assert.Nil(t, reader.Get("nonexistent_key"))
}
