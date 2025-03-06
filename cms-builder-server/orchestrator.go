package orchestrator

import (
	"fmt"

	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	auth "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/resources"
	cliPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	configPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/config"
	dbPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	dbResources "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/resources"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	file "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file/resources"
	loggerPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	rlResources "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/request-logger/resources"
	rmPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	schPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler"
	schResources "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/resources"
	svrPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	svrTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/types"
	storePkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store"
	storeConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store/constants"
	storeTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store/types"
)

const orchestratorVersion = "1.6.3"

type OrchestratorUsers struct {
	God       *authModels.User
	Admin     *authModels.User
	Scheduler *authModels.User
	System    *authModels.User
}

type Orchestrator struct {
	Config          *configPkg.ConfigReader
	DB              *dbTypes.DatabaseConnection
	FirebaseClient  *cliPkg.FirebaseManager
	Logger          *loggerTypes.Logger
	LoggerConfig    *loggerTypes.LoggerConfig
	ResourceManager *rmPkg.ResourceManager
	Scheduler       *schPkg.Scheduler
	Server          *svrTypes.Server
	Store           storeTypes.Store
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
		o.InitUsers,
		o.InitServer,
		o.InitStore,
		o.InitFiles,
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
	config, err := configPkg.NewConfigReader(&configPkg.ReaderConfig{
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
	config := &loggerTypes.LoggerConfig{
		LogLevel:    o.Config.GetString(EnvKeys.LogLevel),
		LogFilePath: o.Config.GetString(EnvKeys.LogFilePath),
		WriteToFile: o.Config.GetBool(EnvKeys.LogWriteToFile),
	}

	logger, err := loggerPkg.NewLogger(config)
	if err != nil {
		return fmt.Errorf("error initializing logger: %w", err)
	}
	o.Logger = logger
	o.LoggerConfig = config
	return nil
}

func (o *Orchestrator) InitDatabase() error {
	config := &dbTypes.DatabaseConfig{
		URL:    o.Config.GetString(EnvKeys.DbUrl),
		Path:   o.Config.GetString(EnvKeys.DbFile),
		Driver: o.Config.GetString(EnvKeys.DbDriver),
	}

	db, err := dbPkg.NewDatabaseConnection(config, o.Logger)
	if err != nil {
		return fmt.Errorf("error initializing database: %w", err)
	}
	o.DB = db

	o.Logger.Info().Interface("config", config).Msg("Database initialized")
	return nil
}

func (o *Orchestrator) InitFirebase() error {
	cfg := &cliPkg.FirebaseConfig{
		Secret: o.Config.GetString(EnvKeys.FirebaseSecret),
	}
	fb, err := cliPkg.NewFirebaseAdmin(cfg)
	if err != nil {
		return fmt.Errorf("error initializing firebase: %w", err)
	}
	o.FirebaseClient = fb

	o.Logger.Info().Msg("Firebase initialized")
	return nil
}

func (o *Orchestrator) InitResourceManager() error {
	o.Logger.Info().Msg("Initializing resource manager")
	o.ResourceManager = rmPkg.NewResourceManager(o.DB, o.Logger)
	return nil
}

func (o *Orchestrator) InitAuth() error {

	var getSystemUser = func() *authModels.User {
		return o.Users.System
	}

	resourceConfig := auth.SetupUserResource(o.FirebaseClient, o.DB, o.Logger, getSystemUser)
	_, err := o.ResourceManager.AddResource(resourceConfig)
	return err
}

func (o *Orchestrator) InitDatabaseLogger() error {
	resourceConfig := dbResources.SetupDBLoggerResource(o.ResourceManager, o.DB, o.Logger)
	_, err := o.ResourceManager.AddResource(resourceConfig)
	return err
}

func (o *Orchestrator) InitRequestLogger() error {
	resourceConfig := rlResources.SetupRequestLoggerResource(o.ResourceManager, o.DB, o.Logger)
	_, err := o.ResourceManager.AddResource(resourceConfig)
	return err
}

func (o *Orchestrator) InitFiles() error {

	storeConfig := o.Store.GetConfig()
	o.Logger.Info().Interface("storeConfig", storeConfig).Msg("Initializing store")

	fileConfig := file.SetupFileResource(o.ResourceManager, o.DB, o.Store, o.Logger, o.Config.GetString(EnvKeys.BaseUrl))
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
	config := &svrTypes.ServerConfig{
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

	server, err := svrPkg.NewServer(config, o.DB, o.Logger)
	if err != nil {
		return fmt.Errorf("error initializing server: %w", err)
	}
	o.Server = server

	return nil
}

func (o *Orchestrator) InitStore() error {
	storeType := o.Config.GetString(EnvKeys.StoreType)
	folder := "media/" + o.Config.GetString(EnvKeys.AppName)
	baseUrl := o.Config.GetString(EnvKeys.BaseUrl)

	storeConfig := &storeTypes.StoreConfig{
		MaxSize:            o.Config.GetInt64(EnvKeys.StoreMaxSize),
		SupportedMimeTypes: o.Config.GetStringSlice(EnvKeys.StoreSupportedMime),
		MediaFolder:        folder,
	}

	o.Logger.Info().Interface("storeConfig", storeConfig).Msg("Initializing store")

	var s storeTypes.Store
	var err error

	switch storeType {
	case string(storeConstants.StoreS3):
		s3Config := &storeTypes.S3Config{
			Bucket:    o.Config.GetString(EnvKeys.AwsBucket),
			Region:    o.Config.GetString(EnvKeys.AwsRegion),
			AccessKey: o.Config.GetString(EnvKeys.AwsAccessKeyId),
			SecretKey: o.Config.GetString(EnvKeys.AwsSecretAccessKey),
			Folder:    folder,
		}
		s, err = storePkg.NewS3Store(storeConfig, s3Config)
	case string(storeConstants.StoreLocal):
		s, err = storePkg.NewLocalStore(storeConfig, folder, baseUrl)
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
	taskConfig := schResources.SetupSchedulerTaskResource()
	_, err := o.ResourceManager.AddResource(taskConfig)
	if err != nil {
		return fmt.Errorf("error adding scheduler task resource: %w", err)
	}

	sch, err := schPkg.NewScheduler(o.DB, o.Users.Scheduler, o.Logger)
	if err != nil {
		return fmt.Errorf("error initializing scheduler: %w", err)
	}
	o.Scheduler = sch

	jobConfig := schResources.SetupSchedulerJobDefinitionResource(o.ResourceManager, o.DB, sch.JobRegistry)
	_, err = o.ResourceManager.AddResource(jobConfig)
	if err != nil {
		return fmt.Errorf("error adding scheduler job definition resource: %w", err)
	}
	return nil
}

func (o *Orchestrator) Run() error {
	o.Logger.Info().Msg("Starting Server")

	return svrPkg.RunServer(o.Server, o.ResourceManager.GetRoutes, o.Config.GetString(EnvKeys.BaseUrl))
}
