package store_test

import (
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
