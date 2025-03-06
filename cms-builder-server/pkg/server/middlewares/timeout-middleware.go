package server

import (
	"context"
	"net/http"
	"time"

	loggerPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
)

const TimeoutSeconds = 15

// TimeoutMiddleware sets a timeout for requests.
func TimeoutMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx, cancel := context.WithTimeout(r.Context(), TimeoutSeconds*time.Second)
		defer cancel()

		log := loggerPkg.Default

		r = r.WithContext(ctx)

		done := make(chan struct{})

		go func() {
			next.ServeHTTP(w, r)
			close(done)
		}()

		select {
		case <-done:
			// Handler completed successfully
		case <-ctx.Done():
			log.Error().Msg("Request timed out")
			http.Error(w, "Request timed out", http.StatusGatewayTimeout) // Or 504 Gateway Timeout
		}
	})
}
