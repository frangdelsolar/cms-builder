package server

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

var (
	ErrServerConfigNotProvided = errors.New("database config not provided")
	ErrServerNotInitialized    = errors.New("server not initialized")
)

// RouteHandler defines a structure for storing route information.
type RouteHandler struct {
	Route        string           // route is the path for the route. i.e. /users/{id}
	Handler      http.HandlerFunc // handler is the handler for the route
	Name         string           // name is the name of the route
	RequiresAuth bool             // requiresAuth is a flag indicating if the route requires authentication
	Schema       interface{}      // represents the input
	Method       string           // method is the HTTP method for the route
}

// Server defines a structure for managing an HTTP server with middleware and routing capabilities.
type Server struct {
	*http.Server                                     // Server is the underlying HTTP server
	Middlewares    []func(http.Handler) http.Handler // middlewares is a slice of middleware functions
	Routes         []RouteHandler                    // routes is a slice of route handlers
	Root           *mux.Router                       // root is the root handler for the server
	DB             *database.Database                // DB is the database connection
	AllowedOrigins []string
	LoggerConfig   *logger.LoggerConfig
	CsrfToken      string
	GodToken       string
	GodUser        *models.User
	SystemUser     *models.User
	Firebase       *clients.FirebaseManager
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
		Middlewares:  []func(http.Handler) http.Handler{},
		Routes:       []RouteHandler{},
		Root:         r,
		DB:           db,
		LoggerConfig: svrConfig.LoggerConfig,
		GodToken:     svrConfig.GodToken,
		GodUser:      svrConfig.GodUser,
		Firebase:     svrConfig.Firebase,
		SystemUser:   svrConfig.SystemUser,
		CsrfToken:    svrConfig.CsrfToken,
	}

	return svr, nil
}

// Run starts the server and listens for incoming connections on the configured address.
//
// It logs a message indicating the server is running on the specified port,
// applies all registered middleware to the server's handler,
// and finally calls the underlying http.Server's ListenAndServe method.
func (s *Server) Run() error {

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

	log.Info().Msg("Public routes")
	for _, route := range s.Routes {
		if !route.RequiresAuth {
			log.Info().Msgf("Route: %s", route.Route)
			publicRouter.HandleFunc(route.Route, route.Handler).Name(route.Name)
		}
	}

	log.Info().Msg("Authenticated routes")
	// Create separate routers for authenticated and public routes
	authRouter := publicRouter.PathPrefix("/private").Subrouter()
	authRouter.Use(ProtectedRouteMiddleware)

	for _, route := range s.Routes {
		if route.RequiresAuth {
			log.Info().Msgf("Route: /private%s", route.Route)
			authRouter.HandleFunc(route.Route, route.Handler).Name(route.Name)
		}
	}

	log.Info().Msgf("Running server on port %s", s.Addr)
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

// AddRoute adds a new route to the server's routing table.
//
// It takes three arguments:
//   - route: The path for the route (e.g., "/", "/users/{id}").
//   - handler: The function to be called when the route is matched.
//   - name: An optional name for the route (useful for generating URLs)
//   - requiresAuth: A boolean flag indicating whether the route requires authentication
//
// Example:
//
//	AddRoute("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
//	  // Handle user with ID
//	}, "getUser", false)
//
// url, err := r.Get("getUser").URL("id", "123") =>
// "/users/123"
func (s *Server) AddRoute(route string, handler http.HandlerFunc, name string, requiresAuth bool, method string, schema interface{}) {
	// Remove trailing slash if present
	route = strings.TrimSuffix(route, "/")

	s.Routes = append(s.Routes, NewRouteHandler(route, handler, name, requiresAuth, method, schema))
}

// NewRouteHandler creates a new RouteHandler instance.
//
// It takes four arguments:
//   - route: The path for the route (e.g., "/", "/users/{id}").
//   - handler: The function to be called when the route is matched.
//   - name: An optional name for the route (useful for generating URLs)
//   - requiresAuth: A boolean flag indicating whether the route requires authentication
//
// Example:
//
//	routeHandler := NewRouteHandler("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
//	  // Handle user with ID
//	}, "getUser", false)
func NewRouteHandler(route string, handler http.HandlerFunc, name string, requiresAuth bool, method string, schema interface{}) RouteHandler {
	return RouteHandler{
		Route:        route,
		Handler:      handler,
		Name:         name,
		RequiresAuth: requiresAuth,
		Method:       method,
		Schema:       schema,
	}
}

// GetRoutes returns a slice of all registered routes.
//
// The slice is a shallow copy of the server's internal routes slice, so modifying
// the slice or its elements will not affect the server's internal state.
func (s *Server) GetRoutes() []RouteHandler {
	return s.Routes
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	err := ValidateRequestMethod(r, "GET")
	if err != nil {
		SendJsonResponse(w, http.StatusMethodNotAllowed, err, err.Error())
		return
	}

	SendJsonResponse(w, http.StatusOK, nil, "OK")
}
