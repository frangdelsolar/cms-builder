package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// TestCorsMiddleware_AllowedOrigin tests that the middleware allows requests from allowed origins.
func TestCorsMiddleware_AllowedOrigin(t *testing.T) {
	// Create a logger
	logger := zerolog.New(nil)

	allowedOrigins := []string{"https://example.com"}
	middleware := CorsMiddleware(allowedOrigins)

	// Create a test handler
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create a test request with an allowed origin
	req := httptest.NewRequest("GET", "https://example.com", nil)
	req.Header.Set("Origin", "https://example.com")

	// Add the logger to the request context
	ctx := context.WithValue(req.Context(), "logger", &logger)
	req = req.WithContext(ctx)

	// Record the response
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "Content-Type, Authorization, Origin", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
}

// TestCorsMiddleware_DisallowedOrigin tests that the middleware blocks requests from disallowed origins.
func TestCorsMiddleware_DisallowedOrigin(t *testing.T) {
	// Create a logger
	logger := zerolog.New(nil)

	allowedOrigins := []string{"https://example.com"}
	middleware := CorsMiddleware(allowedOrigins)

	// Create a test handler
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create a test request with a disallowed origin
	req := httptest.NewRequest("GET", "https://malicious.com", nil)
	req.Header.Set("Origin", "https://malicious.com")

	// Add the logger to the request context
	ctx := context.WithValue(req.Context(), "logger", &logger)
	req = req.WithContext(ctx)

	// Record the response
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Origin"))
}

// TestCorsMiddleware_WildcardOrigin tests that the middleware allows requests from any origin when "*" is allowed.
func TestCorsMiddleware_WildcardOrigin(t *testing.T) {
	// Create a logger
	logger := zerolog.New(nil)

	allowedOrigins := []string{"*"}
	middleware := CorsMiddleware(allowedOrigins)

	// Create a test handler
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create a test request with an arbitrary origin
	req := httptest.NewRequest("GET", "https://example.com", nil)
	req.Header.Set("Origin", "https://example.com")

	// Add the logger to the request context
	ctx := context.WithValue(req.Context(), "logger", &logger)
	req = req.WithContext(ctx)

	// Record the response
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
}

// TestCorsMiddleware_OptionsRequest tests that the middleware handles OPTIONS requests correctly.
func TestCorsMiddleware_OptionsRequest(t *testing.T) {
	// Create a logger
	logger := zerolog.New(nil)

	allowedOrigins := []string{"https://example.com"}
	middleware := CorsMiddleware(allowedOrigins)

	// Create a test handler
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create a test OPTIONS request
	req := httptest.NewRequest("OPTIONS", "https://example.com", nil)
	req.Header.Set("Origin", "https://example.com")

	// Add the logger to the request context
	ctx := context.WithValue(req.Context(), "logger", &logger)
	req = req.WithContext(ctx)

	// Record the response
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
}

// TestCorsMiddleware_MissingOrigin tests that the middleware does not allow requests without an Origin header.
func TestCorsMiddleware_MissingOrigin(t *testing.T) {
	// Create a logger
	logger := zerolog.New(nil)

	allowedOrigins := []string{"https://example.com"}
	middleware := CorsMiddleware(allowedOrigins)

	// Create a test handler
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create a test request without an Origin header
	req := httptest.NewRequest("GET", "https://example.com", nil)

	// Add the logger to the request context
	ctx := context.WithValue(req.Context(), "logger", &logger)
	req = req.WithContext(ctx)

	// Record the response
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Origin"))
}
