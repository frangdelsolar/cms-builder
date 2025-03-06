package store_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store/utils"
)

func TestDetectContentTypeFromFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Test case 1: PNG file
	t.Run("PNG file", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "testfile.png")
		err := os.WriteFile(filePath, []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01"), 0644)
		assert.NoError(t, err)

		contentType, err := svrUtils.DetectContentTypeFromFile(filePath)
		assert.NoError(t, err)
		assert.Equal(t, "image/png", contentType)
	})

	// Test case 2: JPEG file
	t.Run("JPEG file", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "testfile.jpg")
		err := os.WriteFile(filePath, []byte("\xFF\xD8\xFF\xE0\x00\x10JFIF\x00\x01\x01\x01\x00\x01\x00\x01\x00\x00"), 0644)
		assert.NoError(t, err)

		contentType, err := svrUtils.DetectContentTypeFromFile(filePath)
		assert.NoError(t, err)
		assert.Equal(t, "image/jpeg", contentType)
	})

	// Test case 3: Text file
	t.Run("Text file", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "testfile.txt")
		err := os.WriteFile(filePath, []byte("Hello, World!\nThis is a text file with more content to help DetectContentType identify it correctly."), 0644)
		assert.NoError(t, err)

		contentType, err := svrUtils.DetectContentTypeFromFile(filePath)
		assert.NoError(t, err)

		t.Log(contentType)

		assert.Equal(t, "text/plain; charset=utf-8", contentType)
	})

	// Test case 4: Invalid file path
	t.Run("Invalid file path", func(t *testing.T) {
		_, err := svrUtils.DetectContentTypeFromFile("nonexistent/file/path")
		assert.Error(t, err)
	})
}
