package builder

import (
	"errors"

	"github.com/spf13/viper"
)

var (
	ErrLoggerNotInitialized = errors.New("logger not initialized")
	ErrDBNotInitialized     = errors.New("database not initialized")
	ErrDBConfigNotProvided  = errors.New("database config not provided")
	ErrConfigNotInitialized = errors.New("config file not initialized")
)

var log *Logger

type Builder struct {
	logger       *Logger
	configReader *viper.Viper
	config       *BuilderConfig
	db           *Database
}

type BuilderConfig struct {
	*LoggerConfig // logger configuration
	*ConfigFile   // configfile configuration
	*DBConfig     // database configuration
}

func NewBuilder(cfg *BuilderConfig) *Builder {

	var output = &Builder{}

	// Config
	output.config = cfg

	// Logger
	log = NewLogger(cfg.LoggerConfig)
	output.logger = log

	// Config File
	configReader, err := NewConfigReader(cfg.ConfigFile)
	if err != nil {
		log.Error().Err(err).Msg("Error loading config")
	}
	output.configReader = configReader

	configData := configReader.AllSettings()
	log.Info().Interface("Config", configData).Msg("Loaded Config")

	return output
}

// ConnectDB connects to the database using the provided configuration.
//
// The config parameter is a pointer to a DBConfig struct that contains the database connection settings.
// No return values.
func (b *Builder) ConnectDB(config *DBConfig) error {
	// Remember configuration
	b.config.DBConfig = config

	// Load database
	db, err := LoadDB(config)
	if err != nil {
		log.Error().Err(err).Msg("Error loading database")
		return err
	}

	// Remember connection
	b.db = db

	log.Info().Interface("DBConfig", config).Msg("Connected to database")
	return nil
}

func (b *Builder) SetLoggerConfig(config LoggerConfig) {
	b.config.LoggerConfig = &config
}

// GetLogger returns the logger instance associated with the Builder.
//
// No parameters.
// Returns a pointer to the Logger instance.
func (b *Builder) GetLogger() (*Logger, error) {
	if b.logger == nil {
		return nil, ErrLoggerNotInitialized
	}
	return b.logger, nil
}

// GetConfigReader returns a viper.Viper instance used to read configuration settings.
//
// No parameters.
// Returns a pointer to a viper.Viper instance.
func (builder *Builder) GetConfigReader() (*viper.Viper, error) {
	if !builder.config.UseConfigFile {
		log.Error().Msg("No config file used")
		return nil, ErrConfigNotInitialized
	}
	return builder.configReader, nil
}
