package cms_admin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func List(app Entity, w http.ResponseWriter, r *http.Request) {
	db := config.DB

	// Get the type of the model
	modelType := reflect.TypeOf(app.Model)
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

	// Query the database to find all records
	result := db.Find(entities)
	if result.Error != nil {
		log.Error().Err(result.Error).Msgf("Error fetching %s records", app.Name())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Marshal the results into JSON
	entitiesValue := reflect.ValueOf(entities).Elem()
	response, err := json.Marshal(entitiesValue.Interface())
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

func Detail(app Entity, w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	log.Info().Msgf("Detail %s %s", app.Name(), id)

	db := config.DB

	// Create a new instance of the model
	instanceType := reflect.TypeOf(app.Model)
	instance := reflect.New(instanceType.Elem()).Interface()

	// Query the database to find the record by ID
	result := db.Where("id = ?", id).First(instance)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "Not found", http.StatusNotFound)
		} else {
			log.Error().Err(result.Error).Msgf("Error fetching %s record", app.Name())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Marshal the results into JSON
	response, err := json.Marshal(instance)
	if err != nil {
		log.Error().Err(err).Msgf("Error marshalling %s record to JSON", app.Name())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set the content type to application/json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK
	w.Write(response)            // Send the JSON response
}

func Update(app Entity, w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	log.Info().Msgf("Update %s %s", app.Name(), id)

	db := config.DB

	// Create a new instance of the model to hold the updated data
	instanceType := reflect.TypeOf(app.Model)
	if instanceType.Kind() == reflect.Ptr {
		instanceType = instanceType.Elem()
	}
	instance := reflect.New(instanceType).Interface()

	// Retrieve the existing record from the database
	result := db.Where("id = ?", id).First(instance)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "Not found", http.StatusNotFound)
		} else {
			log.Error().Err(result.Error).Msgf("Error fetching %s record", app.Name())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Read the request body
	defer r.Body.Close()
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msgf("Error reading request body for %s update", app.Name())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Unmarshal the request body into the instance
	err = json.Unmarshal(bodyBytes, instance)
	if err != nil {
		log.Error().Err(err).Msgf("Error unmarshalling request body for %s update", app.Name())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Update the record in the database
	result = db.Save(instance)
	if result.Error != nil {
		log.Error().Err(result.Error).Msgf("Error updating %s record", app.Name())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Marshal the updated instance into JSON
	response, err := json.Marshal(instance)
	if err != nil {
		log.Error().Err(err).Msgf("Error marshalling %s record to JSON", app.Name())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set the content type to application/json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK
	w.Write(response)            // Send the JSON response
}

func Destroy(app Entity, w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	log.Info().Msgf("Destroy %s %s", app.Name(), id)

	db := config.DB

	// Create a new instance of the model
	instanceType := reflect.TypeOf(app.Model)
	if instanceType.Kind() == reflect.Ptr {
		instanceType = instanceType.Elem()
	}
	instance := reflect.New(instanceType).Interface()

	// Delete the record by ID
	result := db.Where("id = ?", id).Delete(instance)
	if result.Error != nil {
		if result.RowsAffected == 0 {
			http.Error(w, "Not found", http.StatusNotFound)
		} else {
			log.Error().Err(result.Error).Msgf("Error deleting %s record", app.Name())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Set the content type to application/json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent) // 204 No Content
}
