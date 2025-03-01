package file_test

// import (
// 	"bytes"
// 	"io"
// 	"mime/multipart"
// 	"net/http"
// 	"testing"

// 	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file"
// 	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
// 	"github.com/stretchr/testify/assert"
// )

// func TestCreateStoredFilesHandler_ValidMimeType(t *testing.T) {
// 	// Setup test environment
// 	testBed := SetupFileTestBed()

// 	// Create a multipart form with a valid file type (PNG)
// 	body := &bytes.Buffer{}
// 	writer := multipart.NewWriter(body)
// 	part, err := writer.CreateFormFile("file", "testfile.png")
// 	assert.NoError(t, err)

// 	// Write a small PNG file header to simulate a valid image
// 	_, err = io.Copy(part, bytes.NewBufferString("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01"))
// 	assert.NoError(t, err)
// 	writer.Close()

// 	// Create a test request with the POST method
// 	req := CreateTestRequest(t, http.MethodPost, "/files", body.String(), true, testBed.AdminUser, testBed.Logger)
// 	req.Header.Set("Content-Type", writer.FormDataContentType())

// 	// Execute the handler
// 	rr := ExecuteHandler(t, CreateStoredFilesHandler(testBed.Db, testBed.Store)(testBed.Src, testBed.Db), req)

// 	// Assertions
// 	assert.Equal(t, http.StatusCreated, rr.Code) // Expect 201 Created
// 	assert.Contains(t, rr.Body.String(), "file created")
// }

// func TestCreateStoredFilesHandler_InvalidMimeType(t *testing.T) {
// 	// Setup test environment
// 	testBed := SetupFileTestBed()

// 	// Create a multipart form with an invalid file type (unsupported MIME type)
// 	body := &bytes.Buffer{}
// 	writer := multipart.NewWriter(body)
// 	part, err := writer.CreateFormFile("file", "testfile.unsupported")
// 	assert.NoError(t, err)

// 	// Write some arbitrary content to simulate an unsupported file
// 	_, err = io.Copy(part, bytes.NewBufferString("unsupported file content"))
// 	assert.NoError(t, err)
// 	writer.Close()

// 	// Create a test request with the POST method
// 	req := CreateTestRequest(t, http.MethodPost, "/files", body.String(), true, testBed.AdminUser, testBed.Logger)
// 	req.Header.Set("Content-Type", writer.FormDataContentType())

// 	// Execute the handler
// 	rr := ExecuteHandler(t, CreateStoredFilesHandler(testBed.Db, testBed.Store)(testBed.Src, testBed.Db), req)

// 	// Assertions
// 	assert.Equal(t, http.StatusBadRequest, rr.Code) // Expect 400 Bad Request
// 	assert.Contains(t, rr.Body.String(), "invalid content type")
// }

// func TestCreateStoredFilesHandler_EmptyFile(t *testing.T) {
// 	// Setup test environment
// 	testBed := SetupFileTestBed()

// 	// Create a multipart form with an empty file
// 	body := &bytes.Buffer{}
// 	writer := multipart.NewWriter(body)
// 	part, err := writer.CreateFormFile("file", "emptyfile.png")
// 	assert.NoError(t, err)

// 	// Write no content to simulate an empty file
// 	writer.Close()

// 	// Create a test request with the POST method
// 	req := CreateTestRequest(t, http.MethodPost, "/files", body.String(), true, testBed.AdminUser, testBed.Logger)
// 	req.Header.Set("Content-Type", writer.FormDataContentType())

// 	// Execute the handler
// 	rr := ExecuteHandler(t, CreateStoredFilesHandler(testBed.Db, testBed.Store)(testBed.Src, testBed.Db), req)

// 	// Assertions
// 	assert.Equal(t, http.StatusBadRequest, rr.Code) // Expect 400 Bad Request
// 	assert.Contains(t, rr.Body.String(), "invalid content type")
// }

// func TestCreateStoredFilesHandler_LargeFile(t *testing.T) {
// 	// Setup test environment
// 	testBed := SetupFileTestBed()

// 	// Create a multipart form with a large file
// 	body := &bytes.Buffer{}
// 	writer := multipart.NewWriter(body)
// 	part, err := writer.CreateFormFile("file", "largefile.png")
// 	assert.NoError(t, err)

// 	// Write content larger than the maximum allowed size
// 	largeContent := make([]byte, testBed.Store.GetConfig().MaxSize+1) // Exceed max size by 1 byte
// 	_, err = part.Write(largeContent)
// 	assert.NoError(t, err)
// 	writer.Close()

// 	// Create a test request with the POST method
// 	req := CreateTestRequest(t, http.MethodPost, "/files", body.String(), true, testBed.AdminUser, testBed.Logger)
// 	req.Header.Set("Content-Type", writer.FormDataContentType())

// 	// Execute the handler
// 	rr := ExecuteHandler(t, CreateStoredFilesHandler(testBed.Db, testBed.Store)(testBed.Src, testBed.Db), req)

// 	// Assertions
// 	assert.Equal(t, http.StatusBadRequest, rr.Code) // Expect 400 Bad Request
// 	assert.Contains(t, rr.Body.String(), "file size exceeds limit")
// }

// func TestCreateStoredFilesHandler_InvalidMethod(t *testing.T) {
// 	// Setup test environment
// 	testBed := SetupFileTestBed()

// 	// Create a test request with an invalid method (GET)
// 	req := CreateTestRequest(t, http.MethodGet, "/files", "", true, testBed.AdminUser, testBed.Logger)

// 	// Execute the handler
// 	rr := ExecuteHandler(t, CreateStoredFilesHandler(testBed.Db, testBed.Store)(testBed.Src, testBed.Db), req)

// 	// Assertions
// 	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
// 	assert.Contains(t, rr.Body.String(), "Method not allowed")
// }

// func TestCreateStoredFilesHandler_PermissionDenied(t *testing.T) {
// 	// Setup test environment
// 	testBed := SetupFileTestBed()

// 	// Create a multipart form with a file
// 	body := &bytes.Buffer{}
// 	writer := multipart.NewWriter(body)
// 	part, err := writer.CreateFormFile("file", "testfile.png")
// 	assert.NoError(t, err)

// 	_, err = io.Copy(part, bytes.NewBufferString("test file content"))
// 	assert.NoError(t, err)
// 	writer.Close()

// 	// Create a test request with the POST method but with a user who doesn't have permissions
// 	req := CreateTestRequest(t, http.MethodPost, "/files", body.String(), true, testBed.NoRoleUser, testBed.Logger)
// 	req.Header.Set("Content-Type", writer.FormDataContentType())

// 	// Execute the handler
// 	rr := ExecuteHandler(t, CreateStoredFilesHandler(testBed.Db, testBed.Store)(testBed.Src, testBed.Db), req)

// 	// Assertions
// 	assert.Equal(t, http.StatusForbidden, rr.Code)
// 	assert.Contains(t, rr.Body.String(), "User is not allowed to read this resource")
// }

// func TestCreateStoredFilesHandler_FileUploadError(t *testing.T) {
// 	// Setup test environment
// 	testBed := SetupFileTestBed()

// 	// Create a multipart form with an unsupported file type
// 	body := &bytes.Buffer{}
// 	writer := multipart.NewWriter(body)
// 	part, err := writer.CreateFormFile("file", "testfile.unsupported")
// 	assert.NoError(t, err)

// 	_, err = io.Copy(part, bytes.NewBufferString("test file content"))
// 	assert.NoError(t, err)
// 	writer.Close()

// 	// Create a test request with the POST method
// 	req := CreateTestRequest(t, http.MethodPost, "/files", body.String(), true, testBed.AdminUser, testBed.Logger)
// 	req.Header.Set("Content-Type", writer.FormDataContentType())

// 	// Execute the handler
// 	rr := ExecuteHandler(t, CreateStoredFilesHandler(testBed.Db, testBed.Store)(testBed.Src, testBed.Db), req)

// 	// Assertions
// 	assert.Equal(t, http.StatusBadRequest, rr.Code)
// 	assert.Contains(t, rr.Body.String(), "invalid content type")
// }
