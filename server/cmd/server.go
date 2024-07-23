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

func GetServer() (*Server, error) {
	log.Warn().Msg("Starting server...")

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
