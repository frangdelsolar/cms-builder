package utils_test

import (
	"testing"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// TestGetStructName tests the GetStructName function.
func TestGetStructName(t *testing.T) {
	// Define a test struct
	type User struct{}

	// Test with a struct
	name := GetStructName(User{})
	assert.Equal(t, "User", name)

	// Test with a pointer to a struct
	name = GetStructName(&User{})
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
