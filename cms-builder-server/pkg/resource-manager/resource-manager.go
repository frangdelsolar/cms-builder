package resourcemanager

import (
	"fmt"
	"net/http"

	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
)

type ResourceManager struct {
	Resources map[string]*Resource
	DB        *dbTypes.DatabaseConnection
	Logger    *loggerTypes.Logger
}

func NewResourceManager(db *dbTypes.DatabaseConnection, log *loggerTypes.Logger) *ResourceManager {
	return &ResourceManager{
		Resources: make(map[string]*Resource),
		DB:        db,
		Logger:    log,
	}
}

type ResourceConfig struct {
	Model           interface{}
	Handlers        *ApiHandlers
	SkipUserBinding bool
	Validators      ValidatorsMap
	Permissions     server.RolePermissionMap
	Routes          []server.Route
}

func (r *ResourceManager) GetResourceByName(name string) (*Resource, error) {
	if resource, ok := r.Resources[name]; ok {
		return resource, nil
	}

	return nil, fmt.Errorf("resource with name %s not found", name)
}

func (r *ResourceManager) GetResource(input interface{}) (*Resource, error) {
	name, err := utils.GetInterfaceName(input)
	if err != nil {
		return nil, err
	}

	if resource, ok := r.Resources[name]; ok {
		return resource, nil
	}

	return nil, fmt.Errorf("resource with name %s not found", name)
}

func (r *ResourceManager) AddResource(input *ResourceConfig) (*Resource, error) {

	resource := &Resource{
		Model:           input.Model,
		Api:             InitializeHandlers(input.Handlers),
		SkipUserBinding: input.SkipUserBinding,
		Permissions:     make(server.RolePermissionMap),
		Validators:      make(ValidatorsMap),
		Routes:          map[string]server.Route{},
		ResourceNames:   ResourceNames{},
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

func (r *ResourceManager) GetRoutes(apiBaseUrl string) []server.Route {

	routes := []server.Route{
		{
			Path:         "/api",
			Handler:      ApiHandler(r, apiBaseUrl),
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
