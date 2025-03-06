package resourcemanager_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	rmPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
	rmValidators "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/validators"
)

func TestEmailValidator_Success(t *testing.T) {
	// Test data
	entity := rmTypes.EntityData{
		"email": "john.doe@example.com",
	}

	// Call the validator
	output := rmPkg.NewFieldValidationError("email")
	result := rmValidators.EmailValidator("email", entity, &output)

	// Assertions
	assert.Equal(t, result.Error, "")
}

func TestEmailValidator_Failure_InvalidEmail(t *testing.T) {
	// Test data
	entity := rmTypes.EntityData{
		"email": "invalid-email",
	}

	// Call the validator
	output := rmPkg.NewFieldValidationError("email")
	result := rmValidators.EmailValidator("email", entity, &output)

	// Assertions
	assert.NotNil(t, result)
	assert.Equal(t, "email has an invalid format", result.Error)
}

func TestEmailValidator_Success_EmptyEmail(t *testing.T) {
	// Test data
	entity := rmTypes.EntityData{
		"email": "",
	}

	// Call the validator
	output := rmPkg.NewFieldValidationError("email")
	result := rmValidators.EmailValidator("email", entity, &output)

	// Assertions
	assert.Equal(t, result.Error, "")
}
