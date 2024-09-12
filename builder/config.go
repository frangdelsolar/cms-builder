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

func (c *ConfigReader) GetString(key string) string {
	return c.Viper.GetString(key)
}

func (c *ConfigReader) GetBool(key string) bool {
	return c.Viper.GetBool(key)
}

func (c *ConfigReader) GetInt(key string) int {
	return c.Viper.GetInt(key)
}

func (c *ConfigReader) GetInt64(key string) int64 {
	return c.Viper.GetInt64(key)
}

func (c *ConfigReader) GetFloat64(key string) float64 {
	return c.Viper.GetFloat64(key)
}

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
