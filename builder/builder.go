package builder

import (
	"errors"
	"fmt"
)

const builderVersion = "1.3.0"

// ConfigKeys define the keys used in the configuration file
type ConfigKeys struct {
	Environment           string `json:"environment"`           // Environment where the app is running
	LogLevel              string `json:"logLevel"`              // Log level
	LogFilePath           string `json:"logFilePath"`           // File path for logging
	LogWriteToFile        string `json:"logWriteToFile"`        // Write logs to file
	DbFile                string `json:"dbFile"`                // Database file
	DbUrl                 string `json:"dbUrl"`                 // Database URL
	ServerHost            string `json:"serverHost"`            // Server host
	ServerPort            string `json:"serverPort"`            // Server port
	CsrfToken             string `json:"csrfToken"`             // CSRF token
	FirebaseSecret        string `json:"firebaseSecret"`        // Firebase secret
	FirebaseApiKey        string `json:"firebaseApiKey"`        // Firebase API key
	UploaderMaxSize       string `json:"uploaderMaxSize"`       // Uploader max size in MB
	UploaderAuthenticate  string `json:"uploaderAuthenticate"`  // Whether files will be public or private accessible
	UploaderSupportedMime string `json:"uploaderSupportedMime"` // Supported mime types for uploaded files
	UploaderFolder        string `json:"uploaderFolder"`        // Uploader folder
	StaticPath            string `json:"staticPath"`            // Static path
	BaseUrl               string `json:"baseUrl"`               // where the app is running
}

// EnvKeys are the keys used in the configuration file
var EnvKeys = ConfigKeys{
	Environment:           "ENVIRONMENT",
	LogLevel:              "LOG_LEVEL",
	LogFilePath:           "LOG_FILE_PATH",
	LogWriteToFile:        "LOG_WRITE_TO_FILE",
	DbFile:                "DB_FILE",
	DbUrl:                 "DB_URL",
	ServerHost:            "SERVER_HOST",
	ServerPort:            "SERVER_PORT",
	CsrfToken:             "CSRF_TOKEN",
	FirebaseSecret:        "FIREBASE_SECRET",
	FirebaseApiKey:        "FIREBASE_API_KEY",
	UploaderMaxSize:       "UPLOADER_MAX_SIZE",
	UploaderAuthenticate:  "UPLOADER_AUTHENTICATE",
	UploaderSupportedMime: "UPLOADER_SUPPORTED_MIME_TYPES",
	UploaderFolder:        "UPLOADER_FOLDER",
	StaticPath:            "STATIC_PATH",
	BaseUrl:               "BASE_URL",
}

// defaultConfig defines the default values for the configuration
var DefaultEnvValues = ConfigKeys{
	Environment:           "development",
	LogLevel:              "debug",
	LogFilePath:           "logs/default.log",
	LogWriteToFile:        "true",
	DbFile:                "database.db",
	DbUrl:                 "",
	ServerHost:            "0.0.0.0",
	ServerPort:            "80",
	CsrfToken:             "someToken",
	FirebaseSecret:        "encoded64-token-thisIsGeneratedByEncodingFirebaseConfigFile",
	FirebaseApiKey:        "apikeyProvidedByFirebaseClient",
	UploaderMaxSize:       "5", // in MB
	UploaderAuthenticate:  "true",
	UploaderSupportedMime: "image/*",
	UploaderFolder:        "uploads",
	StaticPath:            "static",
	BaseUrl:               "http://0.0.0.0:80",
}

type BuilderErrors struct {
	LoggerNotInitialized       error
	ConfigReaderNotInitialized error
	ConfigurationNotProvided   error
}

var builderErr = BuilderErrors{
	LoggerNotInitialized:       errors.New("logger not initialized"),
	ConfigReaderNotInitialized: errors.New("config reader not initialized"),
	ConfigurationNotProvided:   errors.New("configuration not provided"),
}

var log *Logger          // Global variable for the logger instance
var config *ConfigReader // Global variable for the config reader

// init initializes the global logger and config reader instances with default values.
// It checks if the logger and config reader are initialized and panics if not.
// It also logs the version and environment at the info level.
func init() {
	var err error
	log, err = NewLogger(&LoggerConfig{
		LogLevel:    DefaultEnvValues.LogLevel,
		WriteToFile: DefaultEnvValues.LogWriteToFile == "true",
		LogFilePath: DefaultEnvValues.LogFilePath,
	})
	if err != nil {
		fmt.Println("Error initializing logger:", err)
		panic(builderErr.LoggerNotInitialized)
	}
	config, err = NewConfigReader(&ReaderConfig{
		ReadEnv:  true,
		ReadFile: false,
	})
	if err != nil {
		fmt.Println("Error initializing config reader:", err)
		panic(builderErr.ConfigReaderNotInitialized)
	}
}

// Builder defines a central configuration and management structure for building applications.
type Builder struct {
	Admin    *Admin         // Reference to the created Admin instance
	Config   *ConfigReader  // Reference to the Viper instance used for configuration
	DB       *Database      // Reference to the connected database instance
	Firebase *FirebaseAdmin // Reference to the created Firebase instance
	Logger   *Logger        // Reference to the application's logger instance
	Server   *Server        // Reference to the created Server instance
}

// NewBuilderInput defines the input parameters for the Builder constructor.
type NewBuilderInput struct {
	ReadConfigFromEnv    bool   // Whether to read the configuration from environment variables
	ReadConfigFromFile   bool   // Whether to read the configuration from a file
	ReaderConfigFilePath string // Path to the configuration file
}

// NewBuilder creates a new Builder instance.
func NewBuilder(input *NewBuilderInput) (*Builder, error) {

	if input == nil {
		return nil, builderErr.ConfigurationNotProvided
	}

	builder := &Builder{}

	err := builder.InitConfigReader(input)
	if err != nil {
		log.Err(err).Msg("Error initializing config reader")
		return nil, builderErr.ConfigReaderNotInitialized
	}

	// Make configurations available for other modules
	config = builder.Config

	// Logger
	err = builder.InitLogger()
	if err != nil {
		log.Err(err).Msg("Error initializing logger")
		return nil, builderErr.LoggerNotInitialized
	}

	// Make logger available for other modules
	log = builder.Logger

	log.Info().
		Str("version", builderVersion).
		Str("env", config.GetString(EnvKeys.Environment)).
		Msg("Running Builder")

	// Database
	err = builder.InitDatabase()
	if err != nil {
		log.Err(err).Msg("Error initializing database")
		return nil, err
	}

	// Server
	err = builder.InitServer()
	if err != nil {
		log.Err(err).Msg("Error initializing server")
		return nil, err
	}

	// Admin
	builder.InitAdmin()

	// Firebase
	err = builder.InitFirebase()
	if err != nil {
		log.Err(err).Msg("Error initializing firebase")
		return nil, err
	}

	err = builder.InitAuth()
	if err != nil {
		log.Err(err).Msg("Error initializing auth")
		return nil, err
	}

	// Uploader
	err = builder.InitUploader()
	if err != nil {
		log.Err(err).Msg("Error initializing uploader")
		return nil, err
	}

	return builder, nil
}

// InitConfigReader initializes the config reader based on the provided configuration.
//
// It takes a NewBuilderInput pointer as a parameter, which specifies whether to read the configuration from environment variables or a file.
// It returns an error if the config reader cannot be initialized.
func (b *Builder) InitConfigReader(cfg *NewBuilderInput) error {
	readerCfg := &ReaderConfig{
		ConfigFilePath: cfg.ReaderConfigFilePath,
		ReadEnv:        cfg.ReadConfigFromEnv,
		ReadFile:       cfg.ReadConfigFromFile,
	}

	reader, err := NewConfigReader(readerCfg)
	if err != nil {
		return err
	}
	b.Config = reader
	return nil
}

// InitLogger initializes the logger for the Builder instance.
//
// It retrieves the log configuration from the environment variables and uses it to create a new logger.
// If the logger initialization fails, it returns an error. On success, the logger is assigned to the Builder instance.
func (b *Builder) InitLogger() error {
	config := &LoggerConfig{
		LogLevel:    config.GetString(EnvKeys.LogLevel),
		LogFilePath: config.GetString(EnvKeys.LogFilePath),
		WriteToFile: config.GetBool(EnvKeys.LogWriteToFile),
	}

	logger, err := NewLogger(config)
	if err != nil {
		return err
	}
	b.Logger = logger
	return nil
}

// InitDatabase initializes the database based on the provided configuration.
//
// It determines the database configuration to use based on the environment. If the environment is "development" or "test", it uses the file-based configuration. Otherwise, it uses the URL-based configuration.
// It then calls LoadDB to initialize the database and assigns the result to the Builder's DB field.
// If there is an error initializing the database, it returns the error. On success, it logs a message indicating that the database has been initialized and returns nil.
func (b *Builder) InitDatabase() error {
	dbConfig := &DBConfig{}

	env := config.GetString(EnvKeys.Environment)

	if env == "development" || env == "test" {
		dbConfig.Path = config.GetString(EnvKeys.DbFile)
	} else {
		dbConfig.URL = config.GetString(EnvKeys.DbUrl)
	}

	log.Info().Str("path", dbConfig.Path).Str("url", dbConfig.URL).Msg("Initializing database...")

	db, err := LoadDB(dbConfig)
	if err != nil {
		return err
	}
	b.DB = db

	log.Info().Msg("Database initialized")
	return nil
}

// InitServer initializes the server based on the provided configuration.
//
// It takes the server host, port, and CSRF token from the environment variables and uses them to create a new server.
// The server is assigned to the Builder instance.
// If there is an error initializing the server, it returns the error. On success, it returns nil.
func (b *Builder) InitServer() error {
	server, err := NewServer(&ServerConfig{
		Host:      config.GetString(EnvKeys.ServerHost),
		Port:      config.GetString(EnvKeys.ServerPort),
		CSRFToken: config.GetString(EnvKeys.CsrfToken),
		Builder:   b,
	})
	if err != nil {
		return err
	}
	b.Server = server
	return nil
}

// InitAdmin initializes the admin based on the provided configuration.
func (b *Builder) InitAdmin() {
	b.Admin = NewAdmin(b)
}

// InitFirebase initializes the Firebase Admin based on the provided configuration.
//
// It takes the secret for the Firebase Admin from the environment variables, creates a new Firebase Admin instance, and assigns it to the Builder instance.
// If there is an error initializing the Firebase Admin, it returns the error. On success, it returns nil.
func (b *Builder) InitFirebase() error {
	cfg := &FirebaseConfig{
		Secret: config.GetString(EnvKeys.FirebaseSecret),
	}
	fb, err := NewFirebaseAdmin(cfg)
	if err != nil {
		return err
	}
	b.Firebase = fb

	return nil
}

// InitAuth initializes the auth system of the builder by registering the User app, and
// adding a route for user registration.
//
// It also registers two validators for the User model, EmailValidator and NameValidator.
//
// The route for user registration is added to the server with the name "register" and
// the path "/auth/register".
//
// If an error occurs while registering the User app, it logs the error and panics.
func (b *Builder) InitAuth() error {
	admin := b.Admin
	userApp, err := admin.Register(&User{}, true)
	if err != nil {
		log.Error().Err(err).Msg("Error registering user app")
		return err
	}

	userApp.RegisterValidator("email", ValidatorsList{RequiredValidator, EmailValidator})
	userApp.RegisterValidator("name", ValidatorsList{RequiredValidator})

	svr := b.Server
	svr.AddRoute("/auth/register", b.RegisterUserController, "register", false)
	return nil
}

// InitUploader initializes the uploader by setting the configuration,
// registering the Upload app, and adding routes for file operations.
//
// It adds three primary routes:
//   - POST /file: For uploading new files
//   - DELETE /file/{id}/delete: For deleting files by ID
//   - GET /static/{path:.*}: For serving uploaded files
//
// If an error occurs while registering the Upload app, it logs the error and panics.
func (b *Builder) InitUploader() error {
	staticPath := config.GetString(EnvKeys.StaticPath)
	if staticPath == "" {
		staticPath = DefaultEnvValues.StaticPath
	}

	cfg := &UploaderConfig{
		MaxSize:            config.GetInt64(config.GetString(EnvKeys.UploaderMaxSize)),
		SupportedMimeTypes: config.GetStringSlice(EnvKeys.UploaderSupportedMime),
		Folder:             config.GetString(config.GetString(EnvKeys.UploaderFolder)),
		StaticPath:         staticPath,
	}

	// Register the Upload app without authentication
	_, err := b.Admin.Register(&Upload{}, false)
	if err != nil {
		log.Error().Err(err).Msg("Error registering upload app")
		return err
	}

	// Define the base route for file operations
	route := "/file"

	// Add route for uploading new files
	b.Server.AddRoute(
		route,
		b.GetUploadPostHandler(cfg),
		"file-new",
		true, // Requires authentication
	)

	// Add route for deleting files by ID
	b.Server.AddRoute(
		route+"/{id}/delete",
		b.GetUploadDeleteHandler(cfg),
		"file-delete",
		true, // Requires authentication
	)

	// Add route for serving uploaded files
	b.Server.AddRoute(
		"/"+staticPath+"/{path:.*}",
		b.GetStaticHandler(cfg),
		"file-static",
		false, // Authentication based on config
	)

	return nil
}
