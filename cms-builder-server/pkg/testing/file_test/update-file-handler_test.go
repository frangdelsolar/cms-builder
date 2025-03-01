package file_test

import (
	"net/http"
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
	rr := ExecuteHandler(t, UpdateStoredFilesHandler(testBed.Src, testBed.Db), req)

	// Assertions
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Contains(t, rr.Body.String(), "Method not allowed")
}

func TestUpdateStoredFilesHandler_ValidMethod(t *testing.T) {
	// Setup test environment
	testBed := SetupFileTestBed()

	// Create a test request with the PUT method
	req := CreateTestRequest(t, http.MethodPut, "/files/123", "", true, testBed.AdminUser, testBed.Logger)

	// Execute the handler
	rr := ExecuteHandler(t, UpdateStoredFilesHandler(testBed.Src, testBed.Db), req)

	// Assertions
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Contains(t, rr.Body.String(), "You cannot update a file")
}
