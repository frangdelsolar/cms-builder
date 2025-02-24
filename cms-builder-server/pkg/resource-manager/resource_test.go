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

func TestGetInstance(t *testing.T) {
	r := &Resource{Model: TestModel{}}
	instance := r.GetInstance()

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
