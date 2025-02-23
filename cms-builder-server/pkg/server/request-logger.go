package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
)

type WriterWrapper struct {
	http.ResponseWriter
	StatusCode int
	Body       *bytes.Buffer
}

func (w *WriterWrapper) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *WriterWrapper) Write(b []byte) (int, error) {
	if w.Body != nil {
		w.Body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

// RequestLoggerMiddleware assigns a unique ID to each request and adds it to the context.
func RequestLoggerMiddleware(db *database.Database) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			ctx := r.Context()
			ctx = context.WithValue(ctx, CtxRequestStartTime, start)
			r = r.WithContext(ctx)

			wrappedWriter := &WriterWrapper{
				ResponseWriter: w,
				StatusCode:     http.StatusOK,
				Body:           new(bytes.Buffer),
			}

			var err error
			var requestBody string
			var requestHeaders string
			var responseBody string

			requestIdentifier := uuid.New().String()
			r = r.WithContext(context.WithValue(r.Context(), CtxRequestIdentifier, requestIdentifier))

			defer func() {

				duration := time.Since(start)

				statusCode := wrappedWriter.StatusCode

				errorMessage := ""
				if err != nil {
					errorMessage = err.Error()
				}

				query, err := url.QueryUnescape(r.URL.RawQuery)
				if err != nil {
					log.Error().Err(err).Msg("Error unescaping query")
				}

				logEntry := models.RequestLog{
					Timestamp:         start,
					Ip:                r.RemoteAddr,
					UserId:            r.Header.Get(requestedByParamKey.S()),
					Roles:             r.Header.Get(rolesParamKey.S()),
					Method:            r.Method,
					Path:              r.URL.Path,
					Query:             query,
					Duration:          duration.Nanoseconds() / 1e6,
					StatusCode:        fmt.Sprintf("%d", statusCode),
					Origin:            r.Header.Get("Origin"),
					Referer:           r.Header.Get("Referer"),
					Error:             errorMessage,
					Header:            requestHeaders,
					Body:              requestBody,
					Response:          responseBody,
					RequestIdentifier: requestIdentifier,
				}

				err = db.DB.Create(&logEntry).Error
				if err != nil {
					log.Error().Err(err).Msg("Error creating request log")
				}
			}()

			bodyBytes, readErr := io.ReadAll(r.Body)
			if readErr != nil {
				err = readErr // Capture the error
				log.Error().Err(err).Msg("Error reading request body")
			}

			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			// Capture headers before next.ServeHTTP
			headers := make(map[string][]string)
			for name, values := range r.Header {
				headers[name] = values
			}

			headerJSON, marshalErr := json.Marshal(headers)
			if marshalErr != nil {
				err = marshalErr
				log.Error().Err(err).Msg("Error marshaling headers")
			}

			next.ServeHTTP(wrappedWriter, r)

			// Check for errors after the handler has run
			if wrappedWriter.StatusCode >= 400 || readErr != nil || marshalErr != nil {
				// If there was an error during the request, log the body and headers
				if wrappedWriter.StatusCode >= 400 {
					err = errors.New(http.StatusText(wrappedWriter.StatusCode))
					requestHeaders = string(headerJSON)
					requestBody = string(bodyBytes)
					responseBody = wrappedWriter.Body.String()
				}
			}
		})
	}

}
