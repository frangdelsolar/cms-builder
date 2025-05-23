package requestlogger

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

	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	rmModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/request-logger/models"
	svrConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/constants"
	svrTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/types"
	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/utils"
)

// RequestLoggerMiddleware assigns a unique ID to each request and adds it to the context.
func RequestLoggerMiddleware(db *dbTypes.DatabaseConnection) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			ctx := r.Context()
			ctx = context.WithValue(ctx, svrConstants.CtxRequestStartTime, start)
			traceId := uuid.New().String()
			ctx = context.WithValue(ctx, svrConstants.CtxTraceId, traceId)
			r = r.WithContext(ctx)

			log := loggerPkg.Default

			wrappedWriter := &svrTypes.WriterWrapper{
				ResponseWriter: w,
				StatusCode:     http.StatusOK,
				Body:           new(bytes.Buffer),
			}

			var err error
			var requestBody []byte
			var requestHeaders map[string][]string
			var responseBody string

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

				user := svrUtils.GetRequestUser(r)
				userLabel := "Anonymous"
				roles := ""

				userId := uint(0)
				if user.ID != 0 {
					userLabel = user.Email
					roles = user.Roles
					userId = user.ID
				}

				headersJSON, marshalErr := json.Marshal(requestHeaders)
				if marshalErr != nil {
					log.Error().Err(marshalErr).Msg("Error marshaling headers")
				}

				logEntry := rmModels.RequestLog{
					Timestamp:  start,
					Ip:         r.RemoteAddr,
					UserId:     userId,
					UserLabel:  userLabel,
					Roles:      roles,
					Method:     r.Method,
					Path:       r.URL.Path,
					Query:      query,
					Duration:   duration.Nanoseconds() / 1e6,
					StatusCode: fmt.Sprintf("%d", statusCode),
					Origin:     r.Header.Get("Origin"),
					Referer:    r.Header.Get("Referer"),
					Error:      errorMessage,
					Header:     string(headersJSON),
					Body:       string(requestBody),
					Response:   responseBody,
					TraceId:    traceId,
				}

				if db != nil && db.DB != nil {
					if createErr := db.DB.Create(&logEntry).Error; createErr != nil {
						log.Error().Err(createErr).Msg("Error creating request log")
					}
				} else {
					log.Error().Msg("Database or DB instance is nil, cannot create request log")
				}
			}()

			// Read request body
			requestBody, err = io.ReadAll(r.Body)
			if err != nil {
				log.Error().Err(err).Msg("Error reading request body")
			}
			r.Body = io.NopCloser(bytes.NewReader(requestBody))

			// Capture headers
			requestHeaders = r.Header

			next.ServeHTTP(wrappedWriter, r)

			// Capture response body and status code
			if wrappedWriter.StatusCode >= http.StatusBadRequest {
				responseBody = wrappedWriter.Body.String()
				if wrappedWriter.StatusCode >= http.StatusInternalServerError {
					err = errors.New(http.StatusText(wrappedWriter.StatusCode))
				}
			}
		})
	}
}
