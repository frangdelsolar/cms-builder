package server

import (
	"errors"
	"net/http"
	"time"

	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	svrMiddlewares "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/middlewares"
	svrTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/types"
	"github.com/gorilla/mux"
)

var (
	ErrServerConfigNotProvided = errors.New("database config not provided")
	ErrServerNotInitialized    = errors.New("server not initialized")
)

func NewServer(config *svrTypes.ServerConfig, db *dbTypes.DatabaseConnection, log *loggerTypes.Logger) (*svrTypes.Server, error) {
	log.Info().Interface("config", config).Msg("Initializing server")

	if config == nil {
		return nil, ErrServerConfigNotProvided
	}

	r := mux.NewRouter()

	svr := &svrTypes.Server{
		Server: &http.Server{
			Addr:         config.Host + ":" + config.Port,
			Handler:      r,
			WriteTimeout: svrMiddlewares.TimeoutSeconds * time.Second,
			ReadTimeout:  svrMiddlewares.TimeoutSeconds * time.Second,
		},
		ServerConfig: *config,
		Middlewares:  []func(http.Handler) http.Handler{},
		Root:         r,
		DB:           db,
		Logger:       log,
	}

	return svr, nil
}
