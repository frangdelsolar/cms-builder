package resourcemanager_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	tu "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

func TestDefaultCreateHandler(t *testing.T) {
	// Setup
	mockDb := tu.GetTestDB()
	mockUser := tu.GetTestUser()
	mockLogger := tu.GetTestLogger()
	mockStruct := tu.GetMockResource()
	// mockResourceInstance := tu.GetMockResourceInstance()

	// Create request body
	requestBody := `{"field1": "John Doe", "field2": "john.doe@example.com", "createdById": 123}`
	req, err := http.NewRequest(http.MethodPost, "/mock-struct/new", bytes.NewBufferString(requestBody))
	assert.NoError(t, err)

	// Set request context
	ctx := req.Context()
	ctx = context.WithValue(ctx, server.CtxRequestIsAuth, true)
	ctx = context.WithValue(ctx, server.CtxRequestUser, mockUser)
	ctx = context.WithValue(ctx, server.CtxRequestLogger, mockLogger)

	req = req.WithContext(ctx)
	req.Method = http.MethodPost

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler := DefaultCreateHandler(mockStruct, mockDb)
	handler.ServeHTTP(rr, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Assertions
	t.Log(rr.Body.String())
}
