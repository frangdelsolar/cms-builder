package builder

import "github.com/spf13/viper"

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
	if cfg.ConfigFile == nil {
		cfg.ConfigFile = &ConfigFile{
			UseConfigFile: false,
			ConfigPath:    "",
		}
	}
	configReader, err := NewConfigReader(cfg.ConfigFile)
	if err != nil {
		log.Error().Err(err).Msg("Error loading config")
	}
	output.configReader = configReader

	return output
}

// ConnectDB connects to the database using the provided configuration.
//
// The config parameter is a pointer to a DBConfig struct that contains the database connection settings.
// No return values.
func (b *Builder) ConnectDB(config *DBConfig) {
	// Remember configuration
	b.config.DBConfig = config

	// Load database
	db, err := LoadDB(config)
	if err != nil {
		log.Error().Err(err).Msg("Error loading database")
	}

	// Remember connection
	b.db = db

	log.Info().Interface("DBConfig", config).Msg("Connected to database")
}

// GetLogger returns the logger instance associated with the Builder.
//
// No parameters.
// Returns a pointer to the Logger instance.
func (b *Builder) GetLogger() *Logger {
	return b.logger
}

// GetConfigReader returns a viper.Viper instance used to read configuration settings.
//
// No parameters.
// Returns a pointer to a viper.Viper instance.
func (builder *Builder) GetConfigReader() *viper.Viper {
	if !builder.config.UseConfigFile {
		log.Error().Msg("No config file used")
		return nil
	}
	return builder.configReader
}
