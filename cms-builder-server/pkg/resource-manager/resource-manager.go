package resourcemanager

import (
	"fmt"
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
)

type ResourceManager struct {
	Resources map[string]interface{}
	DB        *database.Database
	Logger    *logger.Logger
}

func NewResourceManager(db *database.Database, log *logger.Logger) *ResourceManager {
	return &ResourceManager{
		Resources: make(map[string]interface{}),
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
		return resource.(*Resource), nil
	}

	return nil, fmt.Errorf("resource with name %s not found", name)
}

func (r *ResourceManager) GetResource(input interface{}) (*Resource, error) {
	name, err := utils.GetInterfaceName(input)
	if err != nil {
		return nil, err
	}

	if resource, ok := r.Resources[name]; ok {
		return resource.(*Resource), nil
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

	// TODO: Move this to a better place
	r.DB.DB.AutoMigrate(resource.Model)

	r.Resources[name] = resource

	return nil
}

// TODO: Is this really necessary?
func (r *ResourceManager) AddResource(input *ResourceConfig) (*Resource, error) {

	// initialize default handlers
	handlers := &ApiHandlers{
		List:   DefaultListHandler,
		Detail: DefaultDetailHandler,
		Create: DefaultCreateHandler,
		Update: DefaultUpdateHandler,
		Delete: DefaultDeleteHandler,
	}

	if input.Handlers == nil {
		input.Handlers = &ApiHandlers{}
	}

	// check if custom handlers are provided
	if input.Handlers.List != nil {
		handlers.List = input.Handlers.List
	}

	if input.Handlers.Detail != nil {
		handlers.Detail = input.Handlers.Detail
	}

	if input.Handlers.Create != nil {
		handlers.Create = input.Handlers.Create
	}

	if input.Handlers.Update != nil {
		handlers.Update = input.Handlers.Update
	}

	if input.Handlers.Delete != nil {
		handlers.Delete = input.Handlers.Delete
	}

	validators := input.Validators
	if validators == nil {
		validators = make(ValidatorsMap)
	}

	permissions := input.Permissions
	if permissions == nil {
		permissions = make(server.RolePermissionMap)
	}

	resource := &Resource{
		Model:           input.Model,
		Api:             handlers,
		SkipUserBinding: input.SkipUserBinding,
		Validators:      validators,
		Permissions:     permissions,
		Routes:          make([]server.Route, 0),
	}

	err := r.Register(resource)
	if err != nil {
		return nil, err
	}

	name, _ := resource.GetName()
	baseRoute := "/api/" + name + "/"

	routes := []server.Route{
		{
			Path:         baseRoute,
			Handler:      resource.Api.List(resource, r.DB),
			Name:         fmt.Sprintf("%s-list", input.Model),
			RequiresAuth: true,
			Method:       http.MethodGet,
		},
		{
			Path:         baseRoute + "{id}",
			Handler:      resource.Api.Detail(resource, r.DB),
			Name:         fmt.Sprintf("%s-detail", input.Model),
			RequiresAuth: true,
			Method:       http.MethodGet,
		},
		{
			Path:         baseRoute + "/create",
			Handler:      resource.Api.Create(resource, r.DB),
			Name:         fmt.Sprintf("%s-create", input.Model),
			RequiresAuth: true,
			Method:       http.MethodPost,
		},
		{
			Path:         baseRoute + "/{id}/update",
			Handler:      resource.Api.Update(resource, r.DB),
			Name:         fmt.Sprintf("%s-update", input.Model),
			RequiresAuth: true,
			Method:       http.MethodPut,
		},
		{
			Path:         baseRoute + "/{id}/delete",
			Handler:      resource.Api.Delete(resource, r.DB),
			Name:         fmt.Sprintf("%s-delete", input.Model),
			RequiresAuth: true,
			Method:       http.MethodDelete,
		},
	}

	resource.Routes = append(resource.Routes, routes...)

	return resource, nil
}

func (r *ResourceManager) GetRoutes() []server.Route {
	routes := make([]server.Route, 0)

	for _, v := range r.Resources {
		resource := v.(*Resource)
		routes = append(routes, resource.Routes...)
	}

	return routes
}
