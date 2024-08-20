package builder

import (
	"github.com/spf13/viper"
)

const defaultConfigPath = "config.yaml"

type ConfigFile struct {
	UseConfigFile bool
	ConfigPath    string
}

type ConfigReader struct {
	*viper.Viper
}

// NewConfigReader returns a viper instance with the loaded configuration.
//
// It takes a ConfigFile pointer as a parameter, which specifies whether to use a config file and the path to the config file.
// If the config file path is empty, it defaults to the defaultConfigPath.
// Returns a viper instance and an error if the config file cannot be read.
func NewConfigReader(config *ConfigFile) (*ConfigReader, error) {

	if config == nil {
		config = &ConfigFile{
			UseConfigFile: false,
			ConfigPath:    "",
		}
	}

	if !config.UseConfigFile {
		log.Warn().Msg("No config file used")
		return nil, nil
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
