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

	// Replace spaces with underscores
	name = strings.ReplaceAll(name, " ", "_")

	// Replace forward slashes with underscores
	name = strings.ReplaceAll(name, "/", "_")

	// Replace backslashes with underscores
	name = strings.ReplaceAll(name, "\\", "_")

	// Add the current timestamp to the file name
	now := strconv.FormatInt(time.Now().UnixNano(), 10)
	name = now + "_" + name

	return name + extension
}

func getMimeTypeAndExtension(mime string) (string, string) {
	parts := strings.Split(mime, "/")
	if len(parts) == 0 {
		return "", "" // Or handle this as an error if you prefer
	}
	mimeType := parts[0]
	extension := ""
	if len(parts) > 1 {
		extension = parts[len(parts)-1] // Get the last part as the extension
	}
	return mimeType, extension
}

func ValidateContentType(contentType string, supportedMimeTypes []string) (bool, error) {
	inMimeType, inExtension := getMimeTypeAndExtension(contentType)

	for _, supportedItem := range supportedMimeTypes {
		// "*"
		if supportedItem == "*" {
			return true, nil
		}
		supportedMimeType, supportedExtension := getMimeTypeAndExtension(supportedItem)
		// "*/*"
		if supportedMimeType == "*" {
			return true, nil
		}
		// "audio/*"
		if supportedExtension == "*" && inMimeType == supportedMimeType {
			return true, nil
		}
		// "audio/wav"
		if supportedExtension == inExtension && supportedMimeType == inMimeType {
			return true, nil
		}
	}

	return false, nil
}
