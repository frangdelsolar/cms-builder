package server

import (
	"net/http"
)

func ProtectedRouteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := GetRequestContext(r)

		if !ctx.IsAuthenticated || ctx.User == nil {
			ctx.Logger.Error().Interface("user", ctx.User).Bool("authenticated", ctx.IsAuthenticated).Msg("Not allowed to enter protected route")

			SendJsonResponse(w, http.StatusUnauthorized, nil, "Unauthorized")
			return
		}

		ctx.Logger.Debug().Bool("authenticated", ctx.IsAuthenticated).Msg("Entering protected route")
		next.ServeHTTP(w, r)
	})
}
