package resourcemanager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
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
	Routes          []server.Route           // Other routes that are not the default api routes
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

func (a *Resource) GetOne() interface{} {
	instanceType := reflect.TypeOf(a.Model)
	if instanceType.Kind() == reflect.Ptr {
		instanceType = instanceType.Elem()
	}
	return reflect.New(instanceType).Interface()
}

func (a *Resource) GetName() (string, error) {
	return utils.GetInterfaceName(a.Model)
}

func (a *Resource) GetKeys() []string {

	if a.Model == nil {
		return nil
	}

	modelType := reflect.TypeOf(a.Model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	if modelType.Kind() != reflect.Struct {
		return nil
	}

	keys := make([]string, 0)
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		if field.Tag.Get("json") != "" {
			keys = append(keys, field.Tag.Get("json"))
		}
	}

	return keys
}

func (a *Resource) GetValidatorsForField(fieldName string) ValidatorsList {

	lowerFieldName := strings.ToLower(string(fieldName))
	validators, ok := a.Validators[lowerFieldName]
	if !ok {
		return nil
	}

	return validators
}

func InterfaceToMap(instance interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(instance)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (a *Resource) Validate(instance interface{}) ValidationResult {

	errors := ValidationResult{
		Errors: make([]ValidationError, 0),
	}

	_, err := InterfaceToMap(instance)
	if err != nil {
		return errors
	}

	// for key := range dataMap {
	// 	validators := a.GetValidatorsForField(FieldName(key))

	// 	for _, validator := range validators {
	// 		output := NewFieldValidationError(key)
	// 		validationResult := validator(key, jsonData, &output)
	// 		if validationResult.Error != "" {
	// 			errors.Errors = append(errors.Errors, *validationResult)
	// 		}
	// 	}
	// }

	if len(errors.Errors) == 0 {
		return ValidationResult{}
	}

	return errors
}
