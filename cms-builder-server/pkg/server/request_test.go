package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
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
			logger := GetRequestLogger(req)

			// Assert the result
			if tt.expectedLogger == (&zerolog.Logger{}) {
				assert.NotNil(t, logger, "Expected a non-nil logger")
			} else {
				assert.Equal(t, tt.expectedLogger, logger, "Unexpected logger returned")
			}
		})
	}
}

// TestGetRequestAccessToken tests the GetRequestAccessToken function.
func TestGetRequestAccessToken(t *testing.T) {
	tests := []struct {
		name          string
		header        string
		expectedToken string
	}{
		{
			name:          "valid token",
			header:        "Bearer valid-token",
			expectedToken: "valid-token",
		},
		{
			name:          "empty header",
			header:        "",
			expectedToken: "",
		},
		{
			name:          "malformed header",
			header:        "InvalidHeader",
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			// Set the Authorization header
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			// Call the function
			token := GetRequestAccessToken(req)

			// Verify the result
			assert.Equal(t, tt.expectedToken, token, "Unexpected token")
		})
	}
}

// TestGetRequestId tests the GetRequestId function.
func TestGetRequestId(t *testing.T) {
	tests := []struct {
		name         string
		contextValue interface{}
		expectedId   string
	}{
		{
			name:         "request ID exists in context",
			contextValue: "test-request-id",
			expectedId:   "test-request-id",
		},
		{
			name:         "request ID does not exist in context",
			contextValue: nil,
			expectedId:   "",
		},
		{
			name:         "invalid type in context",
			contextValue: 123, // Not a string
			expectedId:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			// Add the context value to the request
			ctx := context.WithValue(req.Context(), CtxRequestIdentifier, tt.contextValue)
			req = req.WithContext(ctx)

			// Call the function
			requestId := GetRequestId(req)

			// Verify the result
			assert.Equal(t, tt.expectedId, requestId, "Unexpected request ID")
		})
	}
}

// TestGetRequestUser tests the GetRequestUser function.
func TestGetRequestUser(t *testing.T) {
	tests := []struct {
		name         string
		contextValue interface{}
		expectedUser *models.User
	}{
		{
			name:         "user exists in context",
			contextValue: &models.User{ID: 1, Name: "John Doe"},
			expectedUser: &models.User{ID: 1, Name: "John Doe"},
		},
		{
			name:         "user does not exist in context",
			contextValue: nil,
			expectedUser: nil,
		},
		{
			name:         "invalid type in context",
			contextValue: "not-a-user", // Not a *models.User
			expectedUser: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			// Add the context value to the request
			ctx := context.WithValue(req.Context(), CtxRequestUser, tt.contextValue)
			req = req.WithContext(ctx)

			// Call the function
			user := GetRequestUser(req)

			// Verify the result
			assert.Equal(t, tt.expectedUser, user, "Unexpected user")
		})
	}
}

// TestGetRequestIsAuth tests the GetRequestIsAuth function.
func TestGetRequestIsAuth(t *testing.T) {
	tests := []struct {
		name           string
		contextValue   interface{}
		expectedIsAuth bool
	}{
		{
			name:           "authenticated",
			contextValue:   true,
			expectedIsAuth: true,
		},
		{
			name:           "not authenticated",
			contextValue:   false,
			expectedIsAuth: false,
		},
		{
			name:           "invalid type in context",
			contextValue:   "not-a-bool", // Not a bool
			expectedIsAuth: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			// Add the context value to the request
			ctx := context.WithValue(req.Context(), CtxRequestIsAuth, tt.contextValue)
			req = req.WithContext(ctx)

			// Call the function
			isAuth := GetRequestIsAuth(req)

			// Verify the result
			assert.Equal(t, tt.expectedIsAuth, isAuth, "Unexpected authentication status")
		})
	}
}

// TestGetRequestContext tests the GetRequestContext function.
func TestGetRequestContext(t *testing.T) {
	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Add context values to the request
	ctx := context.WithValue(req.Context(), CtxRequestIdentifier, "test-request-id")
	ctx = context.WithValue(ctx, CtxRequestIsAuth, true)
	ctx = context.WithValue(ctx, CtxRequestUser, &models.User{ID: 1, Name: "John Doe"})
	ctx = context.WithValue(ctx, CtxRequestLogger, &zerolog.Logger{})
	req = req.WithContext(ctx)

	// Call the function
	requestContext := GetRequestContext(req)

	// Verify the result
	assert.Equal(t, "test-request-id", requestContext.RequestId, "Unexpected request ID")
	assert.True(t, requestContext.IsAuthenticated, "Expected authenticated")
	assert.Equal(t, &models.User{ID: 1, Name: "John Doe"}, requestContext.User, "Unexpected user")
	assert.NotNil(t, requestContext.Logger, "Expected a non-nil logger")
}

// TestGetIntQueryParam tests the GetIntQueryParam function.
func TestGetIntQueryParam(t *testing.T) {
	tests := []struct {
		name          string
		param         string
		query         url.Values
		expectedValue int
		expectedError string
	}{
		{
			name:          "valid parameter",
			param:         "limit",
			query:         url.Values{"limit": []string{"10"}},
			expectedValue: 10,
			expectedError: "",
		},
		{
			name:          "missing parameter",
			param:         "limit",
			query:         url.Values{},
			expectedValue: 0,
			expectedError: "missing limit parameter",
		},
		{
			name:          "invalid parameter",
			param:         "limit",
			query:         url.Values{"limit": []string{"invalid"}},
			expectedValue: 0,
			expectedError: "invalid limit parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function
			value, err := GetIntQueryParam(tt.param, tt.query)

			// Verify the result
			if tt.expectedError == "" {
				assert.NoError(t, err, "Expected no error")
				assert.Equal(t, tt.expectedValue, value, "Unexpected value")
			} else {
				assert.Error(t, err, "Expected an error")
				assert.Equal(t, tt.expectedError, err.Error(), "Unexpected error message")
			}
		})
	}
}

// TestGetRequestQueryParams tests the GetRequestQueryParams function.
func TestGetRequestQueryParams(t *testing.T) {
	tests := []struct {
		name           string
		query          url.Values
		expectedParams *QueryParams
		expectedError  string
	}{
		{
			name: "valid query parameters",
			query: url.Values{
				"limit": []string{"10"},
				"page":  []string{"2"},
				"order": []string{"name"},
				"query": []string{"test"},
			},
			expectedParams: &QueryParams{
				Limit: 10,
				Page:  2,
				Order: "name",
				Query: map[string]string{"query": "test"},
			},
			expectedError: "",
		},
		{
			name: "missing limit parameter",
			query: url.Values{
				"page": []string{"2"},
			},
			expectedParams: nil,
			expectedError:  "missing limit parameter",
		},
		{
			name: "invalid limit parameter",
			query: url.Values{
				"limit": []string{"invalid"},
				"page":  []string{"2"},
			},
			expectedParams: nil,
			expectedError:  "invalid limit parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.URL = &url.URL{RawQuery: tt.query.Encode()}

			// Call the function
			params, err := GetRequestQueryParams(req)

			// Verify the result
			if tt.expectedError == "" {
				assert.NoError(t, err, "Expected no error")
				assert.Equal(t, tt.expectedParams, params, "Unexpected query parameters")
			} else {
				assert.Error(t, err, "Expected an error")
				assert.Equal(t, tt.expectedError, err.Error(), "Unexpected error message")
			}
		})
	}
}

// TestValidateOrderParam tests the ValidateOrderParam function.
func TestValidateOrderParam(t *testing.T) {
	tests := []struct {
		name          string
		orderParam    string
		expectedOrder string
		expectedError string
	}{
		{
			name:          "valid order parameter",
			orderParam:    "name,-age",
			expectedOrder: "name,age desc",
			expectedError: "",
		},
		{
			name:          "empty order parameter",
			orderParam:    "",
			expectedOrder: "",
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function
			order, err := ValidateOrderParam(tt.orderParam)

			// Verify the result
			if tt.expectedError == "" {
				assert.NoError(t, err, "Expected no error")
				assert.Equal(t, tt.expectedOrder, order, "Unexpected order")
			} else {
				assert.Error(t, err, "Expected an error")
				assert.Equal(t, tt.expectedError, err.Error(), "Unexpected error message")
			}
		})
	}
}
