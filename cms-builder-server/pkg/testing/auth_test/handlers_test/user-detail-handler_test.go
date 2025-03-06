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

func TestUserDetailHandler(t *testing.T) {
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
			method:         http.MethodGet,
			path:           "/user/99999",
			requestBody:    "",
			user:           bed.AdminUser,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Instance not found",
		},
		{
			name:           "Unauthorized User - Read Permission",
			method:         http.MethodGet,
			path:           "/user/123",
			requestBody:    "",
			user:           bed.NoRoleUser,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to access this resource",
		},
		{
			name:           "Anonymous is not allowed",
			method:         http.MethodGet,
			path:           "/user/123",
			requestBody:    "",
			user:           &authModels.User{},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to access this resource",
		},
		{
			name:           "Admin can view himself",
			method:         http.MethodGet,
			path:           "/user/" + bed.AdminUser.StringID(),
			requestBody:    "",
			user:           bed.AdminUser,
			expectedStatus: http.StatusOK,
			expectedBody:   `"ID":` + bed.AdminUser.StringID(),
		},
		{
			name:        "Admin can view others",
			method:      http.MethodGet,
			path:        "",
			requestBody: "",
			user:        bed.AdminUser,
			setup: func() string {
				instance := testPkg.CreateNoRoleUser()
				bed.Db.DB.Create(&instance)
				return "/user/" + instance.StringID()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"ID":`,
		},
		{
			name:           "Visitor can view himself",
			method:         http.MethodGet,
			path:           "/user/" + bed.VisitorUser.StringID(),
			requestBody:    "",
			user:           bed.VisitorUser,
			expectedStatus: http.StatusOK,
			expectedBody:   `"ID":` + bed.VisitorUser.StringID(),
		},
		{
			name:        "Visitor can not view others",
			method:      http.MethodGet,
			path:        "",
			requestBody: "",
			user:        bed.VisitorUser,
			setup: func() string {
				instance := testPkg.CreateNoRoleUser()
				bed.Db.DB.Create(&instance)
				return "/user/" + instance.StringID()
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Instance not found",
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
			router.HandleFunc("/user/{id}", authHandlers.UserDetailHandler(bed.Src, bed.Db))

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
