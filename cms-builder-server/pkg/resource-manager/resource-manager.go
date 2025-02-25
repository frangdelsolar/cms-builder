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

func (r *ResourceManager) Register(resource *Resource) error {

	name, err := resource.GetName()
	if err != nil {
		return err
	}

	if _, ok := r.Resources[name]; ok {
		return fmt.Errorf("resource with name %s already exists", name)
	}

	// Auto-migrate the resource model
	if err := r.DB.DB.AutoMigrate(resource.Model); err != nil {
		return fmt.Errorf("failed to auto-migrate resource model: %w", err)
	}

	r.Resources[name] = resource

	return nil
}

func InitializeHandlers(input *ApiHandlers) *ApiHandlers {
	handlers := &ApiHandlers{
		List:   DefaultListHandler,
		Detail: DefaultDetailHandler,
		Create: DefaultCreateHandler,
		Update: DefaultUpdateHandler,
		Delete: DefaultDeleteHandler,
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
	}

	return handlers
}

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

	log.Debug().Interface("validators", r.Validators).Msg("Initialized validators")

	return nil
}

// TODO: Is this really necessary?
func (r *ResourceManager) AddResource(input *ResourceConfig) (*Resource, error) {

	permissions := input.Permissions
	if permissions == nil {
		permissions = make(server.RolePermissionMap)
	}

	resource := &Resource{
		Model:           input.Model,
		Api:             InitializeHandlers(input.Handlers),
		SkipUserBinding: input.SkipUserBinding,
		Permissions:     permissions,
		Validators:      make(ValidatorsMap),
		Routes:          make([]server.Route, 0),
	}

	err := InitializeValidators(resource, &input.Validators)
	if err != nil {

		return nil, err
	}

	err = r.Register(resource)
	if err != nil {
		return nil, err
	}

	name, _ := resource.GetKebabCaseName()
	baseRoute := "/api/" + name

	routes := []server.Route{
		{
			Path:         baseRoute,
			Handler:      resource.Api.List(resource, r.DB),
			Name:         fmt.Sprintf("%s-list", input.Model),
			RequiresAuth: true,
			Method:       http.MethodGet,
		},
		{
			Path:         baseRoute + "/new",
			Handler:      resource.Api.Create(resource, r.DB),
			Name:         fmt.Sprintf("%s-create", input.Model),
			RequiresAuth: true,
			Method:       http.MethodPost,
		},
		{
			Path:         baseRoute + "/{id}/delete",
			Handler:      resource.Api.Delete(resource, r.DB),
			Name:         fmt.Sprintf("%s-delete", input.Model),
			RequiresAuth: true,
			Method:       http.MethodDelete,
		},
		{
			Path:         baseRoute + "/{id}/update",
			Handler:      resource.Api.Update(resource, r.DB),
			Name:         fmt.Sprintf("%s-update", input.Model),
			RequiresAuth: true,
			Method:       http.MethodPut,
		},
		{
			Path:         baseRoute + "/{id}",
			Handler:      resource.Api.Detail(resource, r.DB),
			Name:         fmt.Sprintf("%s-detail", input.Model),
			RequiresAuth: true,
			Method:       http.MethodGet,
		},
	}

	resource.Routes = append(resource.Routes, routes...)

	return resource, nil
}

func (r *ResourceManager) GetRoutes() []server.Route {
	routes := make([]server.Route, 0)

	for _, resource := range r.Resources {
		routes = append(routes, resource.Routes...)
	}

	return routes
}
