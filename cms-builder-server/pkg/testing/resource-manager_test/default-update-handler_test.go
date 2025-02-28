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

func TestDefaultUpdateHandler(t *testing.T) {
	bed := SetupHandlerTestBed()

	tests := []struct {
		name           string
		method         string
		path           string
		requestBody    string
		user           *models.User
		setup          func() *MockStruct // Optional setup function for specific test cases
		expectedStatus int
		expectedBody   string
		overrideBody   bool
	}{
		{
			name:        "Success",
			method:      http.MethodPut,
			path:        "",
			requestBody: `{"field1": "Updated Name", "field2": "updated@example.com"}`,
			user:        bed.AdminUser,
			setup: func() *MockStruct {
				instance := CreateMockResourceInstance(bed.AdminUser.ID)
				bed.Db.DB.Create(&instance)
				return instance
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "has been updated",
		},
		{
			name:           "Invalid Method",
			method:         http.MethodPost,
			path:           "/mock-struct/123",
			requestBody:    `{"field1": "Updated Name", "field2": "updated@example.com"}`,
			user:           bed.AdminUser,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed",
		},
		{
			name:           "Unauthorized User - Read Permission",
			method:         http.MethodPut,
			path:           "/mock-struct/123",
			requestBody:    `{"field1": "Updated Name", "field2": "updated@example.com"}`,
			user:           bed.NoRoleUser,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to access this resource",
		},
		{
			name:           "Unauthorized User - Update Permission",
			method:         http.MethodPut,
			path:           "/mock-struct/123",
			requestBody:    `{"field1": "Updated Name", "field2": "updated@example.com"}`,
			user:           bed.VisitorUser,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to delete this resource",
		},
		{
			name:           "Resource Not Found",
			method:         http.MethodPut,
			path:           "/mock-struct/99999",
			requestBody:    `{"field1": "Updated Name", "field2": "updated@example.com"}`,
			user:           bed.AdminUser,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Instance not found",
		},

		{
			name:           "Invalid Request Body",
			method:         http.MethodPut,
			path:           "",
			requestBody:    `{"field1": "Updated Name", "field2": "updated@example.com"`, // Malformed JSON
			user:           bed.AdminUser,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request body",
			setup: func() *MockStruct {
				instance := CreateMockResourceInstance(bed.AdminUser.ID)
				bed.Db.DB.Create(&instance)
				return instance
			},
		},

		{
			name:           "Validation Errors",
			method:         http.MethodPut,
			path:           "",
			requestBody:    `{"field1": ""}`, // Missing required field "field1"
			user:           bed.AdminUser,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Validation failed",
			setup: func() *MockStruct {
				instance := CreateMockResourceInstance(bed.AdminUser.ID)
				bed.Db.DB.Create(&instance)
				return instance
			},
		},

		{
			name:        "Admin Bypass User Binding",
			method:      http.MethodPut,
			path:        "/mock-struct/123",
			requestBody: `{"field1": "Updated Name", "field2": "updated@example.com"}`,
			user:        bed.AdminUser,
			setup: func() *MockStruct {
				instance := CreateMockResourceInstance(bed.VisitorUser.ID)
				bed.Db.DB.Create(&instance)
				return instance
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "has been updated",
		},
		{
			name:        "User Cannot Update Resource They Did Not Create",
			method:      http.MethodPut,
			path:        "/mock-struct/123",
			requestBody: `{"field1": "Updated Name", "field2": "updated@example.com"}`,
			user:        bed.VisitorUser,
			setup: func() *MockStruct {
				instance := CreateMockResourceInstance(bed.AdminUser.ID)
				bed.Db.DB.Create(&instance)
				return instance
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to delete this resource",
		},
		{
			name:        "No Changes",
			method:      http.MethodPut,
			path:        "/mock-struct/123",
			requestBody: "", // No changes
			user:        bed.AdminUser,
			setup: func() *MockStruct {
				instance := CreateMockResourceInstance(bed.AdminUser.ID)
				bed.Db.DB.Create(&instance)
				return instance
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "is up to date",
			overrideBody:   true,
		},
		// TODO
		// {
		// 	name:           "User can not set updated_by",
		// 	method:         http.MethodPost,
		// 	path:           "/mock-struct/new",
		// 	requestBody:    `{"CreatedByID": 1, "field1": "` + RandomString(10) + `", "field2": "` + RandomString(10) + `"`,
		// 	user:           bed.AdminUser,
		// 	expectedStatus: http.StatusBadRequest,
		// 	expectedBody:   "Invalid request body",
		// },
		// {
		// 	name:           "User can not set created_by",
		// 	method:         http.MethodPost,
		// 	path:           "/mock-struct/new",
		// 	requestBody:    `{"CreatedByID": 1, "field1": "` + RandomString(10) + `", "field2": "` + RandomString(10) + `"`,
		// 	user:           bed.AdminUser,
		// 	expectedStatus: http.StatusBadRequest,
		// 	expectedBody:   "Invalid request body",
		// },
		// {
		// 	name:           "Invalid Request Body - Extra Field",
		// 	method:         http.MethodPut,
		// 	path:           "/mock-struct/123",
		// 	requestBody:    `{"field1": "Updated Name", "field2": "updated@example.com", "extra": "field"}`, // Unexpected field
		// 	user:           bed.AdminUser,
		// 	expectedStatus: http.StatusBadRequest,
		// 	expectedBody:   "Invalid request body",
		// setup: func() *MockStruct {
		// 	instance := CreateMockResourceInstance(bed.AdminUser.ID)
		// 	bed.Db.DB.Create(&instance)
		// 	return instance
		// },
		// {
		// 	name:        "Database Error",
		// 	method:      http.MethodPut,
		// 	path:        "/mock-struct/123",
		// 	requestBody: `{"field1": "Updated Name", "field2": "updated@example.com"}`,
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
			body := tt.requestBody

			var instance *MockStruct

			// Run setup if provided
			if tt.setup != nil {

				instance = tt.setup()
				path = "/mock-struct/" + instance.StringID()
				if tt.overrideBody {
					body = `{"field1": "` + instance.Field1 + `", "field2": "` + instance.Field2 + `"}`
				}
			}

			// Create a Gorilla Mux router
			router := mux.NewRouter()
			router.HandleFunc("/mock-struct/{id}", DefaultUpdateHandler(bed.Src, bed.Db))

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
