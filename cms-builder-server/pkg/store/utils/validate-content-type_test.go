package store_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store/utils"
)

func TestValidateContentType(t *testing.T) {
	supportedMimeTypes := []string{"image/png", "image/jpeg", "text/*", "application/pdf"}

	// Test case 1: Exact match
	t.Run("Exact match", func(t *testing.T) {
		valid, err := svrUtils.ValidateContentType("image/png", supportedMimeTypes)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	// Test case 2: Wildcard match (text/*)
	t.Run("Wildcard match", func(t *testing.T) {
		valid, err := svrUtils.ValidateContentType("text/plain", supportedMimeTypes)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	// Test case 3: Unsupported MIME type
	t.Run("Unsupported MIME type", func(t *testing.T) {
		valid, err := svrUtils.ValidateContentType("audio/mpeg", supportedMimeTypes)
		assert.NoError(t, err)
		assert.False(t, valid)
	})

	// Test case 4: Global wildcard (*)
	t.Run("Global wildcard", func(t *testing.T) {
		supported := []string{"*"}
		valid, err := svrUtils.ValidateContentType("image/png", supported)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	// Test case 5: Global wildcard (*/*)
	t.Run("Global wildcard (*/*)", func(t *testing.T) {
		supported := []string{"*/*"}
		valid, err := svrUtils.ValidateContentType("application/json", supported)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	// Test case 6: Invalid MIME type format
	t.Run("Invalid MIME type format", func(t *testing.T) {
		valid, err := svrUtils.ValidateContentType("invalid-mime-type", supportedMimeTypes)
		assert.NoError(t, err)
		assert.False(t, valid)
	})
}
