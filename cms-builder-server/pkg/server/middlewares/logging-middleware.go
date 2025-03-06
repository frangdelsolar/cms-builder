package server

import (
	"context"
	"net/http"

	loggerPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	svrConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/constants"
	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/utils"
)

// LoggingMiddleware is a sample middleware function that logs the request URI.
//
// It takes an http.Handler as input and returns a new http.Handler that wraps the original
// handler and logs the request URI before calling the original handler.
func LoggingMiddleware(loggerConfig *loggerTypes.LoggerConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestId := svrUtils.GetRequestId(r)

			log, err := loggerPkg.NewLogger(loggerConfig)
			if err != nil {
				log = loggerPkg.Default
			}

			// Add the request ID to the loggerTypes context
			zero := log.Logger.With().Str("requestId", requestId).Logger()
			requestLog := &loggerTypes.Logger{Logger: &zero}

			// Add the loggerTypes to the request context
			ctx := context.WithValue(r.Context(), svrConstants.CtxRequestLogger, requestLog)

			user := svrUtils.GetRequestUser(r)

			// Log the request
			requestLog.Info().
				Str("user", user.Email).
				Str("roles", user.Roles).
				Msgf("[%s] %s", r.Method, r.RequestURI)

			// Call the next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
