package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	svrConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/constants"
	svrMiddlewares "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/middlewares"
	testPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
)

func TestLoggingMiddleware_LogsRequest(t *testing.T) {

	var logConfig = &loggerTypes.LoggerConfig{
		LogLevel: "info",
	}

	mockUser := testPkg.CreateAdminUser()

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Retrieve the logger from the context
		loggerFromContext := r.Context().Value(svrConstants.CtxRequestLogger)
		assert.NotNil(t, loggerFromContext)

		w.WriteHeader(http.StatusOK)
	})

	// Wrap the handler with the middleware
	middleware := svrMiddlewares.LoggingMiddleware(logConfig)
	wrappedHandler := middleware(handler)

	// Create a test request
	req := httptest.NewRequest("GET", "https://example.com", nil)

	// Add a request ID to the context
	requestId := "test-request-id"

	ctx := req.Context()
	ctx = context.WithValue(ctx, authConstants.CtxRequestIsAuth, true)
	ctx = context.WithValue(ctx, authConstants.CtxRequestUser, mockUser)
	ctx = context.WithValue(ctx, svrConstants.CtxTraceId, requestId)
	req = req.WithContext(ctx)

	// Record the response
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, http.StatusOK, w.Code)
}
