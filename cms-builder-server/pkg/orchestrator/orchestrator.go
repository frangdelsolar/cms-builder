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
	var err error
	o := &Orchestrator{}

	// Respect the order of initialization
	err = o.InitConfigReader()
	if err != nil {
		fmt.Println("Error initializing config reader:", err)
		return nil, err
	}

	err = o.InitLogger() // config
	if err != nil {
		fmt.Println("Error initializing logger:", err)
		return nil, err
	}

	environment := o.Config.GetString(EnvKeys.Environment)
	o.Logger.Info().
		Str("version", orchestratorVersion).
		Str("env", environment).
		Msg("Orchestrator running")

	err = o.InitDatabase() // config and logger
	if err != nil {
		o.Logger.Error().Err(err).Msg("Error initializing database")
		return nil, err
	}

	err = o.InitFirebase() // config, db and logger
	if err != nil {
		o.Logger.Error().Err(err).Msg("Error initializing firebase")
		return nil, err
	}

	o.InitResourceManager()

	// // Init Models
	o.InitAuth()
	o.InitDatabaseLogger()
	o.InitRequestLogger()
	o.InitFiles()

	err = o.InitUsers() // config, db, firebase, history, amdin and logger
	if err != nil {
		o.Logger.Error().Err(err).Msg("Error initializing users")
		return nil, err
	}

	// err = o.InitServer()
	// if err != nil {
	// 	o.Logger.Error().Err(err).Msg("Error initializing server")
	// 	return nil, err
	// }

	// err = o.InitStore()
	// if err != nil {
	// 	o.Logger.Err(err).Msg("Error initializing store")
	// 	return nil, err
	// }

	// err = o.InitScheduler()
	// if err != nil {
	// 	o.Logger.Err(err).Msg("Error initializing scheduler")
	// 	return nil, err
	// }

	return o, nil
}

func (o *Orchestrator) InitFiles() {
	fileConfig := file.SetupFileResource(o.ResourceManager, o.DB, o.Store, o.Logger)
	o.ResourceManager.AddResource(fileConfig)
}

func (o *Orchestrator) InitScheduler() error {
	taskConfig := scheduler.SetupSchedulerTaskResource()
	o.ResourceManager.AddResource(taskConfig)

	jobConfig := scheduler.SetupSchedulerJobDefinitionResource()
	o.ResourceManager.AddResource(jobConfig)

	sch, err := scheduler.NewScheduler(o.DB, o.Users.Scheduler, o.Logger)
	if err != nil {
		return err
	}

	o.Scheduler = sch
	return nil
}

func (o *Orchestrator) InitRequestLogger() {
	resourceConfig := requestLogger.SetupRequestLoggerResource(o.ResourceManager, o.DB, o.Logger)
	o.ResourceManager.AddResource(resourceConfig)
}

func (o *Orchestrator) InitDatabaseLogger() {
	resourceConfig := dbLogger.SetupDBLoggerResource(o.ResourceManager, o.DB, o.Logger)
	o.ResourceManager.AddResource(resourceConfig)
}

func (o *Orchestrator) InitAuth() {
	o.Logger.Info().Msg("Initializing user resource")

	resourceConfig := auth.SetupUserResource(o.FirebaseClient, o.DB, o.Logger)
	o.ResourceManager.AddResource(resourceConfig)
}

func (o *Orchestrator) InitResourceManager() {
	o.Logger.Info().Msg("Initializing resource manager")
	o.ResourceManager = manager.NewResourceManager(o.DB, o.Logger)
}

func (o *Orchestrator) InitStore() error {
	var s store.Store

	storeType := o.Config.GetString(EnvKeys.StoreType)
	folder := o.Config.GetString(EnvKeys.UploaderFolder)
	baseUrl := o.Config.GetString(EnvKeys.BaseUrl)

	storeConfig := &store.StoreConfig{
		MaxSize:            o.Config.GetInt64(EnvKeys.UploaderMaxSize),
		SupportedMimeTypes: o.Config.GetStringSlice(EnvKeys.UploaderSupportedMime),
		Folder:             o.Config.GetString(EnvKeys.UploaderFolder),
	}

	o.Logger.Info().Interface("storeConfig", storeConfig).Msg("Initializing store")

	switch storeType {
	case string(store.StoreS3):

		s3Config := &store.S3Config{
			Bucket:    o.Config.GetString(EnvKeys.AwsBucket),
			Region:    o.Config.GetString(EnvKeys.AwsRegion),
			AccessKey: o.Config.GetString(EnvKeys.AwsAccessKeyId),
			SecretKey: o.Config.GetString(EnvKeys.AwsSecretAccessKey),
			Folder:    folder,
		}

		s3Store, err := store.NewS3Store(storeConfig, s3Config)
		if err != nil {
			o.Logger.Error().Err(err).Msg("Error initializing S3 store")
			return err
		}
		s = s3Store
	case string(store.StoreLocal):
		localStore, err := store.NewLocalStore(storeConfig, folder, baseUrl)
		if err != nil {
			o.Logger.Error().Err(err).Msg("Error initializing local store")
			return err
		}
		s = localStore
	default:
		return fmt.Errorf("unknown store type: %s", storeType)
	}

	o.Store = s
	return nil
}

func (o *Orchestrator) InitUsers() error {
	return o.SetupOrchestratorUsers()
}

func (o *Orchestrator) InitFirebase() error {
	cfg := &clients.FirebaseConfig{
		Secret: o.Config.GetString(EnvKeys.FirebaseSecret),
	}
	fb, err := clients.NewFirebaseAdmin(cfg)
	if err != nil {
		return err
	}
	o.FirebaseClient = fb

	o.Logger.Info().Msg("Firebase initialized")
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
	}

	server, err := server.NewServer(config, o.DB, o.Logger)
	if err != nil {
		o.Logger.Error().Err(err).Msg("Error initializing server")
		return err
	}

	o.Server = server

	return nil
}

func (o *Orchestrator) InitDatabase() error {

	config := &database.DBConfig{}

	config.URL = o.Config.GetString(EnvKeys.DbUrl)
	config.Path = o.Config.GetString(EnvKeys.DbFile)
	config.Driver = o.Config.GetString(EnvKeys.DbDriver)

	db, err := database.LoadDB(config, o.Logger)
	if err != nil {
		o.Logger.Error().Err(err).Msg("Error initializing database")
		return err
	}

	o.DB = db

	o.Logger.Info().Interface("config", config).Msg("Database initialized")

	return nil
}

func (o *Orchestrator) InitConfigReader() error {
	fmt.Println("Initializing config reader")

	config, err := config.NewConfigReader(
		&config.ReaderConfig{
			ReadEnv:  true,
			ReadFile: false,
		},
	)
	if err != nil {
		fmt.Println("Error initializing config reader:", err)
		return err
	}
	o.Config = config
	return nil
}

func (o *Orchestrator) InitLogger() error {
	fmt.Println("Initializing logger")

	config := &logger.LoggerConfig{
		LogLevel:    o.Config.GetString(EnvKeys.LogLevel),
		LogFilePath: o.Config.GetString(EnvKeys.LogFilePath),
		WriteToFile: o.Config.GetBool(EnvKeys.LogWriteToFile),
	}

	logger, err := logger.NewLogger(config)
	if err != nil {
		fmt.Println("Error initializing logger:", err)
		return err
	}
	o.Logger = logger
	o.LoggerConfig = config
	return nil
}
