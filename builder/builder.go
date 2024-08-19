package builder

import (
	"errors"

	"github.com/spf13/viper"
)

var (
	ErrConfigNotInitialized    = errors.New("config file not initialized")
	ErrLoggerNotInitialized    = errors.New("logger not initialized")
	ErrDBNotInitialized        = errors.New("database not initialized")
	ErrDBConfigNotProvided     = errors.New("database config not provided")
	ErrServerNotInitialized    = errors.New("server not initialized")
	ErrServerConfigNotProvided = errors.New("server config not provided")
)

var log *Logger // Global variable for the logger instance

// Builder defines a central configuration and management structure for building applications.
type Builder struct {
	logger       *Logger        // Reference to the application's logger instance
	configReader *viper.Viper   // Reference to the Viper instance used for configuration
	config       *BuilderConfig // Pointer to the main configuration object
	db           *Database      // Reference to the connected database instance (if applicable)
	server       *Server        // Reference to the created Server instance (if applicable)
}

// BuilderConfig defines a nested configuration structure for various aspects of the application.
type BuilderConfig struct {
	*LoggerConfig // Embedded configuration for the logger
	*ConfigFile   // Embedded configuration for the config file
	*DBConfig     // Embedded configuration for the database
	*ServerConfig // Embedded configuration for the server
}

// NewBuilder creates a new Builder instance and initializes its configuration.
//
// It takes a pointer to a BuilderConfig struct as input, which encapsulates all
// application configuration options. If the provided config argument is nil,
// it will return nil.
//
// On successful initialization, it returns a pointer to the newly created Builder instance.
func NewBuilder(cfg *BuilderConfig) *Builder {
	if cfg == nil {
		return nil
	}

	var output = &Builder{
		config: cfg,
	}

	// Initialize Logger
	log = NewLogger(cfg.LoggerConfig)
	output.logger = log

	// Initialize Config Reader
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
// It takes a pointer to a DBConfig struct that contains the database connection settings.
// On successful connection, it returns nil. Otherwise, it returns an error.
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

// SetLoggerConfig allows you to update the logger configuration after Builder initialization.
//
// It takes a LoggerConfig struct as input and updates the internal configuration for the logger.
func (b *Builder) SetLoggerConfig(config LoggerConfig) {
	b.config.LoggerConfig = &config
}

// GetLogger returns the logger instance associated with the Builder.
//
// It checks if the logger is initialized and returns an error if not. Otherwise, it returns a pointer to the logger instance.
func (b *Builder) GetLogger() (*Logger, error) {
	if b.logger == nil {
		return nil, ErrLoggerNotInitialized
	}
	return b.logger, nil
}

// GetConfigReader returns a viper.Viper instance used to read configuration settings.
//
// It checks if the `UseConfigFile` flag is set in the configuration. If not, it logs an error
// and returns an error. Otherwise, it returns a pointer to the viper.Viper instance.
func (builder *Builder) GetConfigReader() (*viper.Viper, error) {
	if !builder.config.UseConfigFile {
		log.Error().Msg("No config file used")
		return nil, ErrConfigNotInitialized
	}
	return builder.configReader, nil
}

// SetServerConfig sets the server configuration and creates a new Server instance.
//
// It takes a ServerConfig struct as input and updates the internal configuration for the server.
// It then creates a new Server instance using the provided configuration and stores it in the Builder.
// On success, it returns nil. On error, it returns an error indicating the problem.
func (builder *Builder) SetServerConfig(config ServerConfig) error {
	builder.config.ServerConfig = &config
	svr, err := NewServer(builder.config.ServerConfig)
	if err != nil {
		log.Error().Err(err).Msg("Error creating server")
		return err
	}
	builder.server = svr
	return nil
}

// GetServer returns the Server instance associated with the Builder.
//
// It checks if the server is initialized and returns an error if not. Otherwise, it returns a pointer to the Server instance.
func (builder *Builder) GetServer() (*Server, error) {
	if builder.server == nil {
		return nil, ErrServerNotInitialized
	}
	return builder.server, nil
}
