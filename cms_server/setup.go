package cms_server

import (
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type Config struct {
	Logger *zerolog.Logger
	DB     *gorm.DB
}

func Setup(config *Config) {
	log := config.Logger
	log.Info().Msg("Setting up CMS server")
}
