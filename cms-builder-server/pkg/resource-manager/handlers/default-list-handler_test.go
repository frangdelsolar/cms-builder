package resourcemanager_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	rmHandlers "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/handlers"
	testPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
)

func TestDefaultListHandler(t *testing.T) {
	bed := testPkg.SetupHandlerTestBed()

	tests := []struct {
		name           string
		method         string
		path           string
		queryParams    map[string]string
		user           *authModels.User
		setup          func() // Optional setup function for specific test cases
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success",
			method:      http.MethodGet,
			path:        "/mock-struct",
			queryParams: map[string]string{"page": "1", "limit": "10"},
			user:        bed.AdminUser,
			setup: func() {
				// Create some mock resources
				for i := 0; i < 15; i++ {
					instance := testPkg.CreateMockResourceInstance(bed.AdminUser.ID)
					bed.Db.DB.Create(&instance)
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "List",
		},
		{
			name:           "Invalid Method",
			method:         http.MethodPost,
			path:           "/mock-struct",
			queryParams:    map[string]string{"page": "1", "limit": "10"},
			user:           bed.AdminUser,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed",
		},
		{
			name:           "Anonymous is not allowed",
			method:         http.MethodGet,
			path:           "/mock-struct",
			queryParams:    map[string]string{"page": "1", "limit": "10"},
			user:           &authModels.User{},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to read this resource",
		},
		{
			name:           "Unauthorized User",
			method:         http.MethodGet,
			path:           "/mock-struct",
			queryParams:    map[string]string{"page": "1", "limit": "10"},
			user:           bed.NoRoleUser,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to read this resource",
		},
		{
			name:        "Admin Bypass User Binding",
			method:      http.MethodGet,
			path:        "/mock-struct",
			queryParams: map[string]string{"page": "1", "limit": "10"},
			user:        bed.AdminUser,
			setup: func() {
				// Create resources owned by another user
				for i := 0; i < 5; i++ {
					instance := testPkg.CreateMockResourceInstance(bed.VisitorUser.ID)
					bed.Db.DB.Create(&instance)
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "List",
		},
		{
			name:        "User Binding",
			method:      http.MethodGet,
			path:        "/mock-struct",
			queryParams: map[string]string{"page": "1", "limit": "10"},
			user:        bed.VisitorUser,
			setup: func() {
				// Create resources owned by the visitor user
				for i := 0; i < 5; i++ {
					instance := testPkg.CreateMockResourceInstance(bed.VisitorUser.ID)
					bed.Db.DB.Create(&instance)
				}
				// Create resources owned by another user (should not be accessible)
				for i := 0; i < 5; i++ {
					instance := testPkg.CreateMockResourceInstance(bed.AdminUser.ID)
					bed.Db.DB.Create(&instance)
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "List",
		},

		// {
		// 	name:        "Database Error",
		// 	method:      http.MethodGet,
		// 	path:        "/mock-struct",
		// 	queryParams: map[string]string{"page": "1", "limit": "10"},
		// 	user:        bed.AdminUser,
		// 	setup: func() {
		// 		bed.Db.Close() // Simulate a database error
		// 	},
		// 	expectedStatus: http.StatusInternalServerError,
		// 	expectedBody:   "Error finding instances",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the test bed for each test case
			bed := testPkg.SetupHandlerTestBed()

			// Run setup if provided
			if tt.setup != nil {
				tt.setup()
			}

			// Create a Gorilla Mux router
			router := mux.NewRouter()
			router.HandleFunc("/mock-struct", rmHandlers.DefaultListHandler(bed.Src, bed.Db))

			// Create and execute request
			req := testPkg.CreateTestRequest(t, tt.method, tt.path, "", true, tt.user, bed.Logger)
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
		})
	}
}
