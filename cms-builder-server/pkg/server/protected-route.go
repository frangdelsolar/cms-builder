package server

import (
	"net/http"
)

func ProtectedRouteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := GetRequestContext(r)

		if !ctx.IsAuthenticated || ctx.User == nil {
			SendJsonResponse(w, http.StatusUnauthorized, nil, "Unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}
