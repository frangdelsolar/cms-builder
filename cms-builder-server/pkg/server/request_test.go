package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// TestValidateRequestMethod_Valid tests that ValidateRequestMethod returns nil for a valid request method.
func TestValidateRequestMethod_Valid(t *testing.T) {
	// Create a test request with a valid method
	req := httptest.NewRequest("GET", "https://example.com", nil)

	// Validate the request method
	err := ValidateRequestMethod(req, "GET")

	// Verify that no error is returned
	assert.NoError(t, err)
}

// TestValidateRequestMethod_Invalid tests that ValidateRequestMethod returns an error for an invalid request method.
func TestValidateRequestMethod_Invalid(t *testing.T) {
	// Create a test request with an invalid method
	req := httptest.NewRequest("POST", "https://example.com", nil)

	// Validate the request method
	err := ValidateRequestMethod(req, "GET")

	// Verify that an error is returned
	assert.Error(t, err)
	assert.Equal(t, "invalid request method: POST", err.Error())
}

// TestGetLoggerFromRequest tests the GetLoggerFromRequest function.
func TestGetLoggerFromRequest(t *testing.T) {
	tests := []struct {
		name           string
		contextValue   interface{}
		expectedLogger *zerolog.Logger
	}{
		{
			name:           "logger exists in context",
			contextValue:   &zerolog.Logger{},
			expectedLogger: &zerolog.Logger{},
		},
		{
			name:           "logger does not exist in context",
			contextValue:   nil,
			expectedLogger: logger.Default.Logger,
		},
		{
			name:           "invalid type in context",
			contextValue:   "not-a-logger",
			expectedLogger: logger.Default.Logger,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new request
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			// Add the context value to the request
			ctx := context.WithValue(req.Context(), CtxRequestLogger, tt.contextValue)
			req = req.WithContext(ctx)

			// Call the function
			logger := GetLoggerFromRequest(req)

			// Assert the result
			if tt.expectedLogger == (&zerolog.Logger{}) {
				assert.NotNil(t, logger, "Expected a non-nil logger")
			} else {
				assert.Equal(t, tt.expectedLogger, logger, "Unexpected logger returned")
			}
		})
	}
}
