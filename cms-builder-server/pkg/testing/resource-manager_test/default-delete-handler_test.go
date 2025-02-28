package resourcemanager_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing/resource-manager_test"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestDefaultDeleteHandler(t *testing.T) {
	bed := SetupHandlerTestBed()

	tests := []struct {
		name           string
		method         string
		path           string
		requestBody    string
		user           *models.User
		setup          func() string // Optional setup function for specific test cases
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success",
			method:      http.MethodDelete,
			path:        "",
			requestBody: "",
			user:        bed.AdminUser,
			setup: func() string {
				instance := CreateMockResourceInstance(bed.AdminUser.ID)
				bed.Db.DB.Create(&instance)
				return "/mock-struct/" + instance.StringID()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "has been deleted",
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
			name:           "Unauthorized User - Read Permission",
			method:         http.MethodDelete,
			path:           "/mock-struct/123",
			requestBody:    "",
			user:           bed.NoRoleUser,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to access this resource",
		},
		{
			name:           "Unauthorized User - Delete Permission",
			method:         http.MethodDelete,
			path:           "/mock-struct/123",
			requestBody:    "",
			user:           bed.VisitorUser,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to delete this resource",
		},
		{
			name:           "Resource Not Found",
			method:         http.MethodDelete,
			path:           "/mock-struct/99999",
			requestBody:    "",
			user:           bed.AdminUser,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Instance not found",
		},

		{
			name:        "Admin Bypass User Binding",
			method:      http.MethodDelete,
			path:        "",
			requestBody: "",
			user:        bed.AdminUser,
			setup: func() string {
				instance := CreateMockResourceInstance(bed.VisitorUser.ID)
				bed.Db.DB.Create(&instance)
				return "/mock-struct/" + instance.StringID()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "has been deleted",
		},

		{
			name:        "User Cannot Delete Resource They Did Not Create",
			method:      http.MethodDelete,
			path:        "",
			requestBody: "",
			user:        bed.VisitorUser,
			setup: func() string {
				instance := CreateMockResourceInstance(bed.AdminUser.ID)
				bed.Db.DB.Create(&instance)
				return "/mock-struct/" + instance.StringID()
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to delete this resource",
		},
		{
			name:        "Database Error",
			method:      http.MethodDelete,
			path:        "",
			requestBody: "",
			user:        bed.AdminUser,
			setup: func() string {
				instance := CreateMockResourceInstance(bed.AdminUser.ID)
				bed.Db.DB.Create(&instance)
				bed.Db = nil
				return "/mock-struct/" + instance.StringID()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Error finding resource",
		},
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
			router.HandleFunc("/mock-struct/{id}", DefaultDeleteHandler(bed.Src, bed.Db))

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
