package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/rs/zerolog"
)

type ContextParamKey string

const (
	CtxRequestIdentifier ContextParamKey = "requestIdentifier"
	CtxRequestStartTime  ContextParamKey = "requestStartTime"
	CtxRequestLogger     ContextParamKey = "requestLogger"
)

const (
	requestedByParamKey RequestParamKey = "requestedBy"
	rolesParamKey       RequestParamKey = "roles"
	authParamKey        RequestParamKey = "auth"
	limitParamKey       RequestParamKey = "limit"
	pageParamKey        RequestParamKey = "page"
)

type RequestParamKey string

func (r RequestParamKey) S() string {
	return string(r)
}

// ValidateRequestMethod returns an error if the request method does not match the given
// method string. The error message will include the actual request method.
func ValidateRequestMethod(r *http.Request, method string) error {
	if r.Method != method {
		return fmt.Errorf("invalid request method: %s", r.Method)
	}
	return nil
}

// GetAccessTokenFromRequest extracts the access token from the Authorization header of the given request.
// The header should be in the format "Bearer <token>".
// If the token is not found, it returns an empty string.
func GetAccessTokenFromRequest(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header != "" {
		token := strings.Split(header, " ")[1]
		if token != "" {
			return token
		}
	}
	return ""
}

// GetRequestId retrieves the request ID from the context of the given request.
// The request ID is set by the RequestLogMiddleware.
// If the request ID is not found, it returns an empty string.
func GetRequestId(r *http.Request) string {
	ctx := r.Context()
	if requestId, ok := ctx.Value(CtxRequestIdentifier).(string); ok {
		return requestId
	}

	return ""
}

func GetLoggerFromRequest(r *http.Request) *zerolog.Logger {
	loggerFromContext := r.Context().Value(CtxRequestLogger)
	log, ok := loggerFromContext.(*zerolog.Logger)
	if !ok {
		return logger.Default.Logger
	}
	return log
}
