package orchestrator

import (
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/config"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
)

type Orchestrator struct {
	Config *config.ConfigReader
	Logger *logger.Logger
}

func NewOrchestrator() (*Orchestrator, error) {
	var err error
	o := &Orchestrator{}

	err = o.InitConfigReader()
	if err != nil {
		return nil, err
	}

	err = o.InitLogger()
	if err != nil {
		return nil, err
	}

	return o, nil
}

func (o *Orchestrator) InitConfigReader() error {
	config, err := config.NewConfigReader(
		&config.ReaderConfig{
			ReadEnv:  true,
			ReadFile: false,
		},
	)
	if err != nil {
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
		return err
	}
	o.Logger = logger
	return nil
}
