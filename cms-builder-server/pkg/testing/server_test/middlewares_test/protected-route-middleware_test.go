package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

// TestProtectedRouteMiddleware tests the ProtectedRouteMiddleware function.
func TestProtectedRouteMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		auth           bool
		user           *authModels.User
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "authorized request",
			auth:           true,
			user:           &authModels.User{ID: 1, Name: "John Doe"},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "unauthorized request - not authenticated",
			auth:           false,
			user:           nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"data":null,"message":"Unauthorized","pagination":null,"success":false}`,
		},
		{
			name:           "unauthorized request - no user",
			auth:           true,
			user:           nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"data":null,"message":"Unauthorized","pagination":null,"success":false}`,
		},
		{
			name:           "unauthorized request - invalid user",
			auth:           true,
			user:           &authModels.User{ID: 0, Name: "Invalid User"},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"data":null,"message":"Unauthorized","pagination":null,"success":false}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Wrap the handler with the middleware
			wrappedHandler := ProtectedRouteMiddleware(handler)

			// Create a test request
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			// Add context values to the request
			ctx := context.WithValue(req.Context(), CtxRequestIsAuth, tt.auth)
			ctx = context.WithValue(ctx, CtxRequestUser, tt.user)
			req = req.WithContext(ctx)

			// Record the response
			w := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(w, req)

			// Check the response status code
			assert.Equal(t, tt.expectedStatus, w.Code, "Unexpected status code")

			// Check the response body
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String(), "Unexpected response body")
			}
		})
	}
}
