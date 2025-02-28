package resourcemanager_test

import (
	"net/http"
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing/resource-manager_test"
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
		// TODO
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
		// 	method:         http.MethodPost,
		// 	path:           "/mock-struct/new",
		// 	requestBody:    `{"field1": "` + RandomString(10) + `", "field2": "` + RandomString(10) + `", "extra": "` + RandomString(10) + `"}`,
		// 	user:           bed.AdminUser,
		// 	expectedStatus: http.StatusBadRequest,
		// 	expectedBody:   "Invalid request body",
		// },
		// {
		// 	name:           "Database Error",
		// 	method:         http.MethodPost,
		// 	path:           "/mock-struct/new",
		// 	requestBody:    `{"field1": "` + RandomString(10) + `", "field2": "` + RandomString(10) + `"}`,
		// 	user:           bed.AdminUser,
		// 	expectedStatus: http.StatusInternalServerError,
		// 	expectedBody:   "Error creating resource",
		// 	setup:          func() { bed.Db.Close() },
		// },
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
