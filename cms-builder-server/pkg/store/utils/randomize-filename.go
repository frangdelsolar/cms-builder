package store

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// randomizeFileName takes a file name and returns a new file name that is
// randomized using the current timestamp. The file name is also sanitized
// to replace spaces, forward slashes, and backslashes with underscores.
func RandomizeFileName(fileName string) string {
	// Extract the base file name (without the directory)
	baseName := filepath.Base(fileName)

	// Split the base name and extension
	extension := filepath.Ext(baseName)
	name := strings.TrimSuffix(baseName, extension)

	// Replace spaces and slashes with underscores
	name = strings.NewReplacer(" ", "_", "/", "_", "\\", "_").Replace(name)

	// Add the current timestamp to the file name
	now := strconv.FormatInt(time.Now().UnixNano(), 10)
	name = now + "_" + name

	return name + extension
}
