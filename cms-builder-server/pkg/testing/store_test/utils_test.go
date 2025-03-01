package store_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store"
	"github.com/stretchr/testify/assert"
)

func TestRandomizeFileName(t *testing.T) {
	// Helper function to extract the timestamp part and the static part of the filename
	getTimestampAndNameParts := func(fileName string) (timestamp string, name string, extension string) {
		// Extract the file extension and name
		ext := filepath.Ext(fileName)
		baseName := fileName[:len(fileName)-len(ext)]
		parts := strings.SplitN(baseName, "_", 2)

		if len(parts) == 2 {
			timestamp = parts[0]
			name = parts[1]
		}

		extension = ext
		return
	}

	// Test cases
	tests := []struct {
		name     string
		input    string
		expected string // Expected output (excluding timestamp, which is dynamic)
	}{
		{
			name:     "Test with spaces in file name",
			input:    "my file name.txt",
			expected: "my_file_name.txt",
		},
		{
			name:     "Test with slashes in file name",
			input:    "folder/my/file/name.txt",
			expected: "name.txt",
		},
		{
			name:     "Test with backslashes in file name",
			input:    "folder\\my\\file\\name.txt",
			expected: "folder_my_file_name.txt",
		},
		{
			name:     "Test with no spaces, slashes, or backslashes",
			input:    "myfile.txt",
			expected: "myfile.txt",
		},
		{
			name:     "Test with mixed spaces and slashes",
			input:    "file/with spaces\\and\\slashes.txt",
			expected: "with_spaces_and_slashes.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := RandomizeFileName(tt.input)

			// Extract the dynamic timestamp part and check the static part of the filename
			timestamp, name, extension := getTimestampAndNameParts(actual)

			// Ensure the timestamp is valid (a positive integer)
			_, err := strconv.ParseInt(timestamp, 10, 64)
			assert.NoError(t, err, "Expected timestamp to be a valid number")

			// Check that the name part of the filename matches the expected result
			assert.Equal(t, tt.expected, name+extension, "Expected transformed file name")
		})
	}
}

func TestDetectContentTypeFromFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Test case 1: PNG file
	t.Run("PNG file", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "testfile.png")
		err := os.WriteFile(filePath, []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01"), 0644)
		assert.NoError(t, err)

		contentType, err := DetectContentTypeFromFile(filePath)
		assert.NoError(t, err)
		assert.Equal(t, "image/png", contentType)
	})

	// Test case 2: JPEG file
	t.Run("JPEG file", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "testfile.jpg")
		err := os.WriteFile(filePath, []byte("\xFF\xD8\xFF\xE0\x00\x10JFIF\x00\x01\x01\x01\x00\x01\x00\x01\x00\x00"), 0644)
		assert.NoError(t, err)

		contentType, err := DetectContentTypeFromFile(filePath)
		assert.NoError(t, err)
		assert.Equal(t, "image/jpeg", contentType)
	})

	// Test case 3: Text file
	t.Run("Text file", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "testfile.txt")
		err := os.WriteFile(filePath, []byte("Hello, World!\nThis is a text file with more content to help DetectContentType identify it correctly."), 0644)
		assert.NoError(t, err)

		contentType, err := DetectContentTypeFromFile(filePath)
		assert.NoError(t, err)

		t.Log(contentType)

		assert.Equal(t, "text/plain; charset=utf-8", contentType)
	})

	// Test case 4: Invalid file path
	t.Run("Invalid file path", func(t *testing.T) {
		_, err := DetectContentTypeFromFile("nonexistent/file/path")
		assert.Error(t, err)
	})
}

func TestGetContentTypeFromExtension(t *testing.T) {
	// Test case 1: PNG file
	t.Run("PNG file", func(t *testing.T) {
		contentType := GetContentTypeFromExtension("testfile.png")
		assert.Equal(t, "image/png", contentType)
	})

	// Test case 2: JPEG file
	t.Run("JPEG file", func(t *testing.T) {
		contentType := GetContentTypeFromExtension("testfile.jpg")
		assert.Equal(t, "image/jpeg", contentType)
	})

	// Test case 3: Text file
	t.Run("Text file", func(t *testing.T) {
		contentType := GetContentTypeFromExtension("testfile.txt")
		assert.Equal(t, "text/plain; charset=utf-8", contentType)
	})

	// Test case 4: Unsupported file extension
	t.Run("Unsupported file extension", func(t *testing.T) {
		contentType := GetContentTypeFromExtension("testfile.unknown")
		assert.Equal(t, "application/octet-stream", contentType)
	})
}

func TestValidateContentType(t *testing.T) {
	supportedMimeTypes := []string{"image/png", "image/jpeg", "text/*", "application/pdf"}

	// Test case 1: Exact match
	t.Run("Exact match", func(t *testing.T) {
		valid, err := ValidateContentType("image/png", supportedMimeTypes)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	// Test case 2: Wildcard match (text/*)
	t.Run("Wildcard match", func(t *testing.T) {
		valid, err := ValidateContentType("text/plain", supportedMimeTypes)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	// Test case 3: Unsupported MIME type
	t.Run("Unsupported MIME type", func(t *testing.T) {
		valid, err := ValidateContentType("audio/mpeg", supportedMimeTypes)
		assert.NoError(t, err)
		assert.False(t, valid)
	})

	// Test case 4: Global wildcard (*)
	t.Run("Global wildcard", func(t *testing.T) {
		supported := []string{"*"}
		valid, err := ValidateContentType("image/png", supported)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	// Test case 5: Global wildcard (*/*)
	t.Run("Global wildcard (*/*)", func(t *testing.T) {
		supported := []string{"*/*"}
		valid, err := ValidateContentType("application/json", supported)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	// Test case 6: Invalid MIME type format
	t.Run("Invalid MIME type format", func(t *testing.T) {
		valid, err := ValidateContentType("invalid-mime-type", supportedMimeTypes)
		assert.NoError(t, err)
		assert.False(t, valid)
	})
}
