package resourcemanager_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing/resource-manager_test"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestDefaultCreateHandler(t *testing.T) {
	bed := SetupHandlerTestBed()

	tests := []struct {
		name           string
		method         string
		path           string
		requestBody    string
		user           *models.User
		expectedStatus int
		expectedBody   string
		setup          func()
	}{
		{
			name:           "Success",
			method:         http.MethodPost,
			path:           "/mock-struct/new",
			requestBody:    `{"field1": "` + RandomString(10) + `", "field2": "` + RandomEmail() + `"}`,
			user:           bed.AdminUser,
			expectedStatus: http.StatusCreated,
			expectedBody:   "has been created",
		},
		{
			name:           "Invalid Method",
			method:         http.MethodGet,
			path:           "/mock-struct/new",
			requestBody:    `{"field1": "` + RandomString(10) + `", "field2": "` + RandomString(10) + `"}`,
			user:           bed.AdminUser,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed",
		},
		{
			name:           "Unauthorized User",
			method:         http.MethodPost,
			path:           "/mock-struct/new",
			requestBody:    `{"field1": "` + RandomString(10) + `", "field2": "` + RandomString(10) + `"}`,
			user:           bed.NoRoleUser,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to create this resource",
		},
		{
			name:           "Invalid Request Body",
			method:         http.MethodPost,
			path:           "/mock-struct/new",
			requestBody:    `{"field1": "` + RandomString(10) + `", "field2": "` + RandomString(10) + `"`,
			user:           bed.AdminUser,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request body",
		},

		{
			name:           "Validation Errors",
			method:         http.MethodPost,
			path:           "/mock-struct/new",
			requestBody:    `{"field2": "` + RandomString(10) + `"}`,
			user:           bed.AdminUser,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			req := CreateTestRequest(t, tt.method, tt.path, tt.requestBody, true, tt.user, bed.Logger)
			rr := ExecuteHandler(t, DefaultCreateHandler(bed.Src, bed.Db), req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)
		})
	}
}

func TestUserCannotCreateRestrictedFields(t *testing.T) {
	bed := SetupHandlerTestBed()

	// Create a Gorilla Mux router
	router := mux.NewRouter()
	path := "/mock-struct/new"
	router.HandleFunc(path, DefaultCreateHandler(bed.Src, bed.Db))

	rand := RandomUint()

	tests := []struct {
		name string
		body map[string]interface{}
	}{
		{
			name: "UpdatedByID",
			body: map[string]interface{}{
				"UpdatedByID": rand,
				"field1":      "First Update",
			},
		},
		{
			name: "CreatedByID",
			body: map[string]interface{}{
				"CreatedByID": rand,
				"field1":      "Second Update",
			},
		},
		{
			name: "ID",
			body: map[string]interface{}{
				"ID":     rand,
				"field1": "Third Update",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			stringBody, err := json.Marshal(tt.body)
			assert.NoError(t, err)

			// Create and execute request
			req := CreateTestRequest(t, http.MethodPost, path, string(stringBody), true, bed.AdminUser, bed.Logger)
			rr := httptest.NewRecorder()

			// Serve the request using the router
			router.ServeHTTP(rr, req)

			// Assertions
			assert.Equal(t, http.StatusCreated, rr.Code)

			// Check that the instance was not updated
			var response server.Response

			// unmarshall body[data] into createdInstance
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			assert.NoError(t, err)

			// Check that the instance was not updated
			assert.NotEqual(t, rand, response.Data.(map[string]interface{})["UpdatedByID"])
			assert.NotEqual(t, rand, response.Data.(map[string]interface{})["CreatedByID"])
			assert.NotEqual(t, rand, response.Data.(map[string]interface{})["ID"])

			assert.NotEqual(t, bed.AdminUser.ID, response.Data.(map[string]interface{})["UpdatedByID"])
			assert.NotEqual(t, bed.AdminUser.ID, response.Data.(map[string]interface{})["CreatedByID"])

		})
	}
}
