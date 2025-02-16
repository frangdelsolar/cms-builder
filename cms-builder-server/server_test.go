package builder_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	builder "github.com/frangdelsolar/cms-builder/cms-builder-server"
	th "github.com/frangdelsolar/cms-builder/cms-builder-server/test_helpers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// TestNewServer_ValidConfig tests the NewServer function with a valid configuration.
//
// It tests the following:
//   - The returned server is not nil.
//   - The server address is correctly set.
//   - The server root is not nil.
//   - The server middlewares are not nil.
func TestNewServer_ValidConfig(t *testing.T) {
	t.Log("Testing NewServer")
	config := &builder.ServerConfig{
		Host:      "localhost",
		Port:      "8080",
		CSRFToken: "secret",
		Builder:   nil,
	}
	server, err := builder.NewServer(config)

	assert.NoError(t, err)
	assert.Equal(t, "localhost:8080", server.Addr)
	assert.NotNil(t, server.Root)
	assert.NotNil(t, server.Middlewares)
}

// TestAuthenticationMiddleware tests the authentication middleware by registering a user,
// logging in with that user, and verifying that the middleware adds the "auth" header
// to the request.
func TestAuthenticationMiddleware(t *testing.T) {

	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("auth")
		if authHeader != "true" {
			t.Errorf("missing auth header")
		}
	})

	handlerToTest := e.Engine.AuthMiddleware(nextHandler)

	req := httptest.NewRequest("GET", "http://testing", nil)

	userData := th.RandomUserData()
	_, rollback := th.RegisterTestUser(userData)
	defer rollback()

	accessToken, err := th.LoginUser(userData)
	assert.NoError(t, err, "Error logging in user")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	handlerToTest.ServeHTTP(httptest.NewRecorder(), req)
}

func TestRateLimitMiddleware(t *testing.T) {
	rl := builder.NewRateLimiter(2, time.Second) // 2 requests per second
	r := mux.NewRouter()
	r.Use(builder.RateLimitMiddleware(rl))
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	})

	ts := httptest.NewServer(r) // Create a test server
	defer ts.Close()

	// Make requests
	resp, err := http.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

	time.Sleep(2 * time.Second) // Wait for the window to reset

	resp, err = http.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestEngineRateLimitMiddlewareSetup(t *testing.T) {

	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	r := e.Server.Root
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	})

	ts := httptest.NewServer(r) // Create a test server
	defer ts.Close()

	// Make requests within the rate limit
	for i := 0; i < builder.RequestsPerMinute; i++ {
		resp, err := http.Get(ts.URL)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// Make a request outside the rate limit
	resp, err := http.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}

func TestRecoveryMiddleware(t *testing.T) {

	// Create a mock handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("Test panic")
	})

	// Wrap the panic handler with the RecoveryMiddleware
	recoveredHandler := builder.RecoveryMiddleware(panicHandler)

	// Create a test recorder to capture the response
	recorder := httptest.NewRecorder()

	// Create a test request
	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(context.Background()) // Add a context (important!)

	// Serve the request
	recoveredHandler.ServeHTTP(recorder, req)

	// Check the response status code
	assert.Equal(t, http.StatusInternalServerError, recorder.Code, "Expected 500 status code")

	// Check the response body (optional)
	assert.Equal(t, "Internal Server Error\n", recorder.Body.String(), "Expected error message")

}

func TestTimeoutMiddleware(t *testing.T) {

	// Create a mock handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secs := builder.TimeoutSeconds + 1
		t.Log("Testing timeout middleware waiting for", secs, "seconds")
		time.Sleep(time.Duration(secs) * time.Second)
	})

	// Wrap the panic handler with the RecoveryMiddleware
	recoveredHandler := builder.TimeoutMiddleware(panicHandler)

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
