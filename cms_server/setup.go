package cms_server

import (
	"fmt"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

var log *zerolog.Logger
var entities []Entity = []Entity{}
var config *Config

type Config struct {
	Logger  *zerolog.Logger
	DB      *gorm.DB
	RootDir string
}

func Setup(cfg *Config) error {
	// Validate input
	if cfg.Logger == nil {
		return fmt.Errorf("logger is required")
	}

	if cfg.DB == nil {
		return fmt.Errorf("database is required")
	}

	if cfg.RootDir == "" {
		return fmt.Errorf("root directory is required")
	}

	// Initialize variables
	config = cfg
	log = cfg.Logger

	log.Debug().Msg("Setting up CMS server")

	// Load templates -> Maybe there is a better place to do this.
	LoadTemplateConfiguration()
	LoadTemplates()

	return nil
}

func Register(model interface{}) {
	entity := Entity{
		Model: model,
	}

	entities = append(entities, entity)

	log.Debug().Interface("entity", entity).Msgf("Model %s registered", entity.Name())
}

func GetEntities() []Entity {
	return entities
}
