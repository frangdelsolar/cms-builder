package cms_server

import "github.com/gorilla/mux"

func Routes(r *mux.Router) {
	// Define the group route for admin
	adminRouter := r.PathPrefix("/admin").Subrouter()

	// Admin routes
	adminRouter.HandleFunc("/dashboard", Dashboard)

}
