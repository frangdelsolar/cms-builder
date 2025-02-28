package resourcemanager_test

import (
	"net/http"
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	tu "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

func TestDefaultDetailHandler_Success(t *testing.T) {
	mockDb, mockUser, mockLogger, mockStruct := setupTest(t)

	// Create a resource to retrieve
	instance := tu.MockStruct{
		SystemData: models.SystemData{
			CreatedByID: mockUser.ID,
			UpdatedByID: mockUser.ID,
		},
		Field1: "John Doe",
		Field2: "john.doe@example.com",
	}
	err := mockDb.DB.Create(&instance).Error
	assert.NoError(t, err)

	// Create and execute request
	req := createTestRequest(t, http.MethodGet, "/mock-struct/"+instance.StringID(), "", mockUser, mockLogger)
	rr := executeHandler(t, DefaultDetailHandler(mockStruct, mockDb), req)

	// Assertions
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Detail")
	assert.Contains(t, rr.Body.String(), "John Doe")
}

func TestDefaultDetailHandler_InvalidMethod(t *testing.T) {
	mockDb, mockUser, mockLogger, mockStruct := setupTest(t)

	// Create and execute request with invalid method
	req := createTestRequest(t, http.MethodPost, "/mock-struct/123", "", mockUser, mockLogger)
	rr := executeHandler(t, DefaultDetailHandler(mockStruct, mockDb), req)

	// Assertions
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Contains(t, rr.Body.String(), "Method not allowed")
}

func TestDefaultDetailHandler_UnauthorizedUser(t *testing.T) {
	mockDb, _, mockLogger, mockStruct := setupTest(t)

	// Create a user with insufficient permissions
	unauthorizedUser := &models.User{
		ID:    uint(999),
		Name:  "Test User",
		Email: "YHs7r@example.com",
		Roles: "invalid", // User with insufficient permissions
	}

	// Create and execute request
	req := createTestRequest(t, http.MethodGet, "/mock-struct/123", "", unauthorizedUser, mockLogger)
	rr := executeHandler(t, DefaultDetailHandler(mockStruct, mockDb), req)

	// Assertions
	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "User is not allowed to access this resource")
}

func TestDefaultDetailHandler_ResourceNotFound(t *testing.T) {
	mockDb, mockUser, mockLogger, mockStruct := setupTest(t)

	// Create and execute request for a non-existent resource
	req := createTestRequest(t, http.MethodGet, "/mock-struct/99999", "", mockUser, mockLogger)
	rr := executeHandler(t, DefaultDetailHandler(mockStruct, mockDb), req)

	// Assertions
	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "Instance not found")
}

func TestDefaultDetailHandler_DatabaseError(t *testing.T) {
	mockDb, mockUser, mockLogger, mockStruct := setupTest(t)

	// Intentionally break the database connection
	mockDb.Close()

	// Create and execute request
	req := createTestRequest(t, http.MethodGet, "/mock-struct/123", "", mockUser, mockLogger)
	rr := executeHandler(t, DefaultDetailHandler(mockStruct, mockDb), req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Error finding instance")
}

func TestDefaultDetailHandler_AdminBypassUserBinding(t *testing.T) {
	mockDb, _, mockLogger, mockStruct := setupTest(t)

	// Create an admin user
	adminUser := &models.User{
		ID:    uint(999),
		Name:  "Admin User",
		Email: "admin@example.com",
		Roles: models.AdminRole.S(),
	}

	// Create a resource owned by another user
	instance := tu.MockStruct{
		SystemData: models.SystemData{
			CreatedByID: uint(77), // Different from adminUser
			UpdatedByID: uint(77),
		},
		Field1: "John Doe",
		Field2: "john.doe@example.com",
	}
	err := mockDb.DB.Create(&instance).Error
	assert.NoError(t, err)

	// Create and execute request
	req := createTestRequest(t, http.MethodGet, "/mock-struct/"+instance.StringID(), "", adminUser, mockLogger)
	rr := executeHandler(t, DefaultDetailHandler(mockStruct, mockDb), req)

	// Assertions
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Detail")
	assert.Contains(t, rr.Body.String(), "John Doe")
}

func TestDefaultDetailHandler_UserCannotAccessResourceTheyDidNotCreate(t *testing.T) {
	mockDb, _, mockLogger, mockStruct := setupTest(t)

	// Create a resource owned by another user
	otherUser := &models.User{
		ID:    uint(1000),
		Name:  "Other User",
		Email: "other@example.com",
		Roles: "editor",
	}
	instance := tu.MockStruct{
		SystemData: models.SystemData{
			CreatedByID: otherUser.ID,
			UpdatedByID: otherUser.ID,
		},
		Field1: "John Doe",
		Field2: "john.doe@example.com",
	}
	err := mockDb.DB.Create(&instance).Error
	assert.NoError(t, err)

	// Create a request from a different user
	currentUser := &models.User{
		ID:    uint(999),
		Name:  "Test User",
		Email: "test@example.com",
		Roles: "editor", // Same role but not the creator
	}

	// Create and execute request
	req := createTestRequest(t, http.MethodGet, "/mock-struct/"+instance.StringID(), "", currentUser, mockLogger)
	rr := executeHandler(t, DefaultDetailHandler(mockStruct, mockDb), req)

	// Assertions
	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "User is not allowed to access this resource")
}
