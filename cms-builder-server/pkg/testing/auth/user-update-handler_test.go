package auth_test

import (
	"encoding/json"
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

func TestUserUpdateHandler(t *testing.T) {
	bed := SetupHandlerTestBed()

	tests := []struct {
		name           string
		method         string
		path           string
		requestBody    string
		user           *models.User
		setup          func() *models.User // Optional setup function for specific test cases
		expectedStatus int
		expectedBody   string
		overrideBody   bool
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
			method:         http.MethodPut,
			path:           "/user/12345234523",
			requestBody:    "",
			user:           bed.AdminUser,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Instance not found",
		},
		{
			name:           "Unauthorized User - Read Permission",
			method:         http.MethodPut,
			path:           "/user/123",
			requestBody:    "",
			user:           bed.NoRoleUser,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to access this resource",
		},
		{
			name:           "Unauthorized User - Update Permission",
			method:         http.MethodPut,
			path:           "/user/123",
			requestBody:    "",
			user:           bed.VisitorUser,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to update this resource",
		},
		{
			name:        "Anonymous is not allowed",
			method:      http.MethodPut,
			path:        "",
			requestBody: "",
			user:        &models.User{},
			setup: func() *models.User {
				instance := CreateNoRoleUser()
				bed.Db.DB.Create(&instance)

				t.Log("instance", instance.ID)

				return instance
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to access this resource",
		},
		{
			name:           "Admin can update himself",
			method:         http.MethodPut,
			path:           "/user/" + bed.AdminUser.StringID(),
			requestBody:    `{"name": "` + RandomString(10) + `", "email": "` + RandomEmail() + `"}`,
			user:           bed.AdminUser,
			expectedStatus: http.StatusOK,
			expectedBody:   "has been updated",
		},
		{
			name:        "Admin can update others",
			method:      http.MethodPut,
			path:        "",
			requestBody: `{"name": "` + RandomString(10) + `", "email": "` + RandomEmail() + `"}`,
			user:        bed.AdminUser,
			setup: func() *models.User {
				instance := CreateNoRoleUser()
				bed.Db.DB.Create(&instance)
				return instance
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "has been updated",
		},
		{
			name:           "Visitor can not update himself",
			method:         http.MethodPut,
			path:           "/user/" + bed.VisitorUser.StringID(),
			requestBody:    `{"name": "` + RandomString(10) + `", "email": "` + RandomEmail() + `"}`,
			user:           bed.VisitorUser,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to update this resource",
		},
		{
			name:        "Visitor can not update others",
			method:      http.MethodPut,
			path:        "",
			requestBody: `{"name": "` + RandomString(10) + `", "email": "` + RandomEmail() + `"}`,
			user:        bed.VisitorUser,
			setup: func() *models.User {
				instance := CreateNoRoleUser()
				bed.Db.DB.Create(&instance)
				return instance
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to update this resource",
		},

		{
			name:           "Invalid Request Body",
			method:         http.MethodPut,
			path:           "",
			requestBody:    `{"name": "Updated Name", "email": "updated@example.com"`, // Malformed JSON
			user:           bed.AdminUser,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request body",
			setup: func() *models.User {
				instance := CreateNoRoleUser()
				bed.Db.DB.Create(&instance)
				return instance
			},
		},
		{
			name:           "Validation Errors",
			method:         http.MethodPut,
			path:           "",
			requestBody:    `{"name": ""}`, // Missing required field "name"
			user:           bed.AdminUser,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Validation failed",
			setup: func() *models.User {
				instance := CreateNoRoleUser()
				bed.Db.DB.Create(&instance)
				return instance
			},
		},

		{
			name:        "No Changes",
			method:      http.MethodPut,
			path:        "",
			requestBody: "", // No changes
			user:        bed.AdminUser,
			setup: func() *models.User {
				instance := CreateNoRoleUser()
				bed.Db.DB.Create(&instance)
				return instance
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "is up to date",
			overrideBody:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the test bed for each test case
			bed := SetupAuthTestBed()

			path := tt.path
			body := tt.requestBody

			var instance *models.User

			// Run setup if provided
			if tt.setup != nil {
				instance = tt.setup()
				path = "/user/" + instance.StringID()
				if tt.overrideBody {
					body = `{"name": "` + instance.Name + `", "email": "` + instance.Email + `"}`
				}

				t.Log("Instance:", instance)
				t.Log("Path:", path)
				t.Log("Body:", body)
			}

			// Create a Gorilla Mux router
			router := mux.NewRouter()
			router.HandleFunc("/user/{id}", DefaultUpdateHandler(bed.Src, bed.Db))

			// Create and execute request
			req := CreateTestRequest(t, tt.method, path, body, true, tt.user, bed.Logger)
			rr := httptest.NewRecorder()

			// Serve the request using the router
			router.ServeHTTP(rr, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)
		})
	}
}

func TestUserCannotUpdateRestrictedFields(t *testing.T) {
	bed := SetupAuthTestBed()

	// Create a mock instance
	instance := CreateNoRoleUser()
	bed.Db.DB.Create(&instance)

	// Create a Gorilla Mux router
	router := mux.NewRouter()
	router.HandleFunc("/user/{id}", DefaultUpdateHandler(bed.Src, bed.Db))

	path := "/user/" + instance.StringID()
	tests := []struct {
		name string
		body map[string]interface{}
	}{
		{
			name: "ID",
			body: map[string]interface{}{
				"ID":    uint(1),
				"name":  RandomName(),
				"email": RandomEmail(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			t.Log(tt.body)

			stringBody, err := json.Marshal(tt.body)
			assert.NoError(t, err)

			// Create and execute request
			req := CreateTestRequest(t, http.MethodPut, path, string(stringBody), true, bed.AdminUser, bed.Logger)
			rr := httptest.NewRecorder()

			// Serve the request using the router
			router.ServeHTTP(rr, req)

			t.Log(rr.Body)

			// Assertions
			assert.Equal(t, http.StatusOK, rr.Code)

			// Check that the instance was not updated
			var updatedInstance models.User
			bed.Db.DB.First(&updatedInstance, instance.ID)
			assert.Equal(t, instance.ID, updatedInstance.ID)
			assert.NotEqual(t, instance.Name, updatedInstance.Name)
		})
	}
}
