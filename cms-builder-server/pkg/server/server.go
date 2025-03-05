package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
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
	LoggerConfig   *loggerTypes.LoggerConfig
	GodToken       string
	GodUser        *models.User
	SystemUser     *models.User
	Firebase       *clients.FirebaseManager
}

type Server struct {
	*http.Server
	ServerConfig
	Middlewares []func(http.Handler) http.Handler
	Root        *mux.Router
	DB          *database.Database
	Logger      *loggerTypes.Logger
}

func NewServer(config *ServerConfig, db *database.Database, log *loggerTypes.Logger) (*Server, error) {
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
	}

	return svr, nil
}

type GetRoutesFunc func(apiBaseUrl string) []Route

func (s *Server) Run(getRoutes GetRoutesFunc, apiBaseUrl string) error {

	routes := getRoutes(apiBaseUrl)

	routes = append(routes,
		Route{
			Path:         "/",
			Handler:      HealthCheck,
			Name:         "health",
			RequiresAuth: false,
			Methods:      []string{http.MethodGet},
		},
	)

	s.Logger.Info().Msgf("Initializing %d routes", len(routes))

	publicRouter := s.Root
	publicRouter.Use(
		RecoveryMiddleware,
		AuthMiddleware(s.GodToken, s.GodUser, s.Firebase, s.DB, s.SystemUser),
		RequestLoggerMiddleware(s.DB),
		LoggingMiddleware(s.LoggerConfig),
		CorsMiddleware(s.AllowedOrigins),
		TimeoutMiddleware,
		RateLimitMiddleware(),
		gziphandler.GzipHandler,
	)

	routesSeen := map[string]bool{}

	for _, route := range routes {

		// Check for duplicates - We shouldn't have two routes with the same path
		if routesSeen[route.Path] {
			panic("Duplicate route: " + route.Path + " - " + route.Name)
		}
		routesSeen[route.Path] = true

		if route.RequiresAuth {
			continue
		}
		s.Logger.Warn().Str("path", route.Path).Msg("Public")

		methods := route.Methods
		methods = append(methods, http.MethodOptions)

		publicRouter.HandleFunc(route.Path, route.Handler).Name(route.Name).Methods(methods...)
	}

	authRouter := publicRouter.PathPrefix("/private").Subrouter()
	authRouter.Use(ProtectedRouteMiddleware)

	for _, route := range routes {
		if !route.RequiresAuth {
			continue
		}

		methods := route.Methods
		methods = append(methods, http.MethodOptions)
		s.Logger.Error().Str("path", "/private"+route.Path).Msg("Private")

		authRouter.HandleFunc(route.Path, route.Handler).Name(route.Name).Methods(methods...)
	}

	s.Logger.Info().Msgf("Running server on port %s", s.Addr)
	return s.ListenAndServe()
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
