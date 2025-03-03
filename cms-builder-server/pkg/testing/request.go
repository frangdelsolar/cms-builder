package testing

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/stretchr/testify/assert"
)

func CreateTestRequest(t *testing.T, method, path, body string, isAuth bool, user *models.User, log *logger.Logger) *http.Request {
	req, err := http.NewRequest(method, path, bytes.NewBufferString(body))
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:3000")

	ctx := req.Context()
	ctx = context.WithValue(ctx, server.CtxRequestIsAuth, isAuth)
	ctx = context.WithValue(ctx, server.CtxRequestUser, user)
	ctx = context.WithValue(ctx, server.CtxRequestLogger, log)

	return req.WithContext(ctx)
}

// ExecuteHandler executes the handler and returns the response recorder
func ExecuteHandler(t *testing.T, handler http.HandlerFunc, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

// HitEndpoint sends a request to the specified endpoint and returns the response.
// HitEndpoint sends a request to the specified endpoint and returns the response.
func HitEndpoint(t *testing.T, handler http.HandlerFunc, method, path, body string, isAuth bool, user *models.User, log *logger.Logger, ip string) *httptest.ResponseRecorder {
	// Create the request
	req := CreateGodRequestWithIp(t, method, path, body, isAuth, user, log, ip)

	// Execute the handler and return the response
	return ExecuteHandler(t, handler, req)
}

// CreateTestRequest creates a test request with a custom IP address.
func CreateGodRequestWithIp(t *testing.T, method, path, body string, isAuth bool, user *models.User, log *logger.Logger, ip string) *http.Request {
	req, err := http.NewRequest(method, path, bytes.NewBufferString(body))
	assert.NoError(t, err)

	// Set the IP address for the request
	req.RemoteAddr = ip + ":12345" // Add a port to make it a valid address

	// Alternatively, you can use the X-Forwarded-For header
	req.Header.Set("X-Forwarded-For", ip)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("X-God-Token", os.Getenv("GOD_TOKEN"))

	req.Body = io.NopCloser(bytes.NewBufferString(body))

	ctx := req.Context()
	ctx = context.WithValue(ctx, server.CtxRequestIsAuth, isAuth)
	ctx = context.WithValue(ctx, server.CtxRequestUser, user)
	ctx = context.WithValue(ctx, server.CtxRequestLogger, log)

	return req.WithContext(ctx)
}
