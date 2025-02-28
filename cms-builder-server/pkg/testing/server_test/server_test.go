package server_test

import (
	"testing"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing/server_test"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	bed := SetupServerTestBed()

	// Check the bed is complete
	assert.NotNil(t, bed.Server)
	assert.NotNil(t, bed.Db)
	assert.NotNil(t, bed.Logger)

	// Test the NewServer function
	svr := bed.Server
	assert.NotNil(t, svr.Server)
	assert.NotNil(t, svr.ServerConfig)
	assert.NotNil(t, svr.Middlewares)
	assert.NotNil(t, svr.Root)
	assert.NotNil(t, svr.DB)
	assert.NotNil(t, svr.Logger)
}
