package orchestrator

import (
	"fmt"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/config"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	dbLogger "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database-logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	requestLogger "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/request-logger"
	manager "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store"
	"github.com/google/uuid"
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
	Store           *store.Store
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

	o.InitAuth()

	o.InitDatabaseLogger() // Needs ResourceManager

	o.InitRequestLogger() // Needs ResourceManager

	err = o.InitUsers() // config, db, firebase, history, amdin and logger
	if err != nil {
		o.Logger.Error().Err(err).Msg("Error initializing users")
		return nil, err
	}

	err = o.InitServer()
	if err != nil {
		o.Logger.Error().Err(err).Msg("Error initializing server")
		return nil, err
	}

	err = o.InitStore()
	if err != nil {
		o.Logger.Err(err).Msg("Error initializing store")
		return nil, err
	}

	// err = b.InitUploader()
	// if err != nil {
	// 	log.Err(err).Msg("Error initializing uploader")
	// 	return nil, err
	// }

	err = o.InitScheduler()
	if err != nil {
		o.Logger.Err(err).Msg("Error initializing scheduler")
		return nil, err
	}

	environment := o.Config.GetString(EnvKeys.Environment)
	o.Logger.Info().
		Str("version", orchestratorVersion).
		Str("env", environment).
		Msg("Orchestrator initialized")

	return o, nil
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
	resourceConfig := requestLogger.SetupRequestLoggerResource(o.ResourceManager, o.DB)
	o.ResourceManager.AddResource(resourceConfig)
}

func (o *Orchestrator) InitDatabaseLogger() {
	resourceConfig := dbLogger.SetupDBLoggerResource(o.ResourceManager, o.DB)
	o.ResourceManager.AddResource(resourceConfig)
}

func (o *Orchestrator) InitAuth() {
	resourceConfig := auth.SetupUserResource(o.FirebaseClient, o.DB, o.Users.System)
	o.ResourceManager.AddResource(resourceConfig)
}

func (o *Orchestrator) InitResourceManager() {
	o.ResourceManager = manager.NewResourceManager()
}

func (o *Orchestrator) InitStore() error {
	var s store.Store

	storeType := o.Config.GetString(EnvKeys.StoreType)
	folder := o.Config.GetString(EnvKeys.UploaderFolder)

	switch storeType {
	case string(store.StoreS3):
		s3Store, err := store.NewS3Store(folder)
		if err != nil {
			o.Logger.Error().Err(err).Msg("Error initializing S3 store")
			return err
		}
		s = s3Store
	case string(store.StoreLocal):
		s = store.NewLocalStore(folder)
	default:
		return fmt.Errorf("unknown store type: %s", storeType)
	}

	o.Store = s
	return nil
}

func (o *Orchestrator) InitUsers() error {

	usersData := []models.RegisterUserInput{
		{
			Name:             "God",
			Email:            "god@" + o.Config.GetString(EnvKeys.Domain),
			Password:         uuid.New().String(),
			Roles:            []models.Role{models.AdminRole},
			RegisterFirebase: false,
		},
		{
			Name:             o.Config.GetString(EnvKeys.AdminName),
			Email:            o.Config.GetString(EnvKeys.AdminEmail),
			Password:         o.Config.GetString(EnvKeys.AdminPassword),
			Roles:            []models.Role{models.AdminRole},
			RegisterFirebase: true,
		},
		{
			Name:             "Scheduler",
			Email:            "scheduler@" + o.Config.GetString(EnvKeys.Domain),
			Password:         uuid.New().String(),
			Roles:            []models.Role{models.SchedulerRole},
			RegisterFirebase: false,
		},
		{
			Name:             "System",
			Email:            "system@" + o.Config.GetString(EnvKeys.Domain),
			Password:         uuid.New().String(),
			Roles:            []models.Role{models.SchedulerRole},
			RegisterFirebase: false,
		},
	}

	requestId := uuid.New().String()
	requestId = "automated::" + requestId

	for _, userData := range usersData {
		_, err := server.CreateUserWithRole(userData, o.FirebaseClient, o.DB, o.Users.System, requestId)
		if err != nil {
			o.Logger.Error().Err(err).Interface("user", userData).Msg("Error creating user")
			return err
		}
	}

	return nil
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

	server, err := server.NewServer(config, o.DB)
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
	return nil
}

func (o *Orchestrator) InitConfigReader() error {
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
