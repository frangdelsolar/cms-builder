package cms_admin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
)

func Routes(r *mux.Router) {
	/*
		No views, just api
	*/
	// Define the group route for admin
	adminRouter := r.PathPrefix("/admin/api").Subrouter()

	entityRoutes := adminRouter.PathPrefix("/entities").Subrouter()
	entityRoutes.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Entities List"))
	})
	entityRoutes.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Entity Detail"))
	})

	// CRUD routes
	apps := GetEntities()

	for _, app := range apps {
		log.Debug().Msgf("Registering %s routes", app.Name())
		appRoutes := adminRouter.PathPrefix("/" + app.Plural()).Subrouter()
		appRoutes.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
			List(app, w, r)
		})
		appRoutes.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
			New(app, w, r)
		})
		appRoutes.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Detail"))

		})
		appRoutes.HandleFunc("/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Update"))
		})
		appRoutes.HandleFunc("/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Delete"))
		})
	}

}

func List(app Entity, w http.ResponseWriter, r *http.Request) {
	db := config.DB

	// Create a slice to hold the results dynamically
	modelType := reflect.TypeOf(app.Model)

	// Create a new slice of the correct type
	entities := reflect.New(modelType).Interface()

	// Query the database to find all records
	result := db.Find(entities)
	if result.Error != nil {
		log.Error().Err(result.Error).Msgf("Error fetching %s records", app.Name())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Marshal the results into JSON
	response, err := json.Marshal(reflect.ValueOf(entities).Elem().Interface())
	if err != nil {
		log.Error().Err(err).Msgf("Error marshalling %s records to JSON", app.Name())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set the content type to application/json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK
	w.Write(response)            // Send the JSON response
}

func New(app Entity, w http.ResponseWriter, r *http.Request) {
	method := r.Method
	if method == "GET" {
		fmt.Fprintf(w, "Not implemented")
		return
	}

	if method == "POST" {
		defer r.Body.Close()

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msgf("Error reading request body for %s creation", app.Name())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		log.Debug().Interface("bodyBytes", string(bodyBytes)).Msg("Request")

		// Create a new instance of the model
		instanceType := reflect.TypeOf(app.Model)
		instance := reflect.New(instanceType).Interface()

		// Unmarshal the bodyBytes into the instance
		err = json.Unmarshal(bodyBytes, instance)
		if err != nil {
			log.Error().Err(err).Msgf("Error unmarshalling request body for %s creation", app.Name())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		log.Debug().Interface("instance", instance).Msgf("Creating new %s", app.Name())

		// Use reflection to get a pointer to the instance
		instancePtr := reflect.ValueOf(instance).Elem().Interface()

		// Perform database operations
		result := config.DB.Create(instancePtr)
		if result.Error != nil {
			log.Error().Err(result.Error).Msgf("Error creating %s in database", app.Name())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Marshal instance into JSON
		response, err := json.Marshal(instance)
		if err != nil {
			log.Error().Err(err).Msgf("Error marshalling response for %s creation", app.Name())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Set the content type to application/json
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated) // 201 Created
		w.Write(response)                 // Send the JSON response
	}
}
