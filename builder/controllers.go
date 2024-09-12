package builder

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"gorm.io/gorm"
)

func List(model App, db *Database, w http.ResponseWriter, r *http.Request) {
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

func New(model App, db *Database, w http.ResponseWriter, r *http.Request) {
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

func Get(id string, model App, db *Database, w http.ResponseWriter, r *http.Request) {
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

func Update(id string, model App, db *Database, w http.ResponseWriter, r *http.Request) {

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

func Delete(id string, model App, db *Database, w http.ResponseWriter, r *http.Request) {
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
