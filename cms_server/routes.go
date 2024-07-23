package cms_server

import "github.com/gorilla/mux"

func Routes(r *mux.Router) {

	// have everything wrapped in /admin group
	adminRouter := r.PathPrefix("/admin").Subrouter()
	adminRouter.HandleFunc("/dashboard", Dashboard)

}
