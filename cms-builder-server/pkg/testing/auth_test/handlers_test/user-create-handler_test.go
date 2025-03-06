package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	authHandlers "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/handlers"
	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	svrTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/types"
	testPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
)

func TestUserCreateHandler(t *testing.T) {
	bed := testPkg.SetupAuthTestBed()

	tests := []struct {
		name           string
		method         string
		path           string
		requestBody    string
		user           *authModels.User
		expectedStatus int
		expectedBody   string
		setup          func()
	}{
		{
			name:           "Success",
			method:         http.MethodPost,
			path:           "/user/new",
			requestBody:    `{"name": "` + testPkg.RandomString(10) + `", "email": "` + testPkg.RandomEmail() + `"}`,
			user:           bed.AdminUser,
			expectedStatus: http.StatusCreated,
			expectedBody:   "has been created",
		},
		{
			name:           "Invalid Method",
			method:         http.MethodGet,
			path:           "/user/new",
			requestBody:    `{"name": "` + testPkg.RandomString(10) + `", "email": "` + testPkg.RandomEmail() + `"}`,
			user:           bed.AdminUser,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed",
		},
		{
			name:           "Anonymous is not allowed",
			method:         http.MethodPost,
			path:           "/mock-struct/new",
			requestBody:    `{"field1": "` + testPkg.RandomString(10) + `", "field2": "` + testPkg.RandomString(10) + `"}`,
			user:           &authModels.User{},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to create this resource",
		},
		{
			name:           "Unauthorized User",
			method:         http.MethodPost,
			path:           "/user/new",
			requestBody:    `{"name": "` + testPkg.RandomString(10) + `", "email": "` + testPkg.RandomEmail() + `"}`,
			user:           bed.VisitorUser,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to create this resource",
		},
		{
			name:           "Invalid Request Body",
			method:         http.MethodPost,
			path:           "/user/new",
			requestBody:    `{"name": "` + testPkg.RandomString(10) + `", "email": "` + testPkg.RandomEmail() + `"`,
			user:           bed.AdminUser,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request body",
		},

		{
			name:           "Validation Errors",
			method:         http.MethodPost,
			path:           "/user/new",
			requestBody:    `{"name": "` + testPkg.RandomString(10) + `"}`,
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

			req := testPkg.CreateTestRequest(t, tt.method, tt.path, tt.requestBody, true, tt.user, bed.Logger)
			rr := testPkg.ExecuteHandler(t, authHandlers.UserCreateHandler(bed.Src, bed.Db), req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)
		})
	}
}

func TestUserCannotCreateRestrictedFields(t *testing.T) {
	bed := testPkg.SetupAuthTestBed()

	// Create a Gorilla Mux router
	router := mux.NewRouter()
	path := "/user/new"
	router.HandleFunc(path, authHandlers.UserCreateHandler(bed.Src, bed.Db))

	rand := testPkg.RandomUint()

	tests := []struct {
		name string
		body map[string]interface{}
	}{
		{
			name: "ID",
			body: map[string]interface{}{
				"ID":    rand,
				"name":  testPkg.RandomName(),
				"email": testPkg.RandomEmail(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			stringBody, err := json.Marshal(tt.body)
			assert.NoError(t, err)

			// Create and execute request
			req := testPkg.CreateTestRequest(t, http.MethodPost, path, string(stringBody), true, bed.AdminUser, bed.Logger)
			rr := httptest.NewRecorder()

			// Serve the request using the router
			router.ServeHTTP(rr, req)

			// Assertions
			assert.Equal(t, http.StatusCreated, rr.Code)

			// Check that the instance was not updated
			var response svrTypes.Response

			// unmarshall body[data] into createdInstance
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			assert.NoError(t, err)

			// Check that the instance was not updated
			assert.NotEqual(t, rand, response.Data.(map[string]interface{})["ID"])
		})
	}
}
