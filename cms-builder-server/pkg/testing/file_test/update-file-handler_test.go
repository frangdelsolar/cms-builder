package file_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
	"github.com/stretchr/testify/assert"
)

func TestUpdateStoredFilesHandler_InvalidMethod(t *testing.T) {
	// Setup test environment
	testBed := SetupFileTestBed()

	// Create a test request with an invalid method (POST)
	req := CreateTestRequest(t, http.MethodPost, "/files/123", "", true, testBed.AdminUser, testBed.Logger)

	// Execute the handler
	rr := httptest.NewRecorder()
	handler := UpdateStoredFilesHandler(testBed.Src, testBed.Db)
	handler.ServeHTTP(rr, req)

	// Assertions
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Contains(t, rr.Body.String(), "Method not allowed")
}

func TestUpdateStoredFilesHandler_ValidMethodButNotAllowed(t *testing.T) {
	// Setup test environment
	testBed := SetupFileTestBed()

	// Create a test request with the PUT method
	req := CreateTestRequest(t, http.MethodPut, "/files/123", "", true, testBed.AdminUser, testBed.Logger)

	// Execute the handler
	rr := httptest.NewRecorder()
	handler := UpdateStoredFilesHandler(testBed.Src, testBed.Db)
	handler.ServeHTTP(rr, req)

	// Assertions
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Contains(t, rr.Body.String(), "You cannot update a file. You may delete and create a new one.")
}

func TestUpdateStoredFilesHandler_OtherMethods(t *testing.T) {
	// Setup test environment
	testBed := SetupFileTestBed()

	// Define test cases for other HTTP methods
	methods := []string{http.MethodGet, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			// Create a test request with the current method
			req := CreateTestRequest(t, method, "/files/123", "", true, testBed.AdminUser, testBed.Logger)

			// Execute the handler
			rr := httptest.NewRecorder()
			handler := UpdateStoredFilesHandler(testBed.Src, testBed.Db)
			handler.ServeHTTP(rr, req)

			// Assertions
			assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
			assert.Contains(t, rr.Body.String(), "Method not allowed")
		})
	}
}
