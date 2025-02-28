package resourcemanager_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	tu "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

// setupTest initializes common test dependencies (DB, logger, resource, etc.)
func setupTest(t *testing.T) (*database.Database, *models.User, *logger.Logger, *Resource) {
	mockDb := tu.GetTestDB()
	mockUser := tu.GetTestUser()
	mockLogger := tu.GetTestLogger()
	mockStruct := tu.GetMockResource()

	// Auto-migrate the mock struct
	err := mockDb.DB.AutoMigrate(tu.MockStruct{})
	assert.NoError(t, err)

	return mockDb, mockUser, mockLogger, mockStruct
}

// createTestResource creates a mock resource in the database for testing
func createTestResource(t *testing.T, db *database.Database, createdByID uint) *tu.MockStruct {
	instance := tu.MockStruct{
		SystemData: models.SystemData{
			CreatedByID: createdByID,
			UpdatedByID: createdByID,
		},
		Field1: "John Doe",
		Field2: "john.doe@example.com",
	}
	err := db.DB.Create(&instance).Error
	assert.NoError(t, err)
	return &instance
}

func createTestRequest(t *testing.T, method, path, body string, user *models.User, log *logger.Logger) *http.Request {
	req, err := http.NewRequest(method, path, bytes.NewBufferString(body))
	assert.NoError(t, err)

	ctx := req.Context()
	ctx = context.WithValue(ctx, server.CtxRequestIsAuth, true)
	ctx = context.WithValue(ctx, server.CtxRequestUser, user)
	ctx = context.WithValue(ctx, server.CtxRequestLogger, log)

	return req.WithContext(ctx)
}

// executeHandler executes the handler and returns the response recorder
func executeHandler(t *testing.T, handler http.HandlerFunc, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func TestDefaultDeleteHandler_Success(t *testing.T) {
	mockDb, mockUser, mockLogger, mockStruct := setupTest(t)

	// Create a resource to delete
	instance := createTestResource(t, mockDb, mockUser.ID)

	// Create and execute request
	req := createTestRequest(t, http.MethodDelete, "/mock-struct/"+instance.StringID(), "", mockUser, mockLogger)
	rr := executeHandler(t, DefaultDeleteHandler(mockStruct, mockDb), req)

	// Assertions
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "has been deleted")
}

func TestDefaultDeleteHandler_InvalidMethod(t *testing.T) {
	mockDb, mockUser, mockLogger, mockStruct := setupTest(t)

	// Create and execute request with invalid method
	req := createTestRequest(t, http.MethodPost, "/mock-struct/123", "", mockUser, mockLogger)
	rr := executeHandler(t, DefaultDeleteHandler(mockStruct, mockDb), req)

	// Assertions
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Contains(t, rr.Body.String(), "Method not allowed")
}

func TestDefaultDeleteHandler_UnauthorizedUser_ReadPermission(t *testing.T) {
	mockDb, _, mockLogger, mockStruct := setupTest(t)

	// Create a user with insufficient permissions
	unauthorizedUser := &models.User{
		ID:    uint(999),
		Name:  "Test User",
		Email: "YHs7r@example.com",
		Roles: "invalid",
	}

	// Create and execute request
	req := createTestRequest(t, http.MethodDelete, "/mock-struct/123", "", unauthorizedUser, mockLogger)
	rr := executeHandler(t, DefaultDeleteHandler(mockStruct, mockDb), req)

	// Assertions
	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "User is not allowed to access this resource")
}

func TestDefaultDeleteHandler_UnauthorizedUser_DeletePermission(t *testing.T) {
	mockDb, _, mockLogger, mockStruct := setupTest(t)

	// Create a user with read but no delete permissions
	unauthorizedUser := &models.User{
		ID:    uint(999),
		Name:  "Test User",
		Email: "YHs7r@example.com",
		Roles: "visitor",
	}

	// Create and execute request
	req := createTestRequest(t, http.MethodDelete, "/mock-struct/123", "", unauthorizedUser, mockLogger)
	rr := executeHandler(t, DefaultDeleteHandler(mockStruct, mockDb), req)

	// Assertions
	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "User is not allowed to delete this resource")
}

func TestDefaultDeleteHandler_ResourceNotFound(t *testing.T) {
	mockDb, mockUser, mockLogger, mockStruct := setupTest(t)

	// Create and execute request for a non-existent resource
	req := createTestRequest(t, http.MethodDelete, "/mock-struct/99999", "", mockUser, mockLogger)
	rr := executeHandler(t, DefaultDeleteHandler(mockStruct, mockDb), req)

	// Assertions
	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "Instance not found")
}

func TestDefaultDeleteHandler_DatabaseError(t *testing.T) {
	mockDb, mockUser, mockLogger, mockStruct := setupTest(t)

	// Intentionally break the database connection
	mockDb.Close()

	// Create and execute request
	req := createTestRequest(t, http.MethodDelete, "/mock-struct/123", "", mockUser, mockLogger)
	rr := executeHandler(t, DefaultDeleteHandler(mockStruct, mockDb), req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Error finding resource")
}

func TestDefaultDeleteHandler_AdminBypassUserBinding(t *testing.T) {
	mockDb, _, mockLogger, mockStruct := setupTest(t)

	// Create an admin user
	adminUser := &models.User{
		ID:    uint(999),
		Name:  "Admin User",
		Email: "admin@example.com",
		Roles: models.AdminRole.S(),
	}

	// Create a resource owned by another user
	instance := createTestResource(t, mockDb, uint(77)) // Different from adminUser

	// Create and execute request
	req := createTestRequest(t, http.MethodDelete, "/mock-struct/"+instance.StringID(), "", adminUser, mockLogger)
	rr := executeHandler(t, DefaultDeleteHandler(mockStruct, mockDb), req)

	// Assertions
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "has been deleted")
}

func TestDefaultDeleteHandler_UserCannotDeleteResourceTheyDidNotCreate(t *testing.T) {
	mockDb, _, mockLogger, mockStruct := setupTest(t)

	// Create a resource owned by another user
	otherUser := &models.User{
		ID:    uint(1000),
		Name:  "Other User",
		Email: "other@example.com",
		Roles: "visitor",
	}
	instance := createTestResource(t, mockDb, otherUser.ID)

	// Create a request from a different user
	currentUser := &models.User{
		ID:    uint(999),
		Name:  "Test User",
		Email: "test@example.com",
		Roles: "visitor", // Same role but not the creator
	}

	// Create and execute request
	req := createTestRequest(t, http.MethodDelete, "/mock-struct/"+instance.StringID(), "", currentUser, mockLogger)
	rr := executeHandler(t, DefaultDeleteHandler(mockStruct, mockDb), req)

	// Assertions
	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "User is not allowed to delete this resource")
}
