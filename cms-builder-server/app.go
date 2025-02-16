package builder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// these are the keys that will be filtered out of the request body
var filterKeys = map[string]bool{
	"id":            true,
	"createdBy":     true,
	"created_by":    true,
	"createdById":   true,
	"created_by_id": true,
	"updatedBy":     true,
	"updated_by":    true,
	"updatedById":   true,
	"updated_by_id": true,
	"deletedBy":     true,
	"deleted_by":    true,
	"deletedById":   true,
	"deleted_by_id": true,
}

type FieldName string

func (f FieldName) S() string {
	return string(f)
}

type ApiFunction func(a *App, db *Database) http.HandlerFunc

type ApiHandlers struct {
	List   ApiFunction // List is a function that takes an ApiInput, a *Database and an *App and returns a *gorm.DB will be called on GET endpoints (e.g. /api/users)
	Detail ApiFunction // Detail is a function that takes an ApiInput, a *Database and an *App and returns a *gorm.DB will be called on GET endpoints (e.g. /api/users/{id})
	Create ApiFunction // Create is a function that takes an ApiInput, a *Database and an *App and returns a *gorm.DB will be called on POST endpoints (e.g. /api/users/new)
	Update ApiFunction // Update is a function that takes an ApiInput, a *Database and an *App and returns a *gorm.DB will be called on PUT endpoints (e.g. /api/users/{id}/update)
	Delete ApiFunction // Delete is a function that takes an ApiInput, a *Database and an *App and returns a *gorm.DB will be called on DELETE endpoints (e.g. /api/users/{id}/delete)
}

var DefaultListHandler ApiFunction = func(a *App, db *Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := ValidateRequestMethod(r, http.MethodGet)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		params := FormatRequestParameters(r, a.Admin.Builder)
		isAllowed := a.Permissions.HasPermission(params.Roles, OperationRead)
		if !isAllowed {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to read this resource")
			return
		}

		limit, err := strconv.Atoi(GetQueryParam("limit", r))
		if err != nil {
			log.Error().Err(err).Msgf("Error converting limit")
			limit = 10
		}

		page, err := strconv.Atoi(GetQueryParam("page", r))
		if err != nil {
			log.Error().Err(err).Msgf("Error converting page")
			page = 1
		}

		orderParam := GetQueryParam("order", r)
		order, err := a.ValidateOrderParam(orderParam)
		if err != nil {
			log.Error().Err(err).Msgf("Error validating order")
			log.Warn().Msg("Using default order")
		}

		// Create slice to store the model instances.
		instances, err := CreateSliceForUndeterminedType(a.Model)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		pagination := &Pagination{
			Total: 0,
			Page:  page,
			Limit: limit,
		}
		query := ""

		if a.SkipUserBinding {
			// Admin
			for _, role := range params.Roles {
				if role == AdminRole {
					db.Find(instances, query, pagination, order)
					SendJsonResponseWithPagination(w, http.StatusOK, instances, a.Name()+" list", pagination)
					return
				}
			}

			response := db.Find(instances, query, pagination, order)
			if response.Error != nil {
				log.Error().Err(response.Error).Msgf("Error finding instances")
				SendJsonResponse(w, http.StatusInternalServerError, nil, response.Error.Error())
				return
			}

		} else {
			// Admin
			for _, role := range params.Roles {
				if role == AdminRole {
					db.Find(instances, "", pagination, order)
					SendJsonResponseWithPagination(w, http.StatusOK, instances, a.Name()+" list", pagination)
					return
				}
			}

			query = "created_by_id = '" + params.RequestedById + "'"
			res := db.Find(instances, query, pagination, order)
			if res.Error != nil {
				SendJsonResponse(w, http.StatusInternalServerError, nil, res.Error.Error())
				return
			}
		}

		SendJsonResponseWithPagination(w, http.StatusOK, instances, a.Name()+" list", pagination)
	}
}

var DefaultDetailHandler ApiFunction = func(a *App, db *Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := ValidateRequestMethod(r, http.MethodGet)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, err, err.Error())
			return
		}

		params := FormatRequestParameters(r, a.Admin.Builder)
		isAllowed := a.Permissions.HasPermission(params.Roles, OperationRead)
		if !isAllowed {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to read this resource")
			return
		}

		// Create a new instance of the model
		instanceId := GetUrlParam("id", r)
		var instance interface{}
		if a.SkipUserBinding {
			instance = CreateInstanceForUndeterminedType(a.Model)

			for _, role := range params.Roles {
				if role == AdminRole {
					db.FindById(instanceId, instance, "")
				}
			}

			query := "id = '" + params.RequestedById + "'"
			db.FindById(instanceId, instance, query)
		} else {
			instance, err = GetInstanceIfAuthorized(a.Model, a.SkipUserBinding, instanceId, db, &params)
			if err != nil {
				SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
				return
			}
		}

		if instance == nil {
			SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
			return
		}

		SendJsonResponse(w, http.StatusOK, instance, a.Name()+" detail")
	}
}

var DefaultCreateHandler ApiFunction = func(a *App, db *Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := ValidateRequestMethod(r, http.MethodPost)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		params := FormatRequestParameters(r, a.Admin.Builder)
		isAllowed := a.Permissions.HasPermission(params.Roles, OperationCreate)
		if !isAllowed {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to create this resource")
			return
		}

		// Create a new instance of the model and parse the request body
		body, err := FormatRequestBody(r, filterKeys)
		if err != nil {
			SendJsonResponse(w, http.StatusBadRequest, nil, err.Error())
			return
		}

		body["CreatedByID"] = params.User.ID
		body["UpdatedByID"] = params.User.ID

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			log.Error().Err(err).Msg("Error marshalling request body")
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		instance := CreateInstanceForUndeterminedType(a.Model)
		err = json.Unmarshal(bodyBytes, &instance)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// Run validations
		validationErrors := a.Validate(instance)
		if len(validationErrors.Errors) > 0 {
			SendJsonResponse(w, http.StatusBadRequest, validationErrors, "Validation failed")
			return
		}

		res := db.Create(instance, params.User)
		if res.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, res.Error.Error())
			return
		}

		SendJsonResponse(w, http.StatusCreated, &instance, a.Name()+" created")
	}
}

var DefaultUpdateHandler ApiFunction = func(a *App, db *Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := ValidateRequestMethod(r, http.MethodPut)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		params := FormatRequestParameters(r, a.Admin.Builder)

		isAllowed := a.Permissions.HasPermission(params.Roles, OperationUpdate)
		if !isAllowed {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to update this resource")
			return
		}

		body, err := FormatRequestBody(r, filterKeys)
		if err != nil {
			SendJsonResponse(w, http.StatusBadRequest, nil, err.Error())
			return
		}

		body["UpdatedByID"] = params.User.ID

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// Create a new instance of the model
		instanceId := GetUrlParam("id", r)
		instance, err := GetInstanceIfAuthorized(a.Model, a.SkipUserBinding, instanceId, db, &params)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}
		if instance == nil {
			SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
			return
		}

		err = json.Unmarshal(bodyBytes, instance)
		if err != nil {
			log.Error().Err(err).Msg("Error unmarshalling request body")
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// Run validations
		validationErrors := a.Validate(instance)
		if len(validationErrors.Errors) > 0 {
			response, err := json.Marshal(validationErrors)

			if err != nil {
				SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
				return
			}

			SendJsonResponse(w, http.StatusBadRequest, response, "Validation failed")
			return
		}

		// Update the record in the database
		res := db.Save(instance, params.User)
		if res.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, res.Error.Error())
			return
		}

		SendJsonResponse(w, http.StatusOK, instance, a.Name()+" updated")
	}
}

var DefaultDeleteHandler ApiFunction = func(a *App, db *Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := ValidateRequestMethod(r, http.MethodDelete)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		params := FormatRequestParameters(r, a.Admin.Builder)
		isAllowed := a.Permissions.HasPermission(params.Roles, OperationDelete)
		if !isAllowed {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to delete this resource")
			return
		}

		instanceId := GetUrlParam("id", r)

		instance, err := GetInstanceIfAuthorized(a.Model, a.SkipUserBinding, instanceId, db, &params)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}
		if instance == nil {
			SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
			return
		}

		res := db.Delete(instance, params.User)
		if res.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, res.Error.Error())
			return
		}

		// Send a 204 No Content response
		SendJsonResponse(w, http.StatusOK, nil, a.Name()+" deleted")
	}
}

// GetInstanceIfAuthorized returns an instance of the given model if the user is authorized to access it.
//
// If the user has the AdminRole, the function returns the instance with the given ID without any additional
// filtering.
//
// If the user does not have the AdminRole, the function returns the instance with the given ID if and only
// if the "created_by_id" field of the instance matches the RequestedById parameter.
//
// If the user is not authorized to access the instance, the function returns nil.
func GetInstanceIfAuthorized(model interface{}, skipUserBinding bool, instanceId string, db *Database, params *RequestParameters) (interface{}, error) {
	var res *gorm.DB
	instance := CreateInstanceForUndeterminedType(model)

	for _, role := range params.Roles {
		if role == AdminRole {
			res = db.FindById(instanceId, instance, "")
			if res.Error != nil {
				return nil, res.Error
			}
			return instance, nil
		}
	}

	q := ""
	if !skipUserBinding {
		q = "created_by_id = '" + params.RequestedById + "'"
	}

	res = db.FindById(instanceId, instance, q)
	if res.Error != nil {
		return nil, res.Error
	}

	return instance, nil
}

type App struct {
	Model           interface{}       // The model struct
	SkipUserBinding bool              // Means that theres a CreatedBy field in the model that will be used for filtering the database query to only include records created by the user
	Admin           *Admin            // The admin instance
	Validators      ValidatorsMap     // A map of field names to validation functions
	Permissions     RolePermissionMap // Key is Role name, value is permission
	Api             *ApiHandlers      // The API struct
}

// Name returns the name of the model as a string, lowercased and without the package name.
func (a *App) Name() string {
	return GetStructName(a.Model)
}

// PluralName returns the plural form of the name of the model as a string.
func (a *App) PluralName() string {
	return Pluralize(a.Name())
}

// KebabName returns the kebab-case form of the model's name.
func (a *App) KebabName() string {
	return KebabCase(a.Name())
}

// KebabPluralName returns the kebab-case form of the plural form of the model's name.
func (a *App) KebabPluralName() string {
	return KebabCase(a.PluralName())
}

// SnakeName returns the snake_case form of the model's name.
func (a *App) SnakeName() string {
	return SnakeCase(a.Name())
}

// SnakePluralName returns the snake_case form of the plural form of the model's name.
func (a *App) SnakePluralName() string {
	return SnakeCase(a.PluralName())
}

// FieldExists returns true if the given field name exists in the model, false otherwise.
// It uses the JSON representation of the model to check if the field exists.
func (a App) FieldExists(fieldName string) bool {
	fieldNameLower := strings.ToLower(string(fieldName))

	jsonData, err := JsonifyInterface(a.Model)
	if err != nil {
		return false
	}

	// Check if the field exists in the model's JSON representation
	for k := range jsonData {
		if strings.ToLower(k) == fieldNameLower {
			return true
		}
	}

	return false
}

// RegisterValidator registers a list of validators for a specific field in the model.
//
// Parameters:
// - fieldName: the name of the field to register the validators for.
// - validators: a list of validators to be registered for the specified field.
//
// Returns:
// - error: an error if the field is not found in the model.
func (a *App) RegisterValidator(fieldName FieldName, validators ValidatorsList) error {
	fieldNameLower := strings.ToLower(string(fieldName))

	fieldExists := a.FieldExists(fieldNameLower)

	// If the field is not found, return an error
	if !fieldExists {
		return fmt.Errorf("field %s not found in model", fieldName)
	}
	// Append the provided validators to the existing list of validators for that field
	a.Validators[fieldNameLower] = append(a.Validators[fieldNameLower], validators...)

	return nil
}

// GetValidatorForField returns the validator function associated with the given field name.
//
// If no validator is associated with the given field name, it returns nil.
//
// Parameters:
// - fieldName: the name of the field to retrieve the validator for.
//
// Returns:
// - FieldValidationFunc: the validator function associated with the given field name, or nil if none is associated.
func (a *App) GetValidatorsForField(fieldName FieldName) ValidatorsList {

	lowerFieldName := strings.ToLower(string(fieldName))
	validators, ok := a.Validators[lowerFieldName]
	if !ok {
		return nil
	}

	return validators
}

// Validate validates the given instance using all the registered validators.
//
// It returns a ValidationResult which contains a slice of FieldValidationError.
// If the slice is empty, it means that the instance is valid. Otherwise, it
// contains the errors that were found during the validation process.
//
// Parameters:
// - instance: the instance to be validated.
//
// Returns:
// - ValidationResult: a ValidationResult which contains a slice of FieldValidationError.
func (a *App) Validate(instance interface{}) ValidationResult {

	// TODO: Find a way to validate nested structs
	// I think I could have a general object that has paths like system_data.id or upload.file_data.id as keys and validate them individually

	errors := ValidationResult{
		Errors: make([]ValidationError, 0),
	}

	jsonData, err := JsonifyInterface(instance)
	if err != nil {
		log.Error().Err(err).Msg("Error converting instance to JSON")
		return errors
	}

	for key := range jsonData {
		validators := a.GetValidatorsForField(FieldName(key))

		for _, validator := range validators {
			output := NewFieldValidationError(key)
			validationResult := validator(key, jsonData, &output)
			if validationResult.Error != "" {
				errors.Errors = append(errors.Errors, *validationResult)
			}
		}
	}

	if len(errors.Errors) == 0 {
		return ValidationResult{}
	}

	return errors
}

// ValidateOrderParam validates the given orderParam string and returns a valid order string for the given model.
//
// Parameters:
// - orderParam: the orderParam string to be validated.
//
// Returns:
// - string: a valid order string for the given model, or an empty string if the orderParam is empty.
// - error: an error if one of the fields in the orderParam is not found in the model.
func (a *App) ValidateOrderParam(orderParam string) (string, error) {
	if orderParam == "" {
		return "", nil
	}

	order := ""
	fields := strings.Split(orderParam, ",")
	for _, field := range fields {

		desc := strings.HasPrefix(field, "-")

		if desc {
			field = strings.TrimPrefix(field, "-")
		}

		field = SnakeCase(field)

		if desc {
			order += field + " desc,"
		} else {
			order += field + ","
		}
	}

	order = strings.TrimSuffix(order, ",")

	return order, nil
}

// JsonifyInterface takes an interface{} and attempts to convert it to a map[string]interface{}
// via JSON marshaling and unmarshaling. If the conversion fails, it returns an error.
func JsonifyInterface(instance interface{}) (map[string]interface{}, error) {
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

/*
	API HANDLERS
*/

// ApiList returns a handler function that responds to GET requests on the
// list endpoint, e.g. /api/users.
//
// The handler function will return a JSON response containing a list of records.
//
// It will also handle errors and return a 500 Internal Server Error if an error
// occurs during the retrieval of records.
func (a *App) ApiList(db *Database) http.HandlerFunc {
	return a.Api.List(a, db)
}

// ApiDetail returns a handler function that responds to GET requests on the
// details endpoint, e.g. /api/users/{id}.
//
// The handler function will return a JSON response containing the record
// matching the given ID.
//
// It will also handle errors and return a 404 Not Found if the error is a
// gorm.ErrRecordNotFound, or a 500 Internal Server Error if the error is
// not a gorm.ErrRecordNotFound.
func (a *App) ApiDetail(db *Database) http.HandlerFunc {
	return a.Api.Detail(a, db)
}

// ApiCreate returns a handler function that responds to POST requests on the
// list endpoint, e.g. /api/users/new.
//
// The handler function will create a new record in the database and return a
// JSON response containing the newly created record.
//
// It will also handle errors and return a 500 Internal Server Error if an error
// occurs during the creation of the record.
func (a *App) ApiCreate(db *Database) http.HandlerFunc {
	return a.Api.Create(a, db)
}

// ApiUpdate returns a handler function that responds to PUT requests on the
// details endpoint, e.g. /api/users/{id}/update.
//
// The handler function will update the record in the database and return a
// JSON response containing the updated record.
//
// It will also handle errors and return a 500 Internal Server Error if an error
// occurs during the update of the record.
func (a *App) ApiUpdate(db *Database) http.HandlerFunc {
	return a.Api.Update(a, db)
}

// ApiDelete returns a handler function that responds to DELETE requests on the
// delete endpoint, e.g. /api/users/{id}/delete.
//
// The handler function will delete the record in the database and return a
// JSON response containing the deleted record.
//
// It will also handle errors and return a 404 Not Found if the error is a
// gorm.ErrRecordNotFound, or a 500 Internal Server Error if the error is
// not a gorm.ErrRecordNotFound.
func (a *App) ApiDelete(db *Database) http.HandlerFunc {
	return a.Api.Delete(a, db)
}

/*
	REFLECT HELPERS
*/

// CreateInstanceForUndeterminedType creates a new instance of the given model type.
//
// It takes a single argument, which can be a struct, a pointer to a struct, or
// a slice of a struct. It returns a new instance of the given type and does not
// report any errors.
func CreateInstanceForUndeterminedType(model interface{}) interface{} {
	instanceType := reflect.TypeOf(model)
	if instanceType.Kind() == reflect.Ptr {
		instanceType = instanceType.Elem()
	}
	return reflect.New(instanceType).Interface()
}

// CreateSliceForUndeterminedType creates a new slice for the given model type.
//
// It takes a single argument, which can be a struct, a pointer to a struct, or
// a slice of a struct. It returns a new slice of the given type and an error if
// the input is not a valid model type.
//
// The function is used by the admin API to create slices for the different
// models that are registered with the admin.
func CreateSliceForUndeterminedType(model interface{}) (interface{}, error) {
	modelType := reflect.TypeOf(model)

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
