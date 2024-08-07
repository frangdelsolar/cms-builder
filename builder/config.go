package builder

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

const configFilePath = "config.yaml"

var config map[string]interface{}

func (builder *Builder) LoadConfig() error {
	log.Debug().Msg("Loading config")
	cfgFile, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	config = make(map[string]interface{}) // Initialize empty map
	err = yaml.Unmarshal(cfgFile, &config)
	if err != nil {
		return fmt.Errorf("error unmarshalling config file: %w", err)
	}

	log.Info().Interface("config", config).Msg("Loaded config")
	return nil
}

// GetKey retrieves a value from the configuration map
func (builder *Builder) GetKey(key string) (interface{}, bool) {
	value, found := config[key]
	return value, found
}
