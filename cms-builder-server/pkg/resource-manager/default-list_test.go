package resourcemanager_test

// import (
// 	"errors"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
// 	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
// 	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
// 	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
// 	"github.com/stretchr/testify/assert"
// 	"gorm.io/driver/sqlite"
// 	"gorm.io/gorm"
// )

// // MockResource is a mock implementation of the Resource interface.
// type MockResource struct {
// 	SkipUserBinding bool
// 	Permissions     server.RolePermissionMap
// }

// func (m *MockResource) GetSlice() (interface{}, error) {
// 	return &[]models.User{}, nil
// }

// func (m *MockResource) GetName() (string, error) {
// 	return "mock-resource", nil
// }

// func TestDefaultListHandler(t *testing.T) {
// 	// Setup in-memory SQLite database
// 	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
// 	assert.NoError(t, err)

// 	// Create a Database instance
// 	database := &database.Database{DB: db}

// 	// Create a mock resource
// 	mockResource := &MockResource{
// 		SkipUserBinding: false,
// 		Permissions: server.RolePermissionMap{
// 			models.AdminRole: []server.CrudOperation{server.OperationRead},
// 			models.VisitorRole: []server.CrudOperation{server.OperationRead},
// 		},
// 	}

// 	// Create a test user
// 	testUser := &models.User{
// 		ID:    uint(2),
// 		Name:  "John Doe",
// 		Roles: models.VisitorRole.S(),
// 	}

// 	// Test cases
// 	t.Run("Successful list with user binding", func(t *testing.T) {
// 		// Create a request with query parameters
// 		req, err := http.NewRequest(http.MethodGet, "/resources?page=1&limit=10&order=asc", nil)
// 		assert.NoError(t, err)

// 		// Add request context
// 		req = req.WithContext(server.SetRequestContext(req.Context(), &server.RequestContext{
// 			User:   testUser,
// 			Logger: server.NewLogger(),
// 		}))

// 		// Create a response recorder
// 		rr := httptest.NewRecorder()

// 		// Call the handler
// 		handler := DefaultListHandler(mockResource, database)
// 		handler.ServeHTTP(rr, req)

// 		// Verify the response
// 		assert.Equal(t, http.StatusOK, rr.Code)
// 	})

// 	t.Run("Successful list without user binding (admin)", func(t *testing.T) {
// 		// Create a request with query parameters
// 		req, err := http.NewRequest(http.MethodGet, "/resources?page=1&limit=10&order=asc", nil)
// 		assert.NoError(t, err)

// 		// Add request context with admin user
// 		adminUser := &models.User{
// 			ID:    "admin-123",
// 			Name:  "Admin User",
// 			Roles: []string{models.AdminRole},
// 		}
// 		req = req.WithContext(server.SetRequestContext(req.Context(), &server.RequestContext{
// 			User:   adminUser,
// 			Logger: server.NewLogger(),
// 		}))

// 		// Create a response recorder
// 		rr := httptest.NewRecorder()

// 		// Call the handler
// 		handler := DefaultListHandler(mockResource, database)
// 		handler.ServeHTTP(rr, req)

// 		// Verify the response
// 		assert.Equal(t, http.StatusOK, rr.Code)
// 	})

// 	t.Run("Permission denied", func(t *testing.T) {
// 		// Create a request with query parameters
// 		req, err := http.NewRequest(http.MethodGet, "/resources?page=1&limit=10&order=asc", nil)
// 		assert.NoError(t, err)

// 		// Add request context with unauthorized user
// 		unauthorizedUser := &models.User{
// 			ID:    "user-456",
// 			Name:  "Unauthorized User",
// 			Roles: []string{"guest"},
// 		}
// 		req = req.WithContext(server.SetRequestContext(req.Context(), &server.RequestContext{
// 			User:   unauthorizedUser,
// 			Logger: server.NewLogger(),
// 		}))

// 		// Create a response recorder
// 		rr := httptest.NewRecorder()

// 		// Call the handler
// 		handler := DefaultListHandler(mockResource, database)
// 		handler.ServeHTTP(rr, req)

// 		// Verify the response
// 		assert.Equal(t, http.StatusForbidden, rr.Code)
// 	})

// 	t.Run("Invalid request method", func(t *testing.T) {
// 		// Create a request with an invalid method (POST instead of GET)
// 		req, err := http.NewRequest(http.MethodPost, "/resources", nil)
// 		assert.NoError(t, err)

// 		// Add request context
// 		req = req.WithContext(server.SetRequestContext(req.Context(), &server.RequestContext{
// 			User:   testUser,
// 			Logger: server.NewLogger(),
// 		}))

// 		// Create a response recorder
// 		rr := httptest.NewRecorder()

// 		// Call the handler
// 		handler := DefaultListHandler(mockResource, database)
// 		handler.ServeHTTP(rr, req)

// 		// Verify the response
// 		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
// 	})

// 	t.Run("Error getting slice", func(t *testing.T) {
// 		// Create a mock resource that returns an error when GetSlice is called
// 		errorResource := &MockResource{
// 			SkipUserBinding: false,
// 			Permissions: models.Permissions{
// 				models.OperationRead: []string{models.AdminRole, models.UserRole},
// 			},
// 		}
// 		errorResource.GetSlice = func() (interface{}, error) {
// 			return nil, errors.New("failed to get slice")
// 		}

// 		// Create a request with query parameters
// 		req, err := http.NewRequest(http.MethodGet, "/resources?page=1&limit=10&order=asc", nil)
// 		assert.NoError(t, err)

// 		// Add request context
// 		req = req.WithContext(server.SetRequestContext(req.Context(), &server.RequestContext{
// 			User:   testUser,
// 			Logger: server.NewLogger(),
// 		}))

// 		// Create a response recorder
// 		rr := httptest.NewRecorder()

// 		// Call the handler
// 		handler := DefaultListHandler(errorResource, database)
// 		handler.ServeHTTP(rr, req)

// 		// Verify the response
// 		assert.Equal(t, http.StatusInternalServerError, rr.Code)
// 	})
// }
