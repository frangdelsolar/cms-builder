package testing

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/stretchr/testify/assert"
)

func CreateTestRequest(t *testing.T, method, path, body string, isAuth bool, user *models.User, log *logger.Logger) *http.Request {

	log.Error().Str("path", path).Msg("Creating test request")

	req, err := http.NewRequest(method, path, bytes.NewBufferString(body))
	assert.NoError(t, err)

	ctx := req.Context()
	ctx = context.WithValue(ctx, server.CtxRequestIsAuth, isAuth)
	ctx = context.WithValue(ctx, server.CtxRequestUser, user)
	ctx = context.WithValue(ctx, server.CtxRequestLogger, log)

	return req.WithContext(ctx)
}

// ExecuteHandler executes the handler and returns the response recorder
func ExecuteHandler(t *testing.T, handler http.HandlerFunc, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}
