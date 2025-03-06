package resourcemanager_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/types"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	rmPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
	rmValidators "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/validators"
	testPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
)

func TestNewResourceManager(t *testing.T) {
	bed := testPkg.SetupHandlerTestBed()
	rm := rmPkg.NewResourceManager(bed.Db, bed.Logger)
	assert.NotNil(t, rm, "NewResourceManager() returned nil")
	assert.NotNil(t, rm.Resources, "NewResourceManager().Resources is nil")
}

func TestResourceManager_AddResource(t *testing.T) {

	bed := testPkg.SetupHandlerTestBed()
	rm := rmPkg.NewResourceManager(bed.Db, bed.Logger)

	type TestResource struct {
		Name string `json:"name"`
	}

	config := &rmTypes.ResourceConfig{
		Model:           TestResource{},
		SkipUserBinding: true,
		Validators: rmTypes.ValidatorsMap{
			"Name": rmTypes.ValidatorsList{rmValidators.RequiredValidator},
		},
		Permissions: authTypes.RolePermissionMap{
			authConstants.AdminRole: []authTypes.CrudOperation{authConstants.OperationCreate},
		},
		Handlers: &rmTypes.ApiHandlers{
			List:   nil, // Use default
			Detail: nil, // Use default
			Create: nil, // Use default
			Update: nil, // Use default
			Delete: nil, // Use default
		},
	}

	resource, err := rm.AddResource(config)
	assert.NoError(t, err, "AddResource() returned error")

	assert.NotNil(t, resource, "AddResource() returned nil")
	assert.Equal(t, TestResource{}, resource.Model, "AddResource() model mismatch")
	assert.True(t, resource.SkipUserBinding, "AddResource() SkipUserBinding mismatch")
	assert.NotNil(t, resource.Validators, "AddResource() validators is nil")
	assert.NotNil(t, resource.Permissions, "AddResource() permissions is nil")
	assert.NotNil(t, resource.Api, "AddResource() api handlers is nil")
	assert.NotNil(t, resource.Api.List, "AddResource() default list handler not initialized")
	assert.NotNil(t, resource.Api.Detail, "AddResource() default detail handler not initialized")
	assert.NotNil(t, resource.Api.Create, "AddResource() default create handler not initialized")
	assert.NotNil(t, resource.Api.Update, "AddResource() default update handler not initialized")
	assert.NotNil(t, resource.Api.Delete, "AddResource() default delete handler not initialized")

	// Test custom handlers
	type TestResource2 struct {
		Name string `json:"name"`
	}

	customListHandler := func(*rmTypes.Resource, *dbTypes.DatabaseConnection) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {}
	}

	config.Handlers.List = customListHandler
	config.Model = TestResource2{}

	resource, err = rm.AddResource(config)
	assert.NoError(t, err, "AddResource() returned error")
	assert.NotNil(t, resource.Api.List, "Custom List handler not set")

	// Test nil stuff
	type TestResource3 struct {
		Name string `json:"name"`
	}
	config = &rmTypes.ResourceConfig{
		Model:           TestResource3{},
		SkipUserBinding: true,
		Handlers: &rmTypes.ApiHandlers{
			List:   nil,
			Detail: nil,
			Create: nil,
			Update: nil,
			Delete: nil,
		},
	}

	resource, err = rm.AddResource(config)
	assert.NoError(t, err, "AddResource() returned error")
	assert.NotNil(t, resource.Validators, "AddResource() validators should be initialized to empty map if nil")
	assert.NotNil(t, resource.Permissions, "AddResource() permissions should be initialized to empty map if nil")
}
