package resourcemanager

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

type ApiFunction func(a *Resource, db *database.Database) http.HandlerFunc

type ApiHandlers struct {
	List   ApiFunction // List is a function that takes an ApiInput, a *Database and an *App and returns a *gorm.DB will be called on GET endpoints (e.g. /api/users)
	Detail ApiFunction // Detail is a function that takes an ApiInput, a *Database and an *App and returns a *gorm.DB will be called on GET endpoints (e.g. /api/users/{id})
	Create ApiFunction // Create is a function that takes an ApiInput, a *Database and an *App and returns a *gorm.DB will be called on POST endpoints (e.g. /api/users/new)
	Update ApiFunction // Update is a function that takes an ApiInput, a *Database and an *App and returns a *gorm.DB will be called on PUT endpoints (e.g. /api/users/{id}/update)
	Delete ApiFunction // Delete is a function that takes an ApiInput, a *Database and an *App and returns a *gorm.DB will be called on DELETE endpoints (e.g. /api/users/{id}/delete)
}

type Resource struct {
	Model           interface{}              // The model struct
	SkipUserBinding bool                     // Means that theres a CreatedBy field in the model that will be used for filtering the database query to only include records created by the user
	Validators      ValidatorsMap            // A map of field names to validation functions
	Permissions     server.RolePermissionMap // Key is Role name, value is permission
	Api             *ApiHandlers             // The API struct
}

func (a *Resource) GetSlice() (interface{}, error) {
	modelType := reflect.TypeOf(a.Model)

	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	if modelType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model must be a struct or a pointer to a struct")
	}

	sliceType := reflect.SliceOf(modelType)
	entities := reflect.New(sliceType).Interface()

	return entities, nil
}

func (a *Resource) GetInstance() interface{} {
	instanceType := reflect.TypeOf(a.Model)
	if instanceType.Kind() == reflect.Ptr {
		instanceType = instanceType.Elem()
	}
	return reflect.New(instanceType).Interface()
}

func (a *Resource) GetName() (string, error) {
	modelType := reflect.TypeOf(a.Model)

	if modelType == nil {
		return "", fmt.Errorf("model cannot be nil")
	}

	// If it's a pointer, get the element type
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// Ensure it's a struct before returning its name
	if modelType.Kind() != reflect.Struct {
		return "", fmt.Errorf("model must be a struct or a pointer to a struct")
	}

	return modelType.Name(), nil
}
