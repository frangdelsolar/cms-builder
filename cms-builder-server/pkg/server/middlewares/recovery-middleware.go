package server

import (
	"net/http"
	"runtime/debug"

	loggerPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/utils"
)

// RecoveryMiddleware catches panics and logs them.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				loggerPkg.Default.Error().Interface("panic", err).Bytes("stack", debug.Stack()).Msg("Panic recovered")
				svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, "Internal Server Error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
