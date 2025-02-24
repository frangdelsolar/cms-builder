package server

import (
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
)

func ProtectedRouteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Context().Value(CtxRequestIsAuth).(bool)
		user := r.Context().Value(CtxRequestUser).(*models.User)

		if !auth || user == nil || user.ID == 0 {
			SendJsonResponse(w, http.StatusUnauthorized, nil, "Unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}
