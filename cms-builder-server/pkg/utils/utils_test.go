package utils_test

import (
	"reflect"
	"testing"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// TestGetStructName tests the GetStructName function.
func TestGetStructName(t *testing.T) {
	// Define a test struct
	type User struct{}

	// Test with a struct
	name := GetInterfaceName(User{})
	assert.Equal(t, "User", name)

	// Test with a pointer to a struct
	name = GetInterfaceName(&User{})
	assert.Equal(t, "User", name)
}

// TestPluralize tests the Pluralize function.
func TestPluralize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user", "users"},
		{"category", "categories"},
		{"box", "boxes"},
		{"child", "children"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Pluralize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSnakeCase tests the SnakeCase function.
func TestSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"User", "user"},
		{"UserProfile", "user_profile"},
		{"HTTPRequest", "http_request"},
		{"camelCase", "camel_case"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SnakeCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestKebabCase tests the KebabCase function.
func TestKebabCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"User", "user"},
		{"UserProfile", "user-profile"},
		{"HTTPRequest", "http-request"},
		{"camelCase", "camel-case"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := KebabCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompareInterfaces(t *testing.T) {
	type TestCase struct {
		name     string
		a        interface{}
		b        interface{}
		expected interface{}
	}

	testCases := []TestCase{
		{
			name:     "Both nil",
			a:        nil,
			b:        nil,
			expected: map[string]interface{}{},
		},
		{
			name:     "One nil",
			a:        nil,
			b:        map[string]interface{}{"key": "value"},
			expected: map[string]interface{}{"value": []interface{}{nil, map[string]interface{}{"key": "value"}}},
		},
		{
			name:     "Equal values",
			a:        map[string]interface{}{"key": "value"},
			b:        map[string]interface{}{"key": "value"},
			expected: map[string]interface{}{},
		},
		{
			name:     "Different values",
			a:        map[string]interface{}{"key": "value"},
			b:        map[string]interface{}{"key": "different value"},
			expected: map[string]interface{}{"key": []interface{}{"value", "different value"}},
		},
		{
			name:     "Nested maps",
			a:        map[string]interface{}{"key": map[string]interface{}{"nestedKey": "nestedValue"}},
			b:        map[string]interface{}{"key": map[string]interface{}{"nestedKey": "differentNestedValue"}},
			expected: map[string]interface{}{"key": map[string]interface{}{"nestedKey": []interface{}{"nestedValue", "differentNestedValue"}}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CompareInterfaces(tc.a, tc.b)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}
