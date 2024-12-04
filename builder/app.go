package builder

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type App struct {
	model           interface{}   // The model struct
	skipUserBinding bool          // Means that theres a CreatedBy field in the model that will be used for filtering the database query to only include records created by the user
	admin           *Admin        // The admin instance
	validators      ValidatorsMap // A map of field names to validation functions
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
func (a *App) RegisterValidator(fieldName string, validators ValidatorsList) {
	lowerFieldName := strings.ToLower(fieldName)

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
func (a *App) GetValidatorsForField(fieldName string) ValidatorsList {

	lowerFieldName := strings.ToLower(fieldName)
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
		validators := a.GetValidatorsForField(key)

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

		// Retrieve parameters from the request
		param_limit := "10"

		if r.URL != nil && r.URL.Query().Get("limit") != "" {
			param_limit = r.URL.Query().Get("limit")
		}
		param_page := "1"

		if r.URL != nil && r.URL.Query().Get("page") != "" {
			param_page = r.URL.Query().Get("page")
		}

		limit, err := strconv.Atoi(param_limit)
		if err != nil {
			limit = 10
		}

		page, err := strconv.Atoi(param_page)
		if err != nil {
			page = 1
		}

		err = validateRequestMethod(r, http.MethodGet)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		userId := getRequestUserId(r, a)

		// Create slice to store the model instances.
		instances, err := createSliceForUndeterminedType(a.model)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		var result *gorm.DB
		pagination := &Pagination{
			Total: 0,
			Page:  page,
			Limit: limit,
		}

		if a.skipUserBinding {
			result = db.FindAll(instances, pagination)
		} else {
			result = db.FindAllByUserId(instances, userId, pagination)
		}

		if result.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, result.Error.Error())
			return
		}

		SendJsonResponseWithPagination(w, http.StatusOK, instances, a.Name()+" list", pagination)
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

		// Retrieve parameters from the request
		instanceId := mux.Vars(r)["id"]
		userId := getRequestUserId(r, a)

		// Create a new instance of the model
		instance := createInstanceForUndeterminedType(a.model)

		// Query the database to find the record by ID
		result := db.FindById(instanceId, instance, userId, a.skipUserBinding)
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
func (a *App) ApiNew(db *Database) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := validateRequestMethod(r, http.MethodPost)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
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
		if !a.skipUserBinding {
			bodyBytes, err = appendUserDataToRequestBody(bodyBytes, r, true, a)
			if err != nil {
				SendJsonResponse(w, http.StatusUnauthorized, err, err.Error())
				return
			}
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

		// Perform database operation
		result := db.Create(instance)
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

		// Retrieve parameters from the request
		instanceId := mux.Vars(r)["id"]
		userId := getRequestUserId(r, a)

		// Create a new instance of the model
		instance := createInstanceForUndeterminedType(a.model)

		// Query the database to find the record by ID
		result := db.FindById(instanceId, instance, userId, a.skipUserBinding)
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

		// Update SystemData fields
		if !a.skipUserBinding {
			bodyBytes, err = appendUserDataToRequestBody(bodyBytes, r, false, a)
			if err != nil {
				SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
				return
			}
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
		result = db.Save(instance)
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

		// Retrieve parameters from the request
		instanceId := mux.Vars(r)["id"]
		userId := getRequestUserId(r, a)

		// Create a new instance of the model
		instance := createInstanceForUndeterminedType(a.model)

		// Query the database to find the record by ID
		dbResponse := db.FindById(instanceId, instance, userId, a.skipUserBinding)
		if dbResponse.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, dbResponse.Error.Error())
			return
		}

		// Delete the record by ID
		result := db.Delete(instance)
		if result.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, result.Error.Error())
			return
		}

		log.Info().Msgf("Deleted %s with ID %s", a.Name(), instanceId)

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

// validateRequestMethod returns an error if the request method does not match the given
// method string. The error message will include the actual request method.
func validateRequestMethod(r *http.Request, method string) error {
	if r.Method != method {
		return fmt.Errorf("invalid request method: %s", r.Method)
	}
	return nil
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

// readRequestBody reads the entire request body and returns the contents as a byte slice.
// It defers closing the request body until the function returns.
// It returns an error if there is a problem reading the request body.
func readRequestBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	return io.ReadAll(r.Body)
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
func appendUserDataToRequestBody(bytes []byte, r *http.Request, isNewRecord bool, a *App) ([]byte, error) {
	jsonData, err := unmarshalRequestBody(bytes)
	if err != nil {
		log.Error().Err(err).Msgf("Error unmarshalling request body for creation")
		return nil, err
	}

	// Retrieve requested_by from request header
	userId := getRequestUserId(r, a)

	if userId == "" || userId == "0" {
		log.Error().Msgf("No userId found for authorization header")
		return nil, fmt.Errorf("user not authenticated")
	}

	convertedUserId, err := strconv.ParseUint(userId, 10, 64)
	if err != nil {
		log.Error().Err(err).Msgf("Error converting userId")
		return nil, err
	}

	if isNewRecord {
		jsonData["CreatedById"] = convertedUserId
	}

	jsonData["UpdatedById"] = convertedUserId

	log.Info().Interface("body", jsonData).Msg("Appending user data to request body")

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
