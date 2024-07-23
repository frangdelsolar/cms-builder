package cms_server

import (
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

var log *zerolog.Logger

type Config struct {
	Logger *zerolog.Logger
	DB     *gorm.DB
}

func Setup(cfg *Config) {
	log = cfg.Logger
	log.Info().Msg("Setting up CMS server")
}

func Register(model interface{}) {
	log.Info().Msgf("Registering model: %s", model)

	entity := Entity{
		Model: model,
	}

	log.Info().Interface("entity", entity).Msgf("Model registered")
	log.Debug().Msgf("Name: %s", entity.Name())
	log.Debug().Msgf("Fields: %s", entity.Fields())

}
