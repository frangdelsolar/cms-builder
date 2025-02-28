package resourcemanager_test

import (
	"net/http"
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	"github.com/stretchr/testify/assert"
)

func TestDefaultCreateHandler_Success(t *testing.T) {
	mockDb, mockUser, mockLogger, mockStruct := setupTest(t)

	// Create request body
	requestBody := `{"field1": "John Doe", "field2": "john.doe@example.com"}`
	req := createTestRequest(t, http.MethodPost, "/mock-struct/new", requestBody, mockUser, mockLogger)

	// Execute handler
	rr := executeHandler(t, DefaultCreateHandler(mockStruct, mockDb), req)

	// Assertions
	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Contains(t, rr.Body.String(), "has been created")
}

func TestDefaultCreateHandler_InvalidMethod(t *testing.T) {
	mockDb, mockUser, mockLogger, mockStruct := setupTest(t)

	// Create request body
	requestBody := `{"field1": "John Doe", "field2": "john.doe@example.com"}`
	req := createTestRequest(t, http.MethodGet, "/mock-struct/new", requestBody, mockUser, mockLogger)

	// Execute handler
	rr := executeHandler(t, DefaultCreateHandler(mockStruct, mockDb), req)

	// Assertions
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Contains(t, rr.Body.String(), "Method not allowed")
}

func TestDefaultCreateHandler_UnauthorizedUser(t *testing.T) {
	mockDb, _, mockLogger, mockStruct := setupTest(t)

	// Create a user with insufficient permissions
	unauthorizedUser := &models.User{
		ID:    uint(999),
		Name:  "Test User",
		Email: "YHs7r@example.com",
		Roles: "visitor", // User with insufficient permissions
	}

	// Create request body
	requestBody := `{"field1": "John Doe", "field2": "john.doe@example.com"}`
	req := createTestRequest(t, http.MethodPost, "/mock-struct/new", requestBody, unauthorizedUser, mockLogger)

	// Execute handler
	rr := executeHandler(t, DefaultCreateHandler(mockStruct, mockDb), req)

	// Assertions
	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "User is not allowed to create this resource")
}

func TestDefaultCreateHandler_InvalidRequestBody(t *testing.T) {
	mockDb, mockUser, mockLogger, mockStruct := setupTest(t)

	// Create request body with invalid JSON
	requestBody := `{"field1": "John Doe", "field2": "john.doe@example.com"`
	req := createTestRequest(t, http.MethodPost, "/mock-struct/new", requestBody, mockUser, mockLogger)

	// Execute handler
	rr := executeHandler(t, DefaultCreateHandler(mockStruct, mockDb), req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid request body")
}

func TestDefaultCreateHandler_ValidationErrors(t *testing.T) {
	mockDb, mockUser, mockLogger, mockStruct := setupTest(t)

	// Create request body with missing required field
	requestBody := `{"field2": "john.doe@example.com"}`
	req := createTestRequest(t, http.MethodPost, "/mock-struct/new", requestBody, mockUser, mockLogger)

	// Execute handler
	rr := executeHandler(t, DefaultCreateHandler(mockStruct, mockDb), req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Validation failed")
}

func TestDefaultCreateHandler_DatabaseError(t *testing.T) {
	mockDb, mockUser, mockLogger, mockStruct := setupTest(t)

	// Intentionally not running AutoMigrate to cause a database error
	mockDb.Close()

	// Create request body
	requestBody := `{"field1": "John Doe", "field2": "john.doe@example.com"}`
	req := createTestRequest(t, http.MethodPost, "/mock-struct/new", requestBody, mockUser, mockLogger)

	// Execute handler
	rr := executeHandler(t, DefaultCreateHandler(mockStruct, mockDb), req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Error creating resource")
}
