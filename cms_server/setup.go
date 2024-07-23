package cms_server

import (
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

var log *zerolog.Logger
var config *Config

type Config struct {
	Logger *zerolog.Logger
	DB     *gorm.DB
	Router *mux.Router
}

func Setup(cfg *Config) {
	log = cfg.Logger
	log.Info().Msg("Setting up CMS server")
	config = cfg
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
