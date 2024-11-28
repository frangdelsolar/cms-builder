package builder

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

const builderVersion = "1.2.3"

var log *Logger // Global variable for the logger instance

// Initializes the global logger instance with a default configuration.
//
// This function is automatically invoked when the package is imported.
//
// If an error occurs while initializing the logger, the program will panic.
func init() {
	// Make sure the logger is initialized
	var err error
	log, err = NewLogger(nil)
	if err != nil {
		fmt.Println("Error initializing logger:", err)
		panic(err)
	}
}

var (
	ErrBuilderConfigNotProvided   = errors.New("builder configuration not provided")
	ErrConfigReaderNotInitialized = errors.New("config reader not initialized")
	ErrLoggerNotInitialized       = errors.New("logger not initialized")
)

// BuilderConfig defines a nested configuration structure for various aspects of the application.
type BuilderConfig struct {
	configFile     *ReaderConfig   // Embedded configuration for the config file (optional)
	loggerConfig   *LoggerConfig   // Embedded configuration for the logger (optional)
	dbConfig       *DBConfig       // Embedded configuration for the database (optional)
	serverConfig   *ServerConfig   // Embedded configuration for the server (optional)
	firebaseConfig *FirebaseConfig // Embedded configuration for firebase (optional)
	UploaderConfig *UploaderConfig // Embedded configuration for the uploader (optional)
}

// Builder defines a central configuration and management structure for building applications.
type Builder struct {
	Config   *BuilderConfig // Pointer to the main configuration object
	reader   *ConfigReader  // Reference to the Viper instance used for configuration
	logger   *Logger        // Reference to the application's logger instance
	db       *Database      // Reference to the connected database instance (if applicable)
	server   *Server        // Reference to the created Server instance (if applicable)
	admin    *Admin         // Reference to the created Admin instance (if applicable)
	firebase *FirebaseAdmin // Reference to the created Firebase instance (if applicable)
}

// NewBuilderInput defines the input parameters for the Builder constructor.
type NewBuilderInput struct {
	ReadConfigFromFile bool   // Whether to read the configuration from a file
	ConfigFilePath     string // Path to the configuration file
	InitializeLogger   bool   // Whether to initialize the logger, needs readConfigFromFile to be true
	InitiliazeDB       bool   // Whether to initialize the database, needs readConfigFromFile to be true
	InitiliazeServer   bool   // Whether to initialize the server, needs readConfigFromFile to be true
	InitiliazeAdmin    bool   // Whether to initialize the admin, needs readConfigFromFile to be true
	InitiliazeFirebase bool   // Whether to initialize the firebase, needs readConfigFromFile to be true
	InitiliazeUploader bool   // Whether to initialize the uploader
}

// NewBuilder creates a new Builder instance.
func NewBuilder(input *NewBuilderInput) (*Builder, error) {

	if input == nil {
		return nil, ErrBuilderConfigNotProvided
	}

	builder := &Builder{
		Config: &BuilderConfig{
			configFile:     &ReaderConfig{},
			loggerConfig:   &LoggerConfig{},
			dbConfig:       &DBConfig{},
			serverConfig:   &ServerConfig{},
			firebaseConfig: &FirebaseConfig{},
		},
	}

	if !input.ReadConfigFromFile {
		return builder, nil
	}

	// Setup reader and read configuration data from file
	builder.InitConfigReader(&ReaderConfig{ConfigPath: input.ConfigFilePath})
	config, err := builder.GetConfigReader()
	if err != nil {
		return nil, err
	}

	// Logger
	if input.InitializeLogger {
		builder.InitLogger(&LoggerConfig{
			LogLevel:    config.GetString("logLevel"),
			WriteToFile: config.GetBool("logWriteToFile"),
			LogFilePath: config.GetString("logfilePath"),
		})
	} else {
		builder.InitLogger(nil) // Use default logger
	}

	log, _ = builder.GetLogger()

	log.Info().Msgf("Running Builder v%s", builderVersion)

	// Database
	if !input.InitiliazeDB {
		return builder, nil
	}

	dbConfig := &DBConfig{
		Path: config.GetString("dbFile"),
		URL:  config.GetString("dbUrl"),
	}
	builder.InitDatabase(dbConfig)

	db, err := builder.GetDatabase()
	if err != nil {
		log.Error().Err(err).Msg("Error getting database")
		return nil, err
	}

	log.Info().Bool("db nil", db == nil).Interface("config", dbConfig).Msg("Database initialized")

	// Server
	if !input.InitiliazeServer {
		return builder, nil
	}
	builder.initServer(&ServerConfig{
		Host:      config.GetString("serverHost"),
		Port:      config.GetString("serverPort"),
		CSRFToken: config.GetString("csrfToken"),
		Builder:   builder,
	})

	// Admin
	if input.InitiliazeAdmin {
		builder.initAdmin()
	}

	// Firebase
	if input.InitiliazeFirebase {
		builder.initFirebase(&FirebaseConfig{
			Secret: config.GetString("firebaseSecret"),
		})

		builder.initAuth()
	}

	// Uploader
	if input.InitiliazeUploader {
		builder.initUploader(&UploaderConfig{
			MaxSize:            config.GetInt64("uploaderMaxSize"),
			Authenticate:       config.GetBool("uploaderAuthenticate"),
			SupportedMimeTypes: config.GetStringSlice("uploaderSupportedMimeTypes"),
			Folder:             config.GetString("uploaderFolder"),
		})
	}

	return builder, nil
}

// InitConfigReader initializes the configuration reader based on the provided configuration file.
func (b *Builder) InitConfigReader(configFile *ReaderConfig) error {
	b.Config.configFile = configFile
	reader, err := NewConfigReader(b.Config.configFile)
	if err != nil {
		return err
	}
	b.reader = reader
	return nil
}

// GetConfigReader returns the configuration reader.
func (b *Builder) GetConfigReader() (*ConfigReader, error) {
	if b.reader == nil {
		return nil, ErrConfigReaderNotInitialized
	}
	return b.reader, nil
}

// InitLogger initializes the logger based on the provided configuration.
func (b *Builder) InitLogger(config *LoggerConfig) error {
	b.Config.loggerConfig = config
	logger, err := NewLogger(config)
	if err != nil {
		return err
	}
	b.logger = logger
	return nil
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

// InitDatabase initializes the database based on the provided configuration.
func (b *Builder) InitDatabase(config *DBConfig) error {
	log.Debug().Str("path", config.Path).Str("url", config.URL).Msg("Initializing database...")

	b.Config.dbConfig = config
	db, err := LoadDB(config)
	if err != nil {
		return err
	}
	b.db = db

	log.Info().Bool("db nil", b.db == nil).Msg("Database initialized")
	return nil
}

// GetDatabase returns the database instance associated with the Builder.
//
// It checks if the database is initialized and returns an error if not. Otherwise, it returns a pointer to the database instance.
func (b *Builder) GetDatabase() (*Database, error) {
	if b.db == nil {
		return nil, ErrDBNotInitialized
	}
	return b.db, nil
}

// initServer initializes the server based on the provided configuration.
func (b *Builder) initServer(config *ServerConfig) error {
	b.Config.serverConfig = config
	server, err := NewServer(config)
	if err != nil {
		return err
	}
	b.server = server
	return nil
}

// GetServer returns the server instance associated with the Builder.
//
// It checks if the server is initialized and returns an error if not. Otherwise, it returns a pointer to the server instance.
func (b *Builder) GetServer() (*Server, error) {
	if b.server == nil {
		return nil, ErrServerNotInitialized
	}
	return b.server, nil
}

// initAdmin initializes the admin based on the provided configuration.
func (b *Builder) initAdmin() {
	admin := NewAdmin(b.db, b.server, b)
	b.admin = admin
}

// GetAdmin returns the admin instance associated with the Builder.
//
// It checks if the admin is initialized and returns an error if not. Otherwise, it returns a pointer to the admin instance.
func (b *Builder) GetAdmin() (*Admin, error) {
	if b.admin == nil {
		return nil, ErrAdminNotInitialized
	}
	return b.admin, nil
}

// initFirebase initializes the Firebase Admin based on the provided configuration.
//
// It checks if the Firebase Admin is initialized and returns an error if not. Otherwise, it returns a pointer to the Firebase Admin instance.
func (b *Builder) initFirebase(config *FirebaseConfig) error {
	b.Config.firebaseConfig = config
	fb, err := NewFirebaseAdmin(config)
	if err != nil {
		return err
	}
	b.firebase = fb

	return nil
}

// GetFirebase returns the Firebase Admin instance associated with the Builder.
//
// It checks if the Firebase Admin is initialized and returns an error if not. Otherwise, it returns a pointer to the Firebase Admin instance.
func (b *Builder) GetFirebase() (*FirebaseAdmin, error) {
	if b.firebase == nil {
		return nil, ErrFirebaseNotInitialized
	}
	return b.firebase, nil
}

// initAuth initializes the auth system of the builder by registering the User app, and
// adding a route for user registration.
//
// It also registers two validators for the User model, EmailValidator and NameValidator.
//
// The route for user registration is added to the server with the name "register" and
// the path "/auth/register".
//
// If an error occurs while registering the User app, it logs the error and panics.
func (b *Builder) initAuth() {
	admin := b.admin
	userApp, err := admin.Register(&User{}, true)
	if err != nil {
		log.Error().Err(err).Msg("Error registering user app")
		panic(err)
	}

	userApp.RegisterValidator("email", ValidatorsList{RequiredValidator, EmailValidator})
	userApp.RegisterValidator("name", ValidatorsList{RequiredValidator})

	svr := b.server
	svr.AddRoute("/auth/register", b.RegisterUserController, "register", false)
}

// initUploader initializes the uploader by setting the configuration,
// registering the Upload app, and adding routes for file operations.
//
// It adds three primary routes:
//   - POST /file: For uploading new files
//   - DELETE /file/{id}/delete: For deleting files by ID
//   - GET /static/{path:.*}: For serving uploaded files
//
// If an error occurs while registering the Upload app, it logs the error and panics.
func (b *Builder) initUploader(config *UploaderConfig) {
	// Set the uploader configuration
	b.Config.UploaderConfig = config

	// Register the Upload app without authentication
	_, err := b.admin.Register(&Upload{}, false)
	if err != nil {
		log.Error().Err(err).Msg("Error registering upload app")
		panic(err)
	}

	// Define the base route for file operations
	route := "/file"

	// Add route for uploading new files
	b.server.AddRoute(
		route,
		b.GetUploadPostHandler(config),
		"file-new",
		true, // Requires authentication
	)

	// Add route for deleting files by ID
	b.server.AddRoute(
		route+"/{id}/delete",
		b.GetUploadDeleteHandler(config),
		"file-delete",
		true, // Requires authentication
	)

	// Add route for serving uploaded files
	b.server.AddRoute(
		"/static/{path:.*}",
		b.GetStaticHandler(config),
		"file-static",
		config.Authenticate, // Authentication based on config
	)

	// TODO: Remove - just for testing purposes
	b.server.AddRoute(
		route+"/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			id := mux.Vars(r)["id"]
			uploadApp, err := b.admin.GetApp("upload")
			if err != nil {
				SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
				return
			}
			userId := getRequestUserId(r, &uploadApp)

			var instance Upload
			// Query the database to find the record by ID
			result := b.db.FindById(id, &instance, userId, true)
			if result.Error != nil {
				SendJsonResponse(w, http.StatusInternalServerError, nil, result.Error.Error())
				return
			}

			if instance == (Upload{}) {
				SendJsonResponse(w, http.StatusNotFound, nil, "File not found")
				return
			}

			// Serve the file to the client
			http.ServeFile(w, r, instance.FilePath)
		},
		"file-see",
		true, // Requires authentication
	)
}
