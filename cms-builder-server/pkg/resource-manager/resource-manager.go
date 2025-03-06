package resourcemanager

import (
	"fmt"
	"net/http"

	authTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/types"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	rmHandlers "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/handlers"
	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
	svrTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/types"
	utilsPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
)

type ResourceManager struct {
	Resources map[string]*rmTypes.Resource
	DB        *dbTypes.DatabaseConnection
	Logger    *loggerTypes.Logger
}

func (r *ResourceManager) GetResourceByName(name string) (*rmTypes.Resource, error) {
	if resource, ok := r.Resources[name]; ok {
		return resource, nil
	}

	return nil, fmt.Errorf("resource with name %s not found", name)
}

func (r *ResourceManager) GetResource(input interface{}) (*rmTypes.Resource, error) {
	name, err := utilsPkg.GetInterfaceName(input)
	if err != nil {
		return nil, err
	}

	if resource, ok := r.Resources[name]; ok {
		return resource, nil
	}

	return nil, fmt.Errorf("resource with name %s not found", name)
}

func (r *ResourceManager) GetRoutes(apiBaseUrl string) []svrTypes.Route {

	routes := []svrTypes.Route{
		{
			Path:         "/api",
			Handler:      rmHandlers.ApiHandler(r.Resources, apiBaseUrl),
			Name:         "api",
			RequiresAuth: false,
			Methods:      []string{http.MethodGet},
		},
	}

	for _, resource := range r.Resources {
		for _, route := range resource.Routes {
			routes = append(routes, route)
		}
	}

	return routes
}

func (r *ResourceManager) AddResource(input *rmTypes.ResourceConfig) (*rmTypes.Resource, error) {

	resource := &rmTypes.Resource{
		Model:           input.Model,
		Api:             InitializeHandlers(input.Handlers),
		SkipUserBinding: input.SkipUserBinding,
		Permissions:     make(authTypes.RolePermissionMap),
		Validators:      make(rmTypes.ValidatorsMap),
		Routes:          map[string]svrTypes.Route{},
		ResourceNames:   rmTypes.ResourceNames{},
		JsonSchema:      nil,
	}

	// Validate Name
	resourceNames, err := InitializeResourceNames(resource)
	if err != nil {
		r.Logger.Error().Err(err).Msg("Error initializing resource names")
		return nil, err
	}
	resource.ResourceNames = resourceNames

	name := resourceNames.Singular

	if _, ok := r.Resources[name]; ok {
		r.Logger.Error().Msgf("Resource with name %s already exists", name)
		return nil, fmt.Errorf("resource with name %s already exists", name)
	}

	// TODO: Initialize JsonSchema

	// TODO: Maybe I can initialize FieldNames, so that validations are easier

	// Validate Permissions
	if input.Permissions != nil {
		resource.Permissions = input.Permissions
	}

	// Validate Validators
	err = InitializeValidators(resource, &input.Validators)
	if err != nil {
		return nil, err
	}

	// Validate Routes
	err = InitializeRoutes(resource, input.Routes, r.DB)
	if err != nil {
		return nil, err
	}

	r.Resources[name] = resource

	// Migrate
	err = r.DB.DB.AutoMigrate(resource.Model)
	if err != nil {
		return nil, err
	}

	return resource, nil
}
