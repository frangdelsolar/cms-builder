package resourcemanager_test

import (
	"testing"

	mgr "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	"github.com/stretchr/testify/assert"
)

func TestRequiredValidator_Success(t *testing.T) {
	// Test data
	entity := mgr.EntityData{
		"name": "John Doe",
	}

	// Call the validator
	output := mgr.NewFieldValidationError("name")
	result := mgr.RequiredValidator("name", entity, &output)

	// Assertions
	assert.Equal(t, result.Error, "")
}

func TestRequiredValidator_Failure_MissingField(t *testing.T) {
	// Test data
	entity := mgr.EntityData{}

	// Call the validator
	output := mgr.NewFieldValidationError("name")
	result := mgr.RequiredValidator("name", entity, &output)

	// Assertions
	assert.NotNil(t, result)
	assert.Equal(t, "name is required", result.Error)
}

func TestRequiredValidator_Failure_EmptyField(t *testing.T) {
	// Test data
	entity := mgr.EntityData{
		"name": "",
	}

	// Call the validator
	output := mgr.NewFieldValidationError("name")
	result := mgr.RequiredValidator("name", entity, &output)

	// Assertions
	assert.NotNil(t, result)
	assert.Equal(t, "name is required", result.Error)
}

func TestEmailValidator_Success(t *testing.T) {
	// Test data
	entity := mgr.EntityData{
		"email": "john.doe@example.com",
	}

	// Call the validator
	output := mgr.NewFieldValidationError("email")
	result := mgr.EmailValidator("email", entity, &output)

	// Assertions
	assert.Equal(t, result.Error, "")
}

func TestEmailValidator_Failure_InvalidEmail(t *testing.T) {
	// Test data
	entity := mgr.EntityData{
		"email": "invalid-email",
	}

	// Call the validator
	output := mgr.NewFieldValidationError("email")
	result := mgr.EmailValidator("email", entity, &output)

	// Assertions
	assert.NotNil(t, result)
	assert.Equal(t, "email has an invalid format", result.Error)
}

func TestEmailValidator_Success_EmptyEmail(t *testing.T) {
	// Test data
	entity := mgr.EntityData{
		"email": "",
	}

	// Call the validator
	output := mgr.NewFieldValidationError("email")
	result := mgr.EmailValidator("email", entity, &output)

	// Assertions
	assert.Equal(t, result.Error, "")
}
