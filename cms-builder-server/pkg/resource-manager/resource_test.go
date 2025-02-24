package resourcemanager_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
)

type TestModel struct {
	ID   int
	Name string
}

func TestGetSlice(t *testing.T) {
	r := &Resource{Model: TestModel{}}
	slice, err := r.GetSlice()

	assert.NoError(t, err)

	sliceType := reflect.TypeOf(slice)
	expectedType := reflect.PtrTo(reflect.SliceOf(reflect.TypeOf(TestModel{}))) // Expect a pointer to a slice

	assert.Equal(t, expectedType, sliceType)
}

func TestGetOne(t *testing.T) {
	r := &Resource{Model: TestModel{}}
	instance := r.GetOne()

	instanceType := reflect.TypeOf(instance)
	expectedType := reflect.PtrTo(reflect.TypeOf(TestModel{}))

	assert.Equal(t, expectedType, instanceType)
}

func TestGetName(t *testing.T) {
	type Test1 struct{}
	type test2 struct{}

	tests := []struct {
		name        string
		model       interface{}
		expected    string
		expectError bool
	}{
		{
			name:        "Valid struct model",
			model:       Test1{},
			expected:    "Test1",
			expectError: false,
		},
		{
			name:        "Valid pointer to struct model",
			model:       &test2{},
			expected:    "test2",
			expectError: false,
		},
		{
			name:        "Non-struct model (string)",
			model:       "string",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Nil model",
			model:       nil,
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Resource{Model: tt.model}
			actual, err := a.GetName()

			if tt.expectError {
				assert.Error(t, err, "Expected an error but got nil")
			} else {
				assert.NoError(t, err, "Did not expect an error but got one")
				assert.Equal(t, tt.expected, actual, "Expected and actual model names do not match")
			}
		})
	}
}

// TestResource_GetKeys tests the GetKeys method.
func TestResource_GetKeys(t *testing.T) {
	// Define a test model with JSON tags
	type TestModel struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	// Define a test model without JSON tags
	type TestModelNoTags struct {
		ID    int
		Name  string
		Email string
	}

	// Define test cases
	tests := []struct {
		name          string
		model         interface{}
		expectedKeys  []string
		expectedError bool
	}{
		{
			name:          "struct with JSON tags",
			model:         TestModel{},
			expectedKeys:  []string{"id", "name", "email"},
			expectedError: false,
		},
		{
			name:          "pointer to struct with JSON tags",
			model:         &TestModel{},
			expectedKeys:  []string{"id", "name", "email"},
			expectedError: false,
		},
		{
			name:          "struct without JSON tags",
			model:         TestModelNoTags{},
			expectedKeys:  []string{},
			expectedError: false,
		},
		{
			name:          "non-struct type",
			model:         "not-a-struct",
			expectedKeys:  nil,
			expectedError: true,
		},
		{
			name:          "nil model",
			model:         nil,
			expectedKeys:  nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a Resource instance with the test model
			resource := &Resource{
				Model: tt.model,
			}

			// Call the GetKeys method
			keys := resource.GetKeys()

			// Verify the result
			if tt.expectedError {
				assert.Nil(t, keys, "Expected nil keys for test case: %s", tt.name)
			} else {
				assert.Equal(t, tt.expectedKeys, keys, "Unexpected keys for test case: %s", tt.name)
			}
		})
	}
}
