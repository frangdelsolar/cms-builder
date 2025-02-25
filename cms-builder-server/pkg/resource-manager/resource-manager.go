package resourcemanager

import (
	"fmt"

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

	routes := input.Routes
	if routes == nil {
		routes = make([]server.Route, 0)
	}

	resource := &Resource{
		Model:           input.Model,
		Api:             handlers,
		SkipUserBinding: input.SkipUserBinding,
		Validators:      validators,
		Permissions:     permissions,
		Routes:          routes,
	}

	err := r.Register(resource)
	if err != nil {
		return nil, err
	}

	return resource, nil
}
