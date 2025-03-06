package file_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	fileHandlers "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file/handlers"
	fileModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file/models"
	serverTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/types"
	testPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
)

func UploadMockFile(t *testing.T, testBed *testPkg.TestUtils, content string, fileName string) *fileModels.File {
	// Create a multipart form with a valid file type (PNG)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fileName)
	assert.NoError(t, err)

	// Write a small PNG file header to simulate a valid image
	_, err = io.Copy(part, bytes.NewBufferString(content))
	assert.NoError(t, err)
	writer.Close()

	// Create a test request with the POST method
	req := testPkg.CreateTestRequest(t, http.MethodPost, "/files", body.String(), true, testBed.AdminUser, testBed.Logger)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute the handler
	rr := testPkg.ExecuteHandler(t, fileHandlers.CreateStoredFilesHandler(testBed.Db, testBed.Store, "http://localhost:8080")(testBed.Src, testBed.Db), req)

	// Assertions
	assert.Equal(t, http.StatusCreated, rr.Code) // Expect 201 Created
	assert.Contains(t, rr.Body.String(), "File created")

	// Get the file record from the database
	var response serverTypes.Response
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	var file fileModels.File
	// Since the Response struct might contain a generic Data field,
	// we need to perform a two-step unmarshalling process.
	// 1. Marshal the Data field from the Response struct into a separate byte slice.
	jsonData, err := json.Marshal(response.Data)
	assert.NoError(t, err)

	// 2. Unmarshal the marshalled data (jsonData) into the provided interface (v).
	err = json.Unmarshal(jsonData, &file)
	assert.NoError(t, err)

	return &file

}

func TestDownloadStoredFileHandler_Success(t *testing.T) {
	// Setup test environment
	testBed := testPkg.SetupFileTestBed()

	// Create a test file in the store
	fileName := "testfile.png"
	fileContent := "\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01"
	file := UploadMockFile(t, &testBed, fileContent, fileName)

	// Create a test request with the GET method
	path := "/files/" + file.StringID() + "/download"
	req := testPkg.CreateTestRequest(t, http.MethodGet, path, "", true, testBed.AdminUser, testBed.Logger)

	// Execute the handler
	rr := httptest.NewRecorder()

	route := "/files/{id}/download"
	router := mux.NewRouter()
	router.HandleFunc(route, fileHandlers.DownloadStoredFileHandler(testBed.Mgr, testBed.Db, testBed.Store))
	router.ServeHTTP(rr, req)

	// Assertions
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), fileContent)
	assert.Equal(t, "image/png", rr.Header().Get("Content-Type"))
}

func TestDownloadStoredFileHandler_InvalidMethod(t *testing.T) {
	// Setup test environment
	testBed := testPkg.SetupFileTestBed()

	// Create a test request with an invalid method (POST)
	req := testPkg.CreateTestRequest(t, http.MethodPost, "/files/123", "", true, testBed.AdminUser, testBed.Logger)

	// Execute the handler
	rr := httptest.NewRecorder()
	handler := fileHandlers.DownloadStoredFileHandler(testBed.Mgr, testBed.Db, testBed.Store)
	handler.ServeHTTP(rr, req)

	// Assertions
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Contains(t, rr.Body.String(), "Method not allowed")
}

func TestDownloadStoredFileHandler_PermissionDenied(t *testing.T) {
	// Setup test environment
	testBed := testPkg.SetupFileTestBed()

	// Create a test file in the store
	fileContent := "test file content"
	filePath := "testfile.txt"
	err := os.WriteFile(filePath, []byte(fileContent), 0644)
	assert.NoError(t, err)
	defer os.Remove(filePath)

	// Create a file record in the database
	file := fileModels.File{
		Name:     "testfile.txt",
		Path:     filePath,
		MimeType: "text/plain",
	}
	err = testBed.Db.DB.Create(&file).Error
	assert.NoError(t, err)

	// Create a test request with the GET method but with a user who doesn't have permissions
	req := testPkg.CreateTestRequest(t, http.MethodGet, "/files/"+file.StringID(), "", true, testBed.NoRoleUser, testBed.Logger)

	// Execute the handler
	rr := httptest.NewRecorder()
	handler := fileHandlers.DownloadStoredFileHandler(testBed.Mgr, testBed.Db, testBed.Store)
	handler.ServeHTTP(rr, req)

	// Assertions
	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "User is not allowed to access this resource")
}

func TestDownloadStoredFileHandler_FileNotFound(t *testing.T) {
	// Setup test environment
	testBed := testPkg.SetupFileTestBed()

	// Create a test request with the GET method for a non-existent file
	req := testPkg.CreateTestRequest(t, http.MethodGet, "/files/99999/download", "", true, testBed.AdminUser, testBed.Logger)

	// Execute the handler
	rr := httptest.NewRecorder()

	route := "/files/{id}/download"
	router := mux.NewRouter()
	router.HandleFunc(route, fileHandlers.DownloadStoredFileHandler(testBed.Mgr, testBed.Db, testBed.Store))
	router.ServeHTTP(rr, req)

	// Assertions
	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "Instance not found")
}

func TestDownloadStoredFileHandler_FileOpenError(t *testing.T) {
	// Setup test environment
	testBed := testPkg.SetupFileTestBed()

	// Create a file record in the database with an invalid path
	file := fileModels.File{
		Name:     "nonexistent.txt",
		Path:     "nonexistent.txt",
		MimeType: "text/plain",
	}
	err := testBed.Db.DB.Create(&file).Error
	assert.NoError(t, err)

	// Create a test request with the GET method
	req := testPkg.CreateTestRequest(t, http.MethodGet, "/files/"+file.StringID()+"/download", "", true, testBed.AdminUser, testBed.Logger)

	// Execute the handler
	rr := httptest.NewRecorder()

	route := "/files/{id}/download"
	router := mux.NewRouter()
	router.HandleFunc(route, fileHandlers.DownloadStoredFileHandler(testBed.Mgr, testBed.Db, testBed.Store))
	router.ServeHTTP(rr, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "no such file or directory")
}
