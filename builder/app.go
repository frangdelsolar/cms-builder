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

type RequestData struct {
	Instance   interface{}            // instance where data will be stored
	Pagination *Pagination            // pagination object
	Parameters RequestParameters      // request parameters
	InstanceId string                 // instance id
	Body       map[string]interface{} // request body
}

// CleanBody removes system data fields from the request body.
//
// This function takes no parameters and returns no values.
//
// It iterates over the keys of the SystemData struct and removes any key-value pair
// from the request body where the key matches a key from the SystemData struct.
func (a *RequestData) CleanBody() {
	systemDataInstance := SystemData{}

	for _, key := range systemDataInstance.Keys() {
		delete(a.Body, key)
	}
}

// SetSystemData adds the system data fields to the request body.
//
// The function takes two parameters, isNewRecord and requestedBy.
// isNewRecord is a boolean indicating whether the request is to create a new record or update an existing one.
// requestedBy is the ID of the user making the request.
//
// If isNewRecord is true, the function adds the CreatedById field to the request body with the value of requestedBy.
// The function always adds the UpdatedById field to the request body with the value of requestedBy.
//
// The function returns an error if there is an error converting requestedBy to an unsigned integer.
func (a *RequestData) SetSystemData(isNewRecord bool, requestedBy string) error {
	convertedUserId, err := strconv.ParseUint(requestedBy, 10, 64)
	if err != nil {
		log.Error().Err(err).Msgf("Error converting userId")
		return err
	}
	if isNewRecord {
		a.Body["CreatedById"] = convertedUserId
	}
	a.Body["UpdatedById"] = convertedUserId
	return nil
}

// GetBodyBytes takes the request body as a map[string]interface{} and returns a byte slice representing the JSON encoded body.
//
// If there is an error marshaling the request body, the function returns an error.
func (a *RequestData) GetBodyBytes() ([]byte, error) {
	bodyBytes, err := json.Marshal(a.Body)
	if err != nil {
		log.Error().Err(err).Msgf("Error marshalling request body for creation")
		return nil, err
	}
	return bodyBytes, nil
}

// SetInstanceData sets the values of the instance fields with the given map of values.
//
// This function takes one parameter, values, which is a map of field names to values.
// The function sets the value of each field in the instance that matches a key in the map.
// If a field does not match a key in the map, its value is left unchanged.
// If a field does match a key in the map but is not a string, the function returns an error.
// The function returns an error if the instance is not a struct.
func (a *RequestData) SetInstanceData(values map[string]string) error {
	errMsg := ""

	v := reflect.ValueOf(a.Instance).Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("instance is not a struct")
	}

	for key, value := range values {
		field := v.FieldByName(key)
		if field.IsValid() && field.CanSet() {
			switch field.Kind() {
			case reflect.String:
				field.SetString(value)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				i, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					errMsg += fmt.Sprintf("error parsing %s as int: %s\n", value, err)
				}
				field.SetInt(i)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				u, err := strconv.ParseUint(value, 10, 64)
				if err != nil {
					errMsg += fmt.Sprintf("error parsing %s as uint: %s\n", value, err)
				}
				field.SetUint(u)
			case reflect.Bool:
				b, err := strconv.ParseBool(value)
				if err != nil {
					errMsg += fmt.Sprintf("error parsing %s as bool: %s\n", value, err)
				}
				field.SetBool(b)
			case reflect.Float32, reflect.Float64:
				f, err := strconv.ParseFloat(value, 64)
				if err != nil {
					errMsg += fmt.Sprintf("error parsing %s as float: %s\n", value, err)
				}
				field.SetFloat(f)

			default:
				errMsg += fmt.Sprintf("field %s is not a supported type\n", key)
			}
		} else {
			errMsg += fmt.Sprintf("field %s not found in instance\n", key)
		}

	}

	a.Instance = v.Interface()

	if errMsg != "" {
		return fmt.Errorf("error setting instance data: %s", errMsg)
	}

	return nil
}

// ApiFunction is a function that takes an ApiInput, a *Database and an *App and returns a *gorm.DB.
// The ApiFunction is used to define the behavior of the API endpoints.
type ApiFunction func(input *RequestData, db *Database, app *App) (*gorm.DB, error)

type API struct {
	List   ApiFunction // List is a function that takes an ApiInput, a *Database and an *App and returns a *gorm.DB will be called on GET endpoints (e.g. /api/users)
	Detail ApiFunction // Detail is a function that takes an ApiInput, a *Database and an *App and returns a *gorm.DB will be called on GET endpoints (e.g. /api/users/{id})
	Create ApiFunction // Create is a function that takes an ApiInput, a *Database and an *App and returns a *gorm.DB will be called on POST endpoints (e.g. /api/users/new)
	Update ApiFunction // Update is a function that takes an ApiInput, a *Database and an *App and returns a *gorm.DB will be called on PUT endpoints (e.g. /api/users/{id}/update)
	Delete ApiFunction // Delete is a function that takes an ApiInput, a *Database and an *App and returns a *gorm.DB will be called on DELETE endpoints (e.g. /api/users/{id}/delete)
}

var DefaultList ApiFunction = func(input *RequestData, db *Database, app *App) (*gorm.DB, error) {
	query := ""
	for _, role := range input.Parameters.Roles {
		if role == AdminRole {
			return db.Find(input.Instance, query, input.Pagination), nil
		}
	}

	query = "created_by_id = '" + input.Parameters.RequestedById + "'"
	result := db.Find(input.Instance, query, input.Pagination)

	return result, nil
}

var DefaultDetail ApiFunction = func(input *RequestData, db *Database, app *App) (*gorm.DB, error) {
	queryExtension := ""

	for _, role := range input.Parameters.Roles {
		if role == AdminRole {
			return db.FindById(input.InstanceId, input.Instance, queryExtension), nil
		}
	}

	queryExtension = "created_by_id = '" + input.Parameters.RequestedById + "'"
	return db.FindById(input.InstanceId, input.Instance, queryExtension), nil
}

var DefaultCreate ApiFunction = func(input *RequestData, db *Database, app *App) (*gorm.DB, error) {
	err := input.SetInstanceData(map[string]string{
		"UpdatedByID": input.Parameters.RequestedById,
		"CreatedByID": input.Parameters.RequestedById,
	})
	if err != nil {
		return nil, err
	}

	result := db.Create(input.Instance)
	return result, nil
}

var DefaultUpdate ApiFunction = func(input *RequestData, db *Database, app *App) (*gorm.DB, error) {
	err := input.SetInstanceData(map[string]string{
		"UpdatedByID": input.Parameters.RequestedById,
	})
	if err != nil {
		return nil, err
	}

	result := db.Save(input.Instance)
	return result, nil
}

var DefaultDelete ApiFunction = func(input *RequestData, db *Database, app *App) (*gorm.DB, error) {
	result := db.Delete(input.Instance)
	return result, nil
}

type App struct {
	model           interface{}       // The model struct
	skipUserBinding bool              // Means that theres a CreatedBy field in the model that will be used for filtering the database query to only include records created by the user
	Admin           *Admin            // The admin instance
	Validators      ValidatorsMap     // A map of field names to validation functions
	Permissions     RolePermissionMap // Key is Role name, value is permission
	Api             *API              // The API struct
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
		err := ValidateRequestMethod(r, http.MethodGet)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		params := FormatRequestParameters(r, a.Admin.builder)
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

		listInput := RequestData{
			Instance:   instances,
			Pagination: pagination,
			Parameters: params,
			Body:       FormatRequestBody(r),
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

		SendJsonResponseWithPagination(w, http.StatusOK, listInput.Instance, a.Name()+" list", listInput.Pagination)
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

		err := ValidateRequestMethod(r, http.MethodGet)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, err, err.Error())
			return
		}

		params := FormatRequestParameters(r, a.Admin.builder)
		isAllowed := a.Permissions.HasPermission(params.Roles, OperationRead)
		if !isAllowed {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to read this resource")
			return
		}

		// Create a new instance of the model
		instance := createInstanceForUndeterminedType(a.model)

		detailInput := RequestData{
			Instance:   instance,
			Parameters: params,
			InstanceId: GetUrlParam("id", r),
			Body:       FormatRequestBody(r),
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

		err := ValidateRequestMethod(r, http.MethodPost)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		params := FormatRequestParameters(r, a.Admin.builder)
		isAllowed := a.Permissions.HasPermission(params.Roles, OperationCreate)
		if !isAllowed {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to create this resource")
			return
		}

		// Create a new instance of the model
		instance := createInstanceForUndeterminedType(a.model)

		createInput := RequestData{
			Instance:   instance,
			Parameters: params,
			Body:       FormatRequestBody(r),
		}

		createInput.CleanBody()

		err = createInput.SetSystemData(true, params.RequestedById)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		data, err := createInput.GetBodyBytes()
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// Unmarshal the updated bytes into the instance
		err = json.Unmarshal(data, instance)
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

		err := ValidateRequestMethod(r, http.MethodPut)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		params := FormatRequestParameters(r, a.Admin.builder)

		isAllowed := a.Permissions.HasPermission(params.Roles, OperationUpdate)
		if !isAllowed {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to update this resource")
			return
		}

		// Create a new instance of the model
		instance := createInstanceForUndeterminedType(a.model)

		apiInput := RequestData{
			Instance:   instance,
			Parameters: params,
			InstanceId: GetUrlParam("id", r),
			Body:       FormatRequestBody(r),
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

		apiInput.CleanBody()

		err = apiInput.SetSystemData(false, params.RequestedById)
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		data, err := apiInput.GetBodyBytes()
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// Unmarshal the updated bytes into the instance
		err = json.Unmarshal(data, instance)
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

		err := ValidateRequestMethod(r, http.MethodDelete)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		params := FormatRequestParameters(r, a.Admin.builder)
		isAllowed := a.Permissions.HasPermission(params.Roles, OperationDelete)
		if !isAllowed {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to delete this resource")
			return
		}

		// Create a new instance of the model
		instance := createInstanceForUndeterminedType(a.model)

		apiInput := RequestData{
			Instance:   instance,
			Parameters: params,
			InstanceId: GetUrlParam("id", r),
			Body:       FormatRequestBody(r),
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
