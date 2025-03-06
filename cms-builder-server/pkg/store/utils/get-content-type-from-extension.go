package utils

import (
	"mime"
	"path/filepath"
)

// GetContentTypeFromExtension guesses the MIME type based on the file extension.
func GetContentTypeFromExtension(fileName string) string {
	mimeType := mime.TypeByExtension(filepath.Ext(fileName))
	if mimeType == "" {
		return "application/octet-stream" // Default MIME type
	}
	return mimeType
}
