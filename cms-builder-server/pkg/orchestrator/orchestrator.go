package orchestrator

import (
	"fmt"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/config"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	dbLogger "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database-logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	requestLogger "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/request-logger"
	manager "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store"
)

const orchestratorVersion = "2.0.0"

type OrchestratorUsers struct {
	God       *models.User
	Admin     *models.User
	Scheduler *models.User
	System    *models.User
}

type Orchestrator struct {
	Config          *config.ConfigReader
	DB              *database.Database
	FirebaseClient  *clients.FirebaseManager
	Logger          *logger.Logger
	LoggerConfig    *logger.LoggerConfig
	ResourceManager *manager.ResourceManager
	Scheduler       *scheduler.Scheduler
	Server          *server.Server
	Store           store.Store
	Users           *OrchestratorUsers
}

func NewOrchestrator() (*Orchestrator, error) {
	o := &Orchestrator{}

	if err := o.init(); err != nil {
		return nil, err
	}

	return o, nil
}

func (o *Orchestrator) init() error {
	initializers := []func() error{
		o.InitConfigReader,
		o.InitLogger,
		o.InitDatabase,
		o.InitFirebase,
		o.InitResourceManager,
		o.InitAuth,
		o.InitDatabaseLogger,
		o.InitRequestLogger,
		o.InitFiles,
		o.InitUsers,
		o.InitServer,
		o.InitStore,
		o.InitScheduler,
	}

	for _, init := range initializers {
		if err := init(); err != nil {
			return err
		}
	}

	o.Logger.Info().
		Str("version", orchestratorVersion).
		Str("env", o.Config.GetString(EnvKeys.Environment)).
		Msg("Orchestrator initialized successfully")

	return nil
}

func (o *Orchestrator) InitConfigReader() error {
	config, err := config.NewConfigReader(&config.ReaderConfig{
		ReadEnv:  true,
		ReadFile: false,
	})
	if err != nil {
		return fmt.Errorf("error initializing config reader: %w", err)
	}
	o.Config = config
	return nil
}

func (o *Orchestrator) InitLogger() error {
	config := &logger.LoggerConfig{
		LogLevel:    o.Config.GetString(EnvKeys.LogLevel),
		LogFilePath: o.Config.GetString(EnvKeys.LogFilePath),
		WriteToFile: o.Config.GetBool(EnvKeys.LogWriteToFile),
	}

	logger, err := logger.NewLogger(config)
	if err != nil {
		return fmt.Errorf("error initializing logger: %w", err)
	}
	o.Logger = logger
	o.LoggerConfig = config
	return nil
}

func (o *Orchestrator) InitDatabase() error {
	config := &database.DBConfig{
		URL:    o.Config.GetString(EnvKeys.DbUrl),
		Path:   o.Config.GetString(EnvKeys.DbFile),
		Driver: o.Config.GetString(EnvKeys.DbDriver),
	}

	db, err := database.LoadDB(config, o.Logger)
	if err != nil {
		return fmt.Errorf("error initializing database: %w", err)
	}
	o.DB = db

	o.Logger.Info().Interface("config", config).Msg("Database initialized")
	return nil
}

func (o *Orchestrator) InitFirebase() error {
	cfg := &clients.FirebaseConfig{
		Secret: o.Config.GetString(EnvKeys.FirebaseSecret),
	}
	fb, err := clients.NewFirebaseAdmin(cfg)
	if err != nil {
		return fmt.Errorf("error initializing firebase: %w", err)
	}
	o.FirebaseClient = fb

	o.Logger.Info().Msg("Firebase initialized")
	return nil
}

func (o *Orchestrator) InitResourceManager() error {
	o.Logger.Info().Msg("Initializing resource manager")
	o.ResourceManager = manager.NewResourceManager(o.DB, o.Logger)
	return nil
}

func (o *Orchestrator) InitAuth() error {
	resourceConfig := auth.SetupUserResource(o.FirebaseClient, o.DB, o.Logger)
	_, err := o.ResourceManager.AddResource(resourceConfig)
	return err
}

func (o *Orchestrator) InitDatabaseLogger() error {
	resourceConfig := dbLogger.SetupDBLoggerResource(o.ResourceManager, o.DB, o.Logger)
	_, err := o.ResourceManager.AddResource(resourceConfig)
	return err
}

func (o *Orchestrator) InitRequestLogger() error {
	resourceConfig := requestLogger.SetupRequestLoggerResource(o.ResourceManager, o.DB, o.Logger)
	_, err := o.ResourceManager.AddResource(resourceConfig)
	return err
}

func (o *Orchestrator) InitFiles() error {
	fileConfig := file.SetupFileResource(o.ResourceManager, o.DB, o.Store, o.Logger)
	_, err := o.ResourceManager.AddResource(fileConfig)
	return err
}

func (o *Orchestrator) InitUsers() error {
	if err := o.SetupOrchestratorUsers(); err != nil {
		return fmt.Errorf("error setting up orchestrator users: %w", err)
	}
	o.Logger.Info().Msg("Users initialized successfully")
	return nil
}

func (o *Orchestrator) InitServer() error {
	config := &server.ServerConfig{
		Host:           o.Config.GetString(EnvKeys.ServerHost),
		Port:           o.Config.GetString(EnvKeys.ServerPort),
		CsrfToken:      o.Config.GetString(EnvKeys.CsrfToken),
		AllowedOrigins: o.Config.GetStringSlice(EnvKeys.CorsAllowedOrigins),
		GodToken:       o.Config.GetString(EnvKeys.GodToken),
		GodUser:        o.Users.God,
		SystemUser:     o.Users.System,
		Firebase:       o.FirebaseClient,
		LoggerConfig:   o.LoggerConfig,
	}

	server, err := server.NewServer(config, o.DB, o.Logger)
	if err != nil {
		return fmt.Errorf("error initializing server: %w", err)
	}
	o.Server = server

	return nil
}

func (o *Orchestrator) InitStore() error {
	storeType := o.Config.GetString(EnvKeys.StoreType)
	folder := o.Config.GetString(EnvKeys.UploaderFolder)
	baseUrl := o.Config.GetString(EnvKeys.BaseUrl)

	storeConfig := &store.StoreConfig{
		MaxSize:            o.Config.GetInt64(EnvKeys.UploaderMaxSize),
		SupportedMimeTypes: o.Config.GetStringSlice(EnvKeys.UploaderSupportedMime),
		Folder:             folder,
	}

	o.Logger.Info().Interface("storeConfig", storeConfig).Msg("Initializing store")

	var s store.Store
	var err error

	switch storeType {
	case string(store.StoreS3):
		s3Config := &store.S3Config{
			Bucket:    o.Config.GetString(EnvKeys.AwsBucket),
			Region:    o.Config.GetString(EnvKeys.AwsRegion),
			AccessKey: o.Config.GetString(EnvKeys.AwsAccessKeyId),
			SecretKey: o.Config.GetString(EnvKeys.AwsSecretAccessKey),
			Folder:    folder,
		}
		s, err = store.NewS3Store(storeConfig, s3Config)
	case string(store.StoreLocal):
		s, err = store.NewLocalStore(storeConfig, folder, baseUrl)
	default:
		return fmt.Errorf("unknown store type: %s", storeType)
	}

	if err != nil {
		return fmt.Errorf("error initializing store: %w", err)
	}
	o.Store = s

	return nil
}

func (o *Orchestrator) InitScheduler() error {
	taskConfig := scheduler.SetupSchedulerTaskResource()
	o.ResourceManager.AddResource(taskConfig)

	jobConfig := scheduler.SetupSchedulerJobDefinitionResource()
	o.ResourceManager.AddResource(jobConfig)

	sch, err := scheduler.NewScheduler(o.DB, o.Users.Scheduler, o.Logger)
	if err != nil {
		return fmt.Errorf("error initializing scheduler: %w", err)
	}
	o.Scheduler = sch

	return nil
}

func (o *Orchestrator) Run() error {
	o.Logger.Info().Msg("Starting Server")
	routes := o.ResourceManager.GetRoutes()

	for _, route := range routes {
		o.Server.AddRoute(route)
	}

	return o.Server.Run()
}
