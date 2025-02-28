package database_test

// import (
// 	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
// )

// type TestStruct struct {
// 	ID   string `json:"ID"`
// 	Name string `json:"Name"`
// }

// func TestNewDatabaseLogEntry(t *testing.T) {
// 	tests := []struct {
// 		name       string
// 		action     CRUDAction
// 		user       User
// 		object     interface{}
// 		wantErr    bool
// 		wantDetail string
// 	}{
// 		{
// 			name:   "success",
// 			action: CreateCRUDAction,
// 			user: User{
// 				ID:    uint(1),
// 				Name:  "Test User",
// 				Email: "YHs7r@example.com",
// 			},
// 			object:     &TestStruct{ID: "123", Name: "Test"},
// 			wantErr:    false,
// 			wantDetail: "{\"ID\":\"123\",\"Name\":\"Test\"}",
// 		},
// 		{
// 			name:   "marshal error",
// 			action: CreateCRUDAction,
// 			user: User{
// 				ID:    uint(1),
// 				Name:  "Test User",
// 				Email: "YHs7r@example.com",
// 			},
// 			object:     make(chan int), // cannot be marshaled to JSON
// 			wantErr:    true,
// 			wantDetail: "",
// 		},
// 		{
// 			name:   "unmarshal error",
// 			action: CreateCRUDAction,
// 			user: User{
// 				ID:    uint(1),
// 				Name:  "Test User",
// 				Email: "YHs7r@example.com",
// 			},
// 			object:     "invalid json",
// 			wantErr:    true,
// 			wantDetail: "",
// 		},
// 		{
// 			name:   "no ID",
// 			action: CreateCRUDAction,
// 			user: User{
// 				ID:    uint(1),
// 				Name:  "Test User",
// 				Email: "YHs7r@example.com",
// 			},
// 			object:     &TestStruct{Name: "Test"},
// 			wantErr:    false,
// 			wantDetail: "{\"ID\":\"\",\"Name\":\"Test\"}",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := NewDatabaseLogEntry(tt.action, &tt.user, tt.object, nil, "23")
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("NewLogHistoryEntry() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !tt.wantErr {
// 				if got.Detail != tt.wantDetail {
// 					t.Errorf("NewLogHistoryEntry() detail = %q, want %q", got.Detail, tt.wantDetail)
// 				}
// 				if got.Timestamp == "" {
// 					t.Errorf("NewLogHistoryEntry() timestamp is empty")
// 				}
// 				if got.ResourceName == "" {
// 					t.Errorf("NewLogHistoryEntry() model name is empty")
// 				}
// 			}
// 		})
// 	}
// }
