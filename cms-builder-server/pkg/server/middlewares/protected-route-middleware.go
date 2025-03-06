package server

import (
	"net/http"

	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/utils"
)

func ProtectedRouteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := svrUtils.GetRequestContext(r)

		if !ctx.IsAuthenticated || ctx.User == nil || ctx.User.ID == 0 {
			ctx.Logger.Error().Interface("user", ctx.User).Bool("authenticated", ctx.IsAuthenticated).Msg("Not allowed to enter protected route")

			svrUtils.SendJsonResponse(w, http.StatusUnauthorized, nil, "Unauthorized")
			return
		}

		ctx.Logger.Debug().Bool("authenticated", ctx.IsAuthenticated).Msg("Entering protected route")
		next.ServeHTTP(w, r)
	})
}
