package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware_LogsRequest(t *testing.T) {

	var logConfig = &loggerTypes.LoggerConfig{
		LogLevel: "info",
	}

	mockUser := CreateAdminUser()

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Retrieve the logger from the context
		loggerFromContext := r.Context().Value(CtxRequestLogger)
		assert.NotNil(t, loggerFromContext)

		w.WriteHeader(http.StatusOK)
	})

	// Wrap the handler with the middleware
	middleware := LoggingMiddleware(logConfig)
	wrappedHandler := middleware(handler)

	// Create a test request
	req := httptest.NewRequest("GET", "https://example.com", nil)

	// Add a request ID to the context
	requestId := "test-request-id"

	ctx := req.Context()
	ctx = context.WithValue(ctx, CtxRequestIsAuth, true)
	ctx = context.WithValue(ctx, CtxRequestUser, mockUser)
	ctx = context.WithValue(ctx, CtxTraceId, requestId)
	req = req.WithContext(ctx)

	// Record the response
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, http.StatusOK, w.Code)
}
