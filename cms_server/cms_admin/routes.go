package cms_admin

import (
	"net/http"

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
			Detail(app, w, r)
		})
		appRoutes.HandleFunc("/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
			Update(app, w, r)
		})
		appRoutes.HandleFunc("/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
			Destroy(app, w, r)
		})
	}

}
