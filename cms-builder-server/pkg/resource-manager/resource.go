package resourcemanager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

// ApiFunction defines the signature for API handler functions.
type ApiFunction func(resource *Resource, db *database.Database) http.HandlerFunc

// ApiHandlers holds the handlers for various API operations.
type ApiHandlers struct {
	List   ApiFunction
	Detail ApiFunction
	Create ApiFunction
	Update ApiFunction
	Delete ApiFunction
	Schema func(resource *Resource) http.HandlerFunc
}

type ResourceNames struct {
	Singular      string `json:"singularName"`
	Plural        string `json:"pluralName"`
	SnakeSingular string `json:"snakeName"`
	SnakePlural   string `json:"snakePluralName"`
	KebabSingular string `json:"kebabName"`
	KebabPlural   string `json:"kebabPluralName"`
}

// Resource represents a resource in the system, including its model, validators, and routes.
type Resource struct {
	Model           interface{}              // The model struct
	SkipUserBinding bool                     // Whether to skip user binding for this resource
	Validators      ValidatorsMap            // Map of field validators
	Permissions     server.RolePermissionMap // Role-based permissions
	Api             *ApiHandlers             // API handlers
	Routes          map[string]server.Route  // Custom routes for this resource
	ResourceNames   ResourceNames            `json:"resourceNames"` // Resource names
	JsonSchema      json.RawMessage          `json:"jsonSchema"`    // JSON schema
	FieldNames      map[string]string        `json:"fieldNames"`    // TODO: Map of field names
}

// GetSlice returns a new slice of the resource's model type.
func (r *Resource) GetSlice() (interface{}, error) {
	modelType := reflect.TypeOf(r.Model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	if modelType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model must be a struct or a pointer to a struct")
	}
	sliceType := reflect.SliceOf(modelType)
	return reflect.New(sliceType).Interface(), nil
}

// GetOne returns a new instance of the resource's model.
func (r *Resource) GetOne() interface{} {
	instanceType := reflect.TypeOf(r.Model)
	if instanceType.Kind() == reflect.Ptr {
		instanceType = instanceType.Elem()
	}
	return reflect.New(instanceType).Interface()
}

// GetKeys returns a list of field names in the resource's model.
func (r *Resource) GetKeys() []string {
	if r.Model == nil {
		return nil
	}

	modelType := reflect.TypeOf(r.Model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	if modelType.Kind() != reflect.Struct {
		return nil
	}

	keys := make([]string, 0)
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		keys = append(keys, field.Name) // Use the exact field name
	}
	return keys
}

// HasField checks if the resource's model contains a field with the given name.
func (r *Resource) HasField(fieldName string) bool {
	modelType := reflect.TypeOf(r.Model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	if modelType.Kind() != reflect.Struct {
		return false
	}
	_, ok := modelType.FieldByName(fieldName)
	return ok
}

// AddValidator adds a validator for a specific field in the resource's model.
func (r *Resource) AddValidator(fieldName string, validator Validator) error {
	if r.Validators == nil {
		r.Validators = make(ValidatorsMap)
	}

	if !r.HasField(fieldName) {
		return fmt.Errorf("field %s not found in model", fieldName)
	}

	if r.Validators[fieldName] == nil {
		r.Validators[fieldName] = make(ValidatorsList, 0)
	}

	r.Validators[fieldName] = append(r.Validators[fieldName], validator)
	return nil
}

// GetFieldValidators retrieves the validators for a specific field.
func (r *Resource) GetFieldValidators(fieldName string) (ValidatorsList, error) {
	validators, ok := r.Validators[fieldName]
	if !ok {
		return nil, fmt.Errorf("field %s not found in model", fieldName)
	}
	return validators, nil
}

// getFieldNameFromJSONTag resolves the struct field name from the JSON tag.
func (r *Resource) getFieldNameFromJSONTag(jsonTag string) (string, error) {
	modelType := reflect.TypeOf(r.Model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		tag := field.Tag.Get("json")
		if tag == jsonTag {
			return field.Name, nil
		}
	}

	return "", fmt.Errorf("field with JSON tag %s not found in model", jsonTag)
}

// InterfaceToMap converts an interface to a map using JSON marshaling.
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

// Validate validates the given instance against the resource's validators.
func (r *Resource) Validate(instance interface{}, log *loggerTypes.Logger) ValidationResult {
	errors := ValidationResult{
		Errors: make([]ValidationError, 0),
	}

	jsonData, err := InterfaceToMap(instance)
	if err != nil {
		log.Error().Err(err).Msg("Error converting instance to JSON")
		return errors
	}

	for jsonTag := range jsonData {
		fieldName, err := r.getFieldNameFromJSONTag(jsonTag)
		if err != nil {
			continue
		}

		if !r.HasField(fieldName) {
			continue // Field not found in model
		}

		validators, ok := r.Validators[fieldName]
		if !ok {
			continue // No validators for this field
		}

		for _, validator := range validators {
			output := NewFieldValidationError(jsonTag)
			validationResult := validator(jsonTag, jsonData, &output)
			if validationResult.Error != "" {
				errors.Errors = append(errors.Errors, *validationResult)
			}
		}
	}

	if len(errors.Errors) == 0 {
		return ValidationResult{
			Errors: make([]ValidationError, 0),
		}
	}
	return errors
}

// AddRoute adds a custom route to the resource.
func (r *Resource) AddRoute(route server.Route) error {
	if r.Routes == nil {
		r.Routes = make(map[string]server.Route)
	}

	if _, exists := r.Routes[route.Name]; exists {
		return fmt.Errorf("route with name %s already exists", route.Name)
	}

	r.Routes[route.Name] = route
	return nil
}
