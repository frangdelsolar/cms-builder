package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	authHandlers "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/handlers"
	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	testPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
)

func TestUserDeleteHandler(t *testing.T) {
	bed := testPkg.SetupAuthTestBed()

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
			name:           "Invalid Method",
			method:         http.MethodPost,
			path:           "/user/123",
			requestBody:    "",
			user:           bed.AdminUser,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed",
		},
		{
			name:           "Resource Not Found",
			method:         http.MethodDelete,
			path:           "/user/12345234523",
			requestBody:    "",
			user:           bed.AdminUser,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Instance not found",
		},
		{
			name:           "Unauthorized User - Read Permission",
			method:         http.MethodDelete,
			path:           "/user/123",
			requestBody:    "",
			user:           bed.NoRoleUser,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to access this resource",
		},
		{
			name:           "Unauthorized User - Delete Permission",
			method:         http.MethodDelete,
			path:           "/user/123",
			requestBody:    "",
			user:           bed.VisitorUser,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to delete this resource",
		},
		{
			name:        "Anonymous is not allowed",
			method:      http.MethodDelete,
			path:        "",
			requestBody: "",
			user:        &authModels.User{},
			setup: func() string {
				instance := testPkg.CreateNoRoleUser()
				bed.Db.DB.Create(&instance)
				return "/user/" + instance.StringID()
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to access this resource",
		},
		{
			name:           "Admin can delete himself",
			method:         http.MethodDelete,
			path:           "/user/" + bed.AdminUser.StringID(),
			requestBody:    "",
			user:           bed.AdminUser,
			expectedStatus: http.StatusOK,
			expectedBody:   "has been deleted",
		},
		{
			name:        "Admin can delete others",
			method:      http.MethodDelete,
			path:        "",
			requestBody: "",
			user:        bed.AdminUser,
			setup: func() string {
				instance := testPkg.CreateNoRoleUser()
				bed.Db.DB.Create(&instance)
				return "/user/" + instance.StringID()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "has been deleted",
		},
		{
			name:           "Visitor can not delete himself",
			method:         http.MethodDelete,
			path:           "/user/" + bed.VisitorUser.StringID(),
			requestBody:    "",
			user:           bed.VisitorUser,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to delete this resource",
		},
		{
			name:        "Visitor can not delete others",
			method:      http.MethodDelete,
			path:        "",
			requestBody: "",
			user:        bed.VisitorUser,
			setup: func() string {
				instance := testPkg.CreateNoRoleUser()
				bed.Db.DB.Create(&instance)
				return "/user/" + instance.StringID()
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to delete this resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the test bed for each test case
			bed := testPkg.SetupAuthTestBed()

			path := tt.path

			// Run setup if provided
			if tt.setup != nil {
				path = tt.setup()
			}

			// Create a Gorilla Mux router
			router := mux.NewRouter()
			router.HandleFunc("/user/{id}", authHandlers.UserDeleteHandler(bed.Src, bed.Db))

			// Create and execute request
			req := testPkg.CreateTestRequest(t, tt.method, path, tt.requestBody, true, tt.user, bed.Logger)
			rr := httptest.NewRecorder()

			// Serve the request using the router
			router.ServeHTTP(rr, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)
		})
	}
}
