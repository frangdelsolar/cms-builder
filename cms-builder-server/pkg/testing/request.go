package testing

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	svrConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/constants"
)

func CreateTestRequest(t *testing.T, method, path, body string, isAuth bool, user *authModels.User, log *loggerTypes.Logger) *http.Request {
	req, err := http.NewRequest(method, path, bytes.NewBufferString(body))
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:3000")

	ctx := req.Context()

	if user != (*authModels.User)(nil) {
		ctx = context.WithValue(ctx, authConstants.CtxRequestIsAuth, isAuth)
		ctx = context.WithValue(ctx, authConstants.CtxRequestUser, user)
	}

	ctx = context.WithValue(ctx, svrConstants.CtxRequestLogger, log)

	return req.WithContext(ctx)
}

// testPkg.ExecuteHandler executes the handler and returns the response recorder
func ExecuteHandler(t *testing.T, handler http.HandlerFunc, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

// HitEndpoint sends a request to the specified endpoint and returns the response.
// HitEndpoint sends a request to the specified endpoint and returns the response.
func HitEndpoint(t *testing.T, handler http.HandlerFunc, method, path, body string, isAuth bool, user *authModels.User, log *loggerTypes.Logger, ip string) *httptest.ResponseRecorder {
	// Create the request
	req := CreateGodRequestWithIp(t, method, path, body, isAuth, user, log, ip)

	// Execute the handler and return the response
	return ExecuteHandler(t, handler, req)
}

// testPkg.CreateTestRequest creates a test request with a custom IP address.
func CreateGodRequestWithIp(t *testing.T, method, path, body string, isAuth bool, user *authModels.User, log *loggerTypes.Logger, ip string) *http.Request {
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
	ctx = context.WithValue(ctx, authConstants.CtxRequestIsAuth, isAuth)
	ctx = context.WithValue(ctx, authConstants.CtxRequestUser, user)
	ctx = context.WithValue(ctx, svrConstants.CtxRequestLogger, log)

	return req.WithContext(ctx)
}
