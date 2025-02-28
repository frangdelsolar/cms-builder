package resourcemanager_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
)

// TestRequiredValidator tests the RequiredValidator function.
func TestRequiredValidator(t *testing.T) {
	tests := []struct {
		name          string
		fieldName     string
		entity        EntityData
		expectedError string
	}{
		{
			name:          "field is present and non-empty",
			fieldName:     "name",
			entity:        EntityData{"name": "John Doe"},
			expectedError: "",
		},
		{
			name:          "field is missing",
			fieldName:     "name",
			entity:        EntityData{},
			expectedError: "name is required",
		},
		{
			name:          "field is empty",
			fieldName:     "name",
			entity:        EntityData{"name": ""},
			expectedError: "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new ValidationError
			output := NewFieldValidationError(tt.fieldName)

			// Call the validator
			result := RequiredValidator(tt.fieldName, tt.entity, &output)

			// Check the result
			if tt.expectedError == "" {
				assert.Nil(t, result, "Expected no error")
			} else {
				assert.NotNil(t, result, "Expected an error")
				assert.Equal(t, tt.expectedError, result.Error, "Unexpected error message")
			}
		})
	}
}

// TestEmailValidator tests the EmailValidator function.
func TestEmailValidator(t *testing.T) {
	tests := []struct {
		name          string
		fieldName     string
		entity        EntityData
		expectedError string
	}{
		{
			name:          "valid email",
			fieldName:     "email",
			entity:        EntityData{"email": "test@example.com"},
			expectedError: "",
		},
		{
			name:          "invalid email",
			fieldName:     "email",
			entity:        EntityData{"email": "invalid-email"},
			expectedError: "email has an invalid format",
		},
		{
			name:          "empty email",
			fieldName:     "email",
			entity:        EntityData{"email": ""},
			expectedError: "",
		},
		{
			name:          "email with invalid characters",
			fieldName:     "email",
			entity:        EntityData{"email": "test@example..com"},
			expectedError: "email has an invalid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new ValidationError
			output := NewFieldValidationError(tt.fieldName)

			// Call the validator
			result := EmailValidator(tt.fieldName, tt.entity, &output)

			// Check the result
			if tt.expectedError == "" {
				assert.Nil(t, result, "Expected no error")
			} else {
				assert.NotNil(t, result, "Expected an error")
				assert.Equal(t, tt.expectedError, result.Error, "Unexpected error message")
			}
		})
	}
}
