package resourcemanager_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing/resource-manager_test"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestDefaultDetailHandler(t *testing.T) {
	bed := SetupHandlerTestBed()

	tests := []struct {
		name           string
		method         string
		path           string
		requestBody    string
		user           *authModels.User
		setup          func() string // Optional setup function for specific test cases
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success",
			method:      http.MethodGet,
			path:        "",
			requestBody: "",
			user:        bed.AdminUser,
			setup: func() string {
				instance := CreateMockResourceInstance(bed.AdminUser.ID)
				bed.Db.DB.Create(&instance)
				return "/mock-struct/" + instance.StringID()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Detail",
		},
		{
			name:           "Invalid Method",
			method:         http.MethodPost,
			path:           "/mock-struct/123",
			requestBody:    "",
			user:           bed.AdminUser,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed",
		},
		{
			name:           "Anonymous is not allowed",
			method:         http.MethodGet,
			path:           "/mock-struct/123",
			requestBody:    "",
			user:           &authModels.User{},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to access this resource",
		},
		{
			name:           "Unauthorized User - Read Permission",
			method:         http.MethodGet,
			path:           "/mock-struct/123",
			requestBody:    "",
			user:           bed.NoRoleUser,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to access this resource",
		},
		{
			name:           "Resource Not Found",
			method:         http.MethodGet,
			path:           "/mock-struct/99999",
			requestBody:    "",
			user:           bed.AdminUser,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Instance not found",
		},
		{
			name:        "Admin Bypass User Binding",
			method:      http.MethodGet,
			path:        "/mock-struct/123",
			requestBody: "",
			user:        bed.AdminUser,
			setup: func() string {
				instance := CreateMockResourceInstance(bed.VisitorUser.ID) // Resource owned by another user
				bed.Db.DB.Create(&instance)
				return "/mock-struct/" + instance.StringID()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Detail",
		},
		{
			name:        "User Cannot Access Resource They Did Not Create",
			method:      http.MethodGet,
			path:        "/mock-struct/123",
			requestBody: "",
			user:        CreateAllAllowedUser(),
			setup: func() string {
				instance := CreateMockResourceInstance(bed.AdminUser.ID) // Resource owned by another user
				bed.Db.DB.Create(&instance)
				return "/mock-struct/" + instance.StringID()
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Instance not found",
		},
		// 		{
		// 	name:        "Database Error",
		// 	method:      http.MethodGet,
		// 	path:        "/mock-struct/123",
		// 	requestBody: "",
		// 	user:        bed.AdminUser,
		// 	setup: func() string {
		// 		bed.Db.Close() // Simulate a database error
		// 		return "/mock-struct/123"
		// 	},
		// 	expectedStatus: http.StatusInternalServerError,
		// 	expectedBody:   "Error finding instance",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the test bed for each test case
			bed := SetupHandlerTestBed()

			path := tt.path

			// Run setup if provided
			if tt.setup != nil {
				path = tt.setup()
			}

			// Create a Gorilla Mux router
			router := mux.NewRouter()
			router.HandleFunc("/mock-struct/{id}", DefaultDetailHandler(bed.Src, bed.Db))

			// Create and execute request
			req := CreateTestRequest(t, tt.method, path, tt.requestBody, true, tt.user, bed.Logger)
			rr := httptest.NewRecorder()

			// Serve the request using the router
			router.ServeHTTP(rr, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)
		})
	}
}
