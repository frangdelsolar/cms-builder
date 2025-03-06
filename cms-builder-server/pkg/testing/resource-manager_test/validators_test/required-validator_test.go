package resourcemanager_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	rmPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
	rmValidators "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/validators"
)

func TestRequiredValidator_Success(t *testing.T) {
	// Test data
	entity := rmTypes.EntityData{
		"name": "John Doe",
	}

	// Call the validator
	output := rmPkg.NewFieldValidationError("name")
	result := rmValidators.RequiredValidator("name", entity, &output)

	// Assertions
	assert.Equal(t, result.Error, "")
}

func TestRequiredValidator_Failure_MissingField(t *testing.T) {
	// Test data
	entity := rmTypes.EntityData{}

	// Call the validator
	output := rmPkg.NewFieldValidationError("name")
	result := rmValidators.RequiredValidator("name", entity, &output)

	// Assertions
	assert.NotNil(t, result)
	assert.Equal(t, "name is required", result.Error)
}

func TestRequiredValidator_Failure_EmptyField(t *testing.T) {
	// Test data
	entity := rmTypes.EntityData{
		"name": "",
	}

	// Call the validator
	output := rmPkg.NewFieldValidationError("name")
	result := rmValidators.RequiredValidator("name", entity, &output)

	// Assertions
	assert.NotNil(t, result)
	assert.Equal(t, "name is required", result.Error)
}
