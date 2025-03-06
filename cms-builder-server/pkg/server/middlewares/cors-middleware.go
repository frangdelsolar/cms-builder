package server

import (
	"fmt"
	"net/http"

	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/utils"
)

// CorsMiddleware adds Cross-Origin Resource Sharing headers to the response.
//
// It sets the following headers:
//
// - Access-Control-Allow-Headers: Content-Type, Authorization, Origin
// - Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
// - Access-Control-Allow-Origin: *
//
// It also checks the Origin header against the list of allowed origins
// and returns a 403 Forbidden response if the origin is not allowed.
//
// If the request method is OPTIONS, it returns a 200 OK response immediately.
func CorsMiddleware(allowedOrigins []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Origin")

			origin := r.Header.Get("Origin")

			// Get the log from the context
			log := svrUtils.GetRequestLogger(r)

			// If the Origin header is missing, block the request
			if origin == "" && allowedOrigins[0] != "*" {
				err := fmt.Errorf("missing Origin header")
				log.Warn().Interface("headers", r.Header).Interface("allowedOrigins", allowedOrigins).Msg("CORS: Missing Origin header")
				svrUtils.SendJsonResponse(w, http.StatusForbidden, nil, err.Error())
				return
			}

			// Check if the origin is allowed
			if len(allowedOrigins) > 0 && allowedOrigins[0] == "*" || contains(allowedOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Origin", origin)

				// Handle OPTIONS requests
				if r.Method == "OPTIONS" {
					w.WriteHeader(http.StatusOK)
					return
				}
			} else {
				err := fmt.Errorf("origin '%s' is not allowed", origin)
				log.Warn().Interface("headers", r.Header).Interface("allowedOrigins", allowedOrigins).Interface("origin", origin).Msg("CORS")
				svrUtils.SendJsonResponse(w, http.StatusForbidden, nil, err.Error())
				return
			}

			// Proceed to the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// contains checks if a slice of strings contains a specific string.
//
// It iterates over the slice 's' and returns true if the element 'e' is found;
// otherwise, it returns false.
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
