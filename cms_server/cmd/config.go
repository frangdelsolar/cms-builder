package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

const configFilePath = "config.yaml"

var config *Config

type Config struct {
	AppEnv         string `yaml:"appEnv"`
	LogLevel       string `yaml:"logLevel"`
	FirebaseSecret string `yaml:"firebaseSecret"`
	DbUrl          string `yaml:"dbUrl"`
	Port           string `yaml:"port"`
}

// LoadConfig loads the configuration file from the specified path and returns the parsed configuration.
//
// It takes no parameters and returns a pointer to the Config struct and an error.
// env variables take precedence over the config file
func LoadConfig() (*Config, error) {

	config = &Config{}

	cfgFile, err := os.ReadFile(configFilePath)
	if err != nil {
		fmt.Printf("WARN: Error reading config file: %s\n", err)
		// return nil, err
	} else {
		err = yaml.Unmarshal(cfgFile, &config)
		if err != nil {
			fmt.Printf("Error unmarshalling config file: %s\n", err)
			// return nil, err
		}
	}

	if os.Getenv("APP_ENV") != "" {
		config.AppEnv = os.Getenv("APP_ENV")
	}

	if os.Getenv("LOG_LEVEL") != "" {
		config.LogLevel = os.Getenv("LOG_LEVEL")
	}

	if os.Getenv("FIREBASE_SECRET") != "" {
		config.FirebaseSecret = os.Getenv("FIREBASE_SECRET")
	}

	if os.Getenv("DB_URL") != "" {
		config.DbUrl = os.Getenv("DB_URL")
	}

	if os.Getenv("PORT") != "" {
		config.Port = os.Getenv("PORT")
	}

	return config, nil
}

// GetConfig retrieves the configuration. If the config is nil, it loads it using LoadConfig.
//
// Returns:
// - A pointer to the Config struct and an error.
func GetConfig() (*Config, error) {
	if config == nil {
		return LoadConfig()
	}
	return config, nil
}
