package resourcemanager_test

// func TestNewResourceManager(t *testing.T) {
// 	rm := NewResourceManager()
// 	assert.NotNil(t, rm, "NewResourceManager() returned nil")
// 	assert.NotNil(t, rm.Resources, "NewResourceManager().Resources is nil")
// }

// func TestResourceManager_AddResource(t *testing.T) {
// 	rm := NewResourceManager()
// 	config := &ResourceConfig{
// 		Model:           TestModel{},
// 		SkipUserBinding: true,
// 		Validators: ValidatorsMap{
// 			"Name": ValidatorsList{RequiredValidator},
// 		},
// 		Permissions: server.RolePermissionMap{
// 			models.AdminRole: []server.CrudOperation{server.OperationCreate},
// 		},
// 		Handlers: &ApiHandlers{
// 			List:   nil, // Use default
// 			Detail: nil, // Use default
// 			Create: nil, // Use default
// 			Update: nil, // Use default
// 			Delete: nil, // Use default
// 		},
// 	}

// 	resource, err := rm.AddResource(config)
// 	assert.NoError(t, err, "AddResource() returned error")

// 	assert.NotNil(t, resource, "AddResource() returned nil")
// 	assert.Equal(t, TestModel{}, resource.Model, "AddResource() model mismatch")
// 	assert.True(t, resource.SkipUserBinding, "AddResource() SkipUserBinding mismatch")
// 	assert.NotNil(t, resource.Validators, "AddResource() validators is nil")
// 	assert.NotNil(t, resource.Permissions, "AddResource() permissions is nil")
// 	assert.NotNil(t, resource.Api, "AddResource() api handlers is nil")
// 	assert.NotNil(t, resource.Api.List, "AddResource() default list handler not initialized")
// 	assert.NotNil(t, resource.Api.Detail, "AddResource() default detail handler not initialized")
// 	assert.NotNil(t, resource.Api.Create, "AddResource() default create handler not initialized")
// 	assert.NotNil(t, resource.Api.Update, "AddResource() default update handler not initialized")
// 	assert.NotNil(t, resource.Api.Delete, "AddResource() default delete handler not initialized")

// 	// Test custom handlers
// 	customListHandler := func(*Resource, *database.Database) http.HandlerFunc {
// 		return func(w http.ResponseWriter, r *http.Request) {}
// 	}

// 	config.Handlers.List = customListHandler

// 	resource, err = rm.AddResource(config)
// 	assert.NoError(t, err, "AddResource() returned error")

// 	assert.NotNil(t, resource.Api.List, "Custom List handler not set")

// 	config.Handlers.List = nil

// 	resource, err = rm.AddResource(config)
// 	assert.NoError(t, err, "AddResource() returned error")

// 	assert.NotNil(t, resource.Api.List, "AddResource() default list handler not initialized")

// 	// Test nil validators and permissions
// 	config = &ResourceConfig{
// 		Model:           TestModel{},
// 		SkipUserBinding: true,
// 		Handlers: &ApiHandlers{
// 			List:   nil,
// 			Detail: nil,
// 			Create: nil,
// 			Update: nil,
// 			Delete: nil,
// 		},
// 	}

// 	resource, err = rm.AddResource(config)
// 	assert.NoError(t, err, "AddResource() returned error")
// 	assert.NotNil(t, resource.Validators, "AddResource() validators should be initialized to empty map if nil")
// 	assert.NotNil(t, resource.Permissions, "AddResource() permissions should be initialized to empty map if nil")
// }
