package orchestrator

import (
	"fmt"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/config"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

const orchestratorVersion = "1.6.0"

type Orchestrator struct {
	Config *config.ConfigReader
	Logger *logger.Logger
	DB     *database.Database
	Server *server.Server
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

	err = o.InitLogger()
	if err != nil {
		fmt.Println("Error initializing logger:", err)
		return nil, err
	}

	err = o.InitDatabase()
	if err != nil {
		o.Logger.Error().Err(err).Msg("Error initializing database")
		return nil, err
	}

	err = o.InitServer()
	if err != nil {
		o.Logger.Error().Err(err).Msg("Error initializing server")
		return nil, err
	}

	environment := o.Config.GetString(EnvKeys.Environment)
	o.Logger.Info().
		Str("version", orchestratorVersion).
		Str("env", environment).
		Msg("Orchestrator initialized")

	return o, nil
}

func (o *Orchestrator) InitServer() error {

	config := &server.ServerConfig{
		Host:      o.Config.GetString(EnvKeys.ServerHost),
		Port:      o.Config.GetString(EnvKeys.ServerPort),
		CSRFToken: o.Config.GetString(EnvKeys.CsrfToken),
	}

	server, err := server.NewServer(config)
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
	return nil
}
