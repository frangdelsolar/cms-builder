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
	model      interface{}       // model where data will be stored
	pagination *Pagination       // pagination object
	parameters RequestParameters // request parameters
	instanceId string            // instance id
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
	return db.Find(input.model, query, input.pagination, app.permissions, input.parameters), nil
}

var DefaultDetail ApiFunction = func(input *ApiInput, db *Database, app *App) (*gorm.DB, error) {
	result := db.FindById(input.instanceId, input.model, app.permissions, input.parameters)
	return result, nil
}

var DefaultCreate ApiFunction = func(input *ApiInput, db *Database, app *App) (*gorm.DB, error) {
	result := db.Create(input.model)
	return result, nil
}

var DefaultUpdate ApiFunction = func(input *ApiInput, db *Database, app *App) (*gorm.DB, error) {
	result := db.Save(input.model)
	return result, nil
}

var DefaultDelete ApiFunction = func(input *ApiInput, db *Database, app *App) (*gorm.DB, error) {
	result := db.Delete(input.model)
	return result, nil
}

type App struct {
	model           interface{}       // The model struct
	skipUserBinding bool              // Means that theres a CreatedBy field in the model that will be used for filtering the database query to only include records created by the user
	admin           *Admin            // The admin instance
	validators      ValidatorsMap     // A map of field names to validation functions
	permissions     RolePermissionMap // Key is Role name, value is permission
	api             API               // The API struct
}

// Name returns the name of the model as a string, lowercased and without the package name.
func (a *App) Name() string {
	return GetStructName(a.model)
}

// PluralName returns the plural form of the name of the model as a string.
func (a *App) PluralName() string {
	return Pluralize(a.Name())
}

// RegisterValidator registers a validator function for the given field name.
//
// Parameters:
//   - fieldName: The name of the field to register the validator for.
//     Name should match what the json schema expects. Otherwise, the validator will not be running against it.
//   - validator: The validator function to register.
//
// Returns:
// - nothing
func (a *App) RegisterValidator(fieldName FieldName, validators ValidatorsList) {
	lowerFieldName := strings.ToLower(string(fieldName))

	a.validators[lowerFieldName] = append(a.validators[lowerFieldName], validators...)
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

		requestedBy := getRequestUserId(r, a)
		params := createRequestParameters(r, requestedBy)

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
			model:      instances,
			pagination: pagination,
			parameters: params,
		}

		result, err := a.api.List(&listInput, db, a)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		if result.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, result.Error.Error())
			return
		}

		SendJsonResponseWithPagination(w, http.StatusOK, listInput.model, a.Name()+" list", listInput.pagination)
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

		requestedBy := getRequestUserId(r, a)
		params := createRequestParameters(r, requestedBy)

		// Create a new instance of the model
		instance := createInstanceForUndeterminedType(a.model)

		detailInput := ApiInput{
			model:      instance,
			parameters: params,
			instanceId: getUrlParam("id", r),
		}

		result, err := a.api.Detail(&detailInput, db, a)
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

		requestedBy := getRequestUserId(r, a)
		params := createRequestParameters(r, requestedBy)

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
		bodyBytes, err = appendUserDataToRequestBody(bodyBytes, requestedBy, true)
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
			model:      instance,
			parameters: params,
		}

		result, err := a.api.Create(&createInput, db, a)
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
		requestedBy := getRequestUserId(r, a)
		params := createRequestParameters(r, requestedBy)

		// Create a new instance of the model
		instance := createInstanceForUndeterminedType(a.model)

		apiInput := ApiInput{
			model:      instance,
			parameters: params,
			instanceId: getUrlParam("id", r),
		}

		result, err := a.api.Detail(&apiInput, db, a)
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

		bodyBytes, err = appendUserDataToRequestBody(bodyBytes, requestedBy, false)
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
		result, err = a.api.Update(&apiInput, db, a)
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

		requestedBy := getRequestUserId(r, a)
		params := createRequestParameters(r, requestedBy)

		// Create a new instance of the model
		instance := createInstanceForUndeterminedType(a.model)

		apiInput := ApiInput{
			model:      instance,
			parameters: params,
			instanceId: getUrlParam("id", r),
		}

		result, err := a.api.Detail(&apiInput, db, a)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		if result.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, result.Error.Error())
			return
		}

		// Delete the record by ID
		result, err = a.api.Delete(&apiInput, db, a)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}
		if result.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, result.Error.Error())
			return
		}

		log.Info().Msgf("Deleted %s with ID %s", a.Name(), apiInput.instanceId)

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

// getRequestUserId validates the access token in the Authorization header of the request.
//
// The function first retrieves the access token from the request header, then verifies it
// by calling VerifyUser on the App's admin instance. If the verification fails, it returns
// an empty string. Otherwise, it returns the ID of the verified user as a string.
func getRequestUserId(r *http.Request, a *App) string {
	accessToken := GetAccessTokenFromRequest(r)
	user, err := a.admin.builder.VerifyUser(accessToken)
	if err != nil {
		return ""
	}
	return user.GetIDString()
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
