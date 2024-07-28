package cms_admin

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

func (c *Config) Validate() error {
	if c.Logger == nil {
		return fmt.Errorf("logger is required")
	}

	if c.DB == nil {
		return fmt.Errorf("database is required")
	}

	if c.RootDir == "" {
		return fmt.Errorf("root directory is required")
	}

	return nil
}

func Setup(cfg *Config) error {
	// Validate input
	err := cfg.Validate()
	if err != nil {
		return err
	}

	// Initialize variables
	log = cfg.Logger

	log.Debug().Msg("Setting up CMS server")

	config = cfg

	return nil
}

func Register(model interface{}) {
	entity := Entity{
		Model: model,
	}

	entities = append(entities, entity)

	config.DB.AutoMigrate(model)

	log.Debug().Interface("entity", entity).Msgf("Model %s registered", entity.Name())
}

func GetEntities() []Entity {
	return entities
}
