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

type FieldName string

func (f FieldName) S() string {
	return string(f)
}

type ApiInput struct {
	Model      interface{}       // model where data will be stored
	Pagination *Pagination       // pagination object
	Parameters RequestParameters // request parameters
	InstanceId string            // instance id
}

// ApiFunction is a function that takes an ApiInput, a *Database and an *App and returns a *gorm.DB.
// The ApiFunction is used to define the behavior of the API endpoints.
type ApiFunction func(input *ApiInput, db *Database, app *App) (*gorm.DB, error)

type API struct {
	List   ApiFunction // List is a function that takes an ApiInput, a *Database and an *App and returns a *gorm.DB will be called on GET endpoints (e.g. /api/users)
	Detail ApiFunction // Detail is a function that takes an ApiInput, a *Database and an *App and returns a *gorm.DB will be called on GET endpoints (e.g. /api/users/{id})
	Create ApiFunction // Create is a function that takes an ApiInput, a *Database and an *App and returns a *gorm.DB will be called on POST endpoints (e.g. /api/users/new)
	Update ApiFunction // Update is a function that takes an ApiInput, a *Database and an *App and returns a *gorm.DB will be called on PUT endpoints (e.g. /api/users/{id}/update)
	Delete ApiFunction // Delete is a function that takes an ApiInput, a *Database and an *App and returns a *gorm.DB will be called on DELETE endpoints (e.g. /api/users/{id}/delete)
}

var DefaultList ApiFunction = func(input *ApiInput, db *Database, app *App) (*gorm.DB, error) {
	query := ""
	role := input.Parameters.Roles[0]
	if role == VisitorRole {
		query = "created_by_id = '" + input.Parameters.RequestedById + "'"
	}

	return db.Find(input.Model, query, input.Pagination), nil
}

var DefaultDetail ApiFunction = func(input *ApiInput, db *Database, app *App) (*gorm.DB, error) {
	queryExtension := ""

	role := input.Parameters.Roles[0]
	if role == VisitorRole {
		queryExtension = "created_by_id = '" + input.Parameters.RequestedById + "'"
	}

	result := db.FindById(input.InstanceId, input.Model, queryExtension)
	return result, nil
}

var DefaultCreate ApiFunction = func(input *ApiInput, db *Database, app *App) (*gorm.DB, error) {
	result := db.Create(input.Model)
	return result, nil
}

var DefaultUpdate ApiFunction = func(input *ApiInput, db *Database, app *App) (*gorm.DB, error) {
	result := db.Save(input.Model)
	return result, nil
}

var DefaultDelete ApiFunction = func(input *ApiInput, db *Database, app *App) (*gorm.DB, error) {
	result := db.Delete(input.Model)
	return result, nil
}

type App struct {
	model           interface{}       // The model struct
	skipUserBinding bool              // Means that theres a CreatedBy field in the model that will be used for filtering the database query to only include records created by the user
	admin           *Admin            // The admin instance
	validators      ValidatorsMap     // A map of field names to validation functions
	Permissions     RolePermissionMap // Key is Role name, value is permission
	Api             API               // The API struct
}

// Name returns the name of the model as a string, lowercased and without the package name.
func (a *App) Name() string {
	return GetStructName(a.model)
}

// PluralName returns the plural form of the name of the model as a string.
func (a *App) PluralName() string {
	return Pluralize(a.Name())
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

	jsonData, err := JsonifyInterface(a.model)
	if err != nil {
		return err
	}

	// Check if the field exists in the model's JSON representation
	fieldExists := false
	for k := range jsonData {
		if strings.ToLower(k) == fieldNameLower {
			fieldExists = true
			break
		}
	}

	// If the field is not found, return an error
	if !fieldExists {
		return fmt.Errorf("field %s not found in model", fieldName)
	}

	// Append the provided validators to the existing list of validators for that field
	a.validators[fieldNameLower] = append(a.validators[fieldNameLower], validators...)

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
	validators, ok := a.validators[lowerFieldName]
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

	errors := ValidationResult{
		Errors: make([]ValidationError, 0),
	}

	jsonData, err := JsonifyInterface(instance)

	if err != nil {
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

	return errors
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

// ApiList returns a handler function that responds to GET requests on the list endpoint.
//
// The handler function retrieves the parameters `limit` and `page` from the request, and
// uses them to fetch the corresponding page of records from the database.
//
// The handler function returns a JSON response containing the records.
//
// Parameters:
//   - db: a pointer to a Database instance.
//
// Returns:
// - A HandlerFunc that responds to GET requests on the list endpoint.
func (a *App) ApiList(db *Database) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := validateRequestMethod(r, http.MethodGet)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		params := FormatRequestParameters(r, a.admin.builder)
		isAllowed := a.Permissions.HasPermission(params.Roles, OperationRead)
		if !isAllowed {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to read this resource")
			return
		}

		limit, err := strconv.Atoi(getQueryParam("limit", r))
		if err != nil {
			log.Error().Err(err).Msgf("Error converting limit")
			limit = 10
		}

		page, err := strconv.Atoi(getQueryParam("page", r))
		if err != nil {
			log.Error().Err(err).Msgf("Error converting page")
			page = 1
		}

		// Create slice to store the model instances.
		instances, err := createSliceForUndeterminedType(a.model)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		pagination := &Pagination{
			Total: 0,
			Page:  page,
			Limit: limit,
		}

		listInput := ApiInput{
			Model:      instances,
			Pagination: pagination,
			Parameters: params,
		}

		result, err := a.Api.List(&listInput, db, a)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		if result.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, result.Error.Error())
			return
		}

		SendJsonResponseWithPagination(w, http.StatusOK, listInput.Model, a.Name()+" list", listInput.Pagination)
	}
}

// ApiDetail returns a handler function that responds to GET requests on the
// details endpoint, e.g. /api/users/{id}.
//
// The handler function will return a JSON response containing the requested
// record.
//
// It will also handle errors and return a 404 Not Found if the error is a
// gorm.ErrRecordNotFound, or a 500 Internal Server Error if the error is
// not a gorm.ErrRecordNotFound.
func (a *App) ApiDetail(db *Database) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := validateRequestMethod(r, http.MethodGet)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, err, err.Error())
			return
		}

		params := FormatRequestParameters(r, a.admin.builder)
		isAllowed := a.Permissions.HasPermission(params.Roles, OperationRead)
		if !isAllowed {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to read this resource")
			return
		}

		// Create a new instance of the model
		instance := createInstanceForUndeterminedType(a.model)

		detailInput := ApiInput{
			Model:      instance,
			Parameters: params,
			InstanceId: getUrlParam("id", r),
		}

		result, err := a.Api.Detail(&detailInput, db, a)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}
		if result.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, result.Error, "Failed to get "+a.Name()+" detail")
			return
		}

		SendJsonResponse(w, http.StatusOK, instance, a.Name()+" detail")
	}
}

// apiNew returns a handler function that responds to POST requests on the
// new endpoint, e.g. /api/users/new.
//
// The handler function will return a JSON response containing the newly
// created record.
//
// It will also handle errors and return a 500 Internal Server Error if the
// error is not a gorm.ErrRecordNotFound.
func (a *App) ApiCreate(db *Database) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := validateRequestMethod(r, http.MethodPost)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		params := FormatRequestParameters(r, a.admin.builder)
		isAllowed := a.Permissions.HasPermission(params.Roles, OperationCreate)
		if !isAllowed {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to create this resource")
			return
		}

		// Will read bytes, append user data and pack the bytes again to be processed
		bodyBytes, err := readRequestBody(r)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// Make sure user is not attempting to modify system data fields
		bodyBytes, err = removeSystemDataFieldsFromRequest(bodyBytes)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// Update SystemData fields
		bodyBytes, err = appendUserDataToRequestBody(bodyBytes, params.RequestedById, true)
		if err != nil {
			SendJsonResponse(w, http.StatusUnauthorized, err, err.Error())
			return
		}

		// Create a new instance of the model
		instance := createInstanceForUndeterminedType(a.model)

		// Unmarshal the updated bytes into the instance
		err = json.Unmarshal(bodyBytes, instance)
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

		createInput := ApiInput{
			Model:      instance,
			Parameters: params,
		}

		result, err := a.Api.Create(&createInput, db, a)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// Perform database operation
		if result.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, result.Error.Error())
			return
		}

		SendJsonResponse(w, http.StatusCreated, instance, a.Name()+" created")
	}
}

// ApiUpdate returns a handler function that responds to PUT requests on the
// details endpoint, e.g. /api/users/{id}.
//
// The handler function will update the record in the database and return a
// JSON response containing the updated record.
//
// It will also handle errors and return a 404 Not Found if the error is a
// gorm.ErrRecordNotFound, or a 500 Internal Server Error if the error is
// not a gorm.ErrRecordNotFound.
func (a *App) ApiUpdate(db *Database) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := validateRequestMethod(r, http.MethodPut)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		params := FormatRequestParameters(r, a.admin.builder)
		isAllowed := a.Permissions.HasPermission(params.Roles, OperationUpdate)
		if !isAllowed {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to update this resource")
			return
		}

		// Create a new instance of the model
		instance := createInstanceForUndeterminedType(a.model)

		apiInput := ApiInput{
			Model:      instance,
			Parameters: params,
			InstanceId: getUrlParam("id", r),
		}

		result, err := a.Api.Detail(&apiInput, db, a)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		if result.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, result.Error.Error())
			return
		}

		bodyBytes, err := readRequestBody(r)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// Make sure user is not attempting to modify system data fields
		bodyBytes, err = removeSystemDataFieldsFromRequest(bodyBytes)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		bodyBytes, err = appendUserDataToRequestBody(bodyBytes, params.RequestedById, false)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// Unmarshal the updated bytes into the instance
		err = json.Unmarshal(bodyBytes, instance)
		if err != nil {
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
		result, err = a.Api.Update(&apiInput, db, a)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		if result.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, result.Error.Error())
			return
		}

		SendJsonResponse(w, http.StatusOK, instance, a.Name()+" updated")
	}
}

// ApiDelete returns a handler function that responds to DELETE requests on the
// details endpoint, e.g. /api/users/{id}.
//
// The handler function will delete the record in the database and return a
// JSON response with a 204 No Content status code.
//
// It will also handle errors and return a 404 Not Found if the error is a
// gorm.ErrRecordNotFound, or a 500 Internal Server Error if the error is
// not a gorm.ErrRecordNotFound.
func (a *App) ApiDelete(db *Database) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := validateRequestMethod(r, http.MethodDelete)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		params := FormatRequestParameters(r, a.admin.builder)
		isAllowed := a.Permissions.HasPermission(params.Roles, OperationDelete)
		if !isAllowed {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to delete this resource")
			return
		}

		// Create a new instance of the model
		instance := createInstanceForUndeterminedType(a.model)

		apiInput := ApiInput{
			Model:      instance,
			Parameters: params,
			InstanceId: getUrlParam("id", r),
		}

		result, err := a.Api.Detail(&apiInput, db, a)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		if result.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, result.Error.Error())
			return
		}

		// Delete the record by ID
		result, err = a.Api.Delete(&apiInput, db, a)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}
		if result.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, result.Error.Error())
			return
		}

		log.Info().Msgf("Deleted %s with ID %s", a.Name(), apiInput.InstanceId)

		// Send a 204 No Content response
		SendJsonResponse(w, http.StatusOK, nil, a.Name()+" deleted")
	}
}

/*
	REQUEST HELPERS
*/

// removeSystemDataFieldsFromRequest takes a JSON request body as a byte slice and returns a new byte slice with the system data fields removed.
//
// It takes a JSON request body, unmarshals it into a map[string]interface{}, removes the system data fields from the map, marshals the modified map back into JSON, and returns it as a byte slice.
//
// If the unmarshaling or marshaling fails, it returns an error.
//
// Parameters:
// - bodyBytes: the JSON request body as a byte slice
//
// Returns:
// - []byte: the modified JSON request body as a byte slice
// - error: an error if the unmarshaling or marshaling failed
func removeSystemDataFieldsFromRequest(bodyBytes []byte) ([]byte, error) {

	var data map[string]interface{}
	err := json.Unmarshal(bodyBytes, &data)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal request body: %w", err)
	}

	systemDataInstance := SystemData{}

	for _, key := range systemDataInstance.Keys() {
		delete(data, key)
	}

	modifiedBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal modified data: %w", err)
	}

	return modifiedBytes, nil
}

// unmarshalRequestBody unmarshals the given byte slice into a map[string]interface{}.
// It returns the unmarshalled map and an error if the unmarshalling fails.
func unmarshalRequestBody(data []byte) (map[string]interface{}, error) {
	var jsonData map[string]interface{}
	err := json.Unmarshal(data, &jsonData)
	return jsonData, err
}

// appendUserDataToRequestBody appends the requested_by from the request header to the request body for create and update operations.
// For create operations, the requested_by is added to the CreatedById field.
// For update operations, the requested_by is added to the UpdatedById field.
// The function returns the updated request body as a byte array, or an error if the request body cannot be unmarshalled or marshalled.

func appendUserDataToRequestBody(bytes []byte, requestedBy string, isNewRecord bool) ([]byte, error) {
	jsonData, err := unmarshalRequestBody(bytes)
	if err != nil {
		log.Error().Err(err).Msgf("Error unmarshalling request body for creation")
		return nil, err
	}

	if requestedBy == "" || requestedBy == "0" {
		log.Error().Msgf("No userId found for authorization header")
		return nil, fmt.Errorf("user not authenticated")
	}

	convertedUserId, err := strconv.ParseUint(requestedBy, 10, 64)
	if err != nil {
		log.Error().Err(err).Msgf("Error converting userId")
		return nil, err
	}

	if isNewRecord {
		jsonData["CreatedById"] = convertedUserId
	}

	jsonData["UpdatedById"] = convertedUserId

	bodyBytes, err := json.Marshal(jsonData)
	if err != nil {
		log.Error().Err(err).Msgf("Error marshalling request body for creation")
		return nil, err
	}

	return bodyBytes, err
}

/*
	REFLECT HELPERS
*/

// createInstanceForUndeterminedType creates a new instance of the given model type.
//
// It takes a single argument, which can be a struct, a pointer to a struct, or
// a slice of a struct. It returns a new instance of the given type and does not
// report any errors.
func createInstanceForUndeterminedType(model interface{}) interface{} {
	instanceType := reflect.TypeOf(model)
	if instanceType.Kind() == reflect.Ptr {
		instanceType = instanceType.Elem()
	}
	return reflect.New(instanceType).Interface()
}

// createSliceForUndeterminedType creates a new slice for the given model type.
//
// It takes a single argument, which can be a struct, a pointer to a struct, or
// a slice of a struct. It returns a new slice of the given type and an error if
// the input is not a valid model type.
//
// The function is used by the admin API to create slices for the different
// models that are registered with the admin.
func createSliceForUndeterminedType(model interface{}) (interface{}, error) {
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
