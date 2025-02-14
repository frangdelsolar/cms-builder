package builder_test

import (
	"testing"

	builder "github.com/frangdelsolar/cms-builder/cms-builder-server"
	"github.com/stretchr/testify/assert"
)

// TestNewServer_ValidConfig tests the NewServer function with a valid configuration.
//
// It tests the following:
//   - The returned server is not nil.
//   - The server address is correctly set.
//   - The server root is not nil.
//   - The server middlewares are not nil.
func TestNewServer_ValidConfig(t *testing.T) {
	t.Log("Testing NewServer")
	config := &builder.ServerConfig{
		Host:      "localhost",
		Port:      "8080",
		CSRFToken: "secret",
		Builder:   nil,
	}
	server, err := builder.NewServer(config)

	assert.NoError(t, err)
	assert.Equal(t, "localhost:8080", server.Addr)
	assert.NotNil(t, server.Root)
	assert.NotNil(t, server.Middlewares)
}
