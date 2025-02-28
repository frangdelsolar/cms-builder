package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/stretchr/testify/assert"
)

// TestRecoveryMiddleware_NoPanic tests the RecoveryMiddleware when the handler does not panic.
func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	// Create a test handler that does not panic
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap the handler with the middleware
	middleware := RecoveryMiddleware(handler)

	// Create a test request
	req := httptest.NewRequest("GET", "https://example.com", nil)

	// Record the response
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

// TestRecoveryMiddleware_Panic tests the RecoveryMiddleware when the handler panics.
func TestRecoveryMiddleware_Panic(t *testing.T) {
	// Create a test handler that panics
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})

	// Wrap the handler with the middleware
	middleware := RecoveryMiddleware(handler)

	// Create a test request
	req := httptest.NewRequest("GET", "https://example.com", nil)

	// Record the response
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, `{"success":false,"data":null,"message":"Internal Server Error","pagination":null}`, w.Body.String())
}

// TestRecoveryMiddleware_UnhandledError tests the RecoveryMiddleware when the handler returns an unhandled error.
func TestRecoveryMiddleware_UnhandledError(t *testing.T) {
	// Create a test handler that returns an error response
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate an unhandled error by returning a 500 Internal Server Error
		SendJsonResponse(w, http.StatusInternalServerError, nil, "Something went wrong")
	})

	// Wrap the handler with the middleware
	middleware := RecoveryMiddleware(handler)

	// Create a test request
	req := httptest.NewRequest("GET", "https://example.com", nil)

	// Record the response
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, `{"success":false,"data":null,"message":"Something went wrong","pagination":null}`, w.Body.String())
}
