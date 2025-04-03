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
	name, err := GetInterfaceName(User{})
	assert.NoError(t, err)
	assert.Equal(t, "User", name)

	// Test with a pointer to a struct
	name, err = GetInterfaceName(&User{})
	assert.NoError(t, err)
	assert.Equal(t, "User", name)

	// Test with a non-struct
	name, err = GetInterfaceName(42)
	assert.Error(t, err)
	assert.Equal(t, "", name)

	// Test with a nil interface
	name, err = GetInterfaceName(nil)
	assert.Error(t, err)
	assert.Equal(t, "", name)
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
		{
			name: "Time comparison with one nil",
			a:    map[string]interface{}{"timestamp": nil},
			b:    map[string]interface{}{"timestamp": "1970-01-01T00:00:00Z"},
			expected: map[string]interface{}{
				"timestamp": []interface{}{nil, "1970-01-01T00:00:00Z"},
			},
		},
		{
			name: "Different times",
			a:    map[string]interface{}{"timestamp": "2025-04-03T14:01:58.606034-03:00"},
			b:    map[string]interface{}{"timestamp": "2025-04-03T15:01:58.606034-03:00"},
			expected: map[string]interface{}{
				"timestamp": []interface{}{"2025-04-03T14:01:58.606034-03:00", "2025-04-03T15:01:58.606034-03:00"},
			},
		},
		{
			name:     "Equal times",
			a:        map[string]interface{}{"timestamp": "2025-04-03T14:01:58.606034-03:00"},
			b:        map[string]interface{}{"timestamp": "2025-04-03T14:01:58.606034-03:00"},
			expected: map[string]interface{}{},
		},
		{
			name: "Nil vs zero time",
			a:    map[string]interface{}{"adFreeUntil": nil},
			b:    map[string]interface{}{"adFreeUntil": "1970-01-01T00:00:00Z"},
			expected: map[string]interface{}{
				"adFreeUntil": []interface{}{nil, "1970-01-01T00:00:00Z"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CompareInterfaces(tc.a, tc.b)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Test case '%s' failed.\nExpected: %v\nGot:      %v", tc.name, tc.expected, result)
			}
		})
	}
}
func TestDeepcopyInt(t *testing.T) {
	original := 42
	copied := Deepcopy(original).(int)
	assert.Equal(t, original, copied)
	assert.NotSame(t, &original, &copied)
}

func TestDeepcopyString(t *testing.T) {
	original := "hello"
	copied := Deepcopy(original).(string)
	assert.Equal(t, original, copied)
	assert.NotSame(t, &original, &copied)
}

func TestDeepcopySlice(t *testing.T) {
	original := []int{1, 2, 3}
	copied := Deepcopy(original).([]int)
	assert.Equal(t, original, copied)
	assert.NotSame(t, &original, &copied)
	if len(original) > 0 {
		assert.NotSame(t, &original[0], &copied[0]) //check that elements are deep copied
	}
}

func TestDeepcopyMap(t *testing.T) {
	original := map[string]int{"a": 1, "b": 2}
	copied := Deepcopy(original).(map[string]int)
	assert.Equal(t, original, copied)
	assert.NotSame(t, &original, &copied)
	// No way to test internal map elements directly, but deep copy of the map itself is sufficient
}

func TestDeepcopyStruct(t *testing.T) {
	type MyStruct struct {
		A int
		B string
		C []int
	}
	original := MyStruct{A: 1, B: "test", C: []int{4, 5, 6}}
	copied := Deepcopy(original).(MyStruct)
	assert.Equal(t, original, copied)
	assert.NotSame(t, &original, &copied)
	assert.NotSame(t, &original.C, &copied.C)
	assert.Equal(t, original.C, copied.C)
	if len(original.C) > 0 {
		assert.NotSame(t, &original.C[0], &copied.C[0])
	}
}

func TestDeepcopyNestedStruct(t *testing.T) {
	type Inner struct {
		Value int
	}
	type Outer struct {
		Inner Inner
	}
	original := Outer{Inner: Inner{Value: 10}}
	copied := Deepcopy(original).(Outer)
	assert.Equal(t, original, copied)
	assert.NotSame(t, &original, &copied)
	assert.NotSame(t, &original.Inner, &copied.Inner)
}

func TestDeepcopyPointer(t *testing.T) {
	original := 42
	ptr := &original
	copiedPtr := Deepcopy(ptr).(*int)
	assert.Equal(t, *ptr, *copiedPtr)
	assert.NotSame(t, &ptr, &copiedPtr)
	assert.NotSame(t, &original, &copiedPtr)
}

func TestDeepcopyNilPointer(t *testing.T) {
	var ptr *int
	copiedPtr := Deepcopy(ptr)
	assert.Nil(t, copiedPtr)
}

func TestDeepcopyInterface(t *testing.T) {
	var original interface{} = []int{1, 2, 3}
	copied := Deepcopy(original).([]int)
	assert.Equal(t, original, copied)
	assert.NotSame(t, &original, &copied)
}

func TestDeepcopyEmptyInterface(t *testing.T) {
	var original interface{}
	copied := Deepcopy(original)
	assert.Nil(t, copied)
}
