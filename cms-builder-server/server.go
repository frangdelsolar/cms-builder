package builder

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const RequestsPerMinute = 100
const TimeoutSeconds = 15

var (
	ErrServerConfigNotProvided = errors.New("database config not provided")
	ErrServerNotInitialized    = errors.New("server not initialized")
)

// RouteHandler defines a structure for storing route information.
type RouteHandler struct {
	Route        string      // route is the path for the route. i.e. /users/{id}
	Handler      HandlerFunc // handler is the handler for the route
	Name         string      // name is the name of the route
	RequiresAuth bool        // requiresAuth is a flag indicating if the route requires authentication
	Schema       interface{} // represents the input
	Method       string      // method is the HTTP method for the route
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
func NewServer(svrConfig *ServerConfig) (*Server, error) {

	if svrConfig == nil {
		return nil, ErrServerConfigNotProvided
	}

	if svrConfig.Host == "" {
		svrConfig.Host = config.GetString(EnvKeys.ServerHost)
	}

	if svrConfig.Port == "" {
		svrConfig.Port = config.GetString(EnvKeys.ServerPort)
	}

	if svrConfig.CSRFToken == "" {
		svrConfig.CSRFToken = uuid.New().String()
	}

	r := mux.NewRouter()
	// r.StrictSlash(true)

	svr := &Server{
		Server: &http.Server{
			Addr:         svrConfig.Host + ":" + svrConfig.Port,
			Handler:      r,
			WriteTimeout: TimeoutSeconds * time.Second,
			ReadTimeout:  TimeoutSeconds * time.Second,
		},
		Middlewares: []func(http.Handler) http.Handler{},
		Routes:      []RouteHandler{},
		Root:        r,
		Builder:     svrConfig.Builder,
	}

	// CSRF
	// csrfKey := []byte(config.GetString(EnvKeys.CsrfToken))
	// csrfMiddleware := csrf.Protect(csrfKey)

	// Middlewares
	r.Use(RecoveryMiddleware)
	r.Use(RequestIDMiddleware)
	r.Use(TimeoutMiddleware)

	rateLimiter := NewRateLimiter(RequestsPerMinute, 1*time.Minute)
	r.Use(RateLimitMiddleware(rateLimiter))

	// r.Use(csrfMiddleware)
	r.Use(CorsMiddleware)

	r.Use(LoggingMiddleware)

	// Public Routes
	svr.AddRoute(
		"/",
		func(w http.ResponseWriter, r *http.Request) {
			err := ValidateRequestMethod(r, "GET")
			if err != nil {
				SendJsonResponse(w, http.StatusMethodNotAllowed, err, err.Error())
				return
			}

			SendJsonResponse(w, http.StatusOK, nil, "ok")
		},
		"healthz",
		false,
		http.MethodGet,
		nil,
	)

	return svr, nil
}

// CorsMiddleware adds Cross-Origin Resource Sharing headers to the response.
//
// It sets the following headers:
//
// - Access-Control-Allow-Headers: Content-Type, Authorization, Origin
// - Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
// - Access-Control-Allow-Origin: *
//
// It also checks the Origin header against the list of allowed origins
// and returns a 403 Forbidden response if the origin is not allowed.
//
// If the request method is OPTIONS, it returns a 200 OK response immediately.
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Origin")

		allowedOrigins := config.GetStringSlice(EnvKeys.CorsAllowedOrigins)
		origin := r.Header.Get("Origin")

		if allowedOrigins[0] == "*" || contains(allowedOrigins, origin) {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Origin", "*")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
		} else {
			err := fmt.Errorf("origin '%s' is not allowed", origin)
			log.Warn().Interface("headers", r.Header).Interface("allowedOrigins", allowedOrigins).Interface("origin", origin).Msg("CORS")
			SendJsonResponse(w, http.StatusForbidden, nil, err.Error())
			return
		}

		next.ServeHTTP(w, r)
	})
}

// contains checks if a slice of strings contains a specific string.
//
// It iterates over the slice 's' and returns true if the element 'e' is found;
// otherwise, it returns false.
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// LoggingMiddleware is a sample middleware function that logs the request URI.
//
// It takes an http.Handler as input and returns a new http.Handler that wraps the original
// handler and logs the request URI before calling the original handler.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId := GetRequestID(r)
		// TODO: Make logger unique to each request, I will need to use context for that.
		// Seems like a massive effort
		log.Info().Str("requestId", requestId).Msg(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

type RateLimiter struct {
	mu     sync.Mutex
	rates  map[string]*RateLimit // Key: Client identifier (e.g., IP address)
	limit  int                   // Maximum requests per window
	window time.Duration         // Time window (e.g., 1 minute)
}

type RateLimit struct {
	Count     int       `json:"count"`
	LastReset time.Time `json:"last_reset"`
}

// NewRateLimiter creates a new rate limiter with the given limit and window.
// The limit is the maximum number of requests that can be made within the window.
// The window is the time duration for which the limit applies. It is used to
// determine when the rate limit is reset.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		rates:  make(map[string]*RateLimit),
		limit:  limit,
		window: window,
	}
}

// Allow checks if the client identified by clientIdentifier is allowed to make
// a request according to the rate limiter's rules. It returns true if the
// client is allowed, and false otherwise.
//
// The method is thread-safe.
func (rl *RateLimiter) Allow(clientIdentifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	r, ok := rl.rates[clientIdentifier]

	if !ok || r.LastReset.Add(rl.window).Before(now) {
		rl.rates[clientIdentifier] = &RateLimit{Count: 1, LastReset: now}
		return true
	}

	if r.Count < rl.limit {
		r.Count++
		return true
	}

	return false // Rate limit exceeded
}

// RateLimitMiddleware is a middleware function that rate-limits incoming requests.
//
// The middleware uses the given rate limiter to determine if a request is allowed
// or not. If the request is allowed, it calls the next handler in the chain.
// If the request is not allowed, it returns an HTTP 429 Too Many Requests response.
//
// The middleware is thread-safe.
func RateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			address := r.RemoteAddr
			clientIP := strings.Split(address, ":")[0]
			if !rl.Allow(clientIP) {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequestIDKey is the key used to store the request ID in the context.
type RequestIDKey struct{}

// RequestIDMiddleware assigns a unique ID to each request and adds it to the context.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate a UUID
		requestID := uuid.New().String()

		// Add the request ID to the context
		ctx := r.Context()
		ctx = context.WithValue(ctx, RequestIDKey{}, requestID)

		// Add the request ID to the response headers
		w.Header().Set("X-Request-ID", requestID)

		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RecoveryMiddleware catches panics and logs them.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Error().Interface("panic", err).Bytes("stack", debug.Stack()).Msg("Panic recovered")

				// Optionally, you can customize the error response
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// TimeoutMiddleware sets a timeout for requests.
func TimeoutMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx, cancel := context.WithTimeout(r.Context(), TimeoutSeconds*time.Second)
		defer cancel()

		r = r.WithContext(ctx)

		done := make(chan struct{})

		go func() {
			next.ServeHTTP(w, r)
			close(done)
		}()

		select {
		case <-done:
			// Handler completed successfully
		case <-ctx.Done():
			log.Error().Msg("Request timed out")
			http.Error(w, "Request timed out", http.StatusGatewayTimeout) // Or 504 Gateway Timeout
		}
	})
}

// Run starts the server and listens for incoming connections on the configured address.
//
// It logs a message indicating the server is running on the specified port,
// applies all registered middleware to the server's handler,
// and finally calls the underlying http.Server's ListenAndServe method.
func (s *Server) Run() error {

	// Include schema endpoint
	s.Builder.Admin.AddApiRoute()

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
func (s *Server) AddRoute(route string, handler HandlerFunc, name string, requiresAuth bool, method string, schema interface{}) {
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
func NewRouteHandler(route string, handler HandlerFunc, name string, requiresAuth bool, method string, schema interface{}) RouteHandler {
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
