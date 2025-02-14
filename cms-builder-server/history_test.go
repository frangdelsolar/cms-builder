package builder_test

import (
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
			got, err := builder.NewLogHistoryEntry(tt.action, &tt.user, tt.object)
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
