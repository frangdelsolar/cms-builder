package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

const RequestsPerMinute = 100
const TimeoutSeconds = 15

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
		requestId := GetRequestId(r)
		// TODO: Make logger unique to each request, I will need to use context for that.
		// Seems like a massive effort
		log.Info().Str("requestId", requestId).Str("method", r.Method).Msg(r.RequestURI)
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
type RequestLog struct {
	gorm.Model
	Timestamp         time.Time `gorm:"type:timestamp" json:"timestamp"`
	Duration          int64     `json:"duration"`
	Ip                string    `json:"ip"`
	Origin            string    `json:"origin"`
	Referer           string    `json:"referrer"`
	UserId            string    `gorm:"foreignKey:UserId" json:"userId"`
	User              *User     `json:"user,omitempty"`
	Roles             string    `json:"roles"`
	Method            string    `json:"method"`
	Path              string    `json:"path"`
	Query             string    `json:"query"`
	StatusCode        string    `json:"statusCode"`
	Error             string    `json:"error"`
	Header            string    `json:"header"`
	Body              string    `json:"body"`
	Response          string    `json:"response"`
	RequestIdentifier string    `json:"requestIdentifier"`
}

// RequestLogMiddleware assigns a unique ID to each request and adds it to the context.
func (b *Builder) RequestLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		ctx := r.Context()

		// Add start time to the context for later use if needed
		ctx = context.WithValue(ctx, "requestStartTime", start) // Use a key type if you have one

		r = r.WithContext(ctx) // Important: Update the request with the new context
		wrappedWriter := &WriterWrapper{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
			Body:           new(bytes.Buffer),
		}

		var err error
		var requestBody string
		var requestHeaders string
		var responseBody string

		requestIdentifier := uuid.New().String()
		r = r.WithContext(context.WithValue(r.Context(), "requestIdentifier", requestIdentifier))

		defer func() {

			duration := time.Since(start)

			statusCode := wrappedWriter.StatusCode

			errorMessage := ""
			if err != nil {
				errorMessage = err.Error()
			}

			query, err := url.QueryUnescape(r.URL.RawQuery)
			if err != nil {
				log.Error().Err(err).Msg("Error unescaping query")
			}

			logEntry := RequestLog{
				Timestamp:         start,
				Ip:                r.RemoteAddr,
				UserId:            r.Header.Get(requestedByParamKey.S()),
				Roles:             r.Header.Get(rolesParamKey.S()),
				Method:            r.Method,
				Path:              r.URL.Path,
				Query:             query,
				Duration:          duration.Nanoseconds() / 1e6,
				StatusCode:        fmt.Sprintf("%d", statusCode),
				Origin:            r.Header.Get("Origin"),
				Referer:           r.Header.Get("Referer"),
				Error:             errorMessage,
				Header:            requestHeaders,
				Body:              requestBody,
				Response:          responseBody,
				RequestIdentifier: requestIdentifier,
			}

			b.DB.DB.Create(&logEntry)
		}()

		bodyBytes, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			err = readErr // Capture the error
			log.Error().Err(err).Msg("Error reading request body")
		}

		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		// Capture headers before next.ServeHTTP
		headers := make(map[string][]string)
		for name, values := range r.Header {
			headers[name] = values
		}

		headerJSON, marshalErr := json.Marshal(headers)
		if marshalErr != nil {
			err = marshalErr
			log.Error().Err(err).Msg("Error marshaling headers")
		}

		next.ServeHTTP(wrappedWriter, r)

		// Check for errors after the handler has run
		if wrappedWriter.StatusCode >= 400 || readErr != nil || marshalErr != nil {
			// If there was an error during the request, log the body and headers
			if wrappedWriter.StatusCode >= 400 {
				err = errors.New(http.StatusText(wrappedWriter.StatusCode))
				requestHeaders = string(headerJSON)
				requestBody = string(bodyBytes)
				responseBody = wrappedWriter.Body.String()
			}
		}
	})
}

type WriterWrapper struct {
	http.ResponseWriter
	StatusCode int
	Body       *bytes.Buffer
}

func (w *WriterWrapper) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *WriterWrapper) Write(b []byte) (int, error) {
	if w.Body != nil { // Write to the buffer only if it's initialized
		w.Body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

// RecoveryMiddleware catches panics and logs them.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				log.Error().Interface("panic", err).Bytes("stack", debug.Stack()).Msg("Panic recovered")
				SendJsonResponse(w, http.StatusInternalServerError, nil, "Internal Server Error")
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

func ProtectedRouteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get(authParamKey.S())
		user := r.Header.Get(requestedByParamKey.S())

		if auth != "true" || user == "" || user == "0" {
			SendJsonResponse(w, http.StatusUnauthorized, nil, "Unauthorized")
			return
		}
		next.ServeHTTP(w, r)
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

	// Middlewares
	publicRouter := s.Root
	publicRouter.Use(s.Builder.RequestLogMiddleware)
	publicRouter.Use(RecoveryMiddleware) // should be first
	// CSRF
	// csrfKey := []byte(config.GetString(EnvKeys.CsrfToken))
	// csrfMiddleware := csrf.Protect(csrfKey)

	// Middlewares
	publicRouter.Use(CorsMiddleware) // needs to be before user middleware
	publicRouter.Use(s.Builder.UserMiddleware)
	publicRouter.Use(TimeoutMiddleware)

	rateLimiter := NewRateLimiter(RequestsPerMinute, 1*time.Minute)
	publicRouter.Use(RateLimitMiddleware(rateLimiter))

	// publicRouter.Use(csrfMiddleware)
	publicRouter.Use(LoggingMiddleware)

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
