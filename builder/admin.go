package builder

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

var (
	ErrAdminNotInitialized = errors.New("admin not initialized")
)

type FieldValidationError struct {
	Field string
	Error string
}

func NewFieldValidationError(fieldName string) FieldValidationError {
	return FieldValidationError{
		Field: fieldName,
		Error: "",
	}
}

type FieldValidationFunc func(string) FieldValidationError

type ValidationResult struct {
	Errors []FieldValidationError
}

func (r *ValidationResult) Execute(validationFunc FieldValidationFunc, field string) {
	if err := validationFunc(field); err != (FieldValidationError{}) {
		r.Errors = append(r.Errors, err)
	}
}

type Model interface {
	Validate() ValidationResult
}

type App struct {
	Model           Model
	SkipUserBinding bool // Means that theres a CreatedBy field in the model that will be used for filtering and shit
}

// Name returns the name of the model as a string, lowercased and without the package name.
func (a *App) Name() string {
	return GetStructName(a.Model)
}

// PluralName returns the plural form of the name of the model as a string.
func (a *App) PluralName() string {
	return Pluralize(a.Name())
}

type Admin struct {
	Models []Model
	db     *Database
	server *Server
}

// NewAdmin creates a new instance of the Admin, which is a central
// configuration and management structure for managing applications.
//
// Parameters:
// - db: A pointer to the Database instance to use for database operations.
// - server: A pointer to the Server instance to use for registering API routes.
//
// Returns:
// - *Admin: A pointer to the new Admin instance.
func NewAdmin(db *Database, server *Server) *Admin {
	return &Admin{
		Models: make([]Model, 0),
		db:     db,
		server: server,
	}
}

// Register adds a new App to the Admin instance, applies database migration, and
// registers API routes for CRUD operations.
func (a *Admin) Register(model Model, skipUserBinding bool) {
	log.Debug().Interface("App", model).Msg("Registering app")

	app := App{
		Model:           model,
		SkipUserBinding: skipUserBinding,
	}

	a.Models = append(a.Models, app.Model)
	a.db.Migrate(app.Model)

	a.registerAPIRoutes(app)
}

// registerAPIRoutes registers API routes for the given App.
//
// It takes two arguments:
//   - appName: The plural name of the App, which is used to generate the base route.
//   - app: The App struct to use for generating the API routes.
//
// It registers the following API routes:
//   - GET /{appName}: Returns a list of all App instances.
//   - POST /{appName}/new: Creates a new App instance.
//   - GET /{appName}/{id}: Returns the App instance with the given ID.
//   - DELETE /{appName}/{id}/delete: Deletes the App instance with the given ID.
//   - PUT /{appName}/{id}/update: Updates the App instance with the given ID.
func (a *Admin) registerAPIRoutes(app App) {
	baseRoute := "/api/" + app.PluralName()

	a.server.AddRoute(
		baseRoute,
		apiList(app, a.db),
		app.Name()+"-list",
		true,
	)

	a.server.AddRoute(
		baseRoute+"/new",
		apiNew(app, a.db),
		app.Name()+"-new",
		true,
	)

	a.server.AddRoute(
		baseRoute+"/{id}",
		apiDetail(app, a.db),
		app.Name()+"-get",
		true,
	)

	a.server.AddRoute(
		baseRoute+"/{id}/delete",
		apiDelete(app, a.db),
		app.Name()+"-delete",
		true,
	)

	a.server.AddRoute(
		baseRoute+"/{id}/update",
		apiUpdate(app, a.db),
		app.Name()+"-update",
		true,
	)
}

/*
	API HANDLERS
*/

// apiList returns a handler function that responds to GET requests on the
// list endpoint, e.g. /api/users.
//
// The handler function will return a JSON response containing all the records
// of the given model.
//
// It will also handle errors and return a 500 Internal Server Error if the
// error is not a gorm.ErrRecordNotFound.
func apiList(app App, db *Database) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := validateRequestMethod(r, http.MethodGet)
		if err != nil {
			handleError(w, err, "Invalid request method")
			return
		}

		// Get the user ID from the request.
		userId := r.Header.Get("user_id")

		// Create slice to store the model instances.
		instances, err := createSliceForUndeterminedType(app.Model)
		if err != nil {
			handleError(w, err, "Error creating slice for model")
			return
		}

		var result *gorm.DB
		if app.SkipUserBinding {
			result = db.FindAll(instances)
		} else {
			result = db.FindAllByUserId(instances, userId)
		}

		log.Debug().Interface("instances", instances).Msg("instances")

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

// apiDetail returns a handler function that responds to GET requests on the
// details endpoint, e.g. /api/users/{id}.
//
// The handler function will return a JSON response containing the requested
// record.
//
// It will also handle errors and return a 404 Not Found if the error is a
// gorm.ErrRecordNotFound, or a 500 Internal Server Error if the error is
// not a gorm.ErrRecordNotFound.
func apiDetail(app App, db *Database) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := validateRequestMethod(r, http.MethodGet)
		if err != nil {
			handleError(w, err, err.Error())
			return
		}

		// Retrieve parameters from the request
		instanceId := mux.Vars(r)["id"]
		userId := r.Header.Get("user_id")

		// Create a new instance of the model
		instance := createInstanceForUndeterminedType(app.Model)

		// Query the database to find the record by ID
		result := db.FindById(instanceId, instance, userId, app.SkipUserBinding)
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
func apiNew(app App, db *Database) HandlerFunc {
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

		updatedBytes, err := appendUserDataToRequestBody(bodyBytes, r, true)
		if err != nil {
			handleError(w, err, "Error appending user data to request body")
			return
		}

		// Create a new instance of the model
		instance := createInstanceForUndeterminedType(app.Model)

		// Unmarshal the updated bytes into the instance
		err = json.Unmarshal(updatedBytes, instance)
		if err != nil {
			handleError(w, err, "Error unmarshalling instance")
			return
		}

		// Run validations
		validationErrors, err := validateInterface(instance)
		if err != nil {
			handleError(w, err, "Error validating instance")
			return
		}
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

// apiUpdate returns a handler function that responds to PUT requests on the
// details endpoint, e.g. /api/users/{id}.
//
// The handler function will update the record in the database and return a
// JSON response containing the updated record.
//
// It will also handle errors and return a 404 Not Found if the error is a
// gorm.ErrRecordNotFound, or a 500 Internal Server Error if the error is
// not a gorm.ErrRecordNotFound.
func apiUpdate(app App, db *Database) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := validateRequestMethod(r, http.MethodPut)
		if err != nil {
			handleError(w, err, "Invalid request method")
			return
		}

		// Retrieve parameters from the request
		instanceId := mux.Vars(r)["id"]
		userId := r.Header.Get("user_id")

		// Create a new instance of the model
		instance := createInstanceForUndeterminedType(app.Model)

		// Query the database to find the record by ID
		result := db.FindById(instanceId, app.Model, userId, app.SkipUserBinding)
		if result.Error != nil {
			handleError(w, result.Error, result.Error.Error())
			return
		}

		bodyBytes, err := readRequestBody(r)
		if err != nil {
			handleError(w, err, "Error reading request body")
			return
		}

		updatedBytes, err := appendUserDataToRequestBody(bodyBytes, r, false)
		if err != nil {
			handleError(w, err, "Error appending user data to request body")
			return
		}

		// Unmarshal the updated bytes into the instance
		err = json.Unmarshal(updatedBytes, instance)
		if err != nil {
			handleError(w, err, "Error unmarshalling instance")
			return
		}

		// Run validations
		validationErrors, err := validateInterface(instance)
		if err != nil {
			handleError(w, err, "Error validating instance")
			return
		}
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

// apiDelete returns a handler function that responds to DELETE requests on the
// details endpoint, e.g. /api/users/{id}.
//
// The handler function will delete the record in the database and return a
// JSON response with a 204 No Content status code.
//
// It will also handle errors and return a 404 Not Found if the error is a
// gorm.ErrRecordNotFound, or a 500 Internal Server Error if the error is
// not a gorm.ErrRecordNotFound.
func apiDelete(app App, db *Database) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := validateRequestMethod(r, http.MethodDelete)
		if err != nil {
			handleError(w, err, err.Error())
			return
		}

		// Retrieve parameters from the request
		instanceId := mux.Vars(r)["id"]
		userId := r.Header.Get("user_id")

		// Query the database to find the record by ID
		dbInstance := db.FindById(instanceId, app.Model, userId, app.SkipUserBinding)
		if dbInstance.Error != nil {
			handleError(w, dbInstance.Error, dbInstance.Error.Error())
			return
		}

		// Delete the record by ID
		result := db.Delete(dbInstance)
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

// validateRequestMethod returns an error if the request method does not match the given
// method string. The error message will include the actual request method.
func validateRequestMethod(r *http.Request, method string) error {
	if r.Method != method {
		return fmt.Errorf("invalid request method: %s", r.Method)
	}
	return nil
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

// appendUserDataToRequestBody appends the user_id from the request header to the request body for create and update operations.
// For create operations, the user_id is added to the CreatedById field.
// For update operations, the user_id is added to the UpdatedById field.
// The function returns the updated request body as a byte array, or an error if the request body cannot be unmarshalled or marshalled.
func appendUserDataToRequestBody(bytes []byte, r *http.Request, isNewRecord bool) ([]byte, error) {
	jsonData, err := unmarshalRequestBody(bytes)
	if err != nil {
		log.Error().Err(err).Msgf("Error unmarshalling request body for creation")
		return nil, err
	}

	// Retrieve user_id from request header
	userId := r.Header.Get("user_id")
	convertedUserId, err := strconv.ParseUint(userId, 10, 64)
	if err != nil {
		log.Error().Err(err).Msgf("Error converting user_id")
		return nil, err
	}

	if isNewRecord {
		jsonData["CreatedById"] = convertedUserId
	}

	jsonData["UpdatedById"] = convertedUserId

	log.Debug().Interface("jsonData", jsonData).Msg("Request")

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

/*
	OTHER HELPERS
*/

// validateInterface takes an interface and checks if it implements the Model interface.
// If it does, it calls the Validate() method and returns the validation errors if any.
// If it does not, it returns an error.
// If the instance is valid, it returns nil, nil.
func validateInterface(instance interface{}) (ValidationResult, error) {
	validator, ok := instance.(Model)

	if !ok {
		return ValidationResult{}, fmt.Errorf("invalid model type")
	}
	return validator.Validate(), nil
}
