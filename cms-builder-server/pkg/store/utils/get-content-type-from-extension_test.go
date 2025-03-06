package store_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store/utils"
)

func TestGetContentTypeFromExtension(t *testing.T) {
	// Test case 1: PNG file
	t.Run("PNG file", func(t *testing.T) {
		contentType := svrUtils.GetContentTypeFromExtension("testfile.png")
		assert.Equal(t, "image/png", contentType)
	})

	// Test case 2: JPEG file
	t.Run("JPEG file", func(t *testing.T) {
		contentType := svrUtils.GetContentTypeFromExtension("testfile.jpg")
		assert.Equal(t, "image/jpeg", contentType)
	})

	// Test case 3: Text file
	t.Run("Text file", func(t *testing.T) {
		contentType := svrUtils.GetContentTypeFromExtension("testfile.txt")
		assert.Equal(t, "text/plain; charset=utf-8", contentType)
	})

	// Test case 4: Unsupported file extension
	t.Run("Unsupported file extension", func(t *testing.T) {
		contentType := svrUtils.GetContentTypeFromExtension("testfile.unknown")
		assert.Equal(t, "application/octet-stream", contentType)
	})
}
