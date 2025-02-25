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

// Server defines a structure for managing an HTTP server with middleware and routing capabilities.
type Server struct {
	*http.Server                                     // Server is the underlying HTTP server
	Middlewares    []func(http.Handler) http.Handler // middlewares is a slice of middleware functions
	Root           *mux.Router                       // root is the root handler for the server
	DB             *database.Database                // DB is the database connection
	AllowedOrigins []string
	LoggerConfig   *logger.LoggerConfig
	CsrfToken      string
	GodToken       string
	GodUser        *models.User
	SystemUser     *models.User
	Firebase       *clients.FirebaseManager
	Logger         *logger.Logger
}

// ServerConfig defines the configuration options for creating a new Server.
type ServerConfig struct {
	Host           string // Host is the hostname or IP address to listen on.
	Port           string // Port is the port number to listen on.
	CsrfToken      string // CSRFToken is the CSRF token to use for CSRF protection.
	AllowedOrigins []string
	LoggerConfig   *logger.LoggerConfig
	GodToken       string
	GodUser        *models.User
	SystemUser     *models.User
	Firebase       *clients.FirebaseManager
	Logger         *logger.Logger
}

// NewServer creates a new Server instance with the provided configuration.
//
// It checks for missing configuration (Host and Port) and returns an error if necessary.
// Otherwise, it creates a new Gorilla Mux router, sets up the server address and handler,
// and adds a basic logging middleware by default.
func NewServer(svrConfig *ServerConfig, db *database.Database, log *logger.Logger) (*Server, error) {

	log.Info().Interface("config", svrConfig).Msg("Initializing server")

	if svrConfig == nil {
		return nil, ErrServerConfigNotProvided
	}

	r := mux.NewRouter()

	svr := &Server{
		Server: &http.Server{
			Addr:         svrConfig.Host + ":" + svrConfig.Port,
			Handler:      r,
			WriteTimeout: TimeoutSeconds * time.Second,
			ReadTimeout:  TimeoutSeconds * time.Second,
		},
		Middlewares:    []func(http.Handler) http.Handler{},
		Root:           r,
		DB:             db,
		LoggerConfig:   svrConfig.LoggerConfig,
		GodToken:       svrConfig.GodToken,
		GodUser:        svrConfig.GodUser,
		Firebase:       svrConfig.Firebase,
		SystemUser:     svrConfig.SystemUser,
		CsrfToken:      svrConfig.CsrfToken,
		Logger:         log,
		AllowedOrigins: svrConfig.AllowedOrigins,
	}

	return svr, nil
}

func (s *Server) Run(routes []Route) error {

	// Include schema endpoint
	// s.Builder.Admin.AddApiRoute()

	// CSRF
	// csrfMiddleware := csrf.Protect([]byte(s.CsrfToken))

	// Middlewares
	publicRouter := s.Root
	publicRouter.Use(RecoveryMiddleware)                // graceful shutdown
	publicRouter.Use(RequestLoggerMiddleware(s.DB))     // will generate a request log
	publicRouter.Use(LoggingMiddleware(s.LoggerConfig)) // will store a logger with requestId
	publicRouter.Use(CorsMiddleware(s.AllowedOrigins))  // needs to be before user middleware
	publicRouter.Use(AuthMiddleware(s.GodToken, s.GodUser, s.Firebase, s.DB, s.SystemUser))
	publicRouter.Use(TimeoutMiddleware)
	publicRouter.Use(RateLimitMiddleware())
	// publicRouter.Use(csrfMiddleware)

	// apply custom middlewares
	for _, middleware := range s.Middlewares {
		s.Handler = middleware(s.Handler)
	}

	// Public Routes
	publicRouter.HandleFunc(
		"/",
		HealthCheck,
	).Name("healthcheck")

	s.Logger.Info().Msg("Public routes")
	for _, route := range routes {
		if !route.RequiresAuth {
			s.Logger.Info().Msgf("Route: %s", route.Path)
			publicRouter.HandleFunc(route.Path, route.Handler).Name(route.Name)
		}
	}

	s.Logger.Info().Msg("Authenticated routes")
	// Create separate routers for authenticated and public routes
	authRouter := publicRouter.PathPrefix("/private").Subrouter()
	authRouter.Use(ProtectedRouteMiddleware)

	for _, route := range routes {
		if route.RequiresAuth {
			s.Logger.Info().Msgf("Route: /private%s", route.Path)
			authRouter.HandleFunc(route.Path, route.Handler).Name(route.Name)
		}
	}

	s.Logger.Info().Msgf("Running server on port %s", s.Addr)
	return s.ListenAndServe()
}

// AddMiddleware adds a new middleware function to the server's middleware chain.
//
// Middleware functions are executed sequentially in the order they are added.
// Each middleware function takes an http.Handler as input and returns a new http.Handler
// that can wrap the original handler and perform additional logic before or after
// the original handler is called.
func (s *Server) AddMiddleware(middleware func(http.Handler) http.Handler) {
	s.Middlewares = append(s.Middlewares, middleware)
}

// HandlerFunc is the type of the function that can be used as an http.HandlerFunc.
// It takes an http.ResponseWriter and an *http.Request as input and returns nothing.
type HandlerFunc func(w http.ResponseWriter, r *http.Request)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	err := ValidateRequestMethod(r, "GET")
	if err != nil {
		SendJsonResponse(w, http.StatusMethodNotAllowed, err, err.Error())
		return
	}

	SendJsonResponse(w, http.StatusOK, nil, "OK")
}
