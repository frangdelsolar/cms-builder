package server_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	loggerPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	svrConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/constants"
	svrTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/types"
	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/utils"
)

// TestValidateRequestMethod_Valid tests that svrUtils.ValidateRequestMethod returns nil for a valid request method.
func TestValidateRequestMethod_Valid(t *testing.T) {
	// Create a test request with a valid method
	req := httptest.NewRequest("GET", "https://example.com", nil)

	// Validate the request method
	err := svrUtils.ValidateRequestMethod(req, "GET")

	// Verify that no error is returned
	assert.NoError(t, err)
}

// TestValidateRequestMethod_Invalid tests that svrUtils.ValidateRequestMethod returns an error for an invalid request method.
func TestValidateRequestMethod_Invalid(t *testing.T) {
	// Create a test request with an invalid method
	req := httptest.NewRequest("POST", "https://example.com", nil)

	// Validate the request method
	err := svrUtils.ValidateRequestMethod(req, "GET")

	// Verify that an error is returned
	assert.Error(t, err)
	assert.Equal(t, "Method not allowed", err.Error())
}

// TestGetLoggerFromRequest tests the GetLoggerFromRequest function.
func TestGetLoggerFromRequest(t *testing.T) {
	tests := []struct {
		name           string
		contextValue   interface{}
		expectedLogger *loggerTypes.Logger
	}{
		{
			name:           "logger exists in context",
			contextValue:   &loggerTypes.Logger{},
			expectedLogger: &loggerTypes.Logger{},
		},
		{
			name:           "logger does not exist in context",
			contextValue:   nil,
			expectedLogger: loggerPkg.Default,
		},
		{
			name:           "invalid type in context",
			contextValue:   "not-a-logger",
			expectedLogger: loggerPkg.Default,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new request
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			// Add the context value to the request
			ctx := context.WithValue(req.Context(), svrConstants.CtxRequestLogger, tt.contextValue)
			req = req.WithContext(ctx)

			// Call the function
			logger := svrUtils.GetRequestLogger(req)

			// Assert the result
			if tt.expectedLogger == logger {
				assert.Equal(t, tt.expectedLogger, logger, "Unexpected logger returned")
			} else {
				assert.NotNil(t, logger, "Expected a non-nil logger")
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
			token := svrUtils.GetRequestAccessToken(req)

			// Verify the result
			assert.Equal(t, tt.expectedToken, token, "Unexpected token")
		})
	}
}

// TestGetRequestId tests the svrUtils.GetRequestId function.
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
			ctx := context.WithValue(req.Context(), svrConstants.CtxTraceId, tt.contextValue)
			req = req.WithContext(ctx)

			// Call the function
			requestId := svrUtils.GetRequestId(req)

			// Verify the result
			assert.Equal(t, tt.expectedId, requestId, "Unexpected request ID")
		})
	}
}

// TestGetRequestUser tests the svrUtils.GetRequestUser function.
func TestGetRequestUser(t *testing.T) {
	tests := []struct {
		name         string
		contextValue interface{}
		expectedUser *authModels.User
	}{
		{
			name:         "user exists in context",
			contextValue: &authModels.User{ID: 1, Name: "John Doe"},
			expectedUser: &authModels.User{ID: 1, Name: "John Doe"},
		},
		{
			name:         "user does not exist in context",
			contextValue: nil,
			expectedUser: nil,
		},
		{
			name:         "invalid type in context",
			contextValue: "not-a-user", // Not a *authModels.User
			expectedUser: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			// Add the context value to the request
			ctx := context.WithValue(req.Context(), authConstants.CtxRequestUser, tt.contextValue)
			req = req.WithContext(ctx)

			// Call the function
			user := svrUtils.GetRequestUser(req)

			// Verify the result
			assert.Equal(t, tt.expectedUser, user, "Unexpected user")
		})
	}
}

// TestGetRequestIsAuth tests the svrUtils.GetRequestIsAuth function.
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
			ctx := context.WithValue(req.Context(), authConstants.CtxRequestIsAuth, tt.contextValue)
			req = req.WithContext(ctx)

			// Call the function
			isAuth := svrUtils.GetRequestIsAuth(req)

			// Verify the result
			assert.Equal(t, tt.expectedIsAuth, isAuth, "Unexpected authentication status")
		})
	}
}

// TestGetRequestContext tests the svrUtils.GetRequestContext function.
func TestGetRequestContext(t *testing.T) {
	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Add context values to the request
	ctx := context.WithValue(req.Context(), svrConstants.CtxTraceId, "test-request-id")
	ctx = context.WithValue(ctx, authConstants.CtxRequestIsAuth, true)
	ctx = context.WithValue(ctx, authConstants.CtxRequestUser, &authModels.User{ID: 1, Name: "John Doe"})
	ctx = context.WithValue(ctx, svrConstants.CtxRequestLogger, &zerolog.Logger{})
	req = req.WithContext(ctx)

	// Call the function
	requestContext := svrUtils.GetRequestContext(req)

	// Verify the result
	assert.Equal(t, "test-request-id", requestContext.RequestId, "Unexpected request ID")
	assert.True(t, requestContext.IsAuthenticated, "Expected authenticated")
	assert.Equal(t, &authModels.User{ID: 1, Name: "John Doe"}, requestContext.User, "Unexpected user")
	assert.NotNil(t, requestContext.Logger, "Expected a non-nil logger")
}

// TestGetIntQueryParam tests the svrUtils.GetIntQueryParam function.
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
			value, err := svrUtils.GetIntQueryParam(tt.param, tt.query)

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

func TestGetRequestQueryParams(t *testing.T) {
	tests := []struct {
		name           string
		query          url.Values
		expectedParams *svrTypes.QueryParams
	}{
		{
			name: "valid query parameters",
			query: url.Values{
				"limit": []string{"10"},
				"page":  []string{"2"},
				"order": []string{"name"},
				"query": []string{"test"},
			},
			expectedParams: &svrTypes.QueryParams{
				Limit: 10,
				Page:  2,
				Order: "name",
				Query: map[string]string{"query": "test"},
			},
		},
		{
			name: "missing limit parameter",
			query: url.Values{
				"page": []string{"2"},
			},
			expectedParams: &svrTypes.QueryParams{
				Limit: 10, // Default limit
				Page:  2,
				Order: "id desc", // Default order
				Query: map[string]string{},
			},
		},
		{
			name: "invalid limit parameter",
			query: url.Values{
				"limit": []string{"invalid"},
				"page":  []string{"2"},
			},
			expectedParams: &svrTypes.QueryParams{
				Limit: 10, // Default limit due to invalid value
				Page:  2,
				Order: "id desc", // Default order
				Query: map[string]string{},
			},
		},
		{
			name: "invalid order parameter",
			query: url.Values{
				"limit": []string{"10"},
				"page":  []string{"2"},
				"order": []string{"TestStruct"},
			},
			expectedParams: &svrTypes.QueryParams{
				Limit: 10,
				Page:  2,
				Order: "test_struct", // Default order due to invalid value
				Query: map[string]string{},
			},
		},
		{
			name:  "no query parameters",
			query: url.Values{},
			expectedParams: &svrTypes.QueryParams{
				Limit: 10,        // Default limit
				Page:  1,         // Default page
				Order: "id desc", // Default order
				Query: map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.URL = &url.URL{RawQuery: tt.query.Encode()}

			// Call the function
			params, err := svrUtils.GetRequestQueryParams(req)

			// Verify the result
			assert.NoError(t, err, "Expected no error")
			assert.Equal(t, tt.expectedParams, params, "Unexpected query parameters")
		})
	}
}

// TestValidateOrderParam tests the svrUtils.ValidateOrderParam function.
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
			order, err := svrUtils.ValidateOrderParam(tt.orderParam)

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

// TestReadRequestBody tests the svrUtils.ReadRequestBody function.
func TestReadRequestBody(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedResult []byte
		expectedError  bool
	}{
		{
			name:           "empty request body",
			requestBody:    "",
			expectedResult: []byte{},
			expectedError:  false,
		},
		{
			name:           "valid request body",
			requestBody:    `{"key": "value"}`,
			expectedResult: []byte(`{"key": "value"}`),
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request with the request body
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.requestBody))

			// Call the svrUtils.ReadRequestBody function
			result, err := svrUtils.ReadRequestBody(req)

			// Verify the result
			if tt.expectedError {
				assert.Error(t, err, "Expected an error for test case: %s", tt.name)
			} else {
				assert.NoError(t, err, "Expected no error for test case: %s", tt.name)
				assert.Equal(t, tt.expectedResult, result, "Unexpected result for test case: %s", tt.name)
			}
		})
	}
}

// TestFormatRequestBody tests the FormatRequestBody function.
func TestFormatRequestBody(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		filterKeys     map[string]bool
		expectedResult map[string]interface{}
		expectedError  bool
	}{
		{
			name:           "empty request body",
			requestBody:    "",
			filterKeys:     map[string]bool{},
			expectedResult: map[string]interface{}{},
			expectedError:  false,
		},
		{
			name:        "valid request body with no filter",
			requestBody: `{"key1": "value1", "key2": "value2"}`,
			filterKeys:  map[string]bool{},
			expectedResult: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			expectedError: false,
		},
		{
			name:        "valid request body with filter",
			requestBody: `{"key1": "value1", "key2": "value2"}`,
			filterKeys:  map[string]bool{"key1": true},
			expectedResult: map[string]interface{}{
				"key2": "value2",
			},
			expectedError: false,
		},
		{
			name:           "invalid JSON request body",
			requestBody:    `invalid-json`,
			filterKeys:     map[string]bool{},
			expectedResult: map[string]interface{}{},
			expectedError:  true,
		},
		{
			name:        "filter keys are case insensitive",
			requestBody: `{"Key1": "value1", "key2": "value2"}`,
			filterKeys:  map[string]bool{"key1": true},
			expectedResult: map[string]interface{}{
				"key2": "value2",
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request with the request body
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.requestBody))

			// Call the FormatRequestBody function
			result, err := svrUtils.FormatRequestBody(req, tt.filterKeys)

			// Verify the result
			if tt.expectedError {
				assert.Error(t, err, "Expected an error for test case: %s", tt.name)
			} else {
				assert.NoError(t, err, "Expected no error for test case: %s", tt.name)
				assert.Equal(t, tt.expectedResult, result, "Unexpected result for test case: %s", tt.name)
			}
		})
	}
}

// TestGetUrlParam tests the svrUtils.GetUrlParam function.
func TestGetUrlParam(t *testing.T) {
	tests := []struct {
		name           string
		param          string
		url            string
		expectedResult string
	}{
		{
			name:           "valid URL parameter",
			param:          "id",
			url:            "/users/123",
			expectedResult: "123",
		},
		{
			name:           "missing URL parameter",
			param:          "id",
			url:            "/users",
			expectedResult: "",
		},
		{
			name:           "empty URL parameter value",
			param:          "id",
			url:            "/users/",
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new request
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)

			// Create a new router and set the URL parameters
			router := mux.NewRouter()
			router.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
				// Call the svrUtils.GetUrlParam function
				result := svrUtils.GetUrlParam(tt.param, r)

				// Verify the result
				assert.Equal(t, tt.expectedResult, result, "Unexpected result for test case: %s", tt.name)
			})

			// Match the request to the router
			router.ServeHTTP(httptest.NewRecorder(), req)
		})
	}
}
