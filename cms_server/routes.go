package cms_server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func Routes(r *mux.Router) {
	entities := GetEntities()
	// Define the group route for admin
	adminRouter := r.PathPrefix("/admin").Subrouter()

	// Admin routes
	adminRouter.HandleFunc("/dashboard", Dashboard)

	for _, entity := range entities {
		log.Debug().Msgf("Registering %s routes", entity.Name())
		entityRoutes := adminRouter.PathPrefix("/" + entity.Name()).Subrouter()
		entityRoutes.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		entityRoutes.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		entityRoutes.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		entityRoutes.HandleFunc("/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		entityRoutes.HandleFunc("/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
	}

}
