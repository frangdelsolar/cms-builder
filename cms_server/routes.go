package cms_server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func Routes(r *mux.Router) {

	// Define the group route for admin
	adminRouter := r.PathPrefix("/admin").Subrouter()

	// Admin routes
	adminRouter.HandleFunc("/dashboard", Dashboard)

	// Entities routes
	entities := GetEntities()
	for _, entity := range entities {
		log.Debug().Msgf("Registering %s routes", entity.Name())
		entityRoutes := adminRouter.PathPrefix("/" + entity.Plural()).Subrouter()
		entityRoutes.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("List"))
		})
		entityRoutes.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("New"))
		})
		entityRoutes.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Detail"))
		})
		entityRoutes.HandleFunc("/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Update"))
		})
		entityRoutes.HandleFunc("/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Delete"))
		})
	}

}
