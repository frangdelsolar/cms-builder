package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestDefaultListHandler(t *testing.T) {
	bed := SetupAuthTestBed()

	tests := []struct {
		name           string
		method         string
		path           string
		queryParams    map[string]string
		user           *models.User
		setup          func() // Optional setup function for specific test cases
		expectedStatus int
		expectedBody   string
		expectedCount  int
	}{
		{
			name:           "Invalid Method",
			method:         http.MethodPost,
			path:           "/user",
			queryParams:    map[string]string{"page": "1", "limit": "10"},
			user:           bed.AdminUser,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed",
		},

		{
			name:           "Unauthorized User",
			method:         http.MethodGet,
			path:           "/user",
			queryParams:    map[string]string{"page": "1", "limit": "10"},
			user:           bed.NoRoleUser,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to read this resource",
		},

		{
			name:           "Anonymous is not allowed",
			method:         http.MethodGet,
			path:           "/user",
			queryParams:    map[string]string{"page": "1", "limit": "10"},
			user:           &models.User{},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to read this resource",
		},
		{
			name:        "Admin can list others and his own",
			method:      http.MethodGet,
			path:        "/user",
			queryParams: map[string]string{"page": "1", "limit": "10"},
			user:        bed.AdminUser,
			setup: func() {
				// Create some users
				for i := 0; i < 5; i++ {
					instance := CreateNoRoleUser()
					bed.Db.DB.Create(&instance)
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "List",
			expectedCount:  10, // created + system users
		},
		{
			name:        "Visitors can just list themselves",
			method:      http.MethodGet,
			path:        "/user",
			queryParams: map[string]string{"page": "1", "limit": "10"},
			user:        bed.VisitorUser,
			setup: func() {
				for i := 0; i < 5; i++ {
					instance := CreateNoRoleUser()
					bed.Db.DB.Create(&instance)
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "List",
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the test bed for each test case
			bed := SetupAuthTestBed()

			// Run setup if provided
			if tt.setup != nil {
				tt.setup()
			}

			// Create a Gorilla Mux router
			router := mux.NewRouter()
			router.HandleFunc("/user", auth.UserListHandler(bed.Src, bed.Db))

			// Create and execute request
			req := CreateTestRequest(t, tt.method, tt.path, "", true, tt.user, bed.Logger)
			// Add query parameters to the request
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			rr := httptest.NewRecorder()

			// Serve the request using the router
			router.ServeHTTP(rr, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)

			if tt.expectedStatus != http.StatusOK {
				return
			}

			var response server.Response

			err := json.Unmarshal(rr.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedCount, len(response.Data.([]interface{})))

		})
	}
}
