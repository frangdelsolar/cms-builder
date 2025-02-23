package server

import (
	"context"
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
)

// LoggingMiddleware is a sample middleware function that logs the request URI.
//
// It takes an http.Handler as input and returns a new http.Handler that wraps the original
// handler and logs the request URI before calling the original handler.
func LoggingMiddleware(loggerConfig *logger.LoggerConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestId := GetRequestId(r)

			log, err := logger.NewLogger(loggerConfig)
			if err != nil {
				logger.Default.Error().Err(err).Msg("Error creating request logger")
				// Fall back to the default logger
				log = logger.Default
			}

			// Add the request ID to the logger context
			requestLog := log.Logger.With().Str("requestId", requestId).Logger()

			// Add the logger to the request context
			ctx := context.WithValue(r.Context(), CtxRequestLogger, &requestLog)
			r = r.WithContext(ctx)

			// Log the request
			requestLog.Info().Msgf("[%s] %s", r.Method, r.RequestURI)

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}
