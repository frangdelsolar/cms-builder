package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

var svr *Server
var port = "8080"

type Server struct {
	*http.Server
}

func (s *Server) Router() *mux.Router {
	return s.Handler.(*mux.Router)
}

func (s *Server) Run() error {
	log.Debug().Interface("url", "http://localhost:8080/").Msgf("Running server on port %s", port)

	return s.ListenAndServe()
}

func GetServer() (*Server, error) {

	// Define Router
	r := mux.NewRouter()

	// Middlewares
	r.Use(loggingMiddleware)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Home")
		// log.Debug().Msgf("X-CSRF-Token: %s", r.Header)
	})

	r.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong")
	})

	// Create a new HTTP server with the router
	svr = &Server{&http.Server{
		Addr:    ":" + port,
		Handler: r,
	}}

	return svr, nil
}

// loggingMiddleware middleware function to log the request URI.
//
// Takes in a http.Handler and returns a http.Handler.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}
