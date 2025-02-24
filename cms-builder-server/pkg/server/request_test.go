package server_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/gorilla/mux"
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

// TestUserIsAllowed tests the UserIsAllowed function.
func TestUserIsAllowed(t *testing.T) {

	const EditorRole models.Role = "editor"
	// Define test cases
	tests := []struct {
		name           string
		appPermissions RolePermissionMap
		userRoles      []models.Role
		action         CrudOperation
		expectedResult bool
	}{
		{
			name: "user has permission for the action",
			appPermissions: RolePermissionMap{
				models.AdminRole: {OperationRead, OperationCreate},
				EditorRole:       {OperationRead},
			},
			userRoles:      []models.Role{models.AdminRole},
			action:         OperationRead,
			expectedResult: true,
		},
		{
			name: "user does not have permission for the action",
			appPermissions: RolePermissionMap{
				models.AdminRole: {OperationCreate},
				EditorRole:       {OperationRead},
			},
			userRoles:      []models.Role{EditorRole},
			action:         OperationCreate,
			expectedResult: false,
		},
		{
			name: "user has no roles",
			appPermissions: RolePermissionMap{
				models.AdminRole: {OperationRead, OperationCreate},
				EditorRole:       {OperationRead},
			},
			userRoles:      []models.Role{},
			action:         OperationRead,
			expectedResult: false,
		},
		{
			name: "role has no permissions",
			appPermissions: RolePermissionMap{
				models.AdminRole: {},
				EditorRole:       {OperationRead},
			},
			userRoles:      []models.Role{models.AdminRole},
			action:         OperationRead,
			expectedResult: false,
		},
		{
			name: "role does not exist in permissions",
			appPermissions: RolePermissionMap{
				EditorRole: {OperationRead},
			},
			userRoles:      []models.Role{models.AdminRole},
			action:         OperationRead,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function
			result := UserIsAllowed(tt.appPermissions, tt.userRoles, tt.action)

			// Verify the result
			assert.Equal(t, tt.expectedResult, result, "Unexpected result for test case: %s", tt.name)
		})
	}
}

// TestReadRequestBody tests the ReadRequestBody function.
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

			// Call the ReadRequestBody function
			result, err := ReadRequestBody(req)

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
			result, err := FormatRequestBody(req, tt.filterKeys)

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

// TestGetUrlParam tests the GetUrlParam function.
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
				// Call the GetUrlParam function
				result := GetUrlParam(tt.param, r)

				// Verify the result
				assert.Equal(t, tt.expectedResult, result, "Unexpected result for test case: %s", tt.name)
			})

			// Match the request to the router
			router.ServeHTTP(httptest.NewRecorder(), req)
		})
	}
}
