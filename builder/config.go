package builder

import (
	"errors"

	"github.com/spf13/viper"
)

const defaultConfigPath = "config.yaml"

var ErrConfigFileNotProvided = errors.New("configuration file not provided")

type ReaderConfig struct {
	ConfigPath string
}

type ConfigReader struct {
	*viper.Viper
}

// GetString returns the value for the given key as a string.
func (c *ConfigReader) GetString(key string) string {
	return c.Viper.GetString(key)
}

// GetBool returns the value for the given key as a boolean.
func (c *ConfigReader) GetBool(key string) bool {
	return c.Viper.GetBool(key)
}

// GetInt returns the value for the given key as an integer.
func (c *ConfigReader) GetInt(key string) int {
	return c.Viper.GetInt(key)
}

// GetInt64 returns the value for the given key as an int64.
func (c *ConfigReader) GetInt64(key string) int64 {
	return c.Viper.GetInt64(key)
}

// GetFloat64 returns the value for the given key as a float64.
func (c *ConfigReader) GetFloat64(key string) float64 {
	return c.Viper.GetFloat64(key)
}

// Get returns the value for the given key.
//
// It returns an interface{} type and can be used to retrieve any type of value.
// You may need to type-cast the result to the desired type.
func (c *ConfigReader) Get(key string) interface{} {
	return c.Viper.Get(key)
}

// NewConfigReader returns a viper instance with the loaded configuration.
//
// It takes a ConfigFile pointer as a parameter, which specifies whether to use a config file and the path to the config file.
// If the config file path is empty, it defaults to the defaultConfigPath.
// Returns a viper instance and an error if the config file cannot be read.
func NewConfigReader(config *ReaderConfig) (*ConfigReader, error) {

	if config == nil {
		return nil, ErrConfigFileNotProvided
	}

	path := config.ConfigPath
	if path == "" {
		path = defaultConfigPath
	}

	viper.SetConfigFile(path)

	err := viper.ReadInConfig()
	if err != nil {
		log.Error().Err(err).Msgf("Error reading config file: %s", path)
		return nil, err
	}

	return &ConfigReader{viper.GetViper()}, nil
}
