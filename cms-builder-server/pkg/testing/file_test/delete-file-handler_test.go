package file_test

// TODO: Complete

import (
	"net/http"
	"net/http/httptest"
	"testing"

	f "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestDeleteStoredFilesHandler(t *testing.T) {
	bed := SetupFileTestBed()

	// Create a Gorilla Mux router
	router := mux.NewRouter()
	path := "/files/{id}"
	router.HandleFunc(path, f.DeleteStoredFilesHandler(bed.Db, bed.Store)(bed.Src, bed.Db))

	tests := []struct {
		name           string
		method         string
		path           string
		fileID         string
		user           *models.User
		expectedStatus int
		expectedBody   string
		setup          func() *models.File
	}{
		{
			name:           "Success",
			method:         http.MethodDelete,
			path:           "/files/1",
			fileID:         "1",
			user:           bed.AdminUser,
			expectedStatus: http.StatusOK,
			expectedBody:   "File deleted",
			setup: func() *models.File {
				// Create a file in the database
				file := &models.File{
					SystemData: &models.SystemData{
						CreatedByID: bed.AdminUser.ID,
					},
				}
				bed.Db.DB.Create(file)
				return file
			},
		},
		{
			name:           "Invalid Method",
			method:         http.MethodPost,
			path:           "/files/1",
			fileID:         "1",
			user:           bed.AdminUser,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed",
			setup: func() *models.File {
				// Create a file in the database
				file := &models.File{
					SystemData: &models.SystemData{
						CreatedByID: bed.AdminUser.ID,
					},
				}
				bed.Db.DB.Create(file)
				return file
			},
		},
		{
			name:           "Permission Denied",
			method:         http.MethodDelete,
			path:           "/files/1",
			fileID:         "1",
			user:           bed.NoRoleUser,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "User is not allowed to access this resource",
			setup: func() *models.File {
				// Create a file in the database
				file := &models.File{
					SystemData: &models.SystemData{
						CreatedByID: bed.AdminUser.ID,
					},
				}
				bed.Db.DB.Create(file)
				return file
			},
		},
		{
			name:           "File Not Found",
			method:         http.MethodDelete,
			path:           "/files/999",
			fileID:         "999",
			user:           bed.AdminUser,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Instance not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var path string = tt.path
			var file *models.File
			if tt.setup != nil {
				file = tt.setup()
				path = "/files/" + file.StringID()
			}

			req := CreateTestRequest(t, tt.method, path, "", true, tt.user, bed.Logger)
			req = mux.SetURLVars(req, map[string]string{"id": tt.fileID})

			rr := httptest.NewRecorder()

			// Serve the request using the router
			router.ServeHTTP(rr, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)

			// If the file was deleted, verify it no longer exists in the database
			if tt.expectedStatus == http.StatusOK && file != nil {
				var deletedFile models.File
				err := bed.Db.DB.First(&deletedFile, file.ID).Error
				assert.Error(t, err) // Expect an error because the file should not exist
			}
		})
	}
}
