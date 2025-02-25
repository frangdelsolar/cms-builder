package resourcemanager

import (
	"fmt"
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
	"github.com/rs/zerolog/log"
)

type ResourceManager struct {
	Resources map[string]*Resource
	DB        *database.Database
	Logger    *logger.Logger
}

func NewResourceManager(db *database.Database, log *logger.Logger) *ResourceManager {
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
	}

	// Validate Name
	name, err := resource.GetName()
	if err != nil {
		return nil, err
	}

	if _, ok := r.Resources[name]; ok {
		return nil, fmt.Errorf("resource with name %s already exists", name)
	}

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

	r.Logger.Debug().Str("name", name).Int("routes", len(resource.Routes)).Msg("Resource added to resource manager with")

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

func InitializeRoutes(r *Resource, input []server.Route, db *database.Database) error {
	name, _ := r.GetKebabCasePluralName()
	baseRoute := "/api/" + name

	routes := []server.Route{
		{
			Path:         baseRoute,
			Handler:      r.Api.List(r, db),
			Name:         fmt.Sprintf("%s:list", name),
			RequiresAuth: true,
			Methods:      []string{http.MethodGet},
		},
		{
			Path:         baseRoute + "/schema",
			Handler:      r.Api.Schema(r),
			Name:         fmt.Sprintf("%s:schema", name),
			RequiresAuth: false,
			Methods:      []string{http.MethodGet},
		},
		{
			Path:         baseRoute + "/new",
			Handler:      r.Api.Create(r, db),
			Name:         fmt.Sprintf("%s:create", name),
			RequiresAuth: true,
			Methods:      []string{http.MethodPost},
		},
		{
			Path:         baseRoute + "/{id}/delete",
			Handler:      r.Api.Delete(r, db),
			Name:         fmt.Sprintf("%s:delete", name),
			RequiresAuth: true,
			Methods:      []string{http.MethodDelete},
		},
		{
			Path:         baseRoute + "/{id}/update",
			Handler:      r.Api.Update(r, db),
			Name:         fmt.Sprintf("%s:update", name),
			RequiresAuth: true,
			Methods:      []string{http.MethodPut},
		},
		{
			Path:         baseRoute + "/{id}",
			Handler:      r.Api.Detail(r, db),
			Name:         fmt.Sprintf("%s:detail", name),
			RequiresAuth: true,
			Methods:      []string{http.MethodGet},
		},
	}

	for _, route := range routes {
		err := r.AddRoute(route)
		if err != nil {
			return err
		}
	}

	for _, route := range input {
		err := r.AddRoute(route)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Routes initialized %d routes for resource: %s\n", len(routes), name)

	return nil
}

// InitializeValidators adds validators to the given resource.
// It loops through the given map of validators and adds each one to the resource's Validators map.
func InitializeValidators(r *Resource, input *ValidatorsMap) error {
	for fieldName, validators := range *input {
		for _, validator := range validators {
			err := r.AddValidator(fieldName, validator)
			if err != nil {
				log.Error().Err(err).Str("field", fieldName).Msg("Failed to add validator")
				return err
			}
		}
	}
	return nil
}

// InitializeHandlers returns a new ApiHandlers struct with default handlers.
// If the given input is not nil, it overwrites the default handlers with the given functions.
func InitializeHandlers(input *ApiHandlers) *ApiHandlers {
	handlers := &ApiHandlers{
		List:   DefaultListHandler,
		Detail: DefaultDetailHandler,
		Create: DefaultCreateHandler,
		Update: DefaultUpdateHandler,
		Delete: DefaultDeleteHandler,
		Schema: DefaultSchemaHandler,
	}

	if input != nil {
		if input.List != nil {
			handlers.List = input.List
		}

		if input.Detail != nil {
			handlers.Detail = input.Detail
		}

		if input.Create != nil {
			handlers.Create = input.Create
		}

		if input.Update != nil {
			handlers.Update = input.Update
		}

		if input.Delete != nil {
			handlers.Delete = input.Delete
		}

		if input.Schema != nil {
			handlers.Schema = input.Schema
		}
	}

	return handlers
}
