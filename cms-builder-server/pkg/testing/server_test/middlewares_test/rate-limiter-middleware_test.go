package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	svrMiddlewares "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/middlewares"
)

func TestRateLimitMiddleware(t *testing.T) {
	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the handler with the middleware
	middleware := svrMiddlewares.RateLimitMiddleware()
	wrappedHandler := middleware(handler)

	// Create a test request
	req := httptest.NewRequest("GET", "https://example.com", nil)

	// Record the response
	w := httptest.NewRecorder()

	// Test within the rate limit
	for i := 0; i < svrMiddlewares.MaxRequestsPerMinute; i++ {
		wrappedHandler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		w = httptest.NewRecorder() // reset the recorder.
	}

	// Test exceeding the rate limit
	wrappedHandler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// Test after the window has passed
	t.Log("Waiting for the window to pass for", svrMiddlewares.WaitingSeconds+1, "seconds")
	time.Sleep((svrMiddlewares.WaitingSeconds + 1) * time.Second)
	w = httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test with different IP address
	req2 := httptest.NewRequest("GET", "https://example.com", nil)
	req2.RemoteAddr = "192.168.1.2:1234"
	w2 := httptest.NewRecorder()

	for i := 0; i < svrMiddlewares.MaxRequestsPerMinute; i++ {
		wrappedHandler.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)
		w2 = httptest.NewRecorder()
	}

	wrappedHandler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)

}

func TestRateLimiter_Allow(t *testing.T) {
	rl := svrMiddlewares.NewRateLimiter(5, 1*time.Second) // 5 requests per second
	clientIP := "192.168.1.1"

	// Allow 5 requests
	for i := 0; i < 5; i++ {
		assert.True(t, rl.Allow(clientIP))
	}

	// 6th request should be denied
	assert.False(t, rl.Allow(clientIP))

	// Wait for the window to pass
	time.Sleep(1 * time.Second)

	// First request after the window should be allowed
	assert.True(t, rl.Allow(clientIP))
}
