package builder

import (
	"errors"
	"net/http"

	// "github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

var (
	ErrServerConfigNotProvided = errors.New("database config not provided")
	ErrServerNotInitialized    = errors.New("server not initialized")
)

// const PUBLIC_DIR = "./server/public"

// RouteHandler defines a structure for storing route information.
type RouteHandler struct {
	Route        string      // route is the path for the route. i.e. /users/{id}
	Handler      HandlerFunc // handler is the handler for the route
	Name         string      // name is the name of the route
	RequiresAuth bool        // requiresAuth is a flag indicating if the route requires authentication
}

// Server defines a structure for managing an HTTP server with middleware and routing capabilities.
type Server struct {
	*http.Server                                   // Server is the underlying HTTP server
	Middlewares  []func(http.Handler) http.Handler // middlewares is a slice of middleware functions
	Routes       []RouteHandler                    // routes is a slice of route handlers
	Root         *mux.Router                       // root is the root handler for the server
	Builder      *Builder
}

// ServerConfig defines the configuration options for creating a new Server.
type ServerConfig struct {
	Host      string // Host is the hostname or IP address to listen on.
	Port      string // Port is the port number to listen on.
	CSRFToken string // CSRFToken is the CSRF token to use for CSRF protection.
	Builder   *Builder
}

// NewServer creates a new Server instance with the provided configuration.
//
// It checks for missing configuration (Host and Port) and returns an error if necessary.
// Otherwise, it creates a new Gorilla Mux router, sets up the server address and handler,
// and adds a basic logging middleware by default.
func NewServer(config *ServerConfig) (*Server, error) {

	if config == nil {
		return nil, ErrServerConfigNotProvided
	}

	if config.Host == "" {
		config.Host = "localhost"
	}

	if config.Port == "" {
		config.Port = "8080"
	}

	if config.CSRFToken == "" {
		config.CSRFToken = "secret"
	}

	r := mux.NewRouter()

	svr := &Server{
		Server: &http.Server{
			Addr:    config.Host + ":" + config.Port,
			Handler: r,
		},
		Middlewares: []func(http.Handler) http.Handler{},
		Routes:      []RouteHandler{},
		Root:        r,
		Builder:     config.Builder,
	}

	svr.AddMiddleware(LoggingMiddleware)

	// CSRF
	// csrfKey := []byte(config.CSRFToken) // Replace with a real secret key
	// csrfMiddleware := csrf.Protect(csrfKey, csrf.CookieName("csrftoken"))

	// Middlewares
	r.Use(LoggingMiddleware)
	r.Use(mux.CORSMethodMiddleware(r))
	// r.Use(csrfMiddleware)

	// Public Routes
	svr.AddRoute("/", func(w http.ResponseWriter, r *http.Request) {
		err := ValidateRequestMethod(r, "GET")
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, err, err.Error())
			return
		}

		SendJsonResponse(w, http.StatusOK, nil, "ok")
	}, "healthz", false)

	return svr, nil
}

// LoggingMiddleware is a sample middleware function that logs the request URI.
//
// It takes an http.Handler as input and returns a new http.Handler that wraps the original
// handler and logs the request URI before calling the original handler.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

// Run starts the server and listens for incoming connections on the configured address.
//
// It logs a message indicating the server is running on the specified port,
// applies all registered middleware to the server's handler,
// and finally calls the underlying http.Server's ListenAndServe method.
func (s *Server) Run() error {

	// Create separate routers for authenticated and public routes
	authRouter := s.Root.PathPrefix("/private").Subrouter()
	publicRouter := s.Root

	for _, middleware := range s.Middlewares {
		s.Handler = middleware(s.Handler)
	}

	// Apply authMiddleware only to the authenticated router
	authRouter.Use(s.Builder.AuthMiddleware)

	log.Info().Msg("Public routes")
	for _, route := range s.Routes {
		if !route.RequiresAuth {
			log.Info().Msgf("Route: %s", route.Route)
			publicRouter.HandleFunc(route.Route, route.Handler).Name(route.Name)
		}
	}

	log.Info().Msg("Authenticated routes")
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
func (s *Server) AddRoute(route string, handler HandlerFunc, name string, requiresAuth bool) {
	s.Routes = append(s.Routes, NewRouteHandler(route, handler, name, requiresAuth))
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
func NewRouteHandler(route string, handler HandlerFunc, name string, requiresAuth bool) RouteHandler {
	return RouteHandler{
		Route:        route,
		Handler:      handler,
		Name:         name,
		RequiresAuth: requiresAuth,
	}
}

// GetRoutes returns a slice of all registered routes.
//
// The slice is a shallow copy of the server's internal routes slice, so modifying
// the slice or its elements will not affect the server's internal state.
func (s *Server) GetRoutes() []RouteHandler {
	return s.Routes
}
