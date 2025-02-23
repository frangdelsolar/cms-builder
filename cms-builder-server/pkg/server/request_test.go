package server_test

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

// TestValidateRequestMethod_Valid tests that ValidateRequestMethod returns nil for a valid request method.
func TestValidateRequestMethod_Valid(t *testing.T) {
	// Create a test request with a valid method
	req := httptest.NewRequest("GET", "https://example.com", nil)

	// Validate the request method
	err := ValidateRequestMethod(req, "GET")

	// Verify that no error is returned
	assert.NoError(t, err)
}

// TestValidateRequestMethod_Invalid tests that ValidateRequestMethod returns an error for an invalid request method.
func TestValidateRequestMethod_Invalid(t *testing.T) {
	// Create a test request with an invalid method
	req := httptest.NewRequest("POST", "https://example.com", nil)

	// Validate the request method
	err := ValidateRequestMethod(req, "GET")

	// Verify that an error is returned
	assert.Error(t, err)
	assert.Equal(t, "invalid request method: POST", err.Error())
}
