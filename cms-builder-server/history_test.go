package builder_test

import (
	"reflect"
	"testing"

	builder "github.com/frangdelsolar/cms-builder/cms-builder-server"
)

type TestStruct struct {
	ID   string `json:"ID"`
	Name string `json:"Name"`
}

func TestNewLogHistoryEntry(t *testing.T) {
	tests := []struct {
		name       string
		action     builder.CRUDAction
		user       builder.User
		object     interface{}
		wantErr    bool
		wantDetail string
	}{
		{
			name:   "success",
			action: builder.CreateCRUDAction,
			user: builder.User{
				ID:    uint(1),
				Name:  "Test User",
				Email: "YHs7r@example.com",
			},
			object:     &TestStruct{ID: "123", Name: "Test"},
			wantErr:    false,
			wantDetail: "{\"ID\":\"123\",\"Name\":\"Test\"}",
		},
		{
			name:   "marshal error",
			action: builder.CreateCRUDAction,
			user: builder.User{
				ID:    uint(1),
				Name:  "Test User",
				Email: "YHs7r@example.com",
			},
			object:     make(chan int), // cannot be marshaled to JSON
			wantErr:    true,
			wantDetail: "",
		},
		{
			name:   "unmarshal error",
			action: builder.CreateCRUDAction,
			user: builder.User{
				ID:    uint(1),
				Name:  "Test User",
				Email: "YHs7r@example.com",
			},
			object:     "invalid json",
			wantErr:    true,
			wantDetail: "",
		},
		{
			name:   "no ID",
			action: builder.CreateCRUDAction,
			user: builder.User{
				ID:    uint(1),
				Name:  "Test User",
				Email: "YHs7r@example.com",
			},
			object:     &TestStruct{Name: "Test"},
			wantErr:    false,
			wantDetail: "{\"ID\":\"\",\"Name\":\"Test\"}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builder.NewLogHistoryEntry(tt.action, &tt.user, tt.object, nil, "23")
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLogHistoryEntry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Detail != tt.wantDetail {
					t.Errorf("NewLogHistoryEntry() detail = %q, want %q", got.Detail, tt.wantDetail)
				}
				if got.Timestamp == "" {
					t.Errorf("NewLogHistoryEntry() timestamp is empty")
				}
				if got.ResourceName == "" {
					t.Errorf("NewLogHistoryEntry() model name is empty")
				}
			}
		})
	}
}

func TestGetDiff(t *testing.T) {
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
			result := builder.CompareInterfaces(tc.a, tc.b)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}
