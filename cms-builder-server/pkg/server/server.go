package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/gorilla/mux"
)

var (
	ErrServerConfigNotProvided = errors.New("database config not provided")
	ErrServerNotInitialized    = errors.New("server not initialized")
)

type ServerConfig struct {
	Host           string
	Port           string
	CsrfToken      string
	AllowedOrigins []string
	LoggerConfig   *logger.LoggerConfig
	GodToken       string
	GodUser        *models.User
	SystemUser     *models.User
	Firebase       *clients.FirebaseManager
}

type Server struct {
	*http.Server
	ServerConfig
	Middlewares   []func(http.Handler) http.Handler
	Root          *mux.Router
	DB            *database.Database
	Logger        *logger.Logger
	PublicRoutes  map[string]Route
	PrivateRoutes map[string]Route
}

func NewServer(config *ServerConfig, db *database.Database, log *logger.Logger) (*Server, error) {
	log.Info().Interface("config", config).Msg("Initializing server")

	if config == nil {
		return nil, ErrServerConfigNotProvided
	}

	r := mux.NewRouter()

	svr := &Server{
		Server: &http.Server{
			Addr:         config.Host + ":" + config.Port,
			Handler:      r,
			WriteTimeout: TimeoutSeconds * time.Second,
			ReadTimeout:  TimeoutSeconds * time.Second,
		},
		ServerConfig: *config,
		Middlewares:  []func(http.Handler) http.Handler{},
		Root:         r,
		DB:           db,
		Logger:       log,
		PublicRoutes: map[string]Route{
			"/health": {
				Path:    "/health",
				Name:    "healthcheck",
				Handler: HealthCheck,
			},
		},
		PrivateRoutes: map[string]Route{},
	}

	return svr, nil
}

func (s *Server) Run() error {
	publicRouter := s.Root
	publicRouter.Use(
		RecoveryMiddleware,
		RequestLoggerMiddleware(s.DB),
		LoggingMiddleware(s.LoggerConfig),
		CorsMiddleware(s.AllowedOrigins),
		AuthMiddleware(s.GodToken, s.GodUser, s.Firebase, s.DB, s.SystemUser),
		TimeoutMiddleware,
		RateLimitMiddleware(),
	)

	for path, route := range s.PublicRoutes {
		s.Logger.Warn().Str("path", path).Msg("Public")
		publicRouter.HandleFunc(route.Path, route.Handler).Name(route.Name).Methods(route.Method)
	}

	authRouter := publicRouter.PathPrefix("/private").Subrouter()
	authRouter.Use(ProtectedRouteMiddleware)

	for path, route := range s.PrivateRoutes {
		if route.RequiresAuth {
			s.Logger.Info().Str("path", "/private"+path).Msg("Private")
			authRouter.HandleFunc(route.Path, route.Handler).Name(route.Name).Methods(route.Method)
		}
	}

	s.Logger.Info().Msgf("Running server on port %s", s.Addr)
	return s.ListenAndServe()
}

func (s *Server) AddRoute(route Route) {

	_, ok := s.PublicRoutes[route.Path]
	if ok {
		s.Logger.Warn().Msgf("Route %s already exists", route.Path)
		return
	}

	_, ok = s.PrivateRoutes[route.Path]
	if ok {
		s.Logger.Warn().Msgf("Route %s already exists", route.Path)
		return
	}

	if route.RequiresAuth {
		s.PrivateRoutes[route.Path] = route
	} else {
		s.PublicRoutes[route.Path] = route
	}
}

func (s *Server) AddMiddleware(middleware func(http.Handler) http.Handler) {
	s.Middlewares = append(s.Middlewares, middleware)
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := ValidateRequestMethod(r, "GET"); err != nil {
		SendJsonResponse(w, http.StatusMethodNotAllowed, err, err.Error())
		return
	}

	SendJsonResponse(w, http.StatusOK, nil, "OK")
}
