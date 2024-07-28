package cms_admin

import (
	"fmt"
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
	entities := app.Model
	db.Find(&entities)

	fmt.Fprintf(w, "%+v", entities)
}

func New(app Entity, w http.ResponseWriter, r *http.Request) {
	// if method GET print not implemented
	method := r.Method
	if method == "GET" {
		fmt.Fprintf(w, "Not implemented")
		return
	}

	// if method POST create the thing
	if method == "POST" {
		fmt.Fprintf(w, "POST")
		return
	}

}
