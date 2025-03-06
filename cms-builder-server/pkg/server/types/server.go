package types

import (
	"net/http"

	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	"github.com/gorilla/mux"
)

type GetRoutesFunc func(apiBaseUrl string) []Route

type Server struct {
	*http.Server
	ServerConfig
	Middlewares []func(http.Handler) http.Handler
	Root        *mux.Router
	DB          *dbTypes.DatabaseConnection
	Logger      *loggerTypes.Logger
}

func (s *Server) AddMiddleware(middleware func(http.Handler) http.Handler) {
	s.Middlewares = append(s.Middlewares, middleware)
}
