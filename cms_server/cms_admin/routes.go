package cms_admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type EntityDetail struct {
	Name   string
	Slug   string
	Fields []string
}

func Routes(r *mux.Router) {
	apps := GetEntities()

	// Define the group route for admin
	adminRouter := r.PathPrefix("/admin").Subrouter()

	// API
	adminAPIRoutes := adminRouter.PathPrefix("/api").Subrouter()
	entityRoutes := adminAPIRoutes.PathPrefix("/entities").Subrouter()
	entityRoutes.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {

		es := []EntityDetail{}

		for _, app := range apps {
			es = append(es, EntityDetail{
				Name:   app.Name(),
				Slug:   app.Plural(),
				Fields: app.Fields(),
			})
		}
		response, err := json.Marshal(es)
		if err != nil {
			log.Error().Err(err).Msgf("Error marshalling entities to JSON")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	})
	entityRoutes.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Entity Detail"))
	})

	// CRUD routes
	for _, app := range apps {
		log.Debug().Msgf("Registering %s routes", app.Name())
		appRoutes := adminAPIRoutes.PathPrefix("/" + app.Plural()).Subrouter()
		appRoutes.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
			APIList(app, w, r)
		})
		appRoutes.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
			APINew(app, w, r)
		})
		appRoutes.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
			APIDetail(app, w, r)
		})
		appRoutes.HandleFunc("/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
			APIUpdate(app, w, r)
		})
		appRoutes.HandleFunc("/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
			APIDestroy(app, w, r)
		})
	}

}
