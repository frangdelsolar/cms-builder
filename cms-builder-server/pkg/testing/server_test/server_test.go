package server_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	testPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
)

func TestNewServer(t *testing.T) {
	bed := testPkg.SetupServerTestBed()

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
