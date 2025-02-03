package builder

import (
	"errors"
	"fmt"
	"net/http"
)

const builderVersion = "1.4.1"

// ConfigKeys define the keys used in the configuration file
type ConfigKeys struct {
	AppName               string `json:"appName"`               // App name
	AdminName             string `json:"adminName"`             // Admin name
	AdminEmail            string `json:"adminEmail"`            // Admin email
	AdminPassword         string `json:"adminPassword"`         // Admin password
	CorsAllowedOrigins    string `json:"corsAllowedOrigins"`    // CORS allowed origins
	Environment           string `json:"environment"`           // Environment where the app is running
	LogLevel              string `json:"logLevel"`              // Log level
	LogFilePath           string `json:"logFilePath"`           // File path for logging
	LogWriteToFile        string `json:"logWriteToFile"`        // Write logs to file
	Domain                string `json:"domain"`                // Domain
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
	StoreType             string `json:"storeType"`             // Uploader store type
	AwsBucket             string `json:"awsBucket"`             // AWS bucket
	AwsRegion             string `json:"awsRegion"`             // AWS region
	AwsSecretAccessKey    string `json:"awsSecretAccessKey"`    // AWS secret access key
	AwsAccessKeyId        string `json:"awsAccessKeyId"`        // AWS access key id
	BaseUrl               string `json:"baseUrl"`               // where the app is running
}

// EnvKeys are the keys used in the configuration file
var EnvKeys = ConfigKeys{
	AppName:               "APP_NAME",
	AdminName:             "ADMIN_NAME",
	AdminEmail:            "ADMIN_EMAIL",
	AdminPassword:         "ADMIN_PASSWORD",
	CorsAllowedOrigins:    "CORS_ALLOWED_ORIGINS",
	Environment:           "ENVIRONMENT",
	LogLevel:              "LOG_LEVEL",
	LogFilePath:           "LOG_FILE_PATH",
	LogWriteToFile:        "LOG_WRITE_TO_FILE",
	Domain:                "DOMAIN",
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
	StoreType:             "STORE_TYPE",
	AwsBucket:             "AWS_BUCKET",
	AwsRegion:             "AWS_REGION",
	AwsSecretAccessKey:    "AWS_SECRET_ACCESS_KEY",
	AwsAccessKeyId:        "AWS_ACCESS_KEY_ID",
	BaseUrl:               "BASE_URL",
}

// defaultConfig defines the default values for the configuration
var DefaultEnvValues = ConfigKeys{
	AppName:               "Builder",
	AdminName:             "Admin",
	AdminEmail:            "admin@admin.com",
	AdminPassword:         "admin123admin",
	CorsAllowedOrigins:    "*",
	Environment:           "development",
	LogLevel:              "debug",
	LogFilePath:           "logs/default.log",
	LogWriteToFile:        "true",
	Domain:                "localhost",
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
	StoreType:             "local",
	AwsBucket:             "s3://something",
	AwsRegion:             "us-east-1",
	AwsSecretAccessKey:    "secretAccessKey",
	AwsAccessKeyId:        "accessKeyId",
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
	Admin     *Admin         // Reference to the created Admin instance
	Config    *ConfigReader  // Reference to the Viper instance used for configuration
	DB        *Database      // Reference to the connected database instance
	Firebase  *FirebaseAdmin // Reference to the created Firebase instance
	Logger    *Logger        // Reference to the application's logger instance
	Server    *Server        // Reference to the created Server instance
	Store     Store          // Reference to the created Store instance
	Scheduler *Scheduler     // Reference to the created Scheduler instance
}

// NewBuilderInput defines the input parameters for the Builder constructor.
type NewBuilderInput struct {
	ReadConfigFromEnv    bool   // Whether to read the configuration from environment variables
	ReadConfigFromFile   bool   // Whether to read the configuration from a file
	ReaderConfigFilePath string // Path to the configuration file
	InitializeScheduler  bool   // Whether to initialize the scheduler
}

// NewBuilder creates a new Builder instance.
func NewBuilder(input *NewBuilderInput) (*Builder, error) {
	if input == nil {
		return nil, builderErr.ConfigurationNotProvided
	}

	b := &Builder{}

	err := b.InitConfigReader(input)
	if err != nil {
		log.Err(err).Msg("Error initializing config reader")
		return nil, builderErr.ConfigReaderNotInitialized
	}

	// Make configurations available for other modules
	config = b.Config

	// Logger
	err = b.InitLogger()
	if err != nil {
		log.Err(err).Msg("Error initializing logger")
		return nil, builderErr.LoggerNotInitialized
	}

	// Make logger available for other modules
	log = b.Logger

	log.Info().
		Str("version", builderVersion).
		Str("env", config.GetString(EnvKeys.Environment)).
		Msg("Running Builder")

	// Database
	err = b.InitDatabase()
	if err != nil {
		log.Err(err).Msg("Error initializing database")
		return nil, err
	}

	// Server
	err = b.InitServer()
	if err != nil {
		log.Err(err).Msg("Error initializing server")
		return nil, err
	}

	// Admin
	b.InitAdmin()

	// History
	err = b.InitHistory()
	if err != nil {
		log.Err(err).Msg("Error initializing history")
		return nil, err
	}

	// Firebase
	err = b.InitFirebase()
	if err != nil {
		log.Err(err).Msg("Error initializing firebase")
		return nil, err
	}

	err = b.InitAuth()
	if err != nil {
		log.Err(err).Msg("Error initializing auth")
		return nil, err
	}

	// Store
	err = b.InitStore()
	if err != nil {
		log.Err(err).Msg("Error initializing store")
		return nil, err
	}

	// Uploader
	err = b.InitUploader()
	if err != nil {
		log.Err(err).Msg("Error initializing uploader")
		return nil, err
	}

	if input.InitializeScheduler {
		err = b.InitScheduler()
		if err != nil {
			log.Err(err).Msg("Error initializing scheduler")
			return nil, err
		}
	}

	err = b.RegisterAdminUser()
	if err != nil {
		log.Err(err).Msg("Error registering admin user")
		return nil, err
	}

	return b, nil
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

	if env == "production" || env == "stage" || env == "docker" {
		dbConfig.URL = config.GetString(EnvKeys.DbUrl)
	} else {
		dbConfig.Path = config.GetString(EnvKeys.DbFile)
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

	// FIXME: This doesn't look good to me!
	permissions := RolePermissionMap{
		AdminRole:   AllAllowedAccess,
		VisitorRole: AllAllowedAccess,
	}

	userApp, err := admin.Register(&User{}, true, permissions)
	if err != nil {
		log.Error().Err(err).Msg("Error registering user app")
		return err
	}

	err = userApp.RegisterValidator("email", ValidatorsList{RequiredValidator, EmailValidator})
	if err != nil {
		log.Error().Err(err).Msg("Error registering email validator")
		return err
	}

	err = userApp.RegisterValidator("name", ValidatorsList{RequiredValidator})
	if err != nil {
		log.Error().Err(err).Msg("Error registering name validator")
		return err
	}

	err = userApp.RegisterValidator("roles", ValidatorsList{RequiredValidator})
	if err != nil {
		log.Error().Err(err).Msg("Error registering roles validator")
		return err
	}

	// FIXME: Implement these
	// userApp.Api.Delete = ...
	// userApp.Api.Update = ...
	// userApp.Api.Create = ...
	// userApp.Api.List = ...
	// userApp.Api.Detail = ...

	// FIXME: Create tests so that no user can edit another user unless authorized
	// No user should be able to delete other users
	// No user should be able to delete users, including himself

	svr := b.Server
	svr.AddRoute("/auth/register", b.RegisterVisitorController, "register", false, http.MethodPost, RegisterUserInput{})
	return nil
}

// InitStore initializes the store based on the provided configuration.
//
// It sets the configuration for the store, registers the correct store implementation,
// and assigns the result to the Builder's Store field.
// If an error occurs while initializing the store, it returns the error. On success, it returns nil.
func (b *Builder) InitStore() error {
	var store Store
	switch config.GetString(EnvKeys.StoreType) {
	case string(StoreS3):
		s3Store, err := NewS3Store(config.GetString(EnvKeys.UploaderFolder))
		if err != nil {
			log.Error().Err(err).Msg("Error initializing S3 store")
			return err
		}
		store = s3Store
	case string(StoreLocal):
		store = NewLocalStore(config.GetString(EnvKeys.UploaderFolder))
	default:
		return errors.New("unknown store type: " + config.GetString(EnvKeys.StoreType))
	}

	b.Store = store
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
	cfg := &UploaderConfig{
		MaxSize:            config.GetInt64(config.GetString(EnvKeys.UploaderMaxSize)),
		SupportedMimeTypes: config.GetStringSlice(EnvKeys.UploaderSupportedMime),
		Folder:             config.GetString(config.GetString(EnvKeys.UploaderFolder)),
		StaticPath:         "private/file/",
	}

	permissions := RolePermissionMap{
		AdminRole:   AllAllowedAccess,
		VisitorRole: AllAllowedAccess,
	}

	// Register the Upload app without authentication
	_, err := b.Admin.Register(&Upload{}, false, permissions)
	if err != nil {
		log.Error().Err(err).Msg("Error registering upload app")
		return err
	}

	// Define the base route for file operations
	route := "/file"

	// Add route for uploading new files
	b.Server.AddRoute(
		route+"/upload",
		b.GetFilePostHandler(cfg),
		"file-new",
		true, // Requires authentication
		http.MethodPost,
		"form with file",
	)

	// Add route for deleting files by ID
	b.Server.AddRoute(
		route+"/{id}/delete",
		b.GetFileDeleteHandler(cfg),
		"file-delete",
		true, // Requires authentication
		http.MethodDelete,
		nil,
	)

	// Download route
	b.Server.AddRoute(
		route+"/{path:.*}",
		b.GetStaticHandler(cfg),
		"file-static",
		true,
		http.MethodGet,
		nil,
	)

	return nil
}

func (b *Builder) InitScheduler() error {

	permissions := RolePermissionMap{
		AdminRole:     AllAllowedAccess,
		SchedulerRole: AllAllowedAccess,
	}

	_, err := b.Admin.Register(&SchedulerJobDefinition{}, false, permissions)
	if err != nil {
		log.Error().Err(err).Msg("Error registering job definition app")
		return err
	}

	_, err = b.Admin.Register(&JobFrequency{}, false, permissions)
	if err != nil {
		log.Error().Err(err).Msg("Error registering job frequency app")
		return err
	}

	_, err = b.Admin.Register(&SchedulerTask{}, false, permissions)
	if err != nil {
		log.Error().Err(err).Msg("Error registering scheduler task app")
		return err
	}

	s, err := NewScheduler(b)
	if err != nil {
		log.Error().Err(err).Msg("Error creating scheduler")
		return err
	}

	b.Scheduler = s

	return nil
}

func (b *Builder) RegisterAdminUser() error {
	userData := &RegisterUserInput{
		Name:     config.GetString(EnvKeys.AdminName),
		Email:    config.GetString(EnvKeys.AdminEmail),
		Password: config.GetString(EnvKeys.AdminPassword),
	}

	user, err := b.CreateUserWithRole(*userData, AdminRole, true)
	if err != nil {
		log.Error().Err(err).Msg("Error creating admin user")
		return err
	}

	if user == nil {
		log.Error().Msg("Error creating admin user")
		return err
	}

	return nil
}

func (b *Builder) InitHistory() error {
	permissions := RolePermissionMap{
		AdminRole: AllAllowedAccess,
	}

	_, err := b.Admin.Register(&HistoryEntry{}, false, permissions)
	if err != nil {
		log.Error().Err(err).Msg("Error registering history app")
		return err
	}

	return nil
}
