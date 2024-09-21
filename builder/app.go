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

// FieldValidationFunc is a function that validates a field value.
type FieldValidationFunc func(value interface{}) FieldValidationError

type FieldValidationError struct {
	Field string // The name of the field that failed validation
	Error string // The error message
}

type ValidationResult struct {
	Errors []FieldValidationError // A list of field validation errors
}

type App struct {
	model           interface{}                    // The model struct
	skipUserBinding bool                           // Means that theres a CreatedBy field in the model that will be used for filtering the database query to only include records created by the user
	admin           *Admin                         // The admin instance
	validators      map[string]FieldValidationFunc // A map of field names to validation functions
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
// - fieldName: The name of the field to register the validator for.
// - validator: The validator function to register.
//
// Returns:
// - nothing
func (a *App) RegisterValidator(fieldName string, validator FieldValidationFunc) {
	lowerFieldName := strings.ToLower(fieldName)
	a.validators[lowerFieldName] = validator
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
func (a *App) GetValidatorForField(fieldName string) FieldValidationFunc {

	lowerFieldName := strings.ToLower(fieldName)
	validator, ok := a.validators[lowerFieldName]
	if !ok {
		return nil
	}

	return validator
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
		Errors: make([]FieldValidationError, 0),
	}

	jsonData, err := jsonifyInterface(instance)

	if err != nil {
		log.Error().Err(err).Msg("Failed to jsonify interface")
		return errors
	}

	for key, value := range jsonData {
		validator := a.GetValidatorForField(key)
		if validator == nil {
			continue
		}
		validationResult := validator(value)
		if validationResult != (FieldValidationError{}) {
			errors.Errors = append(errors.Errors, validationResult)
		}
	}

	return errors
}

// jsonifyInterface takes an interface{} and attempts to convert it to a map[string]interface{}
// via JSON marshaling and unmarshaling. If the conversion fails, it returns an error.
func jsonifyInterface(instance interface{}) (map[string]interface{}, error) {
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

// NewFieldValidationError creates a new FieldValidationError with the given field name and an empty error string.
func NewFieldValidationError(fieldName string) FieldValidationError {
	return FieldValidationError{
		Field: fieldName,
		Error: "",
	}
}

// NewIdValidator returns a FieldValidationFunc that checks if the given ID is valid and belongs to the entity
// with the given name. It also checks if the ID is not empty and if it exists in the database.
// If the ID is valid, it returns an empty FieldValidationError. Otherwise, it returns a FieldValidationError with a
// descriptive error message.
func (a *App) NewIdValidator(otherAppName string, fieldName string, id string, requestedBy string) FieldValidationFunc {
	return func(id interface{}) FieldValidationError {
		validationError := NewFieldValidationError(fieldName)

		if id == "" {
			validationError.Error = fieldName + " cannot be empty"
			return validationError
		}

		otherApp, err := a.admin.GetApp(otherAppName)
		if err != nil {
			validationError.Error = fmt.Sprintf("error getting %s: %s", otherAppName, err)
			return validationError
		}

		// Create a new instance of the model
		instance := createInstanceForUndeterminedType(otherApp.model)

		// Query the database to find the record by ID
		result := a.admin.db.FindById(fmt.Sprint(id), instance, requestedBy, otherApp.skipUserBinding)
		if result.Error != nil {
			return FieldValidationError{
				Field: fieldName,
				Error: fmt.Sprintf("error finding %s: %s", otherAppName, result.Error),
			}
		}

		if result.RowsAffected == 0 {
			validationError.Error = fmt.Sprintf("%s with id %s not found", otherAppName, id)
			return validationError
		}

		return FieldValidationError{}
	}
}

/*
	API HANDLERS
*/

// ApiList returns a handler function that responds to GET requests on the
// list endpoint, e.g. /api/users.
//
// The handler function will return a JSON response containing all the records
// of the given model.
//
// It will also handle errors and return a 500 Internal Server Error if the
// error is not a gorm.ErrRecordNotFound.
func (a *App) ApiList(db *Database) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := validateRequestMethod(r, http.MethodGet)
		if err != nil {
			handleError(w, err, "Invalid request method")
			return
		}

		userId := getRequestUserId(r, a)

		// Create slice to store the model instances.
		instances, err := createSliceForUndeterminedType(a.model)
		if err != nil {
			handleError(w, err, "Error creating slice for model")
			return
		}

		var result *gorm.DB
		if a.skipUserBinding {
			result = db.FindAll(instances)
		} else {
			result = db.FindAllByUserId(instances, userId)
		}

		if result.Error != nil {
			handleError(w, result.Error, "Error finding instances")
			return
		}

		response, err := json.Marshal(instances)
		if err != nil {
			handleError(w, err, "Error marshaling instances to JSON")
			return
		}

		writeJsonResponse(w, response, http.StatusOK)
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
			handleError(w, err, err.Error())
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
			handleError(w, result.Error, result.Error.Error())
			return
		}

		// Marshal the results into JSON
		response, err := json.Marshal(instance)
		if err != nil {
			handleError(w, err, err.Error())
			return
		}

		writeJsonResponse(w, response, http.StatusOK)
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
			handleError(w, err, "Invalid request method")
			return
		}

		// Will read bytes, append user data and pack the bytes again to be processed
		bodyBytes, err := readRequestBody(r)
		if err != nil {
			handleError(w, err, "Error reading request body")
			return
		}

		// Make sure user is not attempting to modify system data fields
		bodyBytes, err = removeSystemDataFieldsFromRequest(bodyBytes)
		if err != nil {
			handleError(w, err, "Error making sure user is not modifying system data fields")
			return
		}

		// Update SystemData fields
		if !a.skipUserBinding {
			bodyBytes, err = appendUserDataToRequestBody(bodyBytes, r, true, a)
			if err != nil {
				handleError(w, err, "Error appending user data to request body")
				return
			}
		}

		// Create a new instance of the model
		instance := createInstanceForUndeterminedType(a.model)

		// Unmarshal the updated bytes into the instance
		err = json.Unmarshal(bodyBytes, instance)
		if err != nil {
			handleError(w, err, "Error unmarshalling instance")
			return
		}

		// Run validations
		validationErrors := a.Validate(instance)
		if len(validationErrors.Errors) > 0 {
			handleError(w, fmt.Errorf("validation errors: %v", validationErrors), "Validation failed")
			return
		}

		// Perform database operation
		result := db.Create(instance)
		if result.Error != nil {
			handleError(w, result.Error, "DB error")
			return
		}

		// Convert the instance to JSON and send it
		responseBytes, err := json.Marshal(instance)
		if err != nil {
			handleError(w, err, "Error marshalling response")
			return
		}

		writeJsonResponse(w, responseBytes, http.StatusOK)
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
			handleError(w, err, "Invalid request method")
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
			handleError(w, result.Error, result.Error.Error())
			return
		}

		bodyBytes, err := readRequestBody(r)
		if err != nil {
			handleError(w, err, "Error reading request body")
			return
		}

		// Make sure user is not attempting to modify system data fields
		bodyBytes, err = removeSystemDataFieldsFromRequest(bodyBytes)
		if err != nil {
			handleError(w, err, "Error making sure user is not modifying system data fields")
			return
		}

		// Update SystemData fields
		if !a.skipUserBinding {
			bodyBytes, err = appendUserDataToRequestBody(bodyBytes, r, false, a)
			if err != nil {
				handleError(w, err, "Error appending user data to request body")
				return
			}
		}

		// Unmarshal the updated bytes into the instance
		err = json.Unmarshal(bodyBytes, instance)
		if err != nil {
			handleError(w, err, "Error unmarshalling instance")
			return
		}

		// Run validations
		validationErrors := a.Validate(instance)
		if len(validationErrors.Errors) > 0 {
			response, err := json.Marshal(validationErrors)

			if err != nil {
				handleError(w, err, "Error marshalling response")
				return
			}
			writeJsonResponse(w, response, http.StatusBadRequest)
			return
		}

		// Update the record in the database
		result = db.Save(instance)
		if result.Error != nil {
			handleError(w, result.Error, result.Error.Error())
			return
		}

		// Convert the instance to JSON and send it
		response, err := json.Marshal(instance)
		if err != nil {
			handleError(w, err, "Error marshalling response")
			return
		}

		writeJsonResponse(w, response, http.StatusOK)
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
			handleError(w, err, err.Error())
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
			handleError(w, dbResponse.Error, dbResponse.Error.Error())
			return
		}

		// Delete the record by ID
		result := db.Delete(instance)
		if result.Error != nil {
			handleError(w, result.Error, result.Error.Error())
			return
		}

		// Set the content type to application/json
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
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

// handleError logs the given error and writes an internal server error to the
// response writer.
func handleError(w http.ResponseWriter, err error, msg string) {
	log.Error().Err(err).Msg(msg)

	// write the error to the response writer
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error

	// write the error to the response writer
	response, _ := json.Marshal(map[string]string{"error": err.Error()})
	w.Write(response)
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
	// userId := r.Header.Get("requested_by")

	if userId == "" {
		log.Error().Msgf("No requested_by found in authorization header")
		return bytes, fmt.Errorf("user not authenticated")
	}

	convertedUserId, err := strconv.ParseUint(userId, 10, 64)
	if err != nil {
		log.Error().Err(err).Msgf("Error converting requested_by")
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

// writeJsonResponse writes a JSON response to the given http.ResponseWriter.
// It sets the Content-Type header to application/json, the status code to 200 OK,
// and writes the provided data as the response body.
func writeJsonResponse(w http.ResponseWriter, data []byte, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status) // 200 OK
	w.Write(data)         // Send the JSON response
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
