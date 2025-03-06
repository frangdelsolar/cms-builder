package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	svrMiddlewares "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/middlewares"
)

func TestTimeoutMiddleware(t *testing.T) {

	// Create a mock handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secs := svrMiddlewares.TimeoutSeconds + 1
		t.Log("Testing timeout middleware waiting for", secs, "seconds")
		time.Sleep(time.Duration(secs) * time.Second)
	})

	// Wrap the panic handler with the RecoveryMiddleware
	recoveredHandler := svrMiddlewares.TimeoutMiddleware(panicHandler)

	// Create a test recorder to capture the response
	recorder := httptest.NewRecorder()

	// Create a test request
	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(context.Background()) // Add a context (important!)

	// Serve the request
	recoveredHandler.ServeHTTP(recorder, req)

	// Check the response status code
	assert.Equal(t, http.StatusGatewayTimeout, recorder.Code, "Expected 504 status code")

	// Check the response body (optional)
	assert.Equal(t, "Request timed out\n", recorder.Body.String(), "Expected error message")

}
