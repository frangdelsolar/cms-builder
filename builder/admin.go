package builder

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

var (
	ErrAdminNotInitialized = errors.New("admin not initialized")
)

type App interface{}

type Admin struct {
	Apps   []App
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
		Apps:   make([]App, 0),
		db:     db,
		server: server,
	}
}

// Register adds a new App to the Admin instance, applies database migration, and
// registers API routes for CRUD operations.
func (a *Admin) Register(app App) {
	log.Debug().Interface("App", app).Msg("Registering app")

	a.Apps = append(a.Apps, app)

	// Apply database migration
	a.db.Migrate(app)

	appName := GetStructName(app)
	a.registerAPIRoutes(Pluralize(appName), app)
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
func (a *Admin) registerAPIRoutes(appName string, app interface{}) {
	baseRoute := "/api/" + appName

	a.server.AddRoute(
		baseRoute,
		apiList(app, a.db),
		appName+"-list",
		true,
	)

	a.server.AddRoute(
		baseRoute+"/new",
		apiNew(app, a.db),
		appName+"-new",
		true,
	)

	a.server.AddRoute(
		baseRoute+"/{id}",
		apiGet(app, a.db),
		appName+"-get",
		true,
	)

	a.server.AddRoute(
		baseRoute+"/{id}/delete",
		apiDelete(app, a.db),
		appName+"-delete",
		true,
	)

	a.server.AddRoute(
		baseRoute+"/{id}/update",
		apiUpdate(app, a.db),
		appName+"-update",
		true,
	)
}

// apiList returns a handler function that responds to GET requests on the
// list endpoint, e.g. /api/users.
//
// The handler function will return a JSON response containing all the records
// of the given model.
//
// It will also handle errors and return a 500 Internal Server Error if the
// error is not a gorm.ErrRecordNotFound.
func apiList(model App, db *Database) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get out if not GET
		method := r.Method
		if method != "GET" {
			fmt.Fprintf(w, "Method not implemented")
			return
		}

		// Get the type of the model
		modelType := reflect.TypeOf(model)

		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}
		if modelType.Kind() != reflect.Struct {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Create a slice to hold the results
		sliceType := reflect.SliceOf(modelType)
		entities := reflect.New(sliceType).Interface()

		result := db.GetAll(entities)
		if result.Error != nil {
			log.Error().Err(result.Error).Msgf("Error fetching %s records", model)
			return
		}

		// Marshal the results into JSON
		entitiesValue := reflect.ValueOf(entities).Elem()
		response, err := json.Marshal(entitiesValue.Interface())
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Set the content type to application/json
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		w.Write(response)            // Send the JSON response
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
func apiNew(model App, db *Database) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// get out if not GET
		method := r.Method
		if method != "POST" {
			fmt.Fprintf(w, "Not implemented")
			return
		}

		defer r.Body.Close()

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msgf("Error reading request body for creation")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		log.Debug().Interface("bodyBytes", string(bodyBytes)).Msg("Request")

		// Create a new instance of the model
		instanceType := reflect.TypeOf(model)
		instance := reflect.New(instanceType).Interface()

		// Unmarshal the bodyBytes into the instance
		err = json.Unmarshal(bodyBytes, instance)
		if err != nil {
			log.Error().Err(err).Msgf("Error unmarshalling request body for creation")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		log.Debug().Interface("instance", instance).Msgf("Creating new")

		// Use reflection to get a pointer to the instance
		instancePtr := reflect.ValueOf(instance).Elem().Interface()

		// Perform database operations
		result := db.Create(instancePtr)
		if result.Error != nil {
			log.Error().Err(result.Error).Msgf("Error creating in database")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Marshal instance into JSON
		response, err := json.Marshal(instance)
		if err != nil {
			log.Error().Err(err).Msgf("Error marshalling response for creation")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Set the content type to application/json
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated) // 201 Created
		w.Write(response)                 // Send the JSON response
	}
}

// apiGet returns a handler function that responds to GET requests on the
// details endpoint, e.g. /api/users/{id}.
//
// The handler function will return a JSON response containing the requested
// record.
//
// It will also handle errors and return a 404 Not Found if the error is a
// gorm.ErrRecordNotFound, or a 500 Internal Server Error if the error is
// not a gorm.ErrRecordNotFound.
func apiGet(model App, db *Database) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		// get out if not GET
		method := r.Method
		if method != "GET" {
			fmt.Fprintf(w, "Not implemented")
			return
		}

		// Create a new instance of the model
		instanceType := reflect.TypeOf(model)
		instance := reflect.New(instanceType.Elem()).Interface()

		// Query the database to find the record by ID
		result := db.GetById(id, instance)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				http.Error(w, "Not found", http.StatusNotFound)
			} else {
				log.Error().Err(result.Error).Msgf("Error fetching record")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Marshal the results into JSON
		response, err := json.Marshal(instance)
		if err != nil {
			log.Error().Err(err).Msgf("Error marshalling record to JSON")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Set the content type to application/json
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		w.Write(response)            // Send the JSON response
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
func apiUpdate(model App, db *Database) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		// get out if not PUT
		method := r.Method
		if method != "PUT" {
			fmt.Fprintf(w, "Not implemented")
			return
		}

		log.Info().Msgf("Update %s", id)

		// Create a new instance of the model to hold the updated data
		instanceType := reflect.TypeOf(model)
		if instanceType.Kind() == reflect.Ptr {
			instanceType = instanceType.Elem()
		}
		instance := reflect.New(instanceType).Interface()

		// Retrieve the existing record from the database
		result := db.GetById(id, instance)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				http.Error(w, "Not found", http.StatusNotFound)
			} else {
				log.Error().Err(result.Error).Msgf("Error fetching record")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Read the request body
		defer r.Body.Close()
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msgf("Error reading request body for update")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Unmarshal the request body into the instance
		err = json.Unmarshal(bodyBytes, instance)
		if err != nil {
			log.Error().Err(err).Msgf("Error unmarshalling request body for update")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Update the record in the database

		result = db.DB.Save(instance)
		if result.Error != nil {
			log.Error().Err(result.Error).Msgf("Error updating record")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Marshal the updated instance into JSON
		response, err := json.Marshal(instance)
		if err != nil {
			log.Error().Err(err).Msg("error marshalling record to JSON")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Set the content type to application/json
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		w.Write(response)            // Send the JSON response
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
func apiDelete(model App, db *Database) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		// get out if not DELETE
		method := r.Method
		if method != "DELETE" {
			fmt.Fprintf(w, "Not implemented")
			return
		}

		log.Info().Msgf("Destroy %s", id)

		// Create a new instance of the model
		instanceType := reflect.TypeOf(model)
		if instanceType.Kind() == reflect.Ptr {
			instanceType = instanceType.Elem()
		}
		instance := reflect.New(instanceType).Interface()

		// Delete the record by ID
		item := db.GetById(id, instance)
		result := db.DB.Delete(item)
		if result.Error != nil {
			if result.RowsAffected == 0 {
				http.Error(w, "Not found", http.StatusNotFound)
			} else {
				log.Error().Err(result.Error).Msgf("Error deleting record")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Set the content type to application/json
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
